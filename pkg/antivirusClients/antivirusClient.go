// Антивирус-бот
// Интерфейс клиента антивируса. Осуществляет непосредственную проверку файла
// Жердев Иван.

package antivirusClients

import "antivirus-bot/pkg/common"

type Client interface {
	CheckFile(fileForCheck *common.FileForCheck, checkResult chan *common.AntivirusResult)
}
