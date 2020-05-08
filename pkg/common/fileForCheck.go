// Антивирус-бот
// Жердев Иван.

package common

import botgolang "github.com/mail-ru-im/bot-golang"

// Структура для отправки файла на проверку и получения результата
type FileForCheck struct {
	File *botgolang.File
	Checked chan struct{}
	IsOk bool
	Err error
	Result string
}
