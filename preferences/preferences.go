package preferences

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

var ErrCantParse = errors.New("Can't parse config")

// Prerefences provides fields for representation of ENV variables config
type Preferences struct {
	Port        string `envconfig:"PORT"`
	LogLevel    string `envconfig:"LOG_LEVEL"`
	LogFormat   string `envconfig:"LOG_FORMAT"`
	DatabaseURL string `envconfig:"DATABASE_URL"`
}

// Returnes struct with filled fields
func Get() (*Preferences, error) {
	var p Preferences
	if err := envconfig.Process("", &p); err != nil {
		return nil, ErrCantParse
	}
	return &p, nil
}
