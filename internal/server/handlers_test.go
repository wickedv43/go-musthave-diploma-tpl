package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/config"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/logger"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage/mocks"
	"go.uber.org/mock/gomock"
)

func setupTestServer(t *testing.T, configureMock func(*mocks.MockDataKeeper)) *Server {
	t.Helper()

	ctrl := gomock.NewController(t)

	container := do.New()
	do.Provide(container, func(i do.Injector) (*config.Config, error) {
		return &config.Config{Server: config.Server{RunAddress: ":8080"}}, nil
	})

	do.Provide(container, logger.NewLogger)

	mockKeeper := mocks.NewMockDataKeeper(ctrl)
	configureMock(mockKeeper)

	do.Provide(container, func(i do.Injector) (storage.DataKeeper, error) {
		return mockKeeper, nil
	})

	do.Provide(container, NewServer)
	return do.MustInvoke[*Server](container)
}

func TestOnRegister_Success(t *testing.T) {
	s := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			CreateUser(gomock.Any(), gomock.Any()).
			Return(storage.User{ID: 1}, nil)
	})

	body := map[string]string{"login": "user", "password": "pass"}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestOnLogin_Success(t *testing.T) {
	s := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			CreateUser(gomock.Any(), gomock.Any()).
			Return(storage.User{ID: 1}, nil)
	})

	body := map[string]string{"login": "user", "password": "pass"}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestOnPostOrders_Success(t *testing.T) {
	s := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			CreateOrder(gomock.Any(), gomock.Any()).
			Return(nil)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte("12345678903")))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	// ставим userID в контекст
	c := s.echo.NewContext(req, rec)
	c.Set("userID", 1)

	_ = s.onPostOrders(c)

	assert.Equal(t, http.StatusAccepted, rec.Code)
}

func TestOnGetOrders_WithOrders(t *testing.T) {
	s := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			GetUser(gomock.Any(), 1).
			Return(storage.User{Orders: []storage.Order{{Number: "123"}}}, nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.Set("userID", 1)

	_ = s.onGetOrders(c)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestOnGetUserBalance_Success(t *testing.T) {
	s := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			GetUser(gomock.Any(), 1).
			Return(storage.User{Balance: storage.UserBalance{Current: 100}}, nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.Set("userID", 1)

	_ = s.onGetUserBalance(c)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestOnWithDraw_Success(t *testing.T) {
	s := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			WithdrawFromBalance(gomock.Any(), 1, float32(10)).
			Return(nil)
		mock.EXPECT().
			CreateBill(gomock.Any(), gomock.Any()).
			Return(nil)
	})

	bill := storage.Bill{Order: "79927398713", Sum: 10}
	data, _ := json.Marshal(bill)

	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.Set("userID", 1)

	_ = s.onWithDraw(c)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetUserBills_WithData(t *testing.T) {
	s := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			GetUser(gomock.Any(), 1).
			Return(storage.User{Bills: []storage.Bill{{Order: "abc"}}}, nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.Set("userID", 1)

	_ = s.GetUserBills(c)

	assert.Equal(t, http.StatusOK, rec.Code)
}
