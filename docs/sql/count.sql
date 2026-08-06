-- auto-generated definition
create table interaction_count
(
    id               bigint unsigned auto_increment comment '主键ID'
        primary key,
    object_type      varchar(32)                        not null comment '对象类型',
    object_id        bigint unsigned                    not null comment '对象ID',
    interaction_type varchar(32)                        not null comment '交互类型：like, view, favor等',
    count            bigint   default 0                 not null comment '计数值',
    created_at       datetime default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at       datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    constraint uk_object
        unique (object_type, object_id, interaction_type)
)
    comment '交互计数表' collate = utf8mb4_unicode_ci;

create index idx_interaction_type
    on interaction_count (interaction_type);

create index idx_object_type
    on interaction_count (object_type);

