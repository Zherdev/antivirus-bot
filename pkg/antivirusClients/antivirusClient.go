// Антивирус-бот
// Интерфейс клиента антивируса. Осуществляет непосредственную проверку файла
// main.go

package antivirusClients

import "antivirus-bot/pkg/common"

type Client interface {
	CheckFile(filePath string, checkResult chan *common.FileForCheck)
}