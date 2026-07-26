CREATE TABLE `comment`
(
    `id`            bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `article_id`    bigint unsigned NOT NULL DEFAULT '0' COMMENT '文章ID',
    `user_id`       bigint unsigned NOT NULL DEFAULT '0' COMMENT '评论用户ID',
    `parent_id`     bigint unsigned NOT NULL DEFAULT '0' COMMENT '父评论ID',
    `root_id`       bigint unsigned NOT NULL DEFAULT '0' COMMENT '根评论ID',
    `reply_to_id`   bigint unsigned NOT NULL DEFAULT '0' COMMENT '回复目标用户ID',
    `reply_to_name` varchar(64)     NOT NULL DEFAULT '' COMMENT '回复目标用户名',
    `content`       text            NOT NULL COMMENT '评论内容',
    `like_count`    int unsigned    NOT NULL DEFAULT '0' COMMENT '点赞数',
    `child_count`   int unsigned    NOT NULL DEFAULT '0' COMMENT '子评论数',
    `created_at`    datetime(3)     NULL COMMENT '创建时间',
    `deleted_at`    datetime(3)     NULL COMMENT '软删除时间',
    PRIMARY KEY (`id`),
    INDEX `idx_article_id` (`article_id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_root_id` (`root_id`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='评论表';