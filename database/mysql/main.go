package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql" // MySQL驱动，使用_只导入不直接使用
	"github.com/jmoiron/sqlx"
)

// ============================= 1. 数据库连接配置 ====================
var db *sqlx.DB

// ============================= 2. 数据模型定义 ====================
// Person 用户结构体，使用db标签映射数据库字段
type Person struct {
	UserId   string `db:"id"`      // 用户ID，对应数据库id字段
	Username string `db:"name"`    // 用户名，对应数据库name字段
	Age      int    `db:"age"`     // 年龄，对应数据库age字段
	Address  string `db:"address"` // 地址，对应数据库address字段
}

// ============================= 3. 初始化数据库连接 ====================
func initDB() {
	// 数据源格式：用户名:密码@协议(地址:端口)/数据库名
	conn, err := sqlx.Open("mysql", "root:wyh246859@tcp(127.0.0.1:3306)/test")
	if err != nil {
		panic(fmt.Sprintf("数据库连接失败: %v", err))
	}

	// 测试连接是否成功
	if err := conn.Ping(); err != nil {
		panic(fmt.Sprintf("数据库ping失败: %v", err))
	}

	db = conn
	fmt.Println("数据库连接成功")
}

// ============================= 4. 查询操作 ====================
// 4.1 单条记录查询
func querySingle() {
	var person Person
	// Get用于查询单条记录，结果映射到结构体
	err := db.Get(&person, "SELECT id, name, age, address FROM user WHERE id = ?", "12132")
	if err != nil {
		fmt.Printf("单条查询失败: %v\n", err)
		return
	}
	fmt.Printf("单条查询成功: %+v\n", person)
}

// 4.2 多条记录查询
func queryMultiple() {
	var persons []Person
	// Select用于查询多条记录，结果映射到结构体切片
	err := db.Select(&persons, "SELECT id, name, age, address FROM user")
	if err != nil {
		fmt.Printf("多条查询失败: %v\n", err)
		return
	}
	fmt.Printf("多条查询成功，共%d条记录: %+v\n", len(persons), persons)
}

// ============================= 5. 增删改操作 ====================
// 5.1 插入数据
func insertData() {
	// Exec执行不返回结果的SQL语句
	result, err := db.Exec("INSERT INTO user VALUES (?, ?, ?, ?)",
		"120230", "李四", 12, "广州市")
	if err != nil {
		fmt.Printf("插入失败: %v\n", err)
		return
	}

	// 获取最后插入的ID
	id, err := result.LastInsertId()
	if err != nil {
		fmt.Printf("获取插入ID失败: %v\n", err)
		return
	}
	fmt.Printf("插入成功，插入ID: %d\n", id)
}

// 5.2 更新数据
func updateData() {
	result, err := db.Exec("UPDATE user SET name = ? WHERE id = ?",
		"赵六", "120230")
	if err != nil {
		fmt.Printf("更新失败: %v\n", err)
		return
	}

	// 获取受影响的行数
	affected, err := result.RowsAffected()
	if err != nil {
		fmt.Printf("获取影响行数失败: %v\n", err)
		return
	}
	fmt.Printf("更新成功，影响行数: %d\n", affected)
}

// 5.3 删除数据
func deleteData() {
	result, err := db.Exec("DELETE FROM user WHERE id = ?", "120230")
	if err != nil {
		fmt.Printf("删除失败: %v\n", err)
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		fmt.Printf("获取影响行数失败: %v\n", err)
		return
	}
	fmt.Printf("删除成功，影响行数: %d\n", affected)
}

// ============================= 6. 事务操作 ====================
func transactionDemo() {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("事务开始失败: %v\n", err)
		return
	}

	// 确保事务回滚（如果提交成功，回滚将无效）
	defer tx.Rollback()

	// 在事务中执行多个操作
	_, err = tx.Exec("INSERT INTO user VALUES (?, ?, ?, ?)",
		"99999", "事务测试", 30, "事务地址")
	if err != nil {
		fmt.Printf("事务内插入失败: %v\n", err)
		return
	}

	_, err = tx.Exec("UPDATE user SET age = ? WHERE id = ?", 99, "99999")
	if err != nil {
		fmt.Printf("事务内更新失败: %v\n", err)
		return
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		fmt.Printf("事务提交失败: %v\n", err)
		return
	}

	fmt.Println("事务执行成功")
}

func main() {
	// 初始化数据库连接
	initDB()
	defer db.Close() // 确保程序退出前关闭数据库连接

	// 执行各种数据库操作
	fmt.Println("\n=== 单条查询 ===")
	querySingle()

	fmt.Println("\n=== 多条查询 ===")
	queryMultiple()

	fmt.Println("\n=== 插入数据 ===")
	insertData()

	fmt.Println("\n=== 更新数据 ===")
	updateData()

	fmt.Println("\n=== 删除数据 ===")
	deleteData()

	fmt.Println("\n=== 事务演示 ===")
	transactionDemo()

	fmt.Println("\n=== 最终数据 ===")
	queryMultiple()
}

// ============================= 总结知识点 ====================
/*
1. 数据库驱动:
   - 使用 _ "github.com/go-sql-driver/mysql" 导入MySQL驱动
   - 驱动会自动注册到database/sql

2. 连接数据库:
   - 使用sqlx.Open("mysql", DSN)建立连接
   - DSN格式: 用户名:密码@协议(地址:端口)/数据库名
   - 使用Ping()测试连接

3. 数据映射:
   - 定义结构体，使用db标签映射数据库字段
   - Get()查询单条记录到结构体
   - Select()查询多条记录到结构体切片

4. SQL执行:
   - Exec()执行不返回结果的SQL(INSERT/UPDATE/DELETE)
   - LastInsertId()获取最后插入ID
   - RowsAffected()获取影响行数

5. 事务处理:
   - Begin()开始事务
   - Commit()提交事务
   - Rollback()回滚事务
   - 使用defer确保事务回滚，避免资源泄漏

6. 最佳实践:
   - 及时关闭数据库连接(defer db.Close())
   - 使用预处理语句防止SQL注入
   - 合理的错误处理
   - 事务中确保原子性操作
*/
