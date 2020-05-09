// Антивирус-бот
// Админка для управления приложением
// Жердев Иван.

package admin

import (
	"antivirus-bot/pkg/configuration"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"sync"
	"time"
)

type Admin struct {
	config *configuration.Configuration
	logger *logrus.Logger
}

func NewAdmin(logOut io.Writer, config *configuration.Configuration) (*Admin, error) {
	if config == nil {
		return nil, fmt.Errorf("nil params in NewAdmin")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.TraceLevel)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "30-12-2006 15:04:05",
	})
	logger.SetOutput(logOut)

	return &Admin{
		config: config,
		logger: logger,
	}, nil
}

// Ожидает команды остановки приложения через http POST
func (a *Admin) shutdownHandler(stop context.CancelFunc, w http.ResponseWriter, r *http.Request) {
	method := "Admin.ShutdownHandler"

	if r.Method != http.MethodPost {
		a.logger.
			WithField("method", method).
			Warn("GET method used")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	reqMap := map[string]string{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqMap)
	if err != nil {
		a.logger.
			WithField("method", method).
			Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, ok := reqMap["AdminToken"]
	if !ok {
		a.logger.
			WithField("method", method).
			Warn("no AdminToken")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if token != a.config.AdminToken {
		a.logger.
			WithField("method", method).
			Warn("bad AdminToken")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	stop()
}

// Старт админки
func (a *Admin) Run(stop context.CancelFunc, ctx context.Context, papaWg *sync.WaitGroup) {
	srvMux := http.NewServeMux()
	srvWg := &sync.WaitGroup{}

	srvMux.HandleFunc(a.config.AdminShutdownPath, func(w http.ResponseWriter, r *http.Request) {
		a.shutdownHandler(stop, w, r)
	})

	srv := &http.Server{
		Addr:         a.config.AdminHost + ":" + a.config.AdminPort,
		Handler:      srvMux,
		ReadTimeout:  time.Duration(a.config.AdminTimeout) * time.Second,
		WriteTimeout: time.Duration(a.config.AdminTimeout) * time.Second,
	}

	srvWg.Add(1)
	go func() {
		defer srvWg.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			a.logger.
				WithField("method", "Admin.Run.ListenAndServe").
				Error(err)
			stop()
		}
	}()

	<-ctx.Done()
	if err := srv.Shutdown(ctx); err != nil {
		a.logger.
			WithField("method", "Admin.Run.Shutdown").
			Error(err)
	}
	srvWg.Wait()

	a.logger.Info("admin is done")
	papaWg.Done()
}
