package main

import (
	"log"

	"github.com/DenisNikolsky/url-shortener/internal/config"
	"github.com/DenisNikolsky/url-shortener/internal/handler"
	"github.com/DenisNikolsky/url-shortener/internal/repository"
	"github.com/DenisNikolsky/url-shortener/internal/service"
	"github.com/labstack/echo/v4"
)

func main() {
	cfg := config.Load(".env")

	db, err := repository.NewPostgres(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	urlRepository := repository.NewPostgresURLRepository(db)

	urlService := service.NewURLService(urlRepository)

	urlHandler := handler.NewURLHandler(urlService)

	e := echo.New()

	urlHandler.RegisterRoutes(e)

	e.Logger.Fatal(e.Start(":8080"))
}
