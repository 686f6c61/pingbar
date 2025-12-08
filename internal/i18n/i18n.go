package i18n

// Lang representa el idioma actual
type Lang string

const (
	ES Lang = "es"
	EN Lang = "en"
)

// Messages contiene todos los mensajes traducidos
type Messages struct {
	Open            string
	Closed          string
	Unknown         string
	Today           string
	Tomorrow        string
	ClosesIn        string
	ClosedAgo       string
	Holiday         string
	SpecialHours    string
	NoSchedule      string
	NotFound        string
	Found           string
	MoreResults     string
	ViewAll         string
	Days            []string
	WelcomeTitle    string
	NoAPIKey        string
	GetAPIKey       string
	MoreInfo        string
	ErrorNoAPIKey   string
	ErrorInvalidKey string
	ErrorNoConnection string
	ErrorLimitReached string
	ConfigSet       string
	ConfigGet       string
	CacheCleared    string
	UninstallConfirm string
	UninstallDone   string
	DeleteConfig    string
	DeleteCache     string
	Yes             string
	No              string
}

var translations = map[Lang]Messages{
	ES: {
		Open:            "ABIERTO",
		Closed:          "CERRADO",
		Unknown:         " --:-- ",
		Today:           "Hoy",
		Tomorrow:        "Mañana",
		ClosesIn:        "cierra en %s",
		ClosedAgo:       "cerró hace %s",
		Holiday:         "Hoy es festivo, puede que no esté abierto",
		SpecialHours:    "horario especial",
		NoSchedule:      "Horario no disponible",
		NotFound:        "No se encontraron resultados para \"%s\" en \"%s\"",
		Found:           "Encontrados: %d resultados",
		MoreResults:     "Hay %d resultados más. ¿Ver todos? [Y/N]: ",
		ViewAll:         "... (%d más)",
		Days:            []string{"domingo", "lunes", "martes", "miércoles", "jueves", "viernes", "sábado"},
		WelcomeTitle:    "Bienvenido a pingbar",
		NoAPIKey:        "No se ha configurado una API Key.",
		GetAPIKey:       "1. Ve a https://serper.dev y crea una cuenta gratuita\n2. Copia tu API Key\n3. Ejecuta: pingbar config set apikey TU_API_KEY",
		MoreInfo:        "Más info: pingbar --help",
		ErrorNoAPIKey:   "No se ha configurado una API Key. Ejecuta: pingbar config set apikey TU_KEY",
		ErrorInvalidKey: "API Key inválida o expirada. Verifica tu key en https://serper.dev",
		ErrorNoConnection: "No se pudo conectar. Verifica tu conexión a internet",
		ErrorLimitReached: "Has alcanzado el límite de búsquedas. Más info en https://serper.dev",
		ConfigSet:       "Configuración guardada: %s = %s",
		ConfigGet:       "%s = %s",
		CacheCleared:    "Caché limpiada correctamente",
		UninstallConfirm: "¿Estás seguro de que deseas desinstalar pingbar? [Y/N]: ",
		UninstallDone:   "pingbar ha sido desinstalado correctamente",
		DeleteConfig:    "¿Deseas eliminar la configuración (~/.config/pingbar/)? [Y/N]: ",
		DeleteCache:     "¿Deseas eliminar la caché? [Y/N]: ",
		Yes:             "Y",
		No:              "N",
	},
	EN: {
		Open:            "OPEN",
		Closed:          "CLOSED",
		Unknown:         " --:-- ",
		Today:           "Today",
		Tomorrow:        "Tomorrow",
		ClosesIn:        "closes in %s",
		ClosedAgo:       "closed %s ago",
		Holiday:         "Today is a holiday, it may not be open",
		SpecialHours:    "special hours",
		NoSchedule:      "Schedule not available",
		NotFound:        "No results found for \"%s\" in \"%s\"",
		Found:           "Found: %d results",
		MoreResults:     "There are %d more results. View all? [Y/N]: ",
		ViewAll:         "... (%d more)",
		Days:            []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"},
		WelcomeTitle:    "Welcome to pingbar",
		NoAPIKey:        "No API Key configured.",
		GetAPIKey:       "1. Go to https://serper.dev and create a free account\n2. Copy your API Key\n3. Run: pingbar config set apikey YOUR_API_KEY",
		MoreInfo:        "More info: pingbar --help",
		ErrorNoAPIKey:   "No API Key configured. Run: pingbar config set apikey YOUR_KEY",
		ErrorInvalidKey: "Invalid or expired API Key. Check your key at https://serper.dev",
		ErrorNoConnection: "Could not connect. Check your internet connection",
		ErrorLimitReached: "You have reached the search limit. More info at https://serper.dev",
		ConfigSet:       "Configuration saved: %s = %s",
		ConfigGet:       "%s = %s",
		CacheCleared:    "Cache cleared successfully",
		UninstallConfirm: "Are you sure you want to uninstall pingbar? [Y/N]: ",
		UninstallDone:   "pingbar has been uninstalled successfully",
		DeleteConfig:    "Do you want to delete configuration (~/.config/pingbar/)? [Y/N]: ",
		DeleteCache:     "Do you want to delete cache? [Y/N]: ",
		Yes:             "Y",
		No:              "N",
	},
}

// Get devuelve los mensajes para el idioma especificado
func Get(lang Lang) Messages {
	if msgs, ok := translations[lang]; ok {
		return msgs
	}
	return translations[ES]
}

// GetDay devuelve el nombre del día en el idioma especificado
func GetDay(lang Lang, dayIndex int) string {
	msgs := Get(lang)
	if dayIndex >= 0 && dayIndex < len(msgs.Days) {
		return msgs.Days[dayIndex]
	}
	return ""
}
