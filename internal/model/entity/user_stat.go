package entity

type UserStat struct {
	UserID uint64 `db:"user_id"`

	ArticleCount uint64 `db:"article_count"`

	FollowersCount uint64 `db:"followers_count"`
	FollowingCount uint64 `db:"following_count"`

	LikeCount        uint64 `db:"like_count"`
	ReceiveLikeCount uint64 `db:"receive_like_count"`

	FavorCount        uint64 `db:"favor_count"`
	ReceiveFavorCount uint64 `db:"receive_favor_count"`
}
