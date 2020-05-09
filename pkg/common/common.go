// Антивирус-бот
// Жердев Иван.

package common

import botgolang "github.com/mail-ru-im/bot-golang"

// Структура для отправки файла на проверку и получения результата проверки
type FileForCheck struct {
	File    *botgolang.File
	Checked chan struct{}
	IsOk    bool
	Err     error
	Msg     string
}
