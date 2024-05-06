package message

type User struct {
	ID int64
}

type Message struct {
	UserID int64
	Text   string
}
