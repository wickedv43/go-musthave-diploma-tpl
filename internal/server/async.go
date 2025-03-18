package server

import (
	"context"
	"sync"
	"time"

	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage"
)

func (s *Server) gen(orders ...storage.Order) chan storage.Order {
	outCh := make(chan storage.Order)

	go func() {
		defer close(outCh)
		for _, order := range orders {
			outCh <- order
		}
	}()

	return outCh
}

func (s *Server) check(inCh chan storage.Order) chan storage.Order {
	outCh := make(chan storage.Order)

	go func() {
		defer close(outCh)

		for order := range inCh {
			updatedOrder, err := s.checkOrder(order)
			if err != nil {
				s.logger.Errorln("Failed to check order:", order.Number, err)
				continue
			}

			outCh <- updatedOrder

			time.Sleep(2 * time.Second)
		}
	}()

	return outCh
}

func (s *Server) fanIn(chs ...chan storage.Order) chan storage.Order {
	var wg sync.WaitGroup
	outCh := make(chan storage.Order)

	output := func(c chan storage.Order) {
		for n := range c {
			outCh <- n
		}
		wg.Done()
	}

	wg.Add(len(chs))

	for _, c := range chs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(outCh)
	}()

	return outCh
}

func (s *Server) watch(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				s.logger.Infoln("watch stopped")
				return
			case <-ticker.C:
				orders, err := s.storage.ProcessingOrders(ctx)
				if err != nil {
					s.logger.Errorln("Failed to fetch orders:", err)
					continue
				}
				if len(orders) == 0 {
					s.logger.Infoln("No orders to process")
					continue
				}

				orderCh := s.gen(orders...)
				checkedOrderCh := s.check(orderCh)
				finalCh := s.fanIn(checkedOrderCh)

				for order := range finalCh {
					err = s.storage.UpdateOrder(ctx, order)
					if err != nil {
						s.logger.Errorln("Failed to update order:", order.Number, err)
					}

					s.logger.Infoln("Order updated:", order.Number)

					if order.Status == "PROCESSED" {
						var user storage.User
						user, err = s.storage.GetUser(ctx, order.UserID)
						if err != nil {
							s.logger.Errorln("Failed to get user:", err)
						}

						user.Balance.Current += order.Accrual

						err = s.storage.UpdateUserBalance(ctx, user)
						if err != nil {
							s.logger.Errorln("Failed to update user balance:", err)
						}
					}
				}
			}
		}
	}()
}
