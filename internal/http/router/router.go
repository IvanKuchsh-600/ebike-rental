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

		// Вложенный ресурс: поломки конкретного велосипеда
		// ← ЭТОЙ СТРОКИ НЕ ХВАТАЕТ
		bikes.GET("/:id/breakdowns", h.Breakdown.ListByBike)
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

	return r
}
