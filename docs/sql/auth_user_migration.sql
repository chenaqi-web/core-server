-- Prerequisite: review auth_user_precheck.sql output and resolve duplicates,
-- empty emails, NULL profile values, and invalid roles manually.
-- This migration never deletes users or overwrites legacy field values.
-- Strict mode makes incompatible legacy values fail the migration instead of
-- being truncated or coerced by MySQL.

SET @auth_previous_sql_mode = @@SESSION.sql_mode;
SET SESSION sql_mode = CONCAT_WS(',', NULLIF(@@SESSION.sql_mode, ''), 'STRICT_ALL_TABLES');

SET @auth_status_column_sql = (
    SELECT IF(
        COUNT(*) = 0,
        'ALTER TABLE `user` ADD COLUMN `status` varchar(16) NOT NULL DEFAULT ''active'' AFTER `role`',
        'SELECT ''status column already exists'' AS migration_note'
    )
    FROM `information_schema`.`columns`
    WHERE `table_schema` = DATABASE()
      AND `table_name` = 'user'
      AND `column_name` = 'status'
);
PREPARE auth_status_column_stmt FROM @auth_status_column_sql;
EXECUTE auth_status_column_stmt;
DEALLOCATE PREPARE auth_status_column_stmt;

SET @auth_version_column_sql = (
    SELECT IF(
        COUNT(*) = 0,
        'ALTER TABLE `user` ADD COLUMN `auth_version` bigint unsigned NOT NULL DEFAULT 1 AFTER `status`',
        'SELECT ''auth_version column already exists'' AS migration_note'
    )
    FROM `information_schema`.`columns`
    WHERE `table_schema` = DATABASE()
      AND `table_name` = 'user'
      AND `column_name` = 'auth_version'
);
PREPARE auth_version_column_stmt FROM @auth_version_column_sql;
EXECUTE auth_version_column_stmt;
DEALLOCATE PREPARE auth_version_column_stmt;

-- These statements preserve existing values. They fail instead of silently
-- coercing data when unresolved NULL or incompatible legacy values remain.
ALTER TABLE `user`
    MODIFY COLUMN `name` varchar(64) NOT NULL,
    MODIFY COLUMN `email` varchar(128) NOT NULL,
    MODIFY COLUMN `password` varchar(255) NOT NULL,
    MODIFY COLUMN `phone` varchar(20) NOT NULL DEFAULT '',
    MODIFY COLUMN `avatar` varchar(255) NOT NULL DEFAULT '',
    MODIFY COLUMN `sex` varchar(8) NOT NULL DEFAULT '未知',
    MODIFY COLUMN `age` bigint unsigned NOT NULL DEFAULT 0,
    MODIFY COLUMN `role` varchar(32) NOT NULL DEFAULT 'user',
    MODIFY COLUMN `status` varchar(16) NOT NULL DEFAULT 'active',
    MODIFY COLUMN `auth_version` bigint unsigned NOT NULL DEFAULT 1;

SET @auth_name_index_sql = (
    SELECT IF(
        COUNT(*) > 0,
        'SELECT ''unique username index already exists'' AS migration_note',
        'CREATE UNIQUE INDEX `uk_user_name` ON `user` (`name`)'
    )
    FROM (
        SELECT `index_name`
        FROM `information_schema`.`statistics`
        WHERE `table_schema` = DATABASE()
          AND `table_name` = 'user'
          AND `non_unique` = 0
        GROUP BY `index_name`
        HAVING COUNT(*) = 1 AND MAX(`column_name` = 'name') = 1
    ) AS unique_name_indexes
);
PREPARE auth_name_index_stmt FROM @auth_name_index_sql;
EXECUTE auth_name_index_stmt;
DEALLOCATE PREPARE auth_name_index_stmt;

SET @auth_email_index_sql = (
    SELECT IF(
        COUNT(*) > 0,
        'SELECT ''unique email index already exists'' AS migration_note',
        'CREATE UNIQUE INDEX `uk_user_email` ON `user` (`email`)'
    )
    FROM (
        SELECT `index_name`
        FROM `information_schema`.`statistics`
        WHERE `table_schema` = DATABASE()
          AND `table_name` = 'user'
          AND `non_unique` = 0
        GROUP BY `index_name`
        HAVING COUNT(*) = 1 AND MAX(`column_name` = 'email') = 1
    ) AS unique_email_indexes
);
PREPARE auth_email_index_stmt FROM @auth_email_index_sql;
EXECUTE auth_email_index_stmt;
DEALLOCATE PREPARE auth_email_index_stmt;

SET SESSION sql_mode = @auth_previous_sql_mode;
