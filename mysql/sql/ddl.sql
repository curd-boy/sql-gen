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
    id        int          NOT NULL AUTO_INCREMENT,
    user_id   int          NOT NULL DEFAULT 0  comment '用户id',
    email     varchar(255) NOT NULL DEFAULT '' comment '邮箱',
    address   varchar(255) NOT NULL DEFAULT '' comment '地址'
) ENGINE = InnoDB COMMENT = '信息表';
-- name: GetUser :one 查询用户
-- params:
-- result:
select *
from users
where id = ?;
-- name: GetUsersInfo 查询用户信息
-- params:
-- result:
select t1.* ,t2.email email_address
from users t1
         left join info t2 on t1.id = t2.user_id
where t1.id = ?;

-- name: UpdateUserAge 更新用户
update users set age = ? where id = ? ;

-- name: DeleteInfo 删除信息
delete from info where id = ? ;

-- name: UpdateInfo 更新信息
update info set address = ? where id = ? ;

-- name: DeleteUser 删除用户
delete from users where id = ? ;

-- name: GetAddress 查询地址
select * from info where user_id = ?;

-- name: AddUsers 添加用户
insert into users (id, name, gender, age, update_time, create_time) VALUES (?,?,?,?,?,?),(?,?,?,?,?,?);