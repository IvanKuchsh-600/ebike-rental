package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/breakdown"
	breakdownusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/breakdown"
)

type BreakdownHandler struct {
	usecase breakdownusecase.Usecase
}

func NewBreakdownHandler(usecase breakdownusecase.Usecase) *BreakdownHandler {
	return &BreakdownHandler{usecase: usecase}
}

// Create godoc
// @Summary     Зафиксировать поломку
// @Tags        breakdowns
// @Accept      json
// @Produce     json
// @Param       request body CreateBreakdownRequest true "Данные поломки"
// @Success     201 {object} BreakdownResponse
// @Failure     400 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /breakdowns [post]
func (h *BreakdownHandler) Create(c *gin.Context) {
	var req CreateBreakdownRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	created, err := h.usecase.Create(c.Request.Context(), toCreateBreakdownInput(req))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toBreakdownResponse(created))
}

// GetByID godoc
// @Summary     Получить поломку по ID
// @Tags        breakdowns
// @Produce     json
// @Param       id path int true "ID поломки"
// @Success     200 {object} BreakdownResponse
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /breakdowns/{id} [get]
func (h *BreakdownHandler) GetByID(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	b, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toBreakdownResponse(b))
}

// List godoc
// @Summary     Список всех поломок
// @Tags        breakdowns
// @Produce     json
// @Success     200 {array} BreakdownResponse
// @Router      /breakdowns [get]
func (h *BreakdownHandler) List(c *gin.Context) {
	items, err := h.usecase.List(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toBreakdownResponseList(items))
}

// ListByBike godoc
// @Summary     Список поломок конкретного велосипеда
// @Tags        breakdowns
// @Produce     json
// @Param       bike_id path int true "ID велосипеда"
// @Success     200 {array} BreakdownResponse
// @Router      /bikes/{bike_id}/breakdowns [get]
func (h *BreakdownHandler) ListByBike(c *gin.Context) {
	bikeID, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid bike_id"})
		return
	}

	items, err := h.usecase.ListByBike(c.Request.Context(), bikeID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toBreakdownResponseList(items))
}

// ListUnfixed godoc
// @Summary     Список непочиненных поломок
// @Tags        breakdowns
// @Produce     json
// @Success     200 {array} BreakdownResponse
// @Router      /breakdowns/unfixed [get]
func (h *BreakdownHandler) ListUnfixed(c *gin.Context) {
	items, err := h.usecase.ListUnfixed(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toBreakdownResponseList(items))
}

// Update godoc
// @Summary     Обновить поломку
// @Tags        breakdowns
// @Accept      json
// @Produce     json
// @Param       id path int true "ID поломки"
// @Param       request body UpdateBreakdownRequest true "Данные для обновления"
// @Success     200 {object} BreakdownResponse
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /breakdowns/{id} [put]
func (h *BreakdownHandler) Update(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	var req UpdateBreakdownRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	updated, err := h.usecase.Update(c.Request.Context(), id, toUpdateBreakdownInput(req))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toBreakdownResponse(updated))
}

// MarkFixed godoc
// @Summary     Отметить поломку как починенную
// @Tags        breakdowns
// @Param       id path int true "ID поломки"
// @Success     204
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /breakdowns/{id}/fix [post]
func (h *BreakdownHandler) MarkFixed(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	if err := h.usecase.MarkFixed(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Delete godoc
// @Summary     Удалить поломку
// @Tags        breakdowns
// @Param       id path int true "ID поломки"
// @Success     204
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /breakdowns/{id} [delete]
func (h *BreakdownHandler) Delete(c *gin.Context) {
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

func (h *BreakdownHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, breakdown.ErrNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "breakdown not found"})
	case errors.Is(err, breakdown.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}
