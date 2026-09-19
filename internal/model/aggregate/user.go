package aggregate

import "core-server/internal/model/entity"

type UserAggregate struct {
	// 用户信息表
	User *entity.User

	// 用户总计数表
	Stat *entity.UserStat
}
