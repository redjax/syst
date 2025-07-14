package config

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

var K = koanf.New(".")

func LoadConfig(flagSet *pflag.FlagSet, configFile string) {
	// Load from config file if provided
	if configFile != "" {
		parser, err := parserForFile(configFile)
		if err != nil {
			log.Fatalf("unsupported config file format: %v", err)
		}
		if err := K.Load(file.Provider(configFile), parser); err != nil {
			log.Printf("error loading config file: %v", err)
		}
	}

	// Load from environment variables (prefix "SYST_")
	// This will convert SYST_FOO_BAR to foo.bar
	K.Load(env.Provider("SYST_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "SYST_")), "_", ".", -1)
	}), nil)

	// Load from command-line flags (highest precedence)
	K.Load(posflag.Provider(flagSet, ".", K), nil)
}

func parserForFile(path string) (koanf.Parser, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return yaml.Parser(), nil
	case ".json":
		return json.Parser(), nil
	case ".toml":
		return toml.Parser(), nil
	case ".env":
		return dotenv.Parser(), nil
	default:
		return nil, fmt.Errorf("unknown file extension: %s", ext)
	}
}
