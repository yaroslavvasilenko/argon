package config

import (
	"github.com/joho/godotenv"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"log"
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
}

var cfg = Config{}

func LoadConfig() {
	// Загрузка .env файла
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Ошибка загрузки .env файла: %v", err)
	}

	// Инициализация Koanf
	k := koanf.New(".")

	// Загрузка конфигурации из config.toml
	if err := k.Load(file.Provider("./config/config.toml"), toml.Parser()); err != nil {
		log.Printf("Ошибка загрузки config.toml: %v", err)
	}

	// Загрузка переменных окружения (перекрывают значения из TOML)
	err = k.Load(env.Provider("", "_", func(s string) string {
		// Переводим ключи переменных в нижний регистр для совместимости
		return strings.ToLower(strings.ReplaceAll(s, "_", "."))
	}), nil)
	if err != nil {
		log.Fatalf("Ошибка загрузки переменных окружения: %v", err)
	}

	// Маппинг переменных в структуру Config
	if err := k.Unmarshal("", &cfg); err != nil {
		log.Fatalf("Ошибка маппинга конфигурации: %v", err)
	}
}

func GetConfig() Config {
	return cfg
}
