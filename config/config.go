package config

import (
	"github.com/joho/godotenv"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/env"
	"strings"

	"log"
)

type Config struct {
	Port string
	DB   struct {
		Url string
	}
}

var cfg = Config{}

func LoadConfig() {
	// Загрузка .env файла
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}

	// Инициализация Koanf
	k := koanf.New(".")

	// Загрузка переменных окружения
	err = k.Load(env.Provider("", "_", func(s string) string {
		// Переводим ключи переменных в верхний регистр
		return strings.ToUpper(s)
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
