CREATE TABLE IF NOT EXISTS `interaction_count` (
   `id` VARCHAR(64) NOT NULL COMMENT '主键ID',
   `object_type` VARCHAR(32) NOT NULL COMMENT '对象类型',
   `object_id` VARCHAR(64) NOT NULL COMMENT '对象ID',
   `interaction_type` VARCHAR(32) NOT NULL COMMENT '交互类型：like, view, favor等',
   `count` BIGINT NOT NULL DEFAULT 0 COMMENT '计数值',
   `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
   `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
   PRIMARY KEY (`id`),
   UNIQUE KEY `uk_object` (`object_type`, `object_id`, `interaction_type`),
   KEY `idx_object_type` (`object_type`),
   KEY `idx_interaction_type` (`interaction_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交互计数表';