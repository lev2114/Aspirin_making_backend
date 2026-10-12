package handler

import (
	"errors"
	"net/http"
	"strings"

	"aspirin_back/internal/app/auth"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const aspirinProductionClaimsKey = "aspirin_production_claims"

// Guests have ID 0; authenticated users are identified by verified JWT claims.
func GetCurrentAspirinProductionUserID(ctx *gin.Context) int {
	claims := currentAspirinProductionClaims(ctx)
	if claims == nil {
		return 0
	}
	return claims.ProductionUserID
}

func currentAspirinProductionClaims(ctx *gin.Context) *auth.AspirinProductionClaims {
	value, exists := ctx.Get(aspirinProductionClaimsKey)
	if !exists {
		return nil
	}
	claims, _ := value.(*auth.AspirinProductionClaims)
	return claims
}

// Public reads remain available even with a missing, expired, or revoked token.
func (h *Handler) OptionalAspirinProductionAuthentication(ctx *gin.Context) {
	h.authenticateAspirinProductionRequest(ctx, false)
}
func (h *Handler) RequireAspirinProductionAuthentication(ctx *gin.Context) {
	h.authenticateAspirinProductionRequest(ctx, true)
}
func (h *Handler) authenticateAspirinProductionRequest(ctx *gin.Context, required bool) {
	parts := strings.Fields(ctx.GetHeader("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		if required {
			ctx.Header("WWW-Authenticate", "Bearer")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Требуется аутентификация"})
		}
		return
	}
	claims, err := h.Authentication.Validate(ctx.Request.Context(), parts[1])
	if err != nil {
		if !errors.Is(err, auth.ErrInvalidAspirinProductionToken) {
			logrus.Error("Ошибка проверки JWT: ", err)
			if required {
				ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": "Сервис аутентификации недоступен"})
			}
		} else if required {
			ctx.Header("WWW-Authenticate", "Bearer")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Недействительный или истёкший токен"})
		}
		return
	}
	ctx.Set(aspirinProductionClaimsKey, claims)
}
