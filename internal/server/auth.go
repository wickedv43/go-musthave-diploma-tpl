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
	UserID int `json:"login"`
}

func (s *Server) createJWT(u storage.User) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		UserID: u.ID,
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

func (s *Server) getUserIDFromCookie(cookie *http.Cookie) (int, error) {

	token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, errors.Wrapf(err, "get login from cookie %s", cookieName)
}

func (s *Server) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		//get cookie
		cookie, err := c.Cookie(cookieName)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, "unauthorized")
		}

		s.logger.Info("got cookie %s", cookieName)

		//check cookie exp_at date
		if cookie.Expires.After(time.Now()) {
			return c.JSON(http.StatusUnauthorized, "unauthorized")
		}

		//get login? or another param
		userID, err := s.getUserIDFromCookie(cookie)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Getting user from cookie")
		}

		//set userID into echo context
		c.Set("userID", userID)

		return next(c)
	}
}
