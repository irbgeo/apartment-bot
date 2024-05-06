package message

import (
	"context"
)

type service struct {
	storage userStorage

	messageCh chan Message
}

//go:generate mockery --name userStorage --structname UserStorage
type userStorage interface {
	Users(ctx context.Context) (<-chan User, error)
}

func NewService(
	s userStorage,
) *service {
	e := &service{
		storage:   s,
		messageCh: make(chan Message),
	}

	return e
}

func (s *service) Close() {
	close(s.messageCh)
}

func (s *service) Publish(ctx context.Context, msg Message) error {
	userCh, err := s.storage.Users(ctx)
	if err != nil {
		return err
	}

	for u := range userCh {
		msg.UserID = u.ID
		s.messageCh <- msg
	}

	return nil
}

func (s *service) Watcher() <-chan Message {
	return s.messageCh
}
