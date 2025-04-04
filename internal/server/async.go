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
			for {
				pauseUntil := s.getPause()
				if time.Now().Before(pauseUntil) {
					sleepDur := time.Until(pauseUntil)
					s.logger.Warnf("Paused due to 429, sleeping for %v", sleepDur)
					time.Sleep(sleepDur)
				} else {
					break
				}
			}

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

					continue
				}

				orderCh := s.gen(orders...)
				checkedOrderCh := s.check(orderCh)
				checkedOrderChTwo := s.check(orderCh)
				finalCh := s.fanIn(checkedOrderCh, checkedOrderChTwo)

				for order := range finalCh {
					err = s.storage.UpdateOrder(ctx, order)
					if err != nil {
						s.logger.Errorln("Failed to update order:", order.Number, err)
					}

					if order.Status == "PROCESSED" {
						//update user balance with tx
						err = s.storage.AddToBalanceWithTX(ctx, order.UserID, order.Accrual)
						if err != nil {
							s.logger.Errorln("Failed to update user balance:", err)
						}
					}
				}
			}
		}
	}()
}
