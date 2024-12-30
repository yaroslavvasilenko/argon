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
	// Initialize Koanf
	k := koanf.New(".")

	// Load config from config.toml first
	if err := k.Load(file.Provider("./config/config.toml"), toml.Parser()); err != nil {
		log.Printf("Ошибка загрузки config.toml: %v", err)
	}

	// Load .env file if exists
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Информация: .env файл не найден: %v", err)
	}

	// Load environment variables with prefix APP_
	callback := func(s string) string {
		return strings.ToLower(strings.ReplaceAll(strings.TrimPrefix(s, "APP_"), "_", "."))
	}
	if err := k.Load(env.Provider("APP_", "_", callback), nil); err != nil {
		log.Printf("Ошибка загрузки переменных окружения: %v", err)
	}

	// Unmarshal config into struct
	if err := k.Unmarshal("", &cfg); err != nil {
		log.Fatalf("Ошибка при разборе конфигурации: %v", err)
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
