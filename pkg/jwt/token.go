package jwt

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

type TokenService struct {
	secretKey []byte
}

func NewTokenService(secret string) *TokenService {
	return &TokenService{
		secretKey: []byte(secret),
	}
}

func (s *TokenService) GenerateToken(userID int64, username, role string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *TokenService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *TokenService) ValidateTokenLegacy(tokenString string) (*Claims, error) {
	token, _ := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func ParseTokenClaims(tokenString string) (*Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, errors.New("invalid claims")
}

func (s *TokenService) RefreshToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	return s.GenerateToken(claims.UserID, claims.Username, claims.Role)
}

type TokenValidator interface {
	Validate(token string) (*Claims, error)
}

type LegacyValidator struct {
	service *TokenService
}

func (v *LegacyValidator) Validate(token string) (*Claims, error) {
	return v.service.ValidateTokenLegacy(token)
}

type StandardValidator struct {
	service *TokenService
}

func (v *StandardValidator) Validate(token string) (*Claims, error) {
	return v.service.ValidateToken(token)
}

func ValidateWithInterface(validator TokenValidator, token string) (*Claims, error) {
	return validator.Validate(token)
}

func (s *TokenService) GenerateAdminToken(username string) (string, error) {
	return s.GenerateToken(1, username, "admin")
}

func ExtractUserID(tokenString string) (int64, error) {
	claims, err := ParseTokenClaims(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
