-- Category table (two-level hierarchy).
CREATE TABLE IF NOT EXISTS `category` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `created_at` datetime(3) DEFAULT NULL COMMENT 'created time',
  `updated_at` datetime(3) DEFAULT NULL COMMENT 'updated time',
  `parent_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT 'parent category id',
  `name` varchar(64) NOT NULL COMMENT 'category name',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_category_parent_id_name` (`parent_id`, `name`),
  KEY `idx_category_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='category';
