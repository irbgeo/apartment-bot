package message

import (
	"sync"

	tele "gopkg.in/telebot.v3"

	tgbot "github.com/irbgeo/apartment-bot/internal/client/tg"
)

// Message структура представляет сообщение с его типом.
type Message struct {
	Message *tele.Message
	Type    tgbot.MessageType
}

// Stack представляет структуру стека сообщений для каждого пользователя.
type Stack struct {
	b                *tele.Bot
	messageUserStack sync.Map
}

// NewStack создает новый экземпляр стека сообщений.
func NewStack() *Stack {
	return &Stack{}
}

// SetBot устанавливает бота для стека сообщений.
func (s *Stack) SetBot(b *tele.Bot) {
	s.b = b
}

// Store сохраняет сообщение в стеке для указанного пользователя.
func (s *Stack) Store(userID int64, m *tele.Message, t tgbot.MessageType) {
	userMQ := s.getOrCreateUserMessageQueue(userID)
	userMQ.Store(m, t)
}

// GetOrCleanTill возвращает или удаляет сообщение из стека до определенного типа.
func (s *Stack) GetOrCleanTill(userID int64, goalType, tillType tgbot.MessageType) (*tele.Message, bool, error) {
	userMQ, isExist := s.messageUserStack.Load(userID)
	if !isExist {
		return nil, false, nil
	}
	return userMQ.(*messageUserStack).GetOrCleanTill(goalType, tillType)
}

// Clean очищает стек сообщений для указанного пользователя.
func (s *Stack) Clean(userID int64) error {
	userMQ, isExist := s.messageUserStack.Load(userID)
	if !isExist {
		return nil
	}
	return userMQ.(*messageUserStack).Clean()
}

// CleanTill очищает стек сообщений до определенного типа сообщения для указанного пользователя.
func (s *Stack) CleanTill(userID int64, t tgbot.MessageType) error {
	userMQ, isExist := s.messageUserStack.Load(userID)
	if !isExist {
		return nil
	}
	return userMQ.(*messageUserStack).cleanTill(t)
}

// getOrCreateUserMessageQueue получает или создает очередь сообщений для пользователя.
func (s *Stack) getOrCreateUserMessageQueue(userID int64) *messageUserStack {
	e, isExist := s.messageUserStack.Load(userID)
	if !isExist {
		userMQ := &messageUserStack{
			b: s.b,
		}
		s.messageUserStack.Store(userID, userMQ)
		return userMQ
	}
	return e.(*messageUserStack)
}

// messageUserStack представляет стек сообщений для конкретного пользователя.
type messageUserStack struct {
	b        *tele.Bot
	m        sync.Mutex
	messages []*Message
}

// Store сохраняет сообщение в стеке для данного пользователя.
func (s *messageUserStack) Store(m *tele.Message, t tgbot.MessageType) {
	s.messages = append(s.messages, &Message{
		Message: m,
		Type:    t,
	})
}

// GetOrCleanTill возвращает или удаляет сообщение из стека до определенного типа.
func (s *messageUserStack) GetOrCleanTill(goalType, tillType tgbot.MessageType) (*tele.Message, bool, error) {
	s.m.Lock()
	defer s.m.Unlock()

	message := s.top()

	for message != nil && message.Type != goalType && message.Type != tillType {
		s.pop()
		err := s.b.Delete(message.Message)
		if err != nil {
			return nil, false, err
		}
		message = s.top()
	}

	if message != nil && message.Type == goalType {
		s.pop()
		return message.Message, true, nil
	}
	return nil, false, nil
}

// Clean удаляет все сообщения из стека для данного пользователя.
func (s *messageUserStack) Clean() error {
	s.m.Lock()
	defer s.m.Unlock()

	message := s.pop()

	for message != nil {
		err := s.b.Delete(message.Message)
		if err != nil {
			return err
		}
		message = s.pop()
	}

	return nil
}

// cleanTill удаляет все сообщения из стека до определенного типа для данного пользователя.
func (s *messageUserStack) cleanTill(t tgbot.MessageType) error {
	s.m.Lock()
	defer s.m.Unlock()

	for len(s.messages) != 0 {
		msg := s.messages[len(s.messages)-1]
		if msg.Type == t {
			return nil
		}
		s.messages = s.messages[:len(s.messages)-1]

		err := s.b.Delete(msg.Message)
		if err != nil {
			return err
		}
	}

	return nil
}

// pop удаляет верхнее сообщение из стека.
func (s *messageUserStack) pop() *Message {
	if len(s.messages) == 0 {
		return nil
	}
	msg := s.messages[len(s.messages)-1]
	s.messages = s.messages[:len(s.messages)-1]

	return msg
}

// top возвращает верхнее сообщение в стеке.
func (s *messageUserStack) top() *Message {
	if len(s.messages) == 0 {
		return nil
	}

	return s.messages[len(s.messages)-1]
}
