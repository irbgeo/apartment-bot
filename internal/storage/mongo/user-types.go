package mongo

type user struct {
	ID          int64 `bson:"tg_id"`
	ClientID    int64 `bson:"client_id"`
	IsSuperuser bool  `bson:"is_superuser"`
}
