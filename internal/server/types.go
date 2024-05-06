package server

type User struct {
	ID          int64
	ClientID    int64
	IsSuperuser bool
}

type City struct {
	Name     string
	District map[string]struct{}
}
