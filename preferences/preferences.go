package preferences

type Preferences struct {
	Port        string `envconfig:"PORT"`
	LogLevel    string `envconfig:"LOG_LEVEL"`
	LogFormat   string `envconfig:"LOG_FORMAT"`
	DatabaseURL string `envconfig:"DATABASE_URL"`
}
