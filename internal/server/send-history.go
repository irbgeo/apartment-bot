package server

import (
	"context"
)

func (s *service) startSendHistoryData(ctx context.Context, f Filter, source <-chan Apartment) <-chan Apartment {
	resultCh := make(chan Apartment)

	histCtx, cancel := context.WithCancel(ctx)
	s.historySending.Store(f.ID, cancel)

	go func() {
		defer close(resultCh)
		defer s.stopSendHistoryData(f)

		for {
			select {
			case <-histCtx.Done():
				return
			case a, ok := <-source:
				if !ok {
					return
				}
				if s.checkApartment(histCtx, a) {
					a.Filter = map[int64][]string{
						f.User.ID: {*f.Name},
					}
					resultCh <- a
				}
			}
		}
	}()

	return resultCh
}

func (s *service) stopSendHistoryData(f Filter) {
	cancel, ok := s.historySending.Load(f.ID)
	if !ok {
		return
	}

	cancel.(context.CancelFunc)()
	s.historySending.Delete(f.ID)
}
