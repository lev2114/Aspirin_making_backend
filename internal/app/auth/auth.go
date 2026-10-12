package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const AspirinProductionIssuer = "aspirin-production"

var ErrInvalidAspirinProductionToken = errors.New("недействительный токен пользователя")

type AspirinProductionClaims struct {
	ProductionUserID int `json:"production_user_id"`
	jwt.RegisteredClaims
}

type AspirinProductionTokenBlacklist interface {
	IsRevoked(context.Context, string) (bool, error)
	Revoke(context.Context, string, time.Duration) error
}

type RedisAspirinProductionTokenBlacklist struct{ Client *redis.Client }

func (b RedisAspirinProductionTokenBlacklist) IsRevoked(ctx context.Context, id string) (bool, error) {
	count, err := b.Client.Exists(ctx, "aspirin.production.jwt.blacklist."+id).Result()
	return count > 0, err
}
func (b RedisAspirinProductionTokenBlacklist) Revoke(ctx context.Context, id string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	return b.Client.Set(ctx, "aspirin.production.jwt.blacklist."+id, "revoked", ttl).Err()
}

type AspirinProductionAuthentication struct {
	secret    []byte
	lifetime  time.Duration
	blacklist AspirinProductionTokenBlacklist
}

func New(secret string, lifetime time.Duration, blacklist AspirinProductionTokenBlacklist) (*AspirinProductionAuthentication, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT_SECRET должен содержать не менее 32 байт")
	}
	if lifetime < time.Second {
		return nil, errors.New("JWT_TTL должен быть не менее 1s")
	}
	if blacklist == nil {
		return nil, errors.New("не настроен blacklist JWT")
	}
	return &AspirinProductionAuthentication{[]byte(secret), lifetime, blacklist}, nil
}

func (a *AspirinProductionAuthentication) Issue(productionUserID int) (string, int64, error) {
	if productionUserID < 1 {
		return "", 0, ErrInvalidAspirinProductionToken
	}
	now := time.Now()
	claims := AspirinProductionClaims{
		ProductionUserID: productionUserID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: AspirinProductionIssuer, Subject: strconv.Itoa(productionUserID),
			ID: uuid.NewString(), IssuedAt: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(a.lifetime)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(a.secret)
	return token, int64(a.lifetime / time.Second), err
}

func (a *AspirinProductionAuthentication) Validate(ctx context.Context, raw string) (*AspirinProductionClaims, error) {
	claims := &AspirinProductionClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) { return a.secret, nil },
		jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(AspirinProductionIssuer), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || !token.Valid || claims.ProductionUserID < 1 || claims.Subject != strconv.Itoa(claims.ProductionUserID) || strings.TrimSpace(claims.ID) == "" {
		return nil, ErrInvalidAspirinProductionToken
	}
	revoked, err := a.blacklist.IsRevoked(ctx, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("проверка blacklist: %w", err)
	}
	if revoked {
		return nil, ErrInvalidAspirinProductionToken
	}
	return claims, nil
}

func (a *AspirinProductionAuthentication) Logout(ctx context.Context, claims *AspirinProductionClaims) error {
	return a.blacklist.Revoke(ctx, claims.ID, time.Until(claims.ExpiresAt.Time))
}
