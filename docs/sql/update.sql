CREATE TABLE `user` (
                        `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                        `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
                        `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
                        `deleted_at` DATETIME(3) NULL COMMENT '软删除时间，为空表示未删除',
                        `name` VARCHAR(255) NOT NULL COMMENT '用户名',
                        `password` VARCHAR(255) NOT NULL COMMENT '密码',
                        `phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '手机号',
                        `avatar` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '头像',
                        `email` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '邮箱',
                        `role` VARCHAR(20) NOT NULL DEFAULT 'user' COMMENT '角色',
                        `sex` VARCHAR(6) NOT NULL DEFAULT '' COMMENT '性别',
                        `birthday` DATETIME(3) NULL COMMENT '生日',
                        `status` VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '用户状态',
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `user_stat` (
                             `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                             `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
                             `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
                             `deleted_at` DATETIME(3) NULL COMMENT '删除时间',
                             `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                             `article_count` BIGINT UNSIGNED NULL COMMENT '发帖/文章数量',
                             `followers_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '粉丝数',
                             `following_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关注数',
                             `like_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '点赞总数',
                             `receive_like_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收到的点赞总数',
                             `favor_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收藏的数量',
                             `receive_favor_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '被收藏的总数',
                             PRIMARY KEY (`id`),
                             UNIQUE KEY `user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



CREATE TABLE `category` (
                            `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                            `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
                            `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
                            `parent_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
                            `name` VARCHAR(64) NOT NULL,
                            PRIMARY KEY (`id`),
                            UNIQUE KEY `parent_id_name` (`parent_id`, `name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
