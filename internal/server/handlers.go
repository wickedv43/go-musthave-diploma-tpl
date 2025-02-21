package server

import (
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/entities"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/util"
)

func (s *Server) onRegUser(c echo.Context) error {
	var u storage.User

	if err := c.Bind(&u); err != nil {
		return c.JSON(http.StatusBadRequest, "Bad Request")
	}

	err := s.storage.RegisterUser(c.Request().Context(), u)
	if err != nil {
		if errors.Is(err, entities.ErrConflict) {
			return c.JSON(http.StatusConflict, "login already exists")
		}
		return c.JSON(http.StatusInternalServerError, err)
	}

	c = s.authorize(c, u)

	return c.JSON(http.StatusOK, nil)
}

func (s *Server) onLogin(c echo.Context) error {
	var u storage.User

	if err := c.Bind(&u); err != nil {
		return c.JSON(http.StatusBadRequest, "Bad Request")
	}

	err := s.storage.LoginUser(c.Request().Context(), u)
	if err != nil {
		if errors.Is(err, entities.ErrBadLogin) {
			return c.JSON(http.StatusConflict, "permission denied")
		}

		return c.JSON(http.StatusInternalServerError, err)
	}

	c = s.authorize(c, u)

	return c.JSON(http.StatusOK, nil)
}

func (s *Server) onPostOrders(c echo.Context) error {
	if c.Request().Header.Get("Content-Type") == "application/json" {
		return c.JSON(http.StatusBadRequest, "Bad request")
	}

	body := c.Request().Body

	orderNum, err := io.ReadAll(body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	//create order
	var o storage.Order

	if !util.LuhnCheck(string(orderNum)) {
		return c.JSON(http.StatusUnprocessableEntity, "Unprocessable Entity")
	}

	o.Number = string(orderNum)
	o.UploadedAt = time.Now().Format(time.RFC3339)

	err = s.storage.CreateOrder(c.Request().Context(), o)
	if err != nil {
		if errors.Is(err, entities.ErrConflict) {
			return c.JSON(http.StatusConflict, "order already loaded by another user")
		}
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	return nil
}
