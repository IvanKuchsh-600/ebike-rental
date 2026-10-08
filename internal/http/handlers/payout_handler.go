package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/payout"
	payoutusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/payout"
)

type PayoutHandler struct {
	usecase payoutusecase.Usecase
}

func NewPayoutHandler(usecase payoutusecase.Usecase) *PayoutHandler {
	return &PayoutHandler{usecase: usecase}
}

// Preview godoc
// @Summary     Предпросмотр выплаты за месяц (без сохранения)
// @Tags        payouts
// @Produce     json
// @Param       month query string true "Месяц в формате YYYY-MM" example("2026-10")
// @Success     200 {object} PayoutPreviewResponse
// @Failure     400 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /payouts/preview [get]
func (h *PayoutHandler) Preview(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "month is required"})
		return
	}

	result, err := h.usecase.Preview(c.Request.Context(), month)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toPayoutPreviewResponse(result))
}

// Create godoc
// @Summary     Зафиксировать выплату партнёру
// @Tags        payouts
// @Accept      json
// @Produce     json
// @Param       request body CreatePayoutRequest true "Данные выплаты"
// @Success     201 {object} PayoutResponse
// @Failure     400 {object} ErrorResponse
// @Failure     409 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /payouts [post]
func (h *PayoutHandler) Create(c *gin.Context) {
	var req CreatePayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	created, err := h.usecase.Create(c.Request.Context(), toCreatePayoutInput(req))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toPayoutResponse(created))
}

// List godoc
// @Summary     Список всех выплат
// @Tags        payouts
// @Produce     json
// @Success     200 {array} PayoutResponse
// @Router      /payouts [get]
func (h *PayoutHandler) List(c *gin.Context) {
	items, err := h.usecase.List(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toPayoutResponseList(items))
}

// GetByID godoc
// @Summary     Получить выплату по ID
// @Tags        payouts
// @Produce     json
// @Param       id path int true "ID выплаты"
// @Success     200 {object} PayoutResponse
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /payouts/{id} [get]
func (h *PayoutHandler) GetByID(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	p, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toPayoutResponse(p))
}

// MarkPaid godoc
// @Summary     Отметить выплату как выполненную
// @Tags        payouts
// @Produce     json
// @Param       id path int true "ID выплаты"
// @Success     200 {object} PayoutResponse
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /payouts/{id}/pay [post]
func (h *PayoutHandler) MarkPaid(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	updated, err := h.usecase.MarkPaid(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toPayoutResponse(updated))
}

// Delete godoc
// @Summary     Удалить выплату
// @Tags        payouts
// @Param       id path int true "ID выплаты"
// @Success     204
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /payouts/{id} [delete]
func (h *PayoutHandler) Delete(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *PayoutHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, payout.ErrNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "payout not found"})

	case errors.Is(err, payout.ErrAlreadyExists):
		c.JSON(http.StatusConflict, ErrorResponse{Error: "payout for this period already exists"})

	case errors.Is(err, payout.ErrAlreadyPaid):
		c.JSON(http.StatusConflict, ErrorResponse{Error: "payout already paid"})

	case errors.Is(err, payout.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})

	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}
