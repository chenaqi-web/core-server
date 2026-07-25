-- auto-generated definition
create table blog_article
(
    id            bigint unsigned auto_increment comment '文章ID'
        primary key,
    title         varchar(200)                              not null comment '文章标题',
    summary       varchar(500)    default ''                not null comment '文章摘要',
    content       longtext                                  not null comment '文章正文（Markdown格式）',
    cover_image   varchar(500)    default ''                not null comment '封面图URL',
    author_id     bigint unsigned                           not null comment '作者用户ID',
    category_id   bigint unsigned default '0'               not null comment '分类ID（0表示未分类）',
    is_top        tinyint(1)      default 0                 not null comment '是否置顶：0-否，1-是',
    view_count    bigint unsigned default '0'               not null comment '浏览量',
    like_count    bigint unsigned default '0'               not null comment '点赞量',
    comment_count bigint unsigned default '0'               not null comment '评论量',
    created_at    datetime        default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at    datetime        default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    deleted_at    datetime                                  null comment '软删除时间（NULL表示未删除）'
)
    comment '博客文章表' collate = utf8mb4_unicode_ci;

create index idx_category_id
    on blog_article (category_id);

create index idx_created_at
    on blog_article (created_at);

create index idx_deleted_at
    on blog_article (deleted_at);

create index idx_user_id
    on blog_article (author_id);

