package config

import (
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

func Load(cfgFilePath string) (Config, error) {
	k := koanf.New(".")
	defaults := Default()

	if err := k.Load(structs.Provider(defaults, "koanf"), nil); err != nil {
		return Config{}, err
	}

	if cfgFilePath != "" {
		if err := k.Load(file.Provider(cfgFilePath), yaml.Parser()); err != nil {
			return Config{}, err
		}
	}

	k.Load(env.Provider(".", env.Opt{
		Prefix: "SAMARKAND_",
		TransformFunc: func(k, v string) (string, any) {
			k = strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, "SAMARKAND_")), "_", ".")
			return k, v
		},
	}), nil)

	var cfg Config
	return cfg, k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "koanf"})
}
