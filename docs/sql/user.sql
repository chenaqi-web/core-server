CREATE TABLE `user` (
                        `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                        `name` VARCHAR(50) NOT NULL COMMENT '用户名',
                        `password` VARCHAR(255) NOT NULL COMMENT '密码',
                        `phone` VARCHAR(20) DEFAULT '' COMMENT '手机号',
                        `avatar` VARCHAR(500) DEFAULT '' COMMENT '头像URL',
                        `email` VARCHAR(100) DEFAULT '' COMMENT '邮箱',
                        `role` VARCHAR(20) DEFAULT 'user' COMMENT '角色：admin/user/guest',
                        `sex` VARCHAR(10) DEFAULT '' COMMENT '性别：male/female',
                        `age` BIGINT UNSIGNED DEFAULT 0 COMMENT '年龄',
                        `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                        `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                        `deleted_at` DATETIME DEFAULT NULL COMMENT '软删除时间',
                        PRIMARY KEY (`id`),
                        UNIQUE KEY `uk_name` (`name`),
                        KEY `idx_email` (`email`),
                        KEY `idx_phone` (`phone`),
                        KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';