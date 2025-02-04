package middleware

import (
	"context"
	"errors"
	"go-server/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	jwt.RegisteredClaims

	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func JWTParser(store *utils.PublicKeyStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing Authorization header",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Malformed Authorization header",
			})
		}

		claims, err := ParseJWT(tokenString, store)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid JWT: " + err.Error(),
			})
		}

		c.Locals("user", claims)
		return c.Next()
	}
}

func ParseJWT(tokenString string, store *utils.PublicKeyStore) (*CustomClaims, error) {
	// 1) 토큰 헤더에서 kid 추출
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &CustomClaims{})
	if err != nil {
		return nil, err
	}

	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("kid not found in token header")
	}

	// 2) Store에서 public key 가져오기
	pubKey, err := store.GetKey(context.Background(), kid)
	if err != nil {
		return nil, err
	}

	// 3) 실제 검증
	parsedToken, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return pubKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := parsedToken.Claims.(*CustomClaims); ok && parsedToken.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
