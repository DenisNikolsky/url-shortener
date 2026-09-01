package handler

import (
	"net/http"

	"github.com/DenisNikolsky/url-shortener/internal/service"

	"github.com/labstack/echo/v4"
)

type URLHandler struct {
	service service.URLService
}

func NewURLHandler(service service.URLService) *URLHandler {
	return &URLHandler{
		service: service,
	}
}

type createURLRequest struct {
	URL string `json:"url"`
}

type createURLResponse struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
}

func (h *URLHandler) Create(c echo.Context) error {
	var req createURLRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	url, err := h.service.Create(
		c.Request().Context(),
		req.URL,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	response := createURLResponse{
		ShortCode:   url.ShortCode,
		OriginalURL: url.OriginalURL,
	}

	return c.JSON(http.StatusCreated, response)
}

func (h *URLHandler) Redirect(c echo.Context) error {
	code := c.Param("code")

	url, err := h.service.GetByCode(
		c.Request().Context(),
		code,
	)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "URL not found",
		})
	}

	return c.Redirect(http.StatusFound, url.OriginalURL)
}

func (h *URLHandler) RegisterRoutes(e *echo.Echo) {
	e.POST("/urls", h.Create)
	e.GET("/:code", h.Redirect)
}
