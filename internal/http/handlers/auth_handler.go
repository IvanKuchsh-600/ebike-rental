package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/user"
	"github.com/IvanKuchsh-600/ebike-rental/internal/http/middleware"
	authusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/auth"
)

type AuthHandler struct {
	usecase authusecase.Usecase
}

func NewAuthHandler(usecase authusecase.Usecase) *AuthHandler {
	return &AuthHandler{usecase: usecase}
}

// Login godoc
// @Summary     Вход в систему
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body LoginRequest true "Email и пароль"
// @Success     200 {object} LoginResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Router      /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	result, err := h.usecase.Login(c.Request.Context(), toLoginInput(req))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: result.Token,
		User:  toUserResponse(result.User),
	})
}

// Me godoc
// @Summary     Информация о текущем пользователе
// @Tags        auth
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} UserResponse
// @Failure     401 {object} ErrorResponse
// @Router      /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	u, err := h.usecase.Me(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toUserResponse(u))
}

// ChangePassword godoc
// @Summary     Смена пароля
// @Tags        auth
// @Accept      json
// @Security    BearerAuth
// @Param       request body ChangePasswordRequest true "Старый и новый пароль"
// @Success     204
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Router      /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	err := h.usecase.ChangePassword(c.Request.Context(), userID, toChangePasswordInput(req))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, user.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid email or password"})

	case errors.Is(err, user.ErrNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})

	case errors.Is(err, user.ErrEmailAlreadyExists):
		c.JSON(http.StatusConflict, ErrorResponse{Error: "email already exists"})

	case errors.Is(err, user.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})

	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}
