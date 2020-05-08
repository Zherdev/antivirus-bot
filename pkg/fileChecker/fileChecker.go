// Антивирус-бот
// Проверка файлов.
// main.go

package fileChecker

import (
	"antivirus-bot/pkg/antivirusClients"
	"antivirus-bot/pkg/common"
	"antivirus-bot/pkg/configuration"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Получает файлы от бота и отправляет их на проверку в антивирусы
type FileChecker struct {
	AvClients []antivirusClients.Client // сюда добавляем антивирусы, в которых проверяются файлы

	checkerChan chan *common.FileForCheck // канал, из которого получаем файлы
	logger *logrus.Logger
	downloader http.Client
	config *configuration.Configuration
}

// Создает файл-чекера, который получает файлы из checkerChan
// timeout на скачивание файла
// пишет логи в logOut
func NewFileChecker(
	checkerChan chan *common.FileForCheck,
	config *configuration.Configuration,
	logOut io.Writer) (*FileChecker, error) {

	if checkerChan == nil {
		return nil, fmt.Errorf("checkerChan is nil in NewFileChecker")
	}

	logger := logrus.New().WithField("agent", "FileChecker").Logger
	logger.SetLevel(logrus.TraceLevel)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "30-12-2006 15:04:05",
	})
	logrus.SetOutput(logOut)
	
	downloader := http.Client{
		Timeout:       time.Second * time.Duration(config.DownloadTimeout),
	}

	return &FileChecker{
		checkerChan: checkerChan,
		logger: logger,
		downloader: downloader,
		config: config,
	}, nil
}

// Запускает файл-чекер. Остановка по контексту ctx
func (c *FileChecker) Run(ctx context.Context) {
	for {
		select {
		case fileForCheck := <- c.checkerChan:
			c.checkFile(fileForCheck)

		case <- ctx.Done():
		}
	}
}

// Проверяет файл file. Остановка по контексту ctx
func (c *FileChecker) checkFile(file *common.FileForCheck) {
	filePath, err := c.downloadFile(file)
	if err != nil {
		file.IsOk = false
		file.Err = err
		file.Checked <- struct{}{}
		return
	}

	file.IsOk = true
	file.Err = nil
	resultChan := make(chan *common.FileForCheck, len(c.AvClients))
	msgBuilder := strings.Builder{}
	// Отправляем файлы на проверку в антивирусы
	for _, avClient := range c.AvClients {
		go avClient.CheckFile(filePath, resultChan)
	}
	for range c.AvClients {
		checkResult := <- resultChan
		if checkResult.Err != nil {
			// ошибка при проверке
			file.IsOk = false
			if file.Err == nil {
				file.Err = checkResult.Err
			} else {
				file.Err = fmt.Errorf("%s;\n%s", file.Err.Error(), checkResult.Err.Error())
			}
			continue
		}
		if !checkResult.IsOk {
			// обнаружена угроза
			file.IsOk = false
			msgBuilder.WriteString(";\n")
			msgBuilder.WriteString(checkResult.Msg)
			file.Checked <- struct{}{}
			continue
		}
	}

	file.Checked <- struct{}{}
}

// Возвращает путь к скачанному файлу
func (c *FileChecker) downloadFile(file *common.FileForCheck) (string, error) {
	method := "FileChecker.downloadFile"

	resp, err := c.downloader.Get(file.File.URL)
	if err != nil {
		c.logger.
			WithField("method", method).
			WithField("url", file.File.URL).
			Error(err)
		return "", errors.Wrap(err, "can't GET file in downloadFile")
	}
	defer resp.Body.Close()

	// Создаем файл
	filePath := fmt.Sprintf("%s/%d", c.config.FilesDir, time.Now().Unix())
	out, err := os.Create(filePath)
	if err != nil {
		c.logger.
			WithField("method", method).
			WithField("filePath", filePath).
			Error(err)
		return "", errors.Wrapf(err, "can't create file in %s", method)
	}
	defer out.Close()

	// Записываем в файл
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		c.logger.
			WithField("method", method).
			WithField("filePath", filePath).
			Error(err)
		return "", errors.Wrapf(err, "can't write to file in %s", method)
	}

	return filePath, nil
}
