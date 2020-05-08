// Антивирус-бот
// Жердев Иван.

package antivirusBot

import (
	"antivirus-bot/pkg/common"
	"context"
	"fmt"
	"github.com/mail-ru-im/bot-golang"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
)

// Основная структура бота
type Bot struct {
	api *botgolang.Bot // icq
	logger *logrus.Logger
	fileChecker chan *common.FileForCheck // канал, в который бот отправляет файлы пользователей на проверку
}

// Создает бота. fileBufferSize - размер очереди на проверку (размер буффера канала)
// пишет логи в logOut
func NewBot(token string, fileBufferSize int, logOut io.Writer) (*Bot, error) {
	api, err := botgolang.NewBot(token)
	if err != nil {
		return nil, errors.Wrap(err, "botgolang error in NewBot")
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "30-12-2006 15:04:05",
	})
	logrus.SetOutput(logOut)

	return &Bot{
		api: api,
		logger: logger,
		fileChecker: make(chan *common.FileForCheck, fileBufferSize),
	}, nil
}

// Запускает бота. Остановка по контексту ctx
func (b *Bot) Run(ctx context.Context) {
	updates := b.api.GetUpdatesChannel(ctx)

	for update := range updates {
		switch update.Type {
		case botgolang.NEW_MESSAGE:
			go func(update botgolang.Event, ctx context.Context) {
				b.processMessage(&update, ctx)
			}(update, ctx)

		default:
			// игнорим всё, кроме новых сообщений
		}
	}
}

// Обработка сообщения от пользователя
func (b *Bot) processMessage(update *botgolang.Event, ctx context.Context) {
	if len(update.Payload.Parts) == 0 {
		b.sendText(update, badMessageMsg)
		return
	}

	// Приветственное сообщение
	if update.Payload.Text == startUserMsg {
		b.sendText(update, startMsg)
		return
	}

	wasFiles := false
	for idx, part := range update.Payload.Parts {
		if part.Type == botgolang.FILE {
			b.sendText(update, getFileMsg, idx + 1)
			b.processFile(update, part.Payload.FileID, ctx)
			wasFiles = true
		}
	}

	// пользователь прислал сообщение без файлов
	if !wasFiles {
		b.sendText(update, noFilesMsg)
	}
}

// Обработка файла от пользователя. Отправляем его на проверку
func (b *Bot) processFile(update *botgolang.Event, fileId string, ctx context.Context) {
	file, err := b.api.GetFileInfo(fileId)
	if err != nil {
		b.sendText(update, getFileErrorMsg)
	}

	forCheck := &common.FileForCheck{
		File: file,
		Checked: make(chan struct{}, 1),
	}

	// отправляем файл на проверку
	b.fileChecker <- forCheck

	for {
		select {
		case <-forCheck.Checked: // проверка завершена
			if forCheck.Err != nil {
				b.sendText(update, checkErrorMsg, file.Name)
				b.logger.
					WithField("method", "antivirusBot.processFile").
					Error(err)
			}
			if forCheck.IsOk {
				b.sendText(update, fileIsOkMsg, file.Name)
				return
			}
			b.sendText(update, fileIsInfectedMsg, file.Name, forCheck.Result)
			return

		case <-ctx.Done(): // бота выключили
			b.sendText(update, sorryGoodbyeMsg, file.Name)
			return
		}
	}
}

// Отправляем текст в сообщении пользователю
func (b *Bot) sendText(update *botgolang.Event, text string, args... interface{}) {
	methodName := "antivirusBot.sendText"

	err := update.Payload.Message().Reply(fmt.Sprintf(text, args))
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
