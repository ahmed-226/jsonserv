package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"myserv/internal/config"
	"myserv/internal/server"
	"myserv/internal/store"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  myserv <file.json>           Start server with JSON data")
		fmt.Println("  myserv <file.yaml>           Start server from YAML config (generates empty data)")
		fmt.Println("  myserv init <file.json>       Generate YAML config from JSON")
		fmt.Println("  myserv init <file.json> --empty  Generate YAML config + empty JSON")
		os.Exit(1)
	}

	if os.Args[1] == "init" {
		cmdInit()
		return
	}

	cmdStart()
}

func cmdInit() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: myserv init <file.json> [--empty]")
		os.Exit(1)
	}

	jsonPath := os.Args[2]
	emptyFlag := len(os.Args) > 3 && os.Args[3] == "--empty"

	s, err := store.New(jsonPath)
	if err != nil {
		log.Fatalf("failed to load %s: %v", jsonPath, err)
	}

	data := make(map[string][]map[string]interface{})
	for _, entity := range s.Entities() {
		data[entity] = s.List(entity)
	}

	entities := config.InferFromData(data)
	cfg := config.Default()
	cfg.Entities = entities

	yamlPath := strings.TrimSuffix(jsonPath, ".json") + ".yaml"
	if err := config.Save(yamlPath, cfg); err != nil {
		log.Fatalf("failed to save config: %v", err)
	}
	fmt.Printf("config saved: %s\n", yamlPath)

	if emptyFlag {
		emptyData := make(map[string]interface{})
		for _, entity := range s.Entities() {
			emptyData[entity] = []interface{}{}
		}
		data, _ := json.MarshalIndent(emptyData, "", "  ")
		if err := os.WriteFile(jsonPath, data, 0644); err != nil {
			log.Fatalf("failed to write empty data: %v", err)
		}
		fmt.Printf("data emptied: %s\n", jsonPath)
	}
}

func cmdStart() {
	filePath := os.Args[1]
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".yaml", ".yml":
		startFromYAML(filePath)
	case ".json":
		startFromJSON(filePath)
	default:
		log.Fatalf("unsupported file type: %s (use .json or .yaml)", ext)
	}
}

func startFromYAML(yamlPath string) {
	cfg, err := config.Load(yamlPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	jsonPath := strings.TrimSuffix(yamlPath, ".yaml") + ".json"
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		var entities []string
		for name := range cfg.Entities {
			entities = append(entities, name)
		}
		s := store.NewEmpty(entities)
		s.SetFilePath(jsonPath)

		raw := make(map[string]interface{})
		for _, entity := range s.Entities() {
			raw[entity] = []interface{}{}
		}
		data, _ := json.MarshalIndent(raw, "", "  ")
		if err := os.WriteFile(jsonPath, data, 0644); err != nil {
			log.Fatalf("failed to write initial data: %v", err)
		}
		fmt.Printf("data file created: %s\n", jsonPath)

		srv := server.New(cfg, s)
		log.Fatal(srv.Start())
	} else {
		s, err := store.New(jsonPath)
		if err != nil {
			log.Fatalf("failed to load data: %v", err)
		}
		srv := server.New(cfg, s)
		log.Fatal(srv.Start())
	}
}

func startFromJSON(jsonPath string) {
	s, err := store.New(jsonPath)
	if err != nil {
		log.Fatalf("failed to load %s: %v", jsonPath, err)
	}

	yamlPath := strings.TrimSuffix(jsonPath, ".json") + ".yaml"
	cfg := config.Default()

	if _, err := os.Stat(yamlPath); err == nil {
		cfg, err = config.Load(yamlPath)
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
	} else {
		data := make(map[string][]map[string]interface{})
		for _, entity := range s.Entities() {
			data[entity] = s.List(entity)
		}
		cfg.Entities = config.InferFromData(data)
		if err := config.Save(yamlPath, cfg); err != nil {
			log.Printf("warning: could not save config: %v", err)
		} else {
			fmt.Printf("config auto-generated: %s\n", yamlPath)
		}
	}

	srv := server.New(cfg, s)
	log.Fatal(srv.Start())
}
