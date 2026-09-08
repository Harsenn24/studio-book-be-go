package auth

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"studio-book-be-go/internal/helper"
)

type AuthHandler struct {
	authService AuthService
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var input LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		helper.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	roleFromPath := c.Param("role")

	input.IPAddress = c.GetHeader("x-ip-address")
	input.DeviceID = c.GetHeader("x-device-id")

	response, err := h.authService.Login(c.Request.Context(), input, roleFromPath)
	if err != nil {
		helper.SendError(c, http.StatusUnauthorized, err.Error())
		return
	}

	helper.SendSuccess(c, http.StatusOK, "StatusOk", response)
}

func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var input RegisterRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		helper.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	roleFromPath := c.Param("role")

	response, err := h.authService.Register(c.Request.Context(), input, roleFromPath)
	if err != nil {
		helper.SendError(c, http.StatusBadRequest, err.Error())
		return
	}
	helper.SendSuccess(c, http.StatusOK, "StatusOk", response)
}

func (h *AuthHandler) CheckUserHandler(c *gin.Context) {
	var input CheckUserRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		helper.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.authService.CheckUser(c.Request.Context(), input)
	if err != nil {
		helper.SendError(c, http.StatusBadRequest, err.Error())
		return
	}
	helper.SendSuccess(c, http.StatusOK, "StatusOk", response)
}
