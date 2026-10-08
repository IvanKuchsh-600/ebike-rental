package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/rental"
	rentalusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/rental"
)

type RentalHandler struct {
	usecase rentalusecase.Usecase
}

func NewRentalHandler(usecase rentalusecase.Usecase) *RentalHandler {
	return &RentalHandler{usecase: usecase}
}

// Create godoc
// @Summary     Создать аренду
// @Tags        rentals
// @Accept      json
// @Produce     json
// @Param       request body CreateRentalRequest true "Данные аренды"
// @Success     201 {object} RentalResponse
// @Failure     400 {object} ErrorResponse
// @Failure     409 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /rentals [post]
func (h *RentalHandler) Create(c *gin.Context) {
	var req CreateRentalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	created, err := h.usecase.Create(c.Request.Context(), toCreateRentalInput(req))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toRentalResponse(created))
}

// GetByID godoc
// @Summary     Получить аренду по ID
// @Tags        rentals
// @Produce     json
// @Param       id path int true "ID аренды"
// @Success     200 {object} RentalResponse
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /rentals/{id} [get]
func (h *RentalHandler) GetByID(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	r, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toRentalResponse(r))
}

// List godoc
// @Summary     Список всех аренд
// @Tags        rentals
// @Produce     json
// @Success     200 {array} RentalResponse
// @Router      /rentals [get]
func (h *RentalHandler) List(c *gin.Context) {
	items, err := h.usecase.List(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toRentalResponseList(items))
}

// ListActive godoc
// @Summary     Список активных аренд
// @Tags        rentals
// @Produce     json
// @Success     200 {array} RentalResponse
// @Router      /rentals/active [get]
func (h *RentalHandler) ListActive(c *gin.Context) {
	items, err := h.usecase.ListActive(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toRentalResponseList(items))
}

// ListByBike godoc
// @Summary     Список аренд конкретного велосипеда
// @Tags        rentals
// @Produce     json
// @Param       id path int true "ID велосипеда"
// @Success     200 {array} RentalResponse
// @Router      /bikes/{id}/rentals [get]
func (h *RentalHandler) ListByBike(c *gin.Context) {
	bikeID, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid bike id"})
		return
	}

	items, err := h.usecase.ListByBike(c.Request.Context(), bikeID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toRentalResponseList(items))
}

// Update godoc
// @Summary     Обновить аренду
// @Tags        rentals
// @Accept      json
// @Produce     json
// @Param       id path int true "ID аренды"
// @Param       request body UpdateRentalRequest true "Данные для обновления"
// @Success     200 {object} RentalResponse
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /rentals/{id} [put]
func (h *RentalHandler) Update(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	var req UpdateRentalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	updated, err := h.usecase.Update(c.Request.Context(), id, toUpdateRentalInput(req))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toRentalResponse(updated))
}

// Return godoc
// @Summary     Вернуть велосипед (закрыть аренду)
// @Tags        rentals
// @Produce     json
// @Param       id path int true "ID аренды"
// @Success     200 {object} RentalResponse
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     409 {object} ErrorResponse
// @Router      /rentals/{id}/return [post]
func (h *RentalHandler) Return(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	updated, err := h.usecase.Return(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toRentalResponse(updated))
}

// MarkPaid godoc
// @Summary     Отметить аренду как оплаченную
// @Tags        rentals
// @Produce     json
// @Param       id path int true "ID аренды"
// @Success     200 {object} RentalResponse
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /rentals/{id}/pay [post]
func (h *RentalHandler) MarkPaid(c *gin.Context) {
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

	c.JSON(http.StatusOK, toRentalResponse(updated))
}

// Delete godoc
// @Summary     Удалить аренду
// @Tags        rentals
// @Param       id path int true "ID аренды"
// @Success     204
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /rentals/{id} [delete]
func (h *RentalHandler) Delete(c *gin.Context) {
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

// handleError — маппинг доменных ошибок в HTTP-статусы.
func (h *RentalHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, rental.ErrNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "rental not found"})

	case errors.Is(err, rental.ErrBikeNotFound):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bike not found"})

	case errors.Is(err, rental.ErrBikeNotAvailable):
		c.JSON(http.StatusConflict, ErrorResponse{Error: "bike is not available"})

	case errors.Is(err, rental.ErrAlreadyReturned):
		c.JSON(http.StatusConflict, ErrorResponse{Error: "rental already returned"})

	case errors.Is(err, rental.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})

	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}
