// Антивирус-бот
// Жердев Иван.

package antivirusBot

import (
	"antivirus-bot/pkg/common"
	"antivirus-bot/pkg/configuration"
	"antivirus-bot/pkg/fileChecker"
	"context"
	"fmt"
	"github.com/mail-ru-im/bot-golang"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"net/url"
	"sync"
)

// Основная структура бота
type Bot struct {
	api         *botgolang.Bot // icq
	logger      *logrus.Logger
	checkerChan chan *common.FileForCheck // канал, в который бот отправляет файлы пользователей на проверку
	config      *configuration.Configuration
}

// Создает бота. fileBufferSize - размер очереди на проверку (размер буффера канала)
// пишет логи в logOut
func NewBot(
	checkerChan chan *common.FileForCheck,
	logOut io.Writer,
	config *configuration.Configuration) (*Bot, error) {
	method := "NewBot"

	if checkerChan == nil {
		return nil, fmt.Errorf("checkerChan is nil in %s", method)
	}

	api, err := botgolang.NewBot(config.IcqToken)
	if err != nil {
		return nil, errors.Wrapf(err, "botgolang error in %s", method)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.TraceLevel)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "30-12-2006 15:04:05",
	})
	logger.SetOutput(logOut)

	return &Bot{
		api:         api,
		logger:      logger,
		checkerChan: checkerChan,
		config:      config,
	}, nil
}

// Запускает бота. Остановка по контексту ctx
func (b *Bot) Run(ctx context.Context, papaWg *sync.WaitGroup) {
	updates := b.api.GetUpdatesChannel(ctx)
	wg := &sync.WaitGroup{}

	for {
		select {
		case update := <-updates:
			switch update.Type {
			case botgolang.NEW_MESSAGE:
				wg.Add(1)
				go func(update botgolang.Event, ctx context.Context) {
					b.processMessage(&update, ctx)
					wg.Done()
				}(update, ctx)

			default:
				// игнорим всё, кроме новых сообщений
			}

		case <-ctx.Done():
			wg.Wait()
			b.logger.Info("bot is done")
			papaWg.Done()
			return
		}
	}
}

// Обработка сообщения от пользователя
func (b *Bot) processMessage(update *botgolang.Event, ctx context.Context) {
	// Приветственное сообщение
	if update.Payload.Text == startUserMsg {
		b.sendText(update, startMsg)
		return
	}
	// Помощь, за нее отвечает icq
	if update.Payload.Text == helpUserMsg {
		return
	}

	// Вложений в сообщении нет
	if len(update.Payload.Parts) == 0 {
		// Пользователь мог прислать url в тексте
		fileUrl, err := url.ParseRequestURI(update.Payload.Text)
		if err != nil {
			b.sendText(update, badMessageMsg)
			return
		}

		b.sendText(update, getFileMsg)
		b.processUrl(update, fileUrl, ctx)
		return
	}

	// Иначе ожидаем в его сообщении файл
	wasFiles := false
	for _, part := range update.Payload.Parts {
		if part.Type == botgolang.FILE {
			b.sendText(update, getFileMsg)
			b.processFile(update, part.Payload.FileID, ctx)
			wasFiles = true
		}
	}

	// пользователь прислал сообщение без ссылки/файлов
	if !wasFiles {
		b.sendText(update, noFilesMsg)
	}
}

// Обработка ссылки от пользователя
func (b *Bot) processUrl(update *botgolang.Event, fileUrl *url.URL, ctx context.Context) {
	forCheck := &common.FileForCheck{
		Url: fileUrl.String(),
		Name: fileUrl.String(),
		Checked: make(chan struct{}, 1),
	}

	b.check(update, forCheck, ctx)
}

// Обработка файла от пользователя
func (b *Bot) processFile(update *botgolang.Event, fileId string, ctx context.Context) {
	method := "antivirusBot.processFile"

	file, err := b.api.GetFileInfo(fileId)
	if err != nil {
		b.sendText(update, getFileErrorMsg)
		b.logger.
			WithField("method", method).
			WithField("request", update).
			Error(err)
		return
	}

	if file.Size > b.config.FileMaxSize {
		b.sendText(update, fileTooBigMsg, file.Name, b.config.FileMaxSize)
		return
	}

	forCheck := &common.FileForCheck{
		Url: file.URL,
		Name: file.Name,
		Checked: make(chan struct{}, 1),
	}

	b.check(update, forCheck, ctx)
}

// отправляем файл на проверку
func (b* Bot) check(update *botgolang.Event, forCheck *common.FileForCheck, ctx context.Context) {
	b.checkerChan <- forCheck

	for {
		select {
		case <-forCheck.Checked: // проверка завершена
			if forCheck.Err == fileChecker.ErrFileTooBig {
				b.sendText(update, fileTooBigMsg, forCheck.Name, b.config.FileMaxSize)
				return
			}

			if forCheck.Err != nil {
				b.sendText(update, checkErrorMsg, forCheck.Name)
				b.logger.
					WithField("method", "antivirusBot.processFile").
					WithField("request", update).
					Error(forCheck.Err)
				return
			}
			if forCheck.IsOk {
				b.sendText(update, fileIsOkMsg, forCheck.Name)
				return
			}
			b.sendText(update, fileIsInfectedMsg, forCheck.Name, forCheck.Msg)
			return

		case <-ctx.Done(): // бота выключили
			b.sendText(update, sorryGoodbyeMsg, forCheck.Name)
			return
		}
	}
}

// Отправляем текст в сообщении пользователю
func (b *Bot) sendText(update *botgolang.Event, text string, args ...interface{}) {
	methodName := "antivirusBot.sendText"

	if len(args) > 0 {
		text = fmt.Sprintf(text, args...)
	} else {
		text = fmt.Sprintf(text)
	}
	err := update.Payload.Message().Reply(text)

	if err != nil {
		b.logger.
			WithField("method", methodName).
			WithField("request", update.Payload).
			Error(err)
	}
	b.logger.
		WithField("method", methodName).
		WithField("request", update.Payload).
		WithField("response", text).
		Trace()
}
