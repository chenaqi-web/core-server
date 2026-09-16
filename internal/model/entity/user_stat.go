package entity

type UserStat struct {
	UserID uint64 `db:"user_id"`

	FollowersCount uint64 `db:"followers_count"`
	FollowingCount uint64 `db:"following_count"`

	LikeCount        uint64 `db:"like_count"`
	ReceiveLikeCount uint64 `db:"receive_like_count"`

	ViewCount        uint64 `db:"view_count"`
	ReceiveViewCount uint64 `db:"receive_view_count"`

	FavorCount        uint64 `db:"favor_count"`
	ReceiveFavorCount uint64 `db:"receive_favor_count"`
}
