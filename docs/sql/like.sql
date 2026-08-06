-- =====================================================
-- 点赞交互表
-- =====================================================
CREATE TABLE IF NOT EXISTS `interaction_like` (
                                                  `id` VARCHAR(64) NOT NULL COMMENT '主键ID',
    `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
    `object_type` VARCHAR(32) NOT NULL COMMENT '对象类型（如：post, comment, video等）',
    `object_id` VARCHAR(64) NOT NULL COMMENT '对象ID',
    `status` VARCHAR(32) NOT NULL DEFAULT 'thumb_up' COMMENT '状态：thumb_up-已点赞, nothing-已取消',
    `version` BIGINT NOT NULL DEFAULT 0 COMMENT '版本号（乐观锁）',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_object` (`user_id`, `object_type`, `object_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='点赞交互表';