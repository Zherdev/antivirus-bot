// Антивирус-бот
// Жердев Иван.

package application

import (
	"antivirus-bot/pkg/antivirusBot"
	"antivirus-bot/pkg/antivirusClients"
	"antivirus-bot/pkg/common"
	"antivirus-bot/pkg/configuration"
	"antivirus-bot/pkg/fileChecker"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Основная структура приложения
type App struct {
	bot         *antivirusBot.Bot
	checker     *fileChecker.FileChecker
	config      *configuration.Configuration
	logger      *logrus.Logger
	closeOnExit []*os.File
}

// Инициализация приложения
func NewApp(configFilePath string) *App {
	var err error
	a := &App{closeOnExit: []*os.File{}}

	a.logger = logrus.New()
	a.logger.SetLevel(logrus.TraceLevel)
	a.logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "30-12-2006 15:04:05",
	})
	a.logger.SetOutput(os.Stdout)
	a.logger.Info("Init app...")

	// получаем конфигурацию приложения
	a.config, err = configuration.GetConfig(configFilePath)
	if err != nil {
		a.logger.Fatal("can't get configuration in main(): ", err.Error())
	}

	// сюда пишем логи
	botLog, err := a.logOut(a.config.BotLogDir)
	if err != nil {
		a.logger.Fatal("can't create logfile for bot: ", err.Error())
	}
	checkerLog, err := a.logOut(a.config.CheckerLogDir)
	if err != nil {
		a.done()
		a.logger.Fatal("can't create logfile for checker: ", err.Error())
	}

	// Канал для передачи файлов
	filesChan := make(chan *common.FileForCheck, a.config.FileBufferSize)

	// icq бот
	a.bot, err = antivirusBot.NewBot(filesChan, botLog, a.config)
	if err != nil {
		a.done()
		a.logger.Fatal("can't create antivirus bot: ", err.Error())
	}

	// Проверяет файлы, полученные от бота
	a.checker, err = fileChecker.NewFileChecker(filesChan, checkerLog, a.config)
	if err != nil {
		a.done()
		a.logger.Fatal("can't create checker: ", err.Error())
	}

	// Добавляем клиенты-антивирусы
	_ = a.checker.AddAv(antivirusClients.NewClamavClient())

	return a
}

func (a *App) logOut(dirPath string) (io.Writer, error) {
	newPath := filepath.Join(".", dirPath)
	os.MkdirAll(newPath, os.ModePerm)

	logFile, err := os.Create(fmt.Sprintf("%s/%d", dirPath, time.Now().Unix()))
	if err != nil {
		return nil, err
	}

	a.closeOnExit = append(a.closeOnExit, logFile)

	return io.MultiWriter(logFile, os.Stdout), nil
}

// Запускает приложение. Остановка по контексту ctx
func (a *App) Run(ctx context.Context) {
	a.logger.Info("Start app...")

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go a.bot.Run(ctx, wg)
	go a.checker.Run(ctx, wg)

	<-ctx.Done()
	wg.Wait()
	a.done()
}

// Закрывает открытые файлы
func (a *App) done() {
	for _, file := range a.closeOnExit {
		file.Close()
	}
	a.logger.Info("done")
}
