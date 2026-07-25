package domain

type UserRepo interface {
}

type UserRepoDomain interface {
	ITransaction
	UserRepo
}

// =====================================================================================================================

type UserCacheDomain interface {
}
