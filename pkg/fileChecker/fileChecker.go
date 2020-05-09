// Антивирус-бот
// Проверка файлов.
// Жердев Иван

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
	"sync"
	"time"
)

// Получает файлы от бота и отправляет их на проверку в антивирусы
type FileChecker struct {
	avClients   []antivirusClients.Client // антивирусы, в которых проверяются файлы
	checkerChan chan *common.FileForCheck // канал, из которого получаем файлы
	logger      *logrus.Logger
	downloader  http.Client
	config      *configuration.Configuration
}

// Создает файл-чекера, который получает файлы из checkerChan
// timeout на скачивание файла
// пишет логи в logOut
func NewFileChecker(
	checkerChan chan *common.FileForCheck,
	logOut io.Writer,
	config *configuration.Configuration) (*FileChecker, error) {

	if checkerChan == nil {
		return nil, fmt.Errorf("checkerChan is nil in NewFileChecker")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.TraceLevel)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "30-12-2006 15:04:05",
	})
	logger.SetOutput(logOut)

	downloader := http.Client{
		Timeout: time.Second * time.Duration(config.DownloadTimeout),
	}

	return &FileChecker{
		checkerChan: checkerChan,
		logger:      logger,
		downloader:  downloader,
		config:      config,
	}, nil
}

// Добавить клиент антивируса
func (c *FileChecker) AddAv(client antivirusClients.Client) error {
	if client == nil {
		return fmt.Errorf("client is nil in FileChecker.Add")
	}

	if c.avClients == nil {
		c.avClients = []antivirusClients.Client{}
	}
	c.avClients = append(c.avClients, client)

	return nil
}

// Запускает файл-чекер. Остановка по контексту ctx
func (c *FileChecker) Run(ctx context.Context, papaWg *sync.WaitGroup) {
	for {
		select {
		case fileForCheck := <-c.checkerChan:
			c.checkFile(fileForCheck)

		case <-ctx.Done():
			c.logger.Info("fileChecker is done")
			papaWg.Done()
			return
		}
	}
}

// Проверяет файл file
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
	resultChan := make(chan *common.FileForCheck, len(c.avClients))
	msgBuilder := strings.Builder{}
	// Отправляем файлы на проверку в антивирусы
	for _, avClient := range c.avClients {
		go avClient.CheckFile(filePath, resultChan)
	}
	for range c.avClients {
		checkResult := <-resultChan
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
			msgBuilder.WriteString(checkResult.Msg)
			msgBuilder.WriteString(";\n")
			file.Msg = msgBuilder.String()
			file.Checked <- struct{}{}
			continue
		}
	}

	file.Checked <- struct{}{}

	err = os.Remove(filePath)
	if err != nil {
		c.logger.WithField("method", "checkFile").Error(err)
	}
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
	os.MkdirAll("./"+c.config.FilesDir, os.ModePerm)
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

	c.logger.
		WithField("method", method).
		WithField("filePath", filePath).
		Trace("downloaded")
	return filePath, nil
}
