package aggregate

import "core-server/internal/model/entity"

type UserAggregate struct {
	// 用户信息表
	user *entity.User

	// 用户总计数表
	stat *entity.UserStat
}
