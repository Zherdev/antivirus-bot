// Антивирус-бот
// Старт приложения.
// Жердев Иван.
// main.go

package main

import (
	"antivirus-bot/pkg/application"
)

const (
	configFilePath = "config.json"
)

func main() {
	app := application.NewApp(configFilePath)

	app.Run()
}
