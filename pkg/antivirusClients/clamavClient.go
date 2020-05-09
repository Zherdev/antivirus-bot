// Антивирус-бот
// Клиент антивируса clamav
// Жердев Иван.

package antivirusClients

import (
	"antivirus-bot/pkg/common"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"os/exec"
)

var (
	ok            = []byte("\nInfected files: 0\n")
	virusExitCode = 1
)

// Клиент антивируса clamav
type ClamavClient struct {
}

func NewClamavClient() *ClamavClient {
	return &ClamavClient{}
}

// Проверка файла на угрозы через clamd
func (c *ClamavClient) CheckFile(filePath string, checkResult chan *common.FileForCheck) {
	checkCmd := exec.Command("clamdscan", "--fdpass", "--stream", filePath)
	out, err := checkCmd.Output()

	// ошибка clamdscan
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if !ok || (ok && exitError.ProcessState.ExitCode() != virusExitCode) {
			checkResult <- &common.FileForCheck{
				Err: errors.Wrap(err, "clamd error"),
			}
			return
		}
	}

	// нет угроз
	if bytes.Contains(out, ok) {
		checkResult <- &common.FileForCheck{
			IsOk: true,
		}
		return
	}

	// обнаружена угроза. Получаем ее название
	startPos := bytes.Index(out, []byte(filePath))
	if startPos == -1 {
		checkResult <- &common.FileForCheck{
			Err: fmt.Errorf("parse error in clamavClient"),
		}
	}
	startPos += len(filePath) + 1
	endPos := bytes.IndexRune(out, '\n')
	if endPos == -1 {
		checkResult <- &common.FileForCheck{
			Err: fmt.Errorf("parse error in clamavClient"),
		}
	}
	virusName := out[startPos:endPos]

	checkResult <- &common.FileForCheck{
		IsOk: false,
		Msg:  string(virusName),
	}
}
