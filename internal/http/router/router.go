package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/IvanKuchsh-600/ebike-rental/internal/http/handlers"
	"github.com/IvanKuchsh-600/ebike-rental/internal/http/middleware"
	webpkg "github.com/IvanKuchsh-600/ebike-rental/internal/web"
)

type Handlers struct {
	Bike      *handlers.BikeHandler
	Breakdown *handlers.BreakdownHandler
	Rental    *handlers.RentalHandler
	Payout    *handlers.PayoutHandler
	Auth      *handlers.AuthHandler
	Partner   *handlers.PartnerHandler
	Web       *webpkg.Handler // ← новое
}

type Config struct {
	JWTSecret string
}

func New(db *pgxpool.Pool, h *Handlers, cfg Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	// Static files
	r.Static("/static", "./web/static")
	r.Static("/uploads", "./uploads")

	// Web pages (HTML)
	web := r.Group("/web")
	{
		web.GET("/login", h.Web.LoginPage)
		web.GET("/dashboard", h.Web.DashboardPage)
		web.GET("/partner/dashboard", h.Web.PartnerDashboardPage)

		// Временные заглушки (заменим на настоящие страницы)
		web.GET("/bikes", h.Web.BikesPage)
		web.GET("/rentals", h.Web.RentalsPage)
		web.GET("/breakdowns", h.Web.BreakdownsPage)
		web.GET("/payouts", h.Web.PayoutsPage)

		web.GET("/partner/bikes", h.Web.PartnerBikesPage)
		web.GET("/partner/rentals", h.Web.PartnerRentalsPage)
		web.GET("/partner/breakdowns", h.Web.PartnerBreakdownsPage)
		web.GET("/partner/payouts", h.Web.PartnerPayoutsPage)
	}

	// ============ Public routes ============

	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db unavailable"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	auth := r.Group("/auth")
	{
		auth.POST("/login", h.Auth.Login)
	}

	// Всё ниже требует валидного JWT.
	protected := r.Group("")
	protected.Use(middleware.RequireAuth(cfg.JWTSecret))
	{
		// auth/me, auth/change-password — требуют логина
		protected.GET("/auth/me", h.Auth.Me)
		protected.POST("/auth/change-password", h.Auth.ChangePassword)

		// ============ Owner-only routes ============

		// Bikes
		bikes := protected.Group("/bikes")
		bikes.Use(middleware.RequireRole("owner"))
		{
			bikes.POST("", h.Bike.Create)
			bikes.GET("", h.Bike.List)
			bikes.GET("/:id", h.Bike.GetByID)
			bikes.PUT("/:id", h.Bike.Update)
			bikes.DELETE("/:id", h.Bike.Delete)

			bikes.GET("/:id/breakdowns", h.Breakdown.ListByBike)
			bikes.GET("/:id/rentals", h.Rental.ListByBike)

			bikes.POST("/:id/photo", h.Bike.UploadPhoto)
			bikes.DELETE("/:id/photo", h.Bike.DeletePhoto)
		}

		// Breakdowns
		breakdowns := protected.Group("/breakdowns")
		breakdowns.Use(middleware.RequireRole("owner"))
		{
			breakdowns.POST("", h.Breakdown.Create)
			breakdowns.GET("", h.Breakdown.List)
			breakdowns.GET("/unfixed", h.Breakdown.ListUnfixed)
			breakdowns.GET("/:id", h.Breakdown.GetByID)
			breakdowns.PUT("/:id", h.Breakdown.Update)
			breakdowns.POST("/:id/fix", h.Breakdown.MarkFixed)
			breakdowns.DELETE("/:id", h.Breakdown.Delete)
		}

		// Rentals
		rentals := protected.Group("/rentals")
		rentals.Use(middleware.RequireRole("owner"))
		{
			rentals.POST("", h.Rental.Create)
			rentals.GET("", h.Rental.List)
			rentals.GET("/active", h.Rental.ListActive)
			rentals.GET("/:id", h.Rental.GetByID)
			rentals.PUT("/:id", h.Rental.Update)
			rentals.POST("/:id/return", h.Rental.Return)
			rentals.POST("/:id/pay", h.Rental.MarkPaid)
			rentals.DELETE("/:id", h.Rental.Delete)
		}

		// Payouts
		payouts := protected.Group("/payouts")
		payouts.Use(middleware.RequireRole("owner"))
		{
			payouts.GET("/preview", h.Payout.Preview)
			payouts.POST("", h.Payout.Create)
			payouts.GET("", h.Payout.List)
			payouts.GET("/:id", h.Payout.GetByID)
			payouts.POST("/:id/pay", h.Payout.MarkPaid)
			payouts.DELETE("/:id", h.Payout.Delete)
		}
	}

	// Partner-only routes
	partner := protected.Group("/partner")
	partner.Use(middleware.RequireRole("partner"))
	{
		partner.GET("/dashboard", h.Partner.Dashboard)

		partner.GET("/bikes", h.Partner.ListBikes)
		partner.GET("/bikes/:id", h.Partner.GetBike)
		partner.GET("/bikes/:id/breakdowns", h.Partner.ListBikeBreakdowns)

		partner.GET("/rentals", h.Partner.ListRentals)
		partner.GET("/breakdowns", h.Partner.ListBreakdowns)

		partner.GET("/payouts", h.Partner.ListPayouts)
		partner.GET("/payouts/preview", h.Partner.PreviewPayout)
	}

	return r
}
