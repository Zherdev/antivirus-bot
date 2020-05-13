// Антивирус-бот
// Админка для управления приложением
// Жердев Иван.

package admin

import (
	"antivirus-bot/pkg/configuration"
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/neko-neko/echo-logrus/v2"
	"github.com/neko-neko/echo-logrus/v2/log"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"sync"
	"time"
)

// Так приходят запросы в json'е
type requestBody struct {
	AdminToken string
}

type Admin struct {
	e      *echo.Echo
	config *configuration.Configuration
	logger *logrus.Logger
	stop   context.CancelFunc
}

func NewAdmin(logOut io.Writer, config *configuration.Configuration) (*Admin, error) {
	if config == nil {
		return nil, fmt.Errorf("nil params in NewAdmin")
	}

	logger := log.Logger()
	logger.Logger.SetLevel(logrus.TraceLevel)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logger.SetOutput(logOut)

	e := echo.New()
	e.Server.ReadTimeout = time.Duration(config.AdminTimeout) * time.Second
	e.Server.WriteTimeout = time.Duration(config.AdminTimeout) * time.Second
	e.Logger = logger
	e.Use(middleware.Logger())

	return &Admin{
		e:      e,
		config: config,
		logger: logger.Logger,
	}, nil
}

// Авторизация. Проверяем админский токен
func (a *Admin) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		method := "authMiddleware"

		data := &requestBody{}
		err := ctx.Bind(data)
		if err != nil {
			a.logger.
				WithField("method", method).
				Warn("can't unmarshal json")
			return ctx.JSON(http.StatusBadRequest, "bad request body")
		}

		if data.AdminToken != a.config.AdminToken {
			a.logger.
				WithField("method", method).
				Warn("bad AdminToken")
			return ctx.JSON(http.StatusUnauthorized, "bad auth data")
		}

		ctx.Set("requestBody", data)
		return next(ctx)
	}
}

// Ожидает команды остановки приложения через http POST
func (a *Admin) shutdownHandler(ctx echo.Context) error {
	method := "shutdownHandler"
	a.logger.
		WithField("method", method).
		Info("shut down")

	a.stop()

	return ctx.JSON(http.StatusOK, "shut down")
}

// Старт админки
func (a *Admin) Run(stop context.CancelFunc, ctx context.Context, papaWg *sync.WaitGroup) {
	a.stop = stop

	fileServer := http.FileServer(http.Dir(a.config.AdminLogsPath)) // для получения логов
	echoFileHandler := echo.WrapHandler(http.StripPrefix(a.config.AdminGetLogsPath, fileServer))

	a.e.POST(a.config.AdminShutdownPath, a.shutdownHandler, a.authMiddleware)
	a.e.POST(a.config.AdminGetLogsPath, echoFileHandler, a.authMiddleware)

	go func() {
		if err := a.e.Start(a.config.AdminHost + ":" + a.config.AdminPort); err != nil {
			a.e.Logger.Info("shutting down the server")
		}
	}()

	<-ctx.Done()
	ctxForShutDown, shutDownEcho := context.WithCancel(context.Background())
	defer shutDownEcho()
	if err := a.e.Shutdown(ctxForShutDown); err != nil {
		a.logger.
			WithField("method", "Run").
			Error(err)
	}

	a.logger.Info("admin is done")
	papaWg.Done()
}
