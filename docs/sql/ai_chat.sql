-- AI 助手历史会话表（gateway 简化版）

CREATE TABLE IF NOT EXISTS `ai_chat_session` (
                                                 `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                                                 `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
    `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
    `user_id` varchar(64) NOT NULL COMMENT '用户ID',
    `session_id` varchar(64) NOT NULL COMMENT '会话ID',
    `title` varchar(256) NOT NULL DEFAULT '' COMMENT '会话标题',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_ai_chat_session_id` (`session_id`),
    KEY `idx_ai_chat_session_user_id` (`user_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI助手会话';

CREATE TABLE IF NOT EXISTS `ai_chat_message` (
                                                 `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                                                 `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
    `session_id` varchar(64) NOT NULL COMMENT '会话ID',
    `user_id` varchar(64) NOT NULL COMMENT '用户ID',
    `role` varchar(16) NOT NULL COMMENT '消息角色 user/assistant',
    `content` text NOT NULL COMMENT '消息内容',
    `ai_model` varchar(64) NOT NULL DEFAULT '' COMMENT '使用的模型',
    PRIMARY KEY (`id`),
    KEY `idx_ai_chat_message_session_created` (`session_id`, `created_at`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI助手对话消息';
