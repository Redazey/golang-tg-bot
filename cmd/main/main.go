package main

import (
	"log"
	"tgssn/internal/pkg/bot"
)

func main() {
	application, err := bot.Init()
	if err != nil {
		log.Fatalf("Ошибка при инициализации: %s", err)
	}
	err = application.Run()
	if err != nil {
		log.Fatalf("Ошибка при запуске: %s", err)
	}
}
