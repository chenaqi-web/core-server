package repo

// IUserRepo defines the interface for user-related data access operations.

type UserPepo struct {
	client *DBClient
}

func NewUserPepo(client *DBClient) *UserPepo {
	return &UserPepo{
		client: client,
	}
}

func (u *UserPepo) Login() {
	_ = u.client.GetDB()
}
