package repo

type LikeRepo struct {
	client *DBClient
}

func NewLikeRepo(client *DBClient) *LikeRepo {
	return &LikeRepo{
		client: client,
	}
}

func (l *LikeRepo) Like() {
	_ = l.client.GetDB()
}
