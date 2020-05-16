-- sql 字段顺序 必须先定义not null
CREATE TABLE users
( id int NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL DEFAULT '' comment '名字',
  gender enum('F','M')  NOT NULL DEFAULT 'M' comment '性别',
  age int NOT NULL DEFAULT 0 COMMENT '年龄',
  update_time datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0) COMMENT '更新时间',
  create_time datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '创建时间'
) ENGINE = InnoDB COMMENT = '用户表';