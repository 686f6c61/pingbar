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

// OrganicResult resultado de búsqueda orgánica
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

// BusinessInfo representa la información procesada de un negocio
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
	HoursInfo   string // Información de horario extraída
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

	if limit <= 0 {
		limit = 10
	}

	// Paso 1: Buscar lugares
	places, err := searchPlaces(apiKey, business, city, limit)
	if err != nil {
		return nil, err
	}

	results := make([]BusinessInfo, 0, len(places))

	// Paso 2: Para cada lugar, intentar extraer horarios
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

		// Solo buscar horarios para los primeros 3 resultados (ahorrar créditos)
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
	// Incluir ciudad en el query para forzar resultados locales
	query := fmt.Sprintf("%s %s", business, city)
	
	requestBody := map[string]interface{}{
		"q":        query,
		"gl":       "es",
		"hl":       "es",
		"location": fmt.Sprintf("%s, España", city),
		"num":      limit * 2, // Pedir más para filtrar después
	}

	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", serperPlacesURL, bytes.NewBuffer(jsonBody))
	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &APIError{Type: "connection", Message: "Error de conexión"}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case 401:
		return nil, &APIError{Type: "invalid_key", Message: "API Key inválida"}
	case 429:
		return nil, &APIError{Type: "limit_reached", Message: "Límite de API alcanzado"}
	case 200:
		// OK
	default:
		return nil, &APIError{Type: "unknown", Message: fmt.Sprintf("Error de API: %d", resp.StatusCode)}
	}

	var serperResp SerperPlacesResponse
	if err := json.Unmarshal(body, &serperResp); err != nil {
		return nil, err
	}

	// Filtrar resultados que contengan la ciudad en la dirección
	cityLower := strings.ToLower(city)
	filtered := make([]PlaceResult, 0)
	
	for _, place := range serperResp.Places {
		addressLower := strings.ToLower(place.Address)
		// Incluir si la dirección contiene la ciudad
		if strings.Contains(addressLower, cityLower) {
			filtered = append(filtered, place)
		}
	}
	
	// Si no hay resultados filtrados, devolver los originales
	if len(filtered) == 0 {
		return serperResp.Places, nil
	}
	
	// Limitar al número solicitado
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

	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", serperSearchURL, bytes.NewBuffer(jsonBody))
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

	body, _ := io.ReadAll(resp.Body)

	var searchResp SerperSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return ""
	}

	// Buscar horarios en los snippets
	for _, result := range searchResp.Organic {
		hours := extractHoursFromText(result.Snippet)
		if hours != "" {
			return hours
		}
	}

	return ""
}

// extractHoursFromText extrae información de horario de un texto
func extractHoursFromText(text string) string {
	text = strings.ToLower(text)

	// Patrones comunes de horarios
	patterns := []string{
		// "10:00 - 22:00" o "10:00-22:00"
		`(\d{1,2}:\d{2})\s*[-–a]\s*(\d{1,2}:\d{2})`,
		// "de 10:00 a 22:00"
		`de\s+(\d{1,2}:\d{2})\s+a\s+(\d{1,2}:\d{2})`,
		// "10h - 22h" o "10h-22h"
		`(\d{1,2})h\s*[-–a]\s*(\d{1,2})h`,
		// "lunes a sábado 10:00 a 22:00"
		`(?:lunes|martes|miércoles|jueves|viernes|sábado|domingo).*?(\d{1,2}:\d{2})\s*[-–a]\s*(\d{1,2}:\d{2})`,
		// "abierto de lunes a sábado"
		`abierto.*?(?:lunes|martes|miércoles|jueves|viernes|sábado|domingo)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) >= 3 {
			return fmt.Sprintf("%s - %s", normalizeTime(matches[1]), normalizeTime(matches[2]))
		}
		if len(matches) >= 1 && strings.Contains(pattern, "abierto") {
			// Extraer contexto alrededor del match
			idx := strings.Index(text, matches[0])
			start := idx
			end := idx + len(matches[0]) + 50
			if end > len(text) {
				end = len(text)
			}
			return strings.TrimSpace(text[start:end])
		}
	}

	// Buscar menciones específicas de horario
	if strings.Contains(text, "horario") {
		// Extraer el contexto alrededor de "horario"
		idx := strings.Index(text, "horario")
		start := idx
		end := idx + 60
		if end > len(text) {
			end = len(text)
		}
		segment := text[start:end]

		// Buscar patrón de hora en el segmento
		re := regexp.MustCompile(`(\d{1,2}[:\.]?\d{0,2})\s*[-–a]\s*(\d{1,2}[:\.]?\d{0,2})`)
		matches := re.FindStringSubmatch(segment)
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

	// Si no tiene minutos, añadir :00
	if !strings.Contains(t, ":") {
		t = t + ":00"
	}

	// Asegurar formato HH:MM
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

// isCurrentlyOpen determina si está abierto basado en el horario extraído
func isCurrentlyOpen(hoursInfo string) bool {
	if strings.Contains(strings.ToLower(hoursInfo), "24 horas") {
		return true
	}

	// Extraer horario
	re := regexp.MustCompile(`(\d{1,2}):(\d{2})\s*-\s*(\d{1,2}):(\d{2})`)
	matches := re.FindStringSubmatch(hoursInfo)
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

	// Si cierra después de medianoche
	if closeMins < openMins {
		closeMins += 24 * 60
		if currentMins < openMins {
			currentMins += 24 * 60
		}
	}

	return currentMins >= openMins && currentMins < closeMins
}

// GetRawResponse obtiene la respuesta cruda de la API para cachear
func GetRawResponse(apiKey, business, city string, limit int) (json.RawMessage, error) {
	if apiKey == "" {
		return nil, &APIError{Type: "no_api_key", Message: "API Key no configurada"}
	}

	if limit <= 0 {
		limit = 10
	}

	requestBody := map[string]interface{}{
		"q":        business,
		"gl":       "es",
		"hl":       "es",
		"location": fmt.Sprintf("%s, España", city),
		"num":      limit,
	}

	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", serperPlacesURL, bytes.NewBuffer(jsonBody))
	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &APIError{Type: "connection", Message: "Error de conexión"}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case 401:
		return nil, &APIError{Type: "invalid_key", Message: "API Key inválida"}
	case 429:
		return nil, &APIError{Type: "limit_reached", Message: "Límite de API alcanzado"}
	case 200:
		return json.RawMessage(body), nil
	default:
		return nil, &APIError{Type: "unknown", Message: fmt.Sprintf("Error de API: %d", resp.StatusCode)}
	}
}

// ParseCachedResponse parsea una respuesta cacheada (sin horarios)
func ParseCachedResponse(data json.RawMessage) ([]BusinessInfo, error) {
	var serperResp SerperPlacesResponse
	if err := json.Unmarshal(data, &serperResp); err != nil {
		return nil, err
	}

	results := make([]BusinessInfo, 0, len(serperResp.Places))

	for _, place := range serperResp.Places {
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
		results = append(results, info)
	}

	return results, nil
}
