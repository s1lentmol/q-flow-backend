package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type UserContext struct {
	ID    int64
	Email string
	AppID int
	Name  string
}

func Auth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing authorization header")
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid authorization header")
		}

		token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Name {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "unexpected signing method")
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid claims")
		}

		uid, _ := claims["uid"].(float64)
		email, _ := claims["email"].(string)
		appID, _ := claims["app_id"].(float64)
		name, _ := claims["name"].(string)

		c.Locals("user", &UserContext{
			ID:    int64(uid),
			Email: email,
			AppID: int(appID),
			Name:  name,
		})

		return c.Next()
	}
}

func GetUser(c *fiber.Ctx) *UserContext {
	val := c.Locals("user")
	if val == nil {
		return nil
	}
	user, _ := val.(*UserContext)
	return user
}
