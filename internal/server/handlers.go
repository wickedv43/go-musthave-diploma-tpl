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

func (s *Server) getUserID(c echo.Context) (int, error) {
	uid := c.Get("UserID")

	userID, ok := uid.(int)
	if !ok {
		return 0, errors.New("invalid user id")
	}

	return userID, nil
}

func (s *Server) onRegUser(c echo.Context) error {
	var aud storage.AuthData

	//get log & pass
	if err := c.Bind(&aud); err != nil {
		return c.JSON(http.StatusBadRequest, "Bad Request")
	}

	//reg user
	user, err := s.storage.RegisterUser(c.Request().Context(), aud)
	if err != nil {
		if errors.Is(err, entities.ErrConflict) {
			return c.JSON(http.StatusConflict, "login already exists")
		}
		return c.JSON(http.StatusInternalServerError, err)
	}

	c = s.authorize(c, user)

	return c.JSON(http.StatusOK, nil)
}

func (s *Server) onLogin(c echo.Context) error {
	var aud storage.AuthData

	if err := c.Bind(&aud); err != nil {
		return c.JSON(http.StatusBadRequest, "Bad Request")
	}

	user, err := s.storage.LoginUser(c.Request().Context(), aud)
	if err != nil {
		if errors.Is(err, entities.ErrBadLogin) {
			return c.JSON(http.StatusConflict, "permission denied")
		}

		return c.JSON(http.StatusInternalServerError, err)
	}

	c = s.authorize(c, user)

	return c.JSON(http.StatusOK, nil)
}

func (s *Server) onPostOrders(c echo.Context) error {
	if c.Request().Header.Get("Content-Type") == "application/json" {
		return c.JSON(http.StatusBadRequest, "Bad request")
	}

	//get order num
	body := c.Request().Body

	orderNum, err := io.ReadAll(body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	//validate orderNum
	var order storage.Order

	if !util.LuhnCheck(string(orderNum)) {
		return c.JSON(http.StatusUnprocessableEntity, "Unprocessable Entity")
	}

	order.Number = string(orderNum)
	order.UploadedAt = time.Now().Format(time.RFC3339)

	//create order
	err = s.storage.CreateOrder(c.Request().Context(), order)
	if err != nil {
		//if another user have same order num
		if errors.Is(err, entities.ErrConflict) {
			return c.JSON(http.StatusConflict, "order already loaded by another user")
		}

		//if user already have this order num
		if errors.Is(err, entities.ErrAlreadyExists) {
			return c.JSON(http.StatusOK, "order already exists")
		}
		//another problem
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	return c.JSON(http.StatusAccepted, nil)
}

func (s *Server) onGetOrders(c echo.Context) error {
	//get userID from cookie
	userID, err := s.getUserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	//get user from postgres
	user, err := s.storage.UserData(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Server error")
	}
	//if user haven't orders
	if len(user.Orders) == 0 {
		return c.JSON(http.StatusNoContent, "No content")
	}

	return c.JSON(http.StatusOK, user.Orders)
}

func (s *Server) onGetUserBalance(c echo.Context) error {
	//get userID from cookie
	userID, err := s.getUserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	//get user from postgres
	user, err := s.storage.UserData(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	return c.JSON(http.StatusOK, user.Balance)
}

func (s *Server) onProcessPayment(c echo.Context) error {
	var pr storage.Bill

	//parse req
	if err := c.Bind(&pr); err != nil {
		return c.JSON(http.StatusBadRequest, "Bad Request")
	}

	//process payment
	err := s.storage.ProcessPayment(c.Request().Context(), pr)
	if err != nil {
		//if bad order num
		if errors.Is(err, entities.ErrBadOrder) {
			return c.JSON(http.StatusUnprocessableEntity, "order already loaded")
		}
		//if user have not money
		if errors.Is(err, entities.ErrHaveEnoughMoney) {
			return c.JSON(http.StatusPaymentRequired, "user have enough money")
		}

		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	return c.JSON(http.StatusOK, nil)
}

func (s *Server) GetUserBills(c echo.Context) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	//get user
	user, err := s.storage.UserData(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	return c.JSON(http.StatusOK, user.Bills)
}
