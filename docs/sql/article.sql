CREATE TABLE blog_article (
    -- ========== 主键 ==========
                            id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '文章ID',

    -- ========== 基础信息 ==========
                            title VARCHAR(200) NOT NULL COMMENT '文章标题',
                            summary VARCHAR(500) NOT NULL DEFAULT '' COMMENT '文章摘要',
                            content LONGTEXT NOT NULL COMMENT '文章正文（Markdown格式）',
                            cover_image VARCHAR(500) NOT NULL DEFAULT '' COMMENT '封面图URL',

    -- ========== 用户关联 ==========
                            author_id BIGINT UNSIGNED NOT NULL COMMENT '作者用户ID',

    -- ========== 分类关联 ==========
                            category_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '分类ID（0表示未分类）',

    -- ========== 状态 ==========
                            is_top TINYINT NOT NULL DEFAULT 0 COMMENT '是否置顶：0-否，1-是',

    -- ========== 统计字段 ==========
                            view_count BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '浏览量',
                            like_count BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '点赞量',
                            comment_count BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '评论量',

    -- ========== 时间戳 ==========
                            created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                            updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                            deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间（NULL表示未删除）',

    -- ========== 索引 ==========
                            INDEX idx_user_id (author_id),
                            INDEX idx_category_id (category_id),
                            INDEX idx_created_at (created_at),
                            INDEX idx_deleted_at (deleted_at)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='博客文章表';