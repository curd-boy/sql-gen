# sql-gen
mysql parse

### 使用
    
  ```bash
      go build -o gen  cmd/main.go 
      ./gen help
  ``` 

### 示例

- sql注释风格

  ```mysql
  -- name: GetUser :one/:many
  -- params:  -- 由sql语句反推生成到函数中,直接指定为条件扩展sql
  -- result: id,last_name -- sql反推,指定则定义相应结构体GetUserRes
  select * from users where id = ? ;
  -- DDL_Defind
  CREATE TABLE users
  (
      id         integer       NOT NULL AUTO_INCREMENT PRIMARY KEY comment '',
      first_name varchar(255)  NOT NULL default '' comment '名',
      last_name  varchar(255)  not null default '' comment '姓',
      age        integer       NOT NULL default 0 comment '年龄',
      job_status ENUM ('APPLIED', 'ACCEPTED') NOT NULL default 'APPLIED' comment '职业状态'
  ) ENGINE = InnoDB;
  
  ```
  
- 生成代码如下

   ```go
  // UsersQuery 一张表对应一个查询类型
  type UsersQuery struct {
  	db *sql.Tx
  }
  
  func NewUsersQuery(db *sql.Tx) *UsersQuery {
  	return &UsersQuery{
  		db: db,
  	}
  }
  // JobStatus 枚举类型
  type JobStatus string
  
  const (
  	JobStatusAPPLIED  = "APPLIED"
  	JobStatusACCEPTED = "ACCEPTED"
  )
  // Users 表ddl语句生成,注释自动填充
  type Users struct {
  	Id        int       `json:"id"`         // '',
  	FirstName string    `json:"first_name"` // '名',
  	LastName  string    `json:"last_name"`  // '姓',
  	Age       int       `json:"age"`        // '年龄',
  	JobStatus JobStatus `json:"job_status"` // '职业状态'
  }
  // GetUserRes 由result注释生成的返回结果
  type GetUserRes struct {
  	Id       int
  	LastName string
  }
  // sqlGetUser 去掉注释后的查询语句
  const sqlGetUser = `select * from users where id = ?`
  
  // GetUser 根据查询语句生成的函数,参数由?占位符反推, 自定义参数扩展暂不支持,返回结果根据result注释生成
  func (q *UsersQuery) GetUser(ctx context.Context, id int) ([]GetUserRes, error) {
  	rows, err := q.db.QueryContext(ctx, sqlGetUser, id)
  	if err != nil {
  		return nil, err
  	}
  	defer rows.Close()
  	var items []GetUserRes
  	for rows.Next() {
  		var i GetUserRes
  		if err := rows.Scan(
  			&i.Id,
  			&i.LastName,
  		); err != nil {
  			return nil, err
  		}
  		items = append(items, i)
  	}
  	if err := rows.Close(); err != nil {
  		return nil, err
  	}
  	if err := rows.Err(); err != nil {
  		return nil, err
  	}
  	return items, nil
  }
   ```