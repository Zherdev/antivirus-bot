// Антивирус-бот
// Жердев Иван.

package common

// Структура для отправки файла на проверку и получения в боте результата проверки
type FileForCheck struct {
	Url     string
	Path    string
	Name    string
	Checked chan struct{} // В этот канал FileChecker отправляет struct{}, сообщая боту, что проверка завершена
	IsOk    bool
	Err     error
	Msg     string
}

// Структура, в которой клиент-антивирус отправляет FileChecker'у результат проверки
type AntivirusResult struct {
	IsOk bool
	Err  error
	Msg  string
}
