// Антивирус-бот
// Тексты сообщений, которые пишет бот
// Жердев Иван.

package antivirusBot

const (
	startUserMsg = "/start"

	helpMsg = "Отправьте мне в сообщении файлы, и я проверю их."
	startMsg = "Привет! Я Антивирус-бот.\nПроверю безопасность файлов. Использую антивирусы: ClamAV.\n\n" + helpMsg
	getFileMsg = "Обрабатываю файл #%d."
	fileIsOkMsg = "Угроз в '%s' не обнаружено. OK."
	fileIsInfectedMsg = "Файл '%s' заражен! Обнаружена угроза: %s. INFECTED."
	sorryGoodbyeMsg = "Извините, я завершаю свою работу, поэтому проверка файла '%s' не может быть завершена. Попробуйте отправить его еще раз позднее."

	checkErrorMsg = "Во время проверки файла '%s' произошла ошибка. Попробуйте отправить его еще раз."
	badMessageMsg = "Я не могу понять Ваше сообщение.\n" + helpMsg
	noFilesMsg = "Я не нашел в Вашем сообщении файлов для проверки.\n" + helpMsg
	getFileErrorMsg = "Произошла ошибка при получении файла. Попробуйте отправить его еще раз."
	fileTooBigMsg = "Файл '%s' слишком велик: его размер больше, чем %d байт."
)