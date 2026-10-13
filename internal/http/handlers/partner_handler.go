package handlers

import (
	"errors"
	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/bike"
	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/breakdown"
	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/payout"
	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/rental"
	"net/http"

	"github.com/gin-gonic/gin"

	bikeusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/bike"
	breakdownusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/breakdown"
	partnerusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/partner"
	payoutusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/payout"
	rentalusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/rental"
)

type PartnerHandler struct {
	partnerSvc   partnerusecase.Usecase
	bikeSvc      bikeusecase.Usecase
	rentalSvc    rentalusecase.Usecase
	breakdownSvc breakdownusecase.Usecase
	payoutSvc    payoutusecase.Usecase
}

func NewPartnerHandler(
	partnerSvc partnerusecase.Usecase,
	bikeSvc bikeusecase.Usecase,
	rentalSvc rentalusecase.Usecase,
	breakdownSvc breakdownusecase.Usecase,
	payoutSvc payoutusecase.Usecase,
) *PartnerHandler {
	return &PartnerHandler{
		partnerSvc:   partnerSvc,
		bikeSvc:      bikeSvc,
		rentalSvc:    rentalSvc,
		breakdownSvc: breakdownSvc,
		payoutSvc:    payoutSvc,
	}
}

// Dashboard godoc
// @Summary     Сводка для партнёра
// @Tags        partner
// @Produce     json
// @Security    BearerAuth
// @Param       month query string false "Месяц YYYY-MM (по умолчанию — текущий)"
// @Success     200 {object} PartnerDashboardResponse
// @Failure     401 {object} ErrorResponse
// @Failure     403 {object} ErrorResponse
// @Router      /partner/dashboard [get]
func (h *PartnerHandler) Dashboard(c *gin.Context) {
	month := c.Query("month")

	result, err := h.partnerSvc.Dashboard(c.Request.Context(), month)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toPartnerDashboardResponse(result))
}

// ListBikes godoc
// @Summary     Список велосипедов (для партнёра)
// @Tags        partner
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} PartnerBikeResponse
// @Router      /partner/bikes [get]
func (h *PartnerHandler) ListBikes(c *gin.Context) {
	items, err := h.bikeSvc.List(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toPartnerBikeResponseList(items))
}

// GetBike godoc
// @Summary     Один велосипед (для партнёра)
// @Tags        partner
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "ID велосипеда"
// @Success     200 {object} PartnerBikeResponse
// @Failure     404 {object} ErrorResponse
// @Router      /partner/bikes/{id} [get]
func (h *PartnerHandler) GetBike(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	b, err := h.bikeSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toPartnerBikeResponse(b))
}

// ListBikeBreakdowns godoc
// @Summary     Поломки конкретного велосипеда (для партнёра)
// @Tags        partner
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "ID велосипеда"
// @Success     200 {array} PartnerBreakdownResponse
// @Router      /partner/bikes/{id}/breakdowns [get]
func (h *PartnerHandler) ListBikeBreakdowns(c *gin.Context) {
	bikeID, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid bike id"})
		return
	}

	items, err := h.breakdownSvc.ListByBike(c.Request.Context(), bikeID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toPartnerBreakdownResponseList(items))
}

// ListRentals godoc
// @Summary     Список аренд (для партнёра)
// @Tags        partner
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} PartnerRentalResponse
// @Router      /partner/rentals [get]
func (h *PartnerHandler) ListRentals(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. Загружаем все аренды
	rentals, err := h.rentalSvc.List(ctx)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 2. Загружаем все велосипеды (один запрос вместо N+1)
	bikes, err := h.bikeSvc.List(ctx)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 3. Строим map: bike_id → serial_number
	serialByBikeID := make(map[int64]string, len(bikes))
	for i := range bikes {
		serialByBikeID[bikes[i].ID] = bikes[i].SerialNumber
	}

	// 4. Собираем ответ
	result := make([]PartnerRentalResponse, 0, len(rentals))
	for i := range rentals {
		serial := serialByBikeID[rentals[i].BikeID]
		result = append(result, toPartnerRentalResponse(&rentals[i], serial))
	}

	c.JSON(http.StatusOK, result)
}

// ListBreakdowns godoc
// @Summary     Список всех поломок (для партнёра)
// @Tags        partner
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} PartnerBreakdownResponse
// @Router      /partner/breakdowns [get]
func (h *PartnerHandler) ListBreakdowns(c *gin.Context) {
	items, err := h.breakdownSvc.List(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toPartnerBreakdownResponseList(items))
}

// ListPayouts godoc
// @Summary     Список выплат (для партнёра)
// @Tags        partner
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} PartnerPayoutResponse
// @Router      /partner/payouts [get]
func (h *PartnerHandler) ListPayouts(c *gin.Context) {
	items, err := h.payoutSvc.List(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toPartnerPayoutResponseList(items))
}

// PreviewPayout godoc
// @Summary     Предпросмотр выплаты за месяц
// @Tags        partner
// @Produce     json
// @Security    BearerAuth
// @Param       month query string true "Месяц YYYY-MM"
// @Success     200 {object} PartnerPayoutPreviewResponse
// @Router      /partner/payouts/preview [get]
func (h *PartnerHandler) PreviewPayout(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "month is required"})
		return
	}

	result, err := h.payoutSvc.Preview(c.Request.Context(), month)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toPartnerPayoutPreviewResponse(result))
}

// handleError — маппинг ошибок.
// Для партнёра — общий, без деталей.
func (h *PartnerHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, bike.ErrNotFound),
		errors.Is(err, breakdown.ErrNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})

	case errors.Is(err, bike.ErrInvalidInput),
		errors.Is(err, rental.ErrInvalidInput),
		errors.Is(err, breakdown.ErrInvalidInput),
		errors.Is(err, payout.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})

	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}
