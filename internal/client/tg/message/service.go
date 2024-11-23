package message

import (
	"sync"

	tele "gopkg.in/telebot.v3"

	tgbot "github.com/irbgeo/apartment-bot/internal/client/tg"
)

// service handles message stacks for multiple users.
type service struct {
	bot        *tele.Bot
	userStacks sync.Map
}

func NewService() *service {
	return &service{}
}

func (s *service) SetBot(bot *tele.Bot) {
	s.bot = bot
}

// getUserStack retrieves or creates a message stack for the specified user.
func (s *service) getUserStack(userID int64) *stack {
	st, exists := s.userStacks.Load(userID)
	if !exists {
		st = &stack{
			bot: s.bot,
		}
		s.userStacks.Store(userID, st)
	}
	return st.(*stack)
}

// StoreMessage saves a message in the user's stack.
func (s *service) StoreMessage(userID int64, msg *tele.Message, msgType tgbot.MessageType) {
	stack := s.getUserStack(userID)
	stack.store(&message{
		Content: msg,
		Type:    msgType,
	})
}

// GetOrCleanTill returns or removes messages until finding a specific message type.
func (s *service) GetOrCleanTill(userID int64, targetType, tillType tgbot.MessageType) (*tele.Message, bool, error) {
	stack := s.getUserStack(userID)
	return stack.getOrCleanTill(targetType, tillType)
}

// CleanUserMessages removes all messages for a specific user.
func (s *service) CleanUserMessages(userID int64) error {
	stack := s.getUserStack(userID)
	return stack.clean()
}

// CleanMessagesUntil removes messages until finding a specific type.
func (s *service) CleanMessagesUntil(userID int64, msgType tgbot.MessageType) error {
	stack := s.getUserStack(userID)
	return stack.cleanUntil(msgType)
}
