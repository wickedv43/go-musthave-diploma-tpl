package server

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage"
)

func (s *Server) checkOrder(order *storage.Order) error {

	url := fmt.Sprintf("%s/api/orders/%s", s.cfg.AccrualSystem.URL, order.Number)

	resp, err := s.client.R().Get(url)
	if err != nil {
		return errors.Wrapf(err, "failed to check order %s", order.Number)
	}
	s.logger.Info("[ACCRUAL RESPONSE] ", resp.Status())

	var acOrder storage.Order

	err = json.Unmarshal(resp.Body(), &acOrder)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal order %s", order.Number)
	}

	order.Accrual = acOrder.Accrual
	order.Status = acOrder.Status

	return nil
}
