package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage/mocks"
)

func TestAuthMiddleware_ValidToken(t *testing.T) {
	e := echo.New()
	srv := setupTestServer(t, func(mockDataKeeper *mocks.MockDataKeeper) {})

	user := storage.User{ID: 42}

	token, err := srv.createJWT(user)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:     "auth_token",
		Value:    token,
		HttpOnly: true,
		Expires:  time.Now().Add(1 * time.Hour),
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	next := func(c echo.Context) error {
		called = true
		uid := c.Get("userID")

		require.Equal(t, 42, uid)
		return c.String(http.StatusOK, "ok")
	}

	err = srv.authMiddleware(next)(c)
	require.NoError(t, err)
	require.True(t, called, "next handler was not called")
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	e := echo.New()
	srv := setupTestServer(t, func(mockDataKeeper *mocks.MockDataKeeper) {})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:     "auth_token",
		Value:    "not.valid.jwt", // явно битый токен
		HttpOnly: true,
		Expires:  time.Now().Add(1 * time.Hour),
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	next := func(c echo.Context) error {
		called = true
		return nil
	}

	err := srv.authMiddleware(next)(c)

	require.NoError(t, err)
	require.False(t, called, "next handler should not be called")
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
