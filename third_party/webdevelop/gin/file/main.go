package main

import (
	"fmt"
	"github.com/gin-gonic/gin/binding"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ============================= 1. 用户结构体定义 ====================
type User struct {
	ID       int    `json:"id" form:"id" uri:"id" binding:"required,min=1"`
	Username string `json:"username" form:"username" uri:"username" binding:"required"`
	Email    string `json:"email" form:"email" binding:"required,email"`
	Age      int    `json:"age" form:"age" binding:"omitempty,gte=0,lte=150"`
}

// ============================= 2. 主函数和初始化 ====================
func main() {
	// 创建Gin引擎，Default()包含Logger和Recovery中间件
	router := gin.Default()

	// 设置文件上传最大内存限制 (默认32MB)
	router.MaxMultipartMemory = 8 << 20 // 8MB

	// ============================= 3. 参数解析路由 ====================

	// 3.1 路由参数 - 命名参数
	router.GET("/user/:id/profile/:username", func(c *gin.Context) {
		id := c.Param("id")
		username := c.Param("username")
		c.JSON(200, gin.H{
			"id":       id,
			"username": username,
			"type":     "路由参数",
		})
	})

	// 3.2 路由参数 - 通配符
	router.GET("/static/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		c.JSON(200, gin.H{
			"filepath": filepath,
			"type":     "通配符参数",
		})
	})

	// 3.3 URL查询参数
	router.GET("/search", func(c *gin.Context) {
		keyword := c.Query("keyword")
		page := c.DefaultQuery("page", "1")
		limit := c.DefaultQuery("limit", "10")
		c.JSON(200, gin.H{
			"keyword": keyword,
			"page":    page,
			"limit":   limit,
			"type":    "查询参数",
		})
	})

	// 3.4 表单参数
	router.POST("/register", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		email := c.PostForm("email")
		c.JSON(200, gin.H{
			"username": username,
			"password": password,
			"email":    email,
			"type":     "表单参数",
		})
	})

	// ============================= 4. 数据绑定和验证 ====================

	// 4.1 自动绑定 (根据Content-Type自动推断)
	router.POST("/users/auto", func(c *gin.Context) {
		var user User
		if err := c.ShouldBind(&user); err != nil {
			c.JSON(400, gin.H{
				"error":   "数据绑定失败",
				"details": err.Error(),
			})
			return
		}
		c.JSON(200, gin.H{
			"message": "自动绑定成功",
			"user":    user,
		})
	})

	// 4.2 显式JSON绑定
	router.POST("/users/json", func(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"message": "JSON绑定成功",
			"user":    user,
		})
	})

	// 4.3 URI参数绑定
	router.GET("/users/:id/:username", func(c *gin.Context) {
		var user User
		if err := c.ShouldBindUri(&user); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"message": "URI绑定成功",
			"user":    user,
		})
	})

	// 4.4 多次绑定示例
	router.POST("/multiple-bind", func(c *gin.Context) {
		type FormA struct {
			FieldA string `json:"field_a" binding:"required"`
		}
		type FormB struct {
			FieldB string `json:"field_b" binding:"required"`
		}

		var formA FormA
		var formB FormB

		// 第一次绑定
		if err := c.ShouldBindBodyWith(&formA, binding.JSON); err == nil {
			c.JSON(200, gin.H{"form": "A", "data": formA})
			return
		}

		// 第二次绑定 (复用body)
		if err := c.ShouldBindBodyWith(&formB, binding.JSON); err == nil {
			c.JSON(200, gin.H{"form": "B", "data": formB})
			return
		}

		c.JSON(400, gin.H{"error": "所有绑定都失败"})
	})

	// ============================= 5. 文件操作 ====================

	// 5.1 单文件上传
	router.POST("/upload/single", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(400, gin.H{"error": "文件上传失败: " + err.Error()})
			return
		}

		// 保存文件
		filename := fmt.Sprintf("uploads/%d_%s", time.Now().Unix(), file.Filename)
		if err := c.SaveUploadedFile(file, filename); err != nil {
			c.JSON(500, gin.H{"error": "文件保存失败: " + err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"message":  "文件上传成功",
			"filename": file.Filename,
			"size":     file.Size,
			"saved_as": filename,
		})
	})

	// 5.2 多文件上传
	router.POST("/upload/multiple", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		files := form.File["files"]
		var results []gin.H

		for _, file := range files {
			filename := fmt.Sprintf("uploads/%d_%s", time.Now().Unix(), file.Filename)
			if err := c.SaveUploadedFile(file, filename); err != nil {
				c.JSON(500, gin.H{"error": "文件保存失败: " + err.Error()})
				return
			}
			results = append(results, gin.H{
				"filename": file.Filename,
				"size":     file.Size,
				"saved_as": filename,
			})
		}

		c.JSON(200, gin.H{
			"message":    "多文件上传成功",
			"file_count": len(files),
			"files":      results,
		})
	})

	// 5.3 文件下载
	router.GET("/download/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		filepath := "uploads/" + filename

		// 简单文件下载
		// c.File(filepath)

		// 带附件的下载 (客户端会提示下载)
		c.FileAttachment(filepath, filename)
	})

	// ============================= 6. 响应方法示例 ====================

	// 6.1 JSON响应 (最常用)
	router.GET("/json-response", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "JSON响应示例",
			"data":    "这是响应数据",
		})
	})

	// 6.2 字符串响应
	router.GET("/string-response", func(c *gin.Context) {
		c.String(200, "这是一个纯文本响应，当前时间: %s", time.Now().Format("2006-01-02 15:04:05"))
	})

	// 6.3 HTML响应 (需要先加载模板)
	router.LoadHTMLGlob("templates/*")
	router.GET("/html-response", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"title":   "Gin示例",
			"message": "Hello, Gin!",
		})
	})

	// 6.4 XML响应
	router.GET("/xml-response", func(c *gin.Context) {
		type Response struct {
			Status  string `xml:"status"`
			Message string `xml:"message"`
		}
		c.XML(200, Response{Status: "success", Message: "XML响应示例"})
	})

	// 6.5 重定向
	router.GET("/redirect", func(c *gin.Context) {
		c.Redirect(302, "/json-response")
	})

	// ============================= 7. 异步处理 ====================

	router.GET("/async", func(c *gin.Context) {
		// 创建Context副本用于异步处理
		ctxCopy := c.Copy()

		// 主goroutine立即返回响应
		c.String(200, "请求已接收，正在异步处理...")

		// 异步处理
		go func() {
			// 使用副本，避免竞争条件
			time.Sleep(2 * time.Second)
			log.Printf("异步处理完成: %s", ctxCopy.Request.URL.Path)
		}()
	})

	// ============================= 8. 自定义中间件 ====================

	// 自定义日志中间件
	router.Use(func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录日志
		duration := time.Since(start)
		log.Printf("请求: %s %s - 状态: %d - 耗时: %v",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration)
	})

	// ============================= 9. 启动服务器 ====================

	// 自定义服务器配置
	server := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	fmt.Println("🚀 Gin服务器启动在 http://localhost:8080")
	fmt.Println("📚 可用路由:")
	fmt.Println("  GET  /user/:id/profile/:username")
	fmt.Println("  GET  /static/*filepath")
	fmt.Println("  GET  /search?keyword=xxx&page=1")
	fmt.Println("  POST /register")
	fmt.Println("  POST /users/auto")
	fmt.Println("  POST /users/json")
	fmt.Println("  GET  /users/:id/:username")
	fmt.Println("  POST /upload/single")
	fmt.Println("  POST /upload/multiple")
	fmt.Println("  GET  /download/:filename")
	fmt.Println("  GET  /json-response")
	fmt.Println("  GET  /string-response")
	fmt.Println("  GET  /html-response")
	fmt.Println("  GET  /xml-response")
	fmt.Println("  GET  /redirect")
	fmt.Println("  GET  /async")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ 服务器启动失败: %v", err)
	}
}

// ============================= 总结知识点 ====================
/*
1. 安装和初始化:
   - go get -u github.com/gin-gonic/gin
   - gin.Default(): 带默认中间件
   - gin.New(): 纯净引擎

2. 参数解析三种方式:
   - 路由参数: c.Param() - /user/:id
   - URL参数: c.Query() - /search?q=term
   - 表单参数: c.PostForm() - form-data/x-www-form-urlencoded

3. 数据绑定:
   - ShouldBind(): 自动推断Content-Type
   - ShouldBindJSON(): 显式绑定JSON
   - ShouldBindUri(): 绑定URI参数
   - ShouldBindBodyWith(): 多次绑定

4. 数据验证:
   - 基于go-playground/validator
   - binding标签: required,email,min,max等
   - 结构体字段标签定义数据源

5. 文件操作:
   - FormFile(): 单文件上传
   - MultipartForm(): 多文件上传
   - SaveUploadedFile(): 保存文件
   - FileAttachment(): 文件下载

6. 响应方法:
   - JSON(): JSON响应
   - String(): 文本响应
   - HTML(): HTML响应
   - XML(): XML响应
   - Redirect(): 重定向

7. 异步处理:
   - c.Copy(): 创建Context副本
   - 在goroutine中使用副本避免竞争

8. 重要配置:
   - MaxMultipartMemory: 文件上传内存限制
   - 自定义服务器超时设置
   - 中间件使用

9. 最佳实践:
   - 使用结构体承载数据而非直接解析参数
   - 合理使用数据验证确保数据完整性
   - 异步处理使用Context副本
   - 生产环境配置合适的超时时间和文件大小限制
*/
