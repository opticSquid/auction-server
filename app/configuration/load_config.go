package configuration

import (
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

func LoadConfig(k *koanf.Koanf, logger *zap.Logger) {
	// Load configuration from file
	if err := k.Load(file.Provider("resources/application.yaml"), yaml.Parser()); err != nil {
		logger.Fatal("could not load configuration. exiting application", zap.Error(err))
	}

	// Load configurations from environment variables to override configurations from yaml file
	k.Load(env.Provider(".", env.Opt{
		TransformFunc: func(k, v string) (string, any) {
			// converts env vars to lower case and replaces _ in env var names to . to match yaml key names
			k = strings.ReplaceAll(strings.ToLower(k), "_", ".")

			// If value contain spaces converts it to slices
			if strings.Contains(v, " ") {
				return k, strings.Split(v, " ")
			}
			return k, v
		},
	}), nil)
}
