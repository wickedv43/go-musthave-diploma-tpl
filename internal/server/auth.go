package server

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage"
)

var (
	secretKey  = []byte("supersecretkey")
	cookieName = "auth_token"
)

type Claims struct {
	jwt.RegisteredClaims
	Login string `json:"login"`
}

func (s *Server) createJWT(u storage.User) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		Login: u.Login,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secretKey)
}

// TODO: check it
func (s *Server) authorize(c echo.Context, u storage.User) echo.Context {
	jwtToken, err := s.createJWT(u)
	if err != nil {
		return nil
	}

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
	}

	c.SetCookie(cookie)

	return c
}

func (s *Server) getLoginFromCookie(cookie *http.Cookie) (string, error) {

	token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.Login, nil
	}
	return "", errors.Wrapf(err, "get login from cookie %s", cookieName)
}

func (s *Server) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie(cookieName)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, "unauthorized")
		}

		login, err := s.getLoginFromCookie(cookie)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Getting user from cookie")
		}

		c.Set("login", login)

		return next(c)
	}
}
