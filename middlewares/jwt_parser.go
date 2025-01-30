package middleware

import (
	"go-server/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func JWTParser() fiber.Handler {
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

		claims, err := utils.ParseJWT(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid JWT: " + err.Error(),
			})
		}

		c.Locals("user", claims)
		return c.Next()
	}
}
