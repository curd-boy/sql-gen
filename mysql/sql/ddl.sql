-- sql 字段顺序 必须先定义not null
CREATE TABLE users
(
    id          int            NOT NULL AUTO_INCREMENT,
    name        varchar(255)   NOT NULL DEFAULT '' comment '名字',
    gender      enum ('F','M') NOT NULL DEFAULT 'M' comment '性别',
    age         int            NOT NULL DEFAULT 0 COMMENT '年龄',
    update_time datetime(0)    NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0) COMMENT '更新时间',
    create_time datetime(0)    NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '创建时间'
) ENGINE = InnoDB COMMENT = '用户表';
CREATE TABLE info
(
    id  int          NOT NULL AUTO_INCREMENT,
    age varchar(255) NOT NULL DEFAULT '' comment '名字'
) ENGINE = InnoDB COMMENT = '信息表';
-- name: GetUser :one
-- params:
-- result:
select *
from users
where id = ?;
-- name: GetUsersInfo :one
-- params:
-- result:
select t1.* ,t2.age info_age
from users t1
         left join info t2 on t1.age = t2.age
where t1.id = ?;