package handler

import (
	"errors"
	"net/http"
	"strings"

	"aspirin_back/internal/app/repository"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type RegisterAspirinProductionUserRequest struct {
	Username string `json:"username" binding:"required,max=100" example:"student"`
	Password string `json:"password" binding:"required" example:"aspirin-password"`
}

type AspirinProductionUserResponse struct {
	ProductionUserID   int    `json:"production_user_id"`
	ProductionUsername string `json:"production_username"`
}

type AspirinProductionAuthenticationResponse struct {
	AccessToken string                        `json:"access_token"`
	TokenType   string                        `json:"token_type" example:"Bearer"`
	ExpiresIn   int64                         `json:"expires_in" example:"3600"`
	Data        AspirinProductionUserResponse `json:"data"`
}

type AspirinProductionMessageResponse struct {
	Message string `json:"message"`
}
type AspirinProductionUserDataResponse struct {
	Data AspirinProductionUserResponse `json:"data"`
}

// RegisterAspirinProductionUserAPI godoc
// @Summary Регистрация пользователя производства аспирина
// @Tags Пользователи производства аспирина
// @Accept json
// @Produce json
// @Param aspirin_production_user body RegisterAspirinProductionUserRequest true "Имя и пароль (1–72 байта)"
// @Success 201 {object} AspirinProductionUserDataResponse
// @Failure 400,409,500 {object} AspirinProductionMessageResponse
// @Router /users [post]
func (h *Handler) RegisterAspirinProductionUserAPI(ctx *gin.Context) {
	var request RegisterAspirinProductionUserRequest
	if !bindAspirinProductionCredentials(ctx, &request) {
		return
	}
	user, err := h.Repository.CreateAspirinProductionUser(request.Username, request.Password)
	if err != nil {
		if errors.Is(err, repository.ErrAspirinProductionUserExists) {
			ctx.JSON(http.StatusConflict, gin.H{"message": err.Error()})
			return
		}
		logrus.Error("Ошибка регистрации пользователя: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Ошибка регистрации пользователя"})
		return
	}
	ctx.JSON(http.StatusCreated, AspirinProductionUserDataResponse{Data: aspirinProductionUserResponse(user)})
}

// AuthenticateAspirinProductionUserAPI godoc
// @Summary Аутентификация пользователя производства аспирина
// @Tags Пользователи производства аспирина
// @Accept json
// @Produce json
// @Param aspirin_production_user body RegisterAspirinProductionUserRequest true "Имя и пароль"
// @Success 200 {object} AspirinProductionAuthenticationResponse
// @Failure 400,401,500 {object} AspirinProductionMessageResponse
// @Router /authentication [post]
func (h *Handler) AuthenticateAspirinProductionUserAPI(ctx *gin.Context) {
	var request RegisterAspirinProductionUserRequest
	if !bindAspirinProductionCredentials(ctx, &request) {
		return
	}
	user, err := h.Repository.AuthenticateAspirinProductionUser(request.Username, request.Password)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidAspirinProductionCredentials) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			return
		}
		logrus.Error("Ошибка аутентификации пользователя: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Ошибка аутентификации пользователя"})
		return
	}
	token, expiresIn, err := h.Authentication.Issue(user.ProductionUserID)
	if err != nil {
		logrus.Error("Ошибка создания JWT: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Ошибка создания токена"})
		return
	}
	ctx.Header("Cache-Control", "no-store")
	ctx.JSON(http.StatusOK, AspirinProductionAuthenticationResponse{
		AccessToken: token, TokenType: "Bearer", ExpiresIn: expiresIn, Data: aspirinProductionUserResponse(user),
	})
}

// DeauthenticateAspirinProductionUserAPI godoc
// @Summary Выход пользователя и отзыв текущего JWT
// @Tags Пользователи производства аспирина
// @Produce json
// @Security AspirinProductionBearer
// @Success 200 {object} AspirinProductionMessageResponse
// @Failure 401,503 {object} AspirinProductionMessageResponse
// @Router /deauthentication [post]
func (h *Handler) DeauthenticateAspirinProductionUserAPI(ctx *gin.Context) {
	claims := currentAspirinProductionClaims(ctx)
	if claims == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Требуется аутентификация"})
		return
	}
	if err := h.Authentication.Logout(ctx.Request.Context(), claims); err != nil {
		logrus.Error("Ошибка отзыва JWT: ", err)
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"message": "Ошибка выхода пользователя"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Пользователь вышел из системы"})
}

func bindAspirinProductionCredentials(ctx *gin.Context, request *RegisterAspirinProductionUserRequest) bool {
	if err := ctx.ShouldBindJSON(request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Некорректные данные пользователя"})
		return false
	}
	request.Username = strings.TrimSpace(request.Username)
	// Preserve the password exactly; bcrypt accepts up to 72 bytes.
	if request.Username == "" || strings.TrimSpace(request.Password) == "" || len(request.Password) > 72 {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Укажите имя и пароль длиной от 1 до 72 байт"})
		return false
	}
	return true
}
func aspirinProductionUserResponse(user *repository.AspirinProductionUser) AspirinProductionUserResponse {
	return AspirinProductionUserResponse{ProductionUserID: user.ProductionUserID, ProductionUsername: user.ProductionUsername}
}
