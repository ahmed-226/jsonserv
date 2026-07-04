package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig            `yaml:"server"`
	Entities map[string]EntityConfig `yaml:"entities,omitempty"`
}

type ServerConfig struct {
	Port    int        `yaml:"port"`
	Host    string     `yaml:"host"`
	Cors    CorsConfig `yaml:"cors"`
	Logging bool       `yaml:"logging"`
}

type CorsConfig struct {
	Enabled bool     `yaml:"enabled"`
	Origins []string `yaml:"origins"`
}

type EntityConfig struct {
	Alias    string            `yaml:"alias,omitempty"`
	Filters  FilterConfig      `yaml:"filters,omitempty"`
	Sort     SortConfig        `yaml:"sort,omitempty"`
	Paginate PaginateConfig    `yaml:"paginate,omitempty"`
	Schema   map[string]string `yaml:"schema,omitempty"`
}

type FilterConfig struct {
	Enabled bool     `yaml:"enabled"`
	Fields  []string `yaml:"fields,omitempty"`
}

type SortConfig struct {
	Enabled      bool   `yaml:"enabled"`
	DefaultField string `yaml:"default_field,omitempty"`
	DefaultOrder string `yaml:"default_order,omitempty"`
}

type PaginateConfig struct {
	Enabled      bool `yaml:"enabled"`
	DefaultPage  int  `yaml:"default_page,omitempty"`
	DefaultLimit int  `yaml:"default_limit,omitempty"`
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port:    3000,
			Host:    "0.0.0.0",
			Cors:    CorsConfig{Enabled: true, Origins: []string{"*"}},
			Logging: true,
		},
		Entities: make(map[string]EntityConfig),
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Entities == nil {
		cfg.Entities = make(map[string]EntityConfig)
	}
	return cfg, nil
}

func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (c *Config) EntityConfig(name string) EntityConfig {
	if c.Entities != nil {
		if ecfg, ok := c.Entities[name]; ok {
			return ecfg
		}
	}
	return EntityConfig{}
}

func InferFromData(data map[string][]map[string]interface{}) map[string]EntityConfig {
	entities := make(map[string]EntityConfig)
	for name, items := range data {
		ecfg := EntityConfig{
			Filters:  FilterConfig{Enabled: true},
			Sort:     SortConfig{Enabled: true, DefaultField: "id", DefaultOrder: "asc"},
			Paginate: PaginateConfig{Enabled: true, DefaultPage: 1, DefaultLimit: 10},
			Schema:   inferFields(items),
		}
		entities[name] = ecfg
	}
	return entities
}

func inferFields(items []map[string]interface{}) map[string]string {
	fields := make(map[string]string)
	for _, item := range items {
		for k, v := range item {
			if _, ok := fields[k]; !ok {
				fields[k] = typeName(v)
			}
		}
	}
	return fields
}

func typeName(v interface{}) string {
	switch v.(type) {
	case float64:
		return "number"
	case string:
		return "string"
	case bool:
		return "boolean"
	default:
		return "string"
	}
}
