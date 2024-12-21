package config

import (
	"github.com/joho/godotenv"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"log"
	"os"
	"strings"
)

type Config struct {
	Port string
	DB   struct {
		Url string
	}
	OpenSearch struct {
		Addr        []string
		Login       string
		Password    string
		PosterIndex string `koanf:"poster_index"`
	}
	Logger struct {
		Level string
	}
	CategoriesJson string
}

var cfg = Config{}

func LoadConfig() {
	// Load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Ошибка загрузки .env файла: %v", err)
	}

	// Initialize Koanf
	k := koanf.New(".")

	// Load config from config.toml
	if err := k.Load(file.Provider("./config/config.toml"), toml.Parser()); err != nil {
		log.Printf("Ошибка загрузки config.toml: %v", err)
	}

	// Load environment variables
	err = k.Load(env.Provider("", "_", func(s string) string {
		return strings.ToLower(strings.ReplaceAll(s, "_", "."))
	}), nil)
	if err != nil {
		log.Fatalf("Ошибка загрузки переменных окружения: %v", err)
	}

	// Map configuration to struct
	if err := k.Unmarshal("", &cfg); err != nil {
		log.Fatalf("Ошибка маппинга конфигурации: %v", err)
	}

	// Read categories.json
	categoriesFile, err := os.ReadFile("categories.json")
	if err != nil {
		log.Fatalf("Ошибка чтения файла categories.json: %v", err)
	}

	cfg.CategoriesJson = string(categoriesFile)
}

func GetConfig() Config {
	return cfg
}
