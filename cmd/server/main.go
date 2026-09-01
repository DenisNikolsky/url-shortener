package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DenisNikolsky/url-shortener/internal/config"
	"github.com/DenisNikolsky/url-shortener/internal/handler"
	"github.com/DenisNikolsky/url-shortener/internal/logger"
	"github.com/DenisNikolsky/url-shortener/internal/repository"
	"github.com/DenisNikolsky/url-shortener/internal/service"
	"github.com/labstack/echo/v4/middleware"

	_ "github.com/DenisNikolsky/url-shortener/docs"

	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/labstack/echo/v4"
)

func setupServer(cfg config.Config) (*echo.Echo, *sql.DB, error) {
	db, err := repository.NewPostgres(cfg)
	if err != nil {
		return nil, nil, err
	}

	urlRepository := repository.NewPostgresURLRepository(db)
	urlService := service.NewURLService(urlRepository)
	urlHandler := handler.NewURLHandler(urlService)

	e := echo.New()

	// HTTP request logging.
	e.Use(middleware.Logger())

	// Swagger UI.
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// API routes.
	urlHandler.RegisterRoutes(e)

	return e, db, nil
}

// @title URL Shortener API
// @version 1.0
// @description REST API для сервиса сокращения URL.
// @host localhost:8080
// @BasePath /
func main() {
	log := logger.New()

	cfg := config.Load(".env")

	e, db, err := setupServer(cfg)
	if err != nil {
		log.Error(
			"failed to setup server",
			"error",
			err,
		)
		return
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Error(
				"failed to close database",
				"error",
				err,
			)
		}
	}()

	// Запускаем HTTP-сервер в отдельной goroutine.
	go func() {
		if err := e.Start(":" + cfg.ServerPort); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Error(
				"server error",
				"error",
				err,
			)
		}
	}()

	log.Info(
		"server started",
		"port",
		cfg.ServerPort,
	)

	// Ждём сигнал от операционной системы.
	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-stop

	log.Info("shutting down server")

	// Даём текущим HTTP-запросам до 10 секунд на завершение.
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Error(
			"failed to shutdown server",
			"error",
			err,
		)
	}

	log.Info("server stopped")
}
