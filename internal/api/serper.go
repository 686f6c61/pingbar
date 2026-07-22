package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	serperPlacesURL = "https://google.serper.dev/places"
	serperSearchURL = "https://google.serper.dev/search"

	maxResponseSize = 10 << 20 // 10MB
	maxInputLength  = 200

	// Solo se buscan horarios para los primeros resultados (ahorrar creditos)
	maxHoursLookups = 3
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

// Regex precompiladas para extraccion de horarios. Los nombres de dia van
// sin tilde porque el texto se normaliza antes de aplicarlas.
var (
	reTimeRange     = regexp.MustCompile(`(\d{1,2}:\d{2})\s*[-–a]\s*(\d{1,2}:\d{2})`)
	reTimeFromTo    = regexp.MustCompile(`de\s+(\d{1,2}:\d{2})\s+a\s+(\d{1,2}:\d{2})`)
	reHourRange     = regexp.MustCompile(`(\d{1,2})h\s*[-–a]\s*(\d{1,2})h`)
	reDayTimeRange  = regexp.MustCompile(`(?:lunes|martes|miercoles|jueves|viernes|sabado|domingo).*?(\d{1,2}:\d{2})\s*[-–a]\s*(\d{1,2}:\d{2})`)
	reOpenDay       = regexp.MustCompile(`abierto.*?(?:lunes|martes|miercoles|jueves|viernes|sabado|domingo)`)
	reSegmentTime   = regexp.MustCompile(`(\d{1,2}[:\.]?\d{0,2})\s*[-–a]\s*(\d{1,2}[:\.]?\d{0,2})`)
	reCurrentlyOpen = regexp.MustCompile(`(\d{1,2}):(\d{2})\s*-\s*(\d{1,2}):(\d{2})`)
)

// reDayHours prioriza el horario del dia concreto cuando el snippet lista
// varios dias con horas distintas
var reDayHours = map[time.Weekday]*regexp.Regexp{
	time.Sunday:    dayHoursRegexp("domingo"),
	time.Monday:    dayHoursRegexp("lunes"),
	time.Tuesday:   dayHoursRegexp("martes"),
	time.Wednesday: dayHoursRegexp("miercoles"),
	time.Thursday:  dayHoursRegexp("jueves"),
	time.Friday:    dayHoursRegexp("viernes"),
	time.Saturday:  dayHoursRegexp("sabado"),
}

func dayHoursRegexp(day string) *regexp.Regexp {
	return regexp.MustCompile(day + `[^0-9]{0,30}(\d{1,2}:\d{2})\s*[-–a]\s*(\d{1,2}:\d{2})`)
}

// accentFolder elimina marcas diacriticas (tildes) para comparaciones
var accentFolder = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

// normalizeForCompare pasa a minusculas y quita tildes ("Málaga" -> "malaga")
func normalizeForCompare(s string) string {
	s = strings.ToLower(s)
	if folded, _, err := transform.String(accentFolder, s); err == nil {
		return folded
	}
	return s
}

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

	results := make([]BusinessInfo, len(places))

	var wg sync.WaitGroup
	for i, place := range places {
		results[i] = BusinessInfo{
			Name:        place.Title,
			Address:     place.Address,
			Rating:      place.Rating,
			RatingCount: place.RatingCount,
			Category:    place.Category,
			Phone:       place.PhoneNumber,
			Website:     place.Website,
			IsUnknown:   true,
		}

		if i < maxHoursLookups {
			wg.Add(1)
			go func(idx int, title string) {
				defer wg.Done()
				hoursInfo := searchHours(apiKey, title, city)
				if hoursInfo != "" {
					results[idx].HoursInfo = hoursInfo
					results[idx].IsUnknown = false
					results[idx].TodayHours = hoursInfo
					results[idx].IsOpen = isOpenAt(hoursInfo, time.Now())
				}
			}(i, place.Title)
		}
	}
	wg.Wait()

	return results, nil
}

// RefreshOpenState recalcula el estado abierto/cerrado con la hora actual.
// Necesario al servir resultados desde cache, donde el estado guardado
// corresponde al momento de la consulta original.
func (b *BusinessInfo) RefreshOpenState() {
	if !b.IsUnknown {
		b.IsOpen = isOpenAt(b.HoursInfo, time.Now())
	}
}

// searchPlaces busca lugares con el endpoint /places
func searchPlaces(apiKey, business, city string, limit int) ([]PlaceResult, error) {
	query := fmt.Sprintf("%s %s", business, city)

	requestBody := map[string]any{
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

	resp, err := httpClient.Do(req)
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

	cityNorm := normalizeForCompare(city)
	filtered := make([]PlaceResult, 0)

	for _, place := range serperResp.Places {
		if strings.Contains(normalizeForCompare(place.Address), cityNorm) {
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

	requestBody := map[string]any{
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

	resp, err := httpClient.Do(req)
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

	today := time.Now().Weekday()
	for _, result := range searchResp.Organic {
		hours := extractHoursFromText(result.Snippet, today)
		if hours != "" {
			return hours
		}
	}

	return ""
}

// extractHoursFromText extrae informacion de horario de un texto,
// priorizando el horario del dia indicado si el texto lo menciona
func extractHoursFromText(text string, weekday time.Weekday) string {
	text = normalizeForCompare(text)

	if re, ok := reDayHours[weekday]; ok {
		if m := re.FindStringSubmatch(text); len(m) >= 3 {
			return fmt.Sprintf("%s - %s", normalizeTime(m[1]), normalizeTime(m[2]))
		}
	}

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

// isOpenAt determina si esta abierto en el instante dado segun el horario extraido
func isOpenAt(hoursInfo string, now time.Time) bool {
	if strings.Contains(strings.ToLower(hoursInfo), "24 horas") {
		return true
	}

	matches := reCurrentlyOpen.FindStringSubmatch(hoursInfo)
	if len(matches) < 5 {
		return false
	}

	openH, _ := strconv.Atoi(matches[1])
	openM, _ := strconv.Atoi(matches[2])
	closeH, _ := strconv.Atoi(matches[3])
	closeM, _ := strconv.Atoi(matches[4])

	currentMins := now.Hour()*60 + now.Minute()
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
