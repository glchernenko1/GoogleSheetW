package main

import (
	"GoogleSheetW/internal/app"
	"GoogleSheetW/internal/logger"
)

func main() {
	log := logger.Get()

	application := app.New()

	log.Info("Запуск приложения GoogleSheetW API")

	if err := application.Run(); err != nil {
		log.Fatalw("Ошибка запуска приложения", "error", err)
	}
}
