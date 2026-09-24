create table blog_article
(
    id            bigint unsigned auto_increment primary key comment '文章ID',
    title         varchar(200)                              not null comment '文章标题',
    summary       varchar(500)    default ''                not null comment '文章摘要',
    content       longtext                                  not null comment '文章正文内容',
    cover_image   varchar(500)    default ''                not null comment '封面图片URL',
    author_id     bigint unsigned                           not null comment '作者ID',
    category_id   bigint unsigned default 0                 not null comment '分类ID，0表示未分类',
    is_top        tinyint(1)      default 0                 not null comment '是否置顶：0否 1是',
    view_count    bigint unsigned default 0                 not null comment '浏览量',
    like_count    bigint unsigned default 0                 not null comment '点赞数',
    favor_count   bigint unsigned default 0                 not null comment '收藏数',
    comment_count bigint unsigned default 0                 not null comment '评论数',
    published_at  datetime                                  null comment '发布时间',
    is_published  tinyint(1)      default 1                 not null comment '是否发布：0草稿 1已发布',
    visibility    int             default 1                 not null comment '可见性：1公开 2私密 3仅粉丝',
    created_at    datetime        default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at    datetime        default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    deleted_at    datetime                                  null comment '删除时间（软删除）'
) comment '博客文章表';

create index idx_category_id on blog_article (category_id);
create index idx_user_id on blog_article (author_id);
