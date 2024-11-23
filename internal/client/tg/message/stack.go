package message

import (
	"fmt"
	"sync"

	tele "gopkg.in/telebot.v3"

	tgbot "github.com/irbgeo/apartment-bot/internal/client/tg"
)

// stack represents a thread-safe message stack for a specific user.
type stack struct {
	bot      *tele.Bot
	mutex    sync.Mutex
	messages []*message
}

func (s *stack) store(msg *message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.messages = append(s.messages, msg)
}

func (s *stack) getOrCleanTill(targetType, tillType tgbot.MessageType) (*tele.Message, bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for len(s.messages) > 0 {
		msg := s.peek()
		if msg == nil {
			return nil, false, nil
		}

		if msg.Type == targetType {
			return s.pop().Content, true, nil
		}

		if msg.Type == tillType {
			return nil, false, nil
		}

		popped := s.pop()
		if err := s.deleteMessage(popped.Content); err != nil {
			return nil, false, fmt.Errorf("failed to delete message: %w", err)
		}
	}

	return nil, false, nil
}

func (s *stack) clean() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for len(s.messages) > 0 {
		msg := s.pop()
		if err := s.deleteMessage(msg.Content); err != nil {
			return fmt.Errorf("failed to clean messages: %w", err)
		}
	}
	return nil
}

func (s *stack) cleanUntil(msgType tgbot.MessageType) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for len(s.messages) > 0 {
		msg := s.peek()
		if msg.Type == msgType {
			return nil
		}

		s.pop()
		if err := s.deleteMessage(msg.Content); err != nil {
			return fmt.Errorf("failed to clean messages until type %v: %w", msgType, err)
		}
	}
	return nil
}

func (s *stack) pop() *message {
	if len(s.messages) == 0 {
		return nil
	}
	lastIdx := len(s.messages) - 1
	msg := s.messages[lastIdx]
	s.messages = s.messages[:lastIdx]
	return msg
}

func (s *stack) peek() *message {
	if len(s.messages) == 0 {
		return nil
	}
	return s.messages[len(s.messages)-1]
}

func (s *stack) deleteMessage(msg *tele.Message) error {
	if msg == nil {
		return nil
	}
	return s.bot.Delete(msg)
}
