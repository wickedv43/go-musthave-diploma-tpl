package server

import (
	"fmt"

	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage"
)

func (s *Server) checkOrder(order storage.Order) (storage.Order, error) {
	url := fmt.Sprintf("%s/api/orders/%s", s.cfg.AccrualSystem.URL, order.Number)

	resp, err := s.client.R().Get(url)
	if err != nil {
		return order, err
	}

	s.logger.Info("[ACCRUAL URL]", url)
	s.logger.Info("[ACCRUAL RESPONSE]", resp.String())

	return order, nil
}
