package config

import (
	// use embed macro for JSON schema.
	_ "embed"
	"fmt"
	"sort"
	"strings"

	"github.com/kaptinlin/jsonschema"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

//go:embed config.schema.json
var jsonSchema []byte

type ErrDefaultsLoading struct {
	Err error
}

func (e *ErrDefaultsLoading) Error() string {
	return fmt.Sprintf("failed to load defaults: %v", e.Err)
}

type ErrConfigValidation struct {
	Errors map[string]string
}

func (e *ErrConfigValidation) Error() string {
	paths := make([]string, 0, len(e.Errors))
	for p := range e.Errors {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	msgs := make([]string, 0, len(paths))
	for _, p := range paths {
		msgs = append(msgs, fmt.Sprintf("%s: %s", p, e.Errors[p]))
	}
	return "configuration validation failed: " + strings.Join(msgs, "; ")
}

func Load(fp string) (Config, error) {
	var cfg Config

	k := koanf.New(".")

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(jsonSchema)
	if err != nil {
		return Config{}, err
	}

	// Loads defaults from JSON schema
	if err := schema.Unmarshal(&cfg, []byte("{}")); err != nil {
		return Config{}, &ErrDefaultsLoading{Err: err}
	}

	if fp != "" {
		if err := k.Load(file.Provider(fp), yaml.Parser()); err != nil {
			return Config{}, err
		}
	}

	err = k.Load(env.Provider(".", env.Opt{
		Prefix: "SAMARKAND__",
		TransformFunc: func(k, v string) (string, any) {
			k = strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, "SAMARKAND__")), "_", ".")
			return k, v
		},
	}), nil)
	if err != nil {
		return Config{}, err
	}

	if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		return Config{}, err
	}

	if result := schema.ValidateStruct(cfg); !result.IsValid() {
		return cfg, &ErrConfigValidation{Errors: result.DetailedErrors()}
	}

	return cfg, nil
}
