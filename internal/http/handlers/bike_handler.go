package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/bike"
	bikeusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/bike"
)

type BikeHandler struct {
	usecase bikeusecase.Usecase
}

func NewBikeHandler(usecase bikeusecase.Usecase) *BikeHandler {
	return &BikeHandler{usecase: usecase}
}

// Create godoc
// @Summary     Создать велосипед
// @Description Добавляет новый велосипед в парк
// @Tags        bikes
// @Accept      json
// @Produce     json
// @Param       request body CreateBikeRequest true "Данные велосипеда"
// @Success     201 {object} BikeResponse
// @Failure     400 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /bikes [post]
func (h *BikeHandler) Create(c *gin.Context) {
	var req CreateBikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	created, err := h.usecase.Create(c.Request.Context(), toCreateBikeInput(req))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toBikeResponse(created))
}

// GetByID godoc
// @Summary     Получить велосипед по ID
// @Tags        bikes
// @Produce     json
// @Param       id path int true "ID велосипеда"
// @Success     200 {object} BikeResponse
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /bikes/{id} [get]
func (h *BikeHandler) GetByID(c *gin.Context) {
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

	c.JSON(http.StatusOK, toBikeResponse(b))
}

// List godoc
// @Summary     Список велосипедов
// @Tags        bikes
// @Produce     json
// @Success     200 {array} BikeResponse
// @Failure     500 {object} ErrorResponse
// @Router      /bikes [get]
func (h *BikeHandler) List(c *gin.Context) {
	bikes, err := h.usecase.List(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toBikeResponseList(bikes))
}

// Update godoc
// @Summary     Обновить велосипед
// @Tags        bikes
// @Accept      json
// @Produce     json
// @Param       id path int true "ID велосипеда"
// @Param       request body UpdateBikeRequest true "Данные для обновления"
// @Success     200 {object} BikeResponse
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /bikes/{id} [put]
func (h *BikeHandler) Update(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	var req UpdateBikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	updated, err := h.usecase.Update(c.Request.Context(), id, toUpdateBikeInput(req))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toBikeResponse(updated))
}

// Delete godoc
// @Summary     Удалить велосипед
// @Tags        bikes
// @Param       id path int true "ID велосипеда"
// @Success     204
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /bikes/{id} [delete]
func (h *BikeHandler) Delete(c *gin.Context) {
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

func (h *BikeHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, bike.ErrNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "bike not found"})
	case errors.Is(err, bike.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}

func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
