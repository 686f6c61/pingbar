package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	serperPlacesURL = "https://google.serper.dev/places"
	serperSearchURL = "https://google.serper.dev/search"

	maxResponseSize = 10 << 20 // 10MB
	maxInputLength  = 200
)

// Regex precompiladas para extraccion de horarios
var (
	reTimeRange     = regexp.MustCompile(`(\d{1,2}:\d{2})\s*[-–a]\s*(\d{1,2}:\d{2})`)
	reTimeFromTo    = regexp.MustCompile(`de\s+(\d{1,2}:\d{2})\s+a\s+(\d{1,2}:\d{2})`)
	reHourRange     = regexp.MustCompile(`(\d{1,2})h\s*[-–a]\s*(\d{1,2})h`)
	reDayTimeRange  = regexp.MustCompile(`(?:lunes|martes|miércoles|jueves|viernes|sábado|domingo).*?(\d{1,2}:\d{2})\s*[-–a]\s*(\d{1,2}:\d{2})`)
	reOpenDay       = regexp.MustCompile(`abierto.*?(?:lunes|martes|miércoles|jueves|viernes|sábado|domingo)`)
	reSegmentTime   = regexp.MustCompile(`(\d{1,2}[:\.]?\d{0,2})\s*[-–a]\s*(\d{1,2}[:\.]?\d{0,2})`)
	reCurrentlyOpen = regexp.MustCompile(`(\d{1,2}):(\d{2})\s*-\s*(\d{1,2}):(\d{2})`)
)

// PlaceResult representa un resultado de lugar de la API
type PlaceResult struct {
	Title       string  `json:"title"`
	Address     string  `json:"address"`
	Rating      float64 `json:"rating"`
	RatingCount int     `json:"ratingCount"`
	Category    string  `json:"category"`
	PhoneNumber string  `json:"phoneNumber"`
	Website     string  `json:"website"`
}

// OrganicResult resultado de busqueda organica
type OrganicResult struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

// SerperPlacesResponse respuesta del endpoint /places
type SerperPlacesResponse struct {
	Places []PlaceResult `json:"places"`
}

// SerperSearchResponse respuesta del endpoint /search
type SerperSearchResponse struct {
	Organic []OrganicResult `json:"organic"`
}

// BusinessInfo representa la informacion procesada de un negocio
type BusinessInfo struct {
	Name        string
	Address     string
	Rating      float64
	RatingCount int
	Category    string
	Phone       string
	Website     string
	IsOpen      bool
	IsUnknown   bool
	TodayHours  string
	HoursInfo   string
}

// APIError representa un error de la API
type APIError struct {
	Type    string
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

// Search busca negocios en Serper y extrae horarios de snippets
func Search(apiKey, business, city string, limit int) ([]BusinessInfo, error) {
	if apiKey == "" {
		return nil, &APIError{Type: "no_api_key", Message: "API Key no configurada"}
	}

	if len(business) > maxInputLength || len(city) > maxInputLength {
		return nil, &APIError{Type: "invalid_input", Message: "Nombre de negocio o ciudad demasiado largo"}
	}

	if limit <= 0 {
		limit = 10
	}

	places, err := searchPlaces(apiKey, business, city, limit)
	if err != nil {
		return nil, err
	}

	results := make([]BusinessInfo, 0, len(places))

	for i, place := range places {
		info := BusinessInfo{
			Name:        place.Title,
			Address:     place.Address,
			Rating:      place.Rating,
			RatingCount: place.RatingCount,
			Category:    place.Category,
			Phone:       place.PhoneNumber,
			Website:     place.Website,
			IsUnknown:   true,
		}

		// Solo buscar horarios para los primeros 3 resultados (ahorrar creditos)
		if i < 3 {
			hoursInfo := searchHours(apiKey, place.Title, city)
			if hoursInfo != "" {
				info.HoursInfo = hoursInfo
				info.IsUnknown = false
				info.TodayHours = hoursInfo
				info.IsOpen = isCurrentlyOpen(hoursInfo)
			}
		}

		results = append(results, info)
	}

	return results, nil
}

// searchPlaces busca lugares con el endpoint /places
func searchPlaces(apiKey, business, city string, limit int) ([]PlaceResult, error) {
	query := fmt.Sprintf("%s %s", business, city)

	requestBody := map[string]interface{}{
		"q":   query,
		"gl":  "es",
		"hl":  "es",
		"num": limit,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error al serializar peticion: %w", err)
	}

	req, err := http.NewRequest("POST", serperPlacesURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error al crear peticion: %w", err)
	}
	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &APIError{Type: "connection", Message: "Error de conexion"}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("error al leer respuesta: %w", err)
	}

	switch resp.StatusCode {
	case 401:
		return nil, &APIError{Type: "invalid_key", Message: "API Key invalida"}
	case 429:
		return nil, &APIError{Type: "limit_reached", Message: "Limite de API alcanzado"}
	case 200:
		// OK
	default:
		return nil, &APIError{Type: "unknown", Message: fmt.Sprintf("Error de API: %d", resp.StatusCode)}
	}

	var serperResp SerperPlacesResponse
	if err := json.Unmarshal(body, &serperResp); err != nil {
		return nil, err
	}

	cityLower := strings.ToLower(city)
	filtered := make([]PlaceResult, 0)

	for _, place := range serperResp.Places {
		addressLower := strings.ToLower(place.Address)
		if strings.Contains(addressLower, cityLower) {
			filtered = append(filtered, place)
		}
	}

	if len(filtered) == 0 {
		return serperResp.Places, nil
	}

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

// searchHours busca horarios usando el endpoint /search
func searchHours(apiKey, businessName, city string) string {
	query := fmt.Sprintf("horario %s %s", businessName, city)

	requestBody := map[string]interface{}{
		"q":   query,
		"gl":  "es",
		"hl":  "es",
		"num": 5,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return ""
	}

	req, err := http.NewRequest("POST", serperSearchURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return ""
	}
	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return ""
	}

	var searchResp SerperSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return ""
	}

	for _, result := range searchResp.Organic {
		hours := extractHoursFromText(result.Snippet)
		if hours != "" {
			return hours
		}
	}

	return ""
}

// extractHoursFromText extrae informacion de horario de un texto
func extractHoursFromText(text string) string {
	text = strings.ToLower(text)

	type regexPattern struct {
		re        *regexp.Regexp
		isContext bool // si es patron de contexto (no extrae horas directamente)
	}

	patterns := []regexPattern{
		{reTimeRange, false},
		{reTimeFromTo, false},
		{reHourRange, false},
		{reDayTimeRange, false},
		{reOpenDay, true},
	}

	for _, p := range patterns {
		matches := p.re.FindStringSubmatch(text)
		if !p.isContext && len(matches) >= 3 {
			return fmt.Sprintf("%s - %s", normalizeTime(matches[1]), normalizeTime(matches[2]))
		}
		if p.isContext && len(matches) >= 1 {
			idx := strings.Index(text, matches[0])
			start := idx
			end := idx + len(matches[0]) + 50
			if end > len(text) {
				end = len(text)
			}
			return strings.TrimSpace(text[start:end])
		}
	}

	// Buscar menciones de "horario"
	if strings.Contains(text, "horario") {
		idx := strings.Index(text, "horario")
		start := idx
		end := idx + 60
		if end > len(text) {
			end = len(text)
		}
		segment := text[start:end]

		matches := reSegmentTime.FindStringSubmatch(segment)
		if len(matches) >= 3 {
			return fmt.Sprintf("%s - %s", normalizeTime(matches[1]), normalizeTime(matches[2]))
		}
	}

	// Buscar "24 horas"
	if strings.Contains(text, "24 horas") || strings.Contains(text, "24h") {
		return "Abierto 24 horas"
	}

	return ""
}

// normalizeTime normaliza el formato de hora
func normalizeTime(t string) string {
	t = strings.TrimSpace(t)
	t = strings.ReplaceAll(t, ".", ":")

	if !strings.Contains(t, ":") {
		t = t + ":00"
	}

	parts := strings.Split(t, ":")
	if len(parts) == 2 {
		hour := parts[0]
		min := parts[1]
		if len(hour) == 1 {
			hour = "0" + hour
		}
		if len(min) == 1 {
			min = "0" + min
		}
		return hour + ":" + min
	}

	return t
}

// isCurrentlyOpen determina si esta abierto basado en el horario extraido
func isCurrentlyOpen(hoursInfo string) bool {
	if strings.Contains(strings.ToLower(hoursInfo), "24 horas") {
		return true
	}

	matches := reCurrentlyOpen.FindStringSubmatch(hoursInfo)
	if len(matches) < 5 {
		return false
	}

	now := time.Now()
	currentHour := now.Hour()
	currentMin := now.Minute()

	var openH, openM, closeH, closeM int
	fmt.Sscanf(matches[1], "%d", &openH)
	fmt.Sscanf(matches[2], "%d", &openM)
	fmt.Sscanf(matches[3], "%d", &closeH)
	fmt.Sscanf(matches[4], "%d", &closeM)

	currentMins := currentHour*60 + currentMin
	openMins := openH*60 + openM
	closeMins := closeH*60 + closeM

	if closeMins < openMins {
		closeMins += 24 * 60
		if currentMins < openMins {
			currentMins += 24 * 60
		}
	}

	return currentMins >= openMins && currentMins < closeMins
}
