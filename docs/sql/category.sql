-- 分类表（固定二级结构，不支持三级及以上）
-- 一级：parent_id = 0，如 动作、喜剧
-- 二级：parent_id = 一级分类 id，如 全部、爱情喜剧

CREATE TABLE IF NOT EXISTS `category` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '软删除时间',
  `parent_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '父分类ID：0=一级分类，非0=二级分类（必须指向一级分类）',
  `name` varchar(64) NOT NULL COMMENT '分类名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_category_parent_id_name` (`parent_id`, `name`),
  KEY `idx_category_parent_id` (`parent_id`),
  KEY `idx_category_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='二级分类';
