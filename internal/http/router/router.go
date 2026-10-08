package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/IvanKuchsh-600/ebike-rental/internal/http/handlers"
)

type Handlers struct {
	Bike      *handlers.BikeHandler
	Breakdown *handlers.BreakdownHandler
	Rental    *handlers.RentalHandler
	Payout    *handlers.PayoutHandler
}

func New(db *pgxpool.Pool, h *Handlers) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	// Health
	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db unavailable"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Bikes
	bikes := r.Group("/bikes")
	{
		bikes.POST("", h.Bike.Create)
		bikes.GET("", h.Bike.List)
		bikes.GET("/:id", h.Bike.GetByID)
		bikes.PUT("/:id", h.Bike.Update)
		bikes.DELETE("/:id", h.Bike.Delete)

		// // Вложенные ресурсы
		bikes.GET("/:id/breakdowns", h.Breakdown.ListByBike)
		bikes.GET("/:id/rentals", h.Rental.ListByBike)
	}

	// Breakdowns
	breakdowns := r.Group("/breakdowns")
	{
		breakdowns.POST("", h.Breakdown.Create)
		breakdowns.GET("", h.Breakdown.List)
		breakdowns.GET("/unfixed", h.Breakdown.ListUnfixed) // ← ВАЖНО: до /:id
		breakdowns.GET("/:id", h.Breakdown.GetByID)
		breakdowns.PUT("/:id", h.Breakdown.Update)
		breakdowns.POST("/:id/fix", h.Breakdown.MarkFixed)
		breakdowns.DELETE("/:id", h.Breakdown.Delete)
	}

	// Rentals
	rentals := r.Group("/rentals")
	{
		rentals.POST("", h.Rental.Create)
		rentals.GET("", h.Rental.List)
		rentals.GET("/active", h.Rental.ListActive) // ← ВАЖНО: до /:id
		rentals.GET("/:id", h.Rental.GetByID)
		rentals.PUT("/:id", h.Rental.Update)
		rentals.POST("/:id/return", h.Rental.Return)
		rentals.POST("/:id/pay", h.Rental.MarkPaid)
		rentals.DELETE("/:id", h.Rental.Delete)
	}

	// Payouts
	payouts := r.Group("/payouts")
	{
		payouts.GET("/preview", h.Payout.Preview) // ← ВАЖНО: до /:id
		payouts.POST("", h.Payout.Create)
		payouts.GET("", h.Payout.List)
		payouts.GET("/:id", h.Payout.GetByID)
		payouts.POST("/:id/pay", h.Payout.MarkPaid)
		payouts.DELETE("/:id", h.Payout.Delete)
	}

	return r
}
