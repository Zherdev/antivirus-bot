// Антивирус-бот
// Управление конфигурацией приложения
// Жердев Иван.

package configuration

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
)

// Хранит параметры конфигурации
type Configuration struct {
	Token string

	FileBufferSize int
	FileMaxSize uint64
	DownloadTimeout int64

	BotLogDir string
	CheckerLogDir string
	FilesDir string
}

// Читает конфигурацию из файла
func GetConfig(configFilePath string) (*Configuration, error) {
	file, err := os.Open(configFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "can't open config file in GetConfig")
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	result := &Configuration{}
	err = decoder.Decode(result)
	if err != nil {
		return nil, errors.Wrap(err, "can't decode config file in GetConfig")
	}

	return result, nil
}
