CREATE TABLE `user_stat` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT '统计记录ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `followers_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '粉丝数',
    `following_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关注数',
    `like_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '点赞数',
    `receive_like_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收到的点赞数',
    `view_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '浏览数',
    `receive_view_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收到的浏览数',
    `favor_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收藏数',
    `receive_favor_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收到的收藏数',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_stat_user_id` (`user_id`),
    CONSTRAINT `fk_user_stat_user_id` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户统计表';
