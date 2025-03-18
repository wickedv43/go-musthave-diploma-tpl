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
	uid := c.Get("userID")

	userID, ok := uid.(int)
	if !ok {
		return 0, errors.New("invalid user id")
	}

	return userID, nil
}

func (s *Server) onRegister(c echo.Context) error {
	var aud storage.AuthData

	//get log & pass
	if err := c.Bind(&aud); err != nil {
		return c.JSON(http.StatusBadRequest, "Bad Request")
	}

	//reg user
	user, err := s.storage.CreateUser(c.Request().Context(), aud)
	if err != nil {
		if errors.Is(err, entities.ErrAlreadyExists) {
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
			return c.JSON(http.StatusUnauthorized, "permission denied")
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
		return c.JSON(http.StatusInternalServerError, "read order number")
	}

	//validate orderNum
	if !util.LuhnCheck(string(orderNum)) {
		return c.JSON(http.StatusUnprocessableEntity, "Unprocessable Entity")
	}

	//get userID from cookie
	userID, err := s.getUserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "get user ID")
	}

	order := storage.Order{
		UserID:     userID,
		Number:     string(orderNum),
		Status:     "",
		Accrual:    0,
		UploadedAt: time.Now().Format(time.RFC3339),
	}

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
		return c.JSON(http.StatusInternalServerError, err.Error())
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
	user, err := s.storage.GetUser(c.Request().Context(), userID)
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
	user, err := s.storage.GetUser(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	return c.JSON(http.StatusOK, user.Balance)
}

func (s *Server) onWithDraw(c echo.Context) error {
	//parse req
	var bill storage.Bill

	err := c.Bind(&bill)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Bad Request")
	}

	//get userID
	userID, err := s.getUserID(c)
	if err != nil {
		s.logger.WithError(err).Error("failed to get the user from cookie")
		return c.JSON(http.StatusInternalServerError, "Server error")
	}
	bill.UserID = userID
	bill.ProcessedAt = time.Now().Format(time.RFC3339)

	//check withdraw num
	if !util.LuhnCheck(bill.Order) {
		return c.JSON(http.StatusUnprocessableEntity, "Order doesn't exist")
	}

	//check user
	user, err := s.storage.GetUser(c.Request().Context(), userID)
	if err != nil {
		s.logger.WithError(err).Error("failed to get user from storage")
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	if user.Balance.Current < bill.Sum {
		return c.JSON(http.StatusPaymentRequired, "Not enough balance")
	}
	//create bill
	err = s.storage.CreateBill(c.Request().Context(), bill)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create bill")
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	//withdraw
	user.Balance.Current -= bill.Sum
	user.Balance.Withdrawn += bill.Sum

	err = s.storage.UpdateUserBalance(c.Request().Context(), user)
	if err != nil {
		s.logger.WithError(err).Error("Failed to update user balance")
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, nil)
}

func (s *Server) GetUserBills(c echo.Context) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	//get user
	user, err := s.storage.GetUser(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	if len(user.Bills) == 0 {
		return c.JSON(http.StatusNoContent, "no content")
	}

	return c.JSON(http.StatusOK, user.Bills)
}
