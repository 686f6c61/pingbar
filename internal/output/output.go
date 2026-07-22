package output

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/686f6c61/pingbar/internal/api"
	"github.com/686f6c61/pingbar/internal/i18n"
	"github.com/fatih/color"
)

// Formatter maneja el formateo de salida
type Formatter struct {
	Lang      i18n.Lang
	UseColors bool
	JSONMode  bool
}

// NewFormatter crea un nuevo formateador
func NewFormatter(lang string, colorMode string, jsonMode bool) *Formatter {
	f := &Formatter{
		Lang:     i18n.Lang(lang),
		JSONMode: jsonMode,
	}

	switch colorMode {
	case "on":
		f.UseColors = true
	case "off":
		f.UseColors = false
		color.NoColor = true
	default:
		f.UseColors = !color.NoColor
	}

	return f
}

// PrintResults imprime los resultados
func (f *Formatter) PrintResults(results []api.BusinessInfo, business, city string) {
	if f.JSONMode {
		f.printJSON(results, business, city)
		return
	}
	f.printText(results, business, city)
}

func (f *Formatter) printJSON(results []api.BusinessInfo, business, city string) {
	jsonResults := make([]map[string]any, 0, len(results))

	for _, r := range results {
		item := map[string]any{
			"nombre":    r.Name,
			"direccion": r.Address,
			"rating":    r.Rating,
			"opiniones": r.RatingCount,
			"categoria": r.Category,
			"telefono":  r.Phone,
			"website":   r.Website,
			"abierto":   r.IsOpen,
		}
		if r.HoursInfo != "" {
			item["horario"] = r.HoursInfo
		}
		jsonResults = append(jsonResults, item)
	}

	output := map[string]any{
		"query": map[string]string{
			"negocio": business,
			"ciudad":  city,
		},
		"total":      len(results),
		"resultados": jsonResults,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		fmt.Fprintf(os.Stderr, "Error al escribir JSON: %v\n", err)
	}
}

func (f *Formatter) printText(results []api.BusinessInfo, business, city string) {
	msgs := i18n.Get(f.Lang)

	if len(results) == 0 {
		fmt.Printf(msgs.NotFound+"\n", business, city)
		return
	}

	if len(results) > 1 {
		fmt.Printf(msgs.Found+"\n\n", len(results))
	}

	for i, r := range results {
		f.printBusinessInfo(r)
		if i < len(results)-1 {
			fmt.Println()
		}
	}
}

func (f *Formatter) printBusinessInfo(info api.BusinessInfo) {
	msgs := i18n.Get(f.Lang)
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	white := color.New(color.FgWhite)
	gray := color.New(color.FgHiBlack)

	var statusColor *color.Color
	var statusText string

	if info.IsUnknown {
		statusColor = yellow
		statusText = msgs.Unknown
	} else if info.IsOpen {
		statusColor = green
		statusText = msgs.Open
	} else {
		statusColor = red
		statusText = msgs.Closed
	}

	statusColor.Printf("[%s] ", statusText)
	white.Printf("%s", info.Name)
	if info.Address != "" {
		fmt.Printf(" - %s", info.Address)
	}
	fmt.Println()

	indent := "          "

	if info.HoursInfo != "" {
		now := time.Now()
		dayName := msgs.Days[int(now.Weekday())]
		fmt.Printf("%s%s %s: %s\n", indent, msgs.Today, dayName, info.HoursInfo)
	} else {
		gray.Printf("%s%s\n", indent, msgs.NoSchedule)
	}

	if info.Rating > 0 {
		stars := ""
		fullStars := int(info.Rating)
		for i := 0; i < fullStars; i++ {
			stars += "*"
		}
		for i := fullStars; i < 5; i++ {
			stars += "."
		}
		gray.Printf("%s%s %.1f", indent, stars, info.Rating)
		if info.RatingCount > 0 {
			gray.Printf(" (%d opiniones)", info.RatingCount)
		}
		fmt.Println()
	}

	if info.Category != "" {
		gray.Printf("%s%s\n", indent, info.Category)
	}

	if info.Phone != "" {
		gray.Printf("%sTel: %s\n", indent, info.Phone)
	}
}

// PrintWelcome imprime el mensaje de bienvenida
func PrintWelcome(lang string) {
	msgs := i18n.Get(i18n.Lang(lang))

	fmt.Println(msgs.WelcomeTitle)
	fmt.Println()
	fmt.Println(msgs.NoAPIKey)
	fmt.Println()
	fmt.Println(msgs.GetAPIKey)
	fmt.Println()
	fmt.Println(msgs.MoreInfo)
}

// PrintError imprime un mensaje de error
func PrintError(errType, lang string) {
	msgs := i18n.Get(i18n.Lang(lang))
	red := color.New(color.FgRed)

	var msg string
	switch errType {
	case "no_api_key":
		msg = msgs.ErrorNoAPIKey
	case "invalid_key":
		msg = msgs.ErrorInvalidKey
	case "connection":
		msg = msgs.ErrorNoConnection
	case "limit_reached":
		msg = msgs.ErrorLimitReached
	default:
		msg = errType
	}

	red.Println(msg)
}

// PrintAbout imprime la informacion sobre el programa
func PrintAbout(version string) {
	fmt.Printf("pingbar v%s\n", version)
	fmt.Println()
	fmt.Println("Autor: https://github.com/686f6c61")
	fmt.Println()
	fmt.Println("Porque necesitabas saber si el bar esta abierto")
	fmt.Println("antes de salir de casa.")
}
