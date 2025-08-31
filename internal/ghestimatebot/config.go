package ghestimatebot

import (
	"errors"
	"os"
	"reflect"
)

type Config struct {
	AppId          string `env:"GH_APP_ID"`
	PrivateKeyPath string `env:"GH_PRIVATE_KEY_PATH"`
	WebhookSecret  string `env:"GH_WEBHOOK_SECRET"`

	Port string `env:"PORT" envDefault:"8080"`
}

func LoadConfigFromEnv() (*Config, error) {
	var cfg Config

	val := reflect.ValueOf(&cfg).Elem()
	t := val.Type()

	for i := 0; i < t.NumField(); i++ {
		field := val.Field(i)
		fieldType := t.Field(i)
		envKey := fieldType.Tag.Get("env")

		if envKey == "" {
			continue
		}

		envVal, exists := os.LookupEnv(envKey)
		if !exists {
			envVal = fieldType.Tag.Get("envDefault")

			if envVal == "" {
				return nil, errors.New("env key " + envKey + " not set")
			}
		}

		field.SetString(envVal)
	}

	return &cfg, nil
}
