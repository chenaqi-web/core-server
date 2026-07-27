-- Run this file before auth_user_migration.sql.
-- It is read-only and intentionally does not repair or delete legacy data.

SHOW CREATE TABLE `user`;
SHOW INDEX FROM `user`;

SELECT `name`, COUNT(*) AS duplicate_count
FROM `user`
GROUP BY `name`
HAVING COUNT(*) > 1;

SELECT `email`, COUNT(*) AS duplicate_count
FROM `user`
GROUP BY `email`
HAVING COUNT(*) > 1;

SELECT COUNT(*) AS empty_email_count
FROM `user`
WHERE `email` = '';

SELECT
    SUM(`name` = '') AS empty_name_count,
    SUM(`password` = '') AS empty_password_count,
    SUM(`name` IS NULL) AS null_name_count,
    SUM(`email` IS NULL) AS null_email_count,
    SUM(`password` IS NULL) AS null_password_count,
    SUM(`phone` IS NULL) AS null_phone_count,
    SUM(`avatar` IS NULL) AS null_avatar_count,
    SUM(`sex` IS NULL) AS null_sex_count,
    SUM(`age` IS NULL) AS null_age_count,
    SUM(`role` IS NULL) AS null_role_count
FROM `user`;

SELECT
    SUM(CHAR_LENGTH(`name`) > 64) AS oversized_name_count,
    SUM(CHAR_LENGTH(`email`) > 128) AS oversized_email_count,
    SUM(CHAR_LENGTH(`password`) > 255) AS oversized_password_count,
    SUM(CHAR_LENGTH(`phone`) > 20) AS oversized_phone_count,
    SUM(CHAR_LENGTH(`avatar`) > 255) AS oversized_avatar_count,
    SUM(CHAR_LENGTH(`sex`) > 8) AS oversized_sex_count,
    SUM(CHAR_LENGTH(`role`) > 32) AS oversized_role_count
FROM `user`;

SELECT
    `column_name`,
    `column_type`,
    `is_nullable`,
    `column_default`,
    `extra`
FROM `information_schema`.`columns`
WHERE `table_schema` = DATABASE()
  AND `table_name` = 'user'
ORDER BY `ordinal_position`;

SELECT
    `index_name`,
    `non_unique`,
    `seq_in_index`,
    `column_name`
FROM `information_schema`.`statistics`
WHERE `table_schema` = DATABASE()
  AND `table_name` = 'user'
ORDER BY `index_name`, `seq_in_index`;

SELECT `id`, `role`
FROM `user`
WHERE `role` NOT IN ('user', 'admin') OR `role` IS NULL;

SELECT
    `column_name`,
    `column_type`,
    `is_nullable`,
    `column_default`
FROM `information_schema`.`columns`
WHERE `table_schema` = DATABASE()
  AND `table_name` = 'user'
  AND `column_name` IN ('status', 'auth_version')
ORDER BY `column_name`;
