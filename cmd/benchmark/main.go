package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yaroslavvasilenko/argon/internal/modules/test"
)

func main() {
	// Определение флагов командной строки
	count := flag.Int("count", 1000, "Количество записей для генерации")
	flag.Parse()

	// Проверка аргументов
	if *count <= 0 {
		fmt.Println("Ошибка: количество записей должно быть положительным числом")
		os.Exit(1)
	}

	// Запуск бенчмарка
	fmt.Printf("Запуск бенчмарка с генерацией %d записей...\n", *count)
	err := modules.RunBenchmark(*count)
	if err != nil {
		fmt.Printf("Ошибка при выполнении бенчмарка: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Бенчмарк успешно завершен")
}
