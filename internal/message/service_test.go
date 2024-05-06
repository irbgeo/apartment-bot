package message_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/irbgeo/apartment-bot/internal/message"
	"github.com/irbgeo/apartment-bot/internal/message/mocks"
)

var (
	N        int64 = 10
	testText       = "test text"
	msg            = message.Message{
		Text: testText,
	}
)

func TestService_Publish(t *testing.T) {
	ctx := context.Background()

	var users sync.Map
	userCh := make(chan message.User)

	mockStorage := getStorageMock(t, ctx, userCh)
	s := message.NewService(mockStorage)

	go func() {
		defer close(userCh)

		var i int64
		for i = 1; i < N; i++ {
			user := message.User{
				ID: i,
			}
			users.Store(i, user)

			userCh <- user
		}

		<-time.After(time.Second)
		s.Close()
	}()

	go func() {
		err := s.Publish(ctx, msg)
		assert.NoError(t, err)
	}()

	msgCh := s.Watcher()
	for msg := range msgCh {
		_, ok := users.Load(msg.UserID)
		require.Equal(t, testText, msg.Text, msg.UserID)
		require.True(t, ok, msg.UserID)
		users.Delete(msg.UserID)
	}

	var count int64
	users.Range(func(_, _ any) bool {
		count++
		return true
	})
	require.Equal(t, int64(0), count)
}

func getStorageMock(t *testing.T, ctx context.Context, ch <-chan message.User) *mocks.UserStorage {
	t.Helper()

	mockStorage := mocks.NewUserStorage(t)

	mockStorage.On("Users", ctx).Return(
		ch,
		nil,
	)

	return mockStorage
}
