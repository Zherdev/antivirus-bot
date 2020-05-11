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

// Проверка файла на угрозы через clamdscan
func (c *ClamavClient) CheckFile(
	fileForCheck *common.FileForCheck,
	checkResult chan *common.AntivirusResult) {

	checkCmd := exec.Command("clamdscan", "--fdpass", "--stream", fileForCheck.Path)
	out, err := checkCmd.Output()

	// ошибка clamdscan
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if !ok || (ok && exitError.ProcessState.ExitCode() != virusExitCode) {
			checkResult <- &common.AntivirusResult{
				Err: errors.Wrap(err, "clamd error"),
			}
			return
		}
	}

	// нет угроз
	if bytes.Contains(out, ok) {
		checkResult <- &common.AntivirusResult{
			IsOk: true,
		}
		return
	}

	// обнаружена угроза. Получаем ее название
	startPos := bytes.Index(out, []byte(fileForCheck.Path))
	if startPos == -1 {
		checkResult <- &common.AntivirusResult{
			Err: fmt.Errorf("parse error in clamavClient"),
		}
	}
	startPos += len(fileForCheck.Path) + 1
	endPos := bytes.IndexRune(out, '\n')
	if endPos == -1 {
		checkResult <- &common.AntivirusResult{
			Err: fmt.Errorf("parse error in clamavClient"),
		}
	}
	virusName := out[startPos:endPos]

	checkResult <- &common.AntivirusResult{
		IsOk: false,
		Msg:  string(virusName),
	}
}
