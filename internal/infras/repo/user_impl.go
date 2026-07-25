package repo

type UserRepo struct {
	*DBClient
}

func NewUserRepo(client *DBClient) *UserRepo {
	return &UserRepo{DBClient: client}
}
