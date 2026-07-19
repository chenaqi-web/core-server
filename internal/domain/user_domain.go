package domain

type UserRepo interface {
}

type UserRepoDomain interface {
	ITransaction
	LikeRepo
}

// =====================================================================================================================

type UserCacheDomain interface {
}
