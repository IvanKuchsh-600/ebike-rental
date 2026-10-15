package web

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler — HTTP-хендлеры для веб-интерфейса.
type Handler struct {
	renderer *Renderer
	logger   *slog.Logger
}

func NewHandler(renderer *Renderer, logger *slog.Logger) *Handler {
	return &Handler{
		renderer: renderer,
		logger:   logger,
	}
}

// LoginPage godoc
// @Summary     Страница входа
// @Tags        web
// @Produce     html
// @Success     200 {string} string "HTML"
// @Router      /web/login [get]
func (h *Handler) LoginPage(c *gin.Context) {
	data := map[string]any{
		"Title": "Вход",
	}

	if err := h.renderer.RenderPage(c.Writer, "layout.html", "login.html", data); err != nil {
		h.logger.Error("failed to render login page", "error", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

func (h *Handler) DashboardPage(c *gin.Context) {
	data := map[string]any{
		"Title": "Дашборд",
	}

	if err := h.renderer.RenderPage(c.Writer, "layout.html", "dashboard.html", data); err != nil {
		h.logger.Error("failed to render dashboard page", "error", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

func (h *Handler) PartnerDashboardPage(c *gin.Context) {
	data := map[string]any{
		"Title": "Кабинет партнёра",
	}

	if err := h.renderer.RenderPage(c.Writer, "layout.html", "partner_dashboard.html", data); err != nil {
		h.logger.Error("failed to render partner dashboard", "error", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

func (h *Handler) WIPPage(c *gin.Context) {
	data := map[string]any{
		"Title": "В разработке",
	}

	if err := h.renderer.RenderPage(c.Writer, "layout.html", "wip.html", data); err != nil {
		h.logger.Error("failed to render wip page", "error", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

func (h *Handler) BikesPage(c *gin.Context) {
	data := map[string]any{
		"Title": "Велосипеды",
	}

	if err := h.renderer.RenderPage(c.Writer, "layout.html", "bikes.html", data); err != nil {
		h.logger.Error("failed to render bikes page", "error", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

func (h *Handler) RentalsPage(c *gin.Context) {
	data := map[string]any{
		"Title": "Аренды",
	}

	if err := h.renderer.RenderPage(c.Writer, "layout.html", "rentals.html", data); err != nil {
		h.logger.Error("failed to render rentals page", "error", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

func (h *Handler) BreakdownsPage(c *gin.Context) {
	data := map[string]any{
		"Title": "Поломки",
	}

	if err := h.renderer.RenderPage(c.Writer, "layout.html", "breakdowns.html", data); err != nil {
		h.logger.Error("failed to render breakdowns page", "error", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

func (h *Handler) PayoutsPage(c *gin.Context) {
	data := map[string]any{
		"Title": "Выплаты",
	}

	if err := h.renderer.RenderPage(c.Writer, "layout.html", "payouts.html", data); err != nil {
		h.logger.Error("failed to render payouts page", "error", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

func (h *Handler) PartnerBikesPage(c *gin.Context) {
	data := map[string]any{
		"Title": "Велосипеды (партнёр)",
	}

	if err := h.renderer.RenderPage(c.Writer, "layout.html", "partner_bikes.html", data); err != nil {
		h.logger.Error("failed to render partner bikes page", "error", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

func (h *Handler) PartnerRentalsPage(c *gin.Context) {
	data := map[string]any{"Title": "Аренды (партнёр)"}
	if err := h.renderer.RenderPage(c.Writer, "layout.html", "partner_rentals.html", data); err != nil {
		h.logger.Error("failed to render partner rentals", "error", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

func (h *Handler) PartnerBreakdownsPage(c *gin.Context) {
	data := map[string]any{"Title": "Поломки (партнёр)"}
	if err := h.renderer.RenderPage(c.Writer, "layout.html", "partner_breakdowns.html", data); err != nil {
		h.logger.Error("failed to render partner breakdowns", "error", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

func (h *Handler) PartnerPayoutsPage(c *gin.Context) {
	data := map[string]any{"Title": "Выплаты (партнёр)"}
	if err := h.renderer.RenderPage(c.Writer, "layout.html", "partner_payouts.html", data); err != nil {
		h.logger.Error("failed to render partner payouts", "error", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}
