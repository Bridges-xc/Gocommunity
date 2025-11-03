package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// ============================= 1. 中间件定义 ====================
// 1.1 全局中间件 - 请求计时器
func TimeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next() // 执行后续中间件和处理器
		duration := time.Since(start)
		log.Printf("请求 %s %s 用时: %v", c.Request.Method, c.Request.URL.Path, duration)
	}
}

// 1.2 全局中间件 - 跨域处理
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// 1.3 路由组中间件 - 认证检查
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
			c.Abort()
			return
		}
		// 这里可以添加token验证逻辑
		c.Set("user_id", "123") // 模拟设置用户ID
		c.Next()
	}
}

// 1.4 路由组中间件 - API版本检查
func VersionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("api_version", "v1")
		c.Next()
	}
}

// ============================= 2. 处理器函数 ====================
func HelloHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello Gin!",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func LoginHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// 简单的登录验证
	if username == "admin" && password == "123456" {
		session := sessions.Default(c)
		session.Set("username", username)
		session.Save()

		c.JSON(http.StatusOK, gin.H{
			"message": "登录成功",
			"user":    username,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户名或密码错误",
		})
	}
}

func ProfileHandler(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")

	c.JSON(http.StatusOK, gin.H{
		"username": username,
		"user_id":  c.MustGet("user_id"),
		"version":  c.MustGet("api_version"),
	})
}

func UpdateHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"user_id": c.MustGet("user_id"),
	})
}

func DeleteHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// ============================= 3. 404和405处理 ====================
func Handle404(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"error":   "页面不存在",
		"path":    c.Request.URL.Path,
		"method":  c.Request.Method,
		"message": "请检查请求路径和方法是否正确",
	})
}

func Handle405(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, gin.H{
		"error":  "方法不允许",
		"path":   c.Request.URL.Path,
		"method": c.Request.Method,
	})
}

// ============================= 4. 主函数和路由配置 ====================
func main() {
	// 4.1 创建Gin引擎
	router := gin.New()

	// 4.2 启用405方法不允许处理
	router.HandleMethodNotAllowed = true

	// 4.3 配置会话存储
	store := cookie.NewStore([]byte("secret-key"))
	router.Use(sessions.Sessions("mysession", store))

	// 4.4 注册全局中间件
	router.Use(gin.Recovery())   // 恢复panic
	router.Use(CorsMiddleware()) // 跨域中间件
	router.Use(TimeMiddleware()) // 计时中间件

	// 4.5 配置静态文件服务
	router.Static("/static", "./static")
	router.StaticFile("/favicon.ico", "./static/favicon.ico")

	// ============================= 5. 路由分组管理 ====================

	// 5.1 公开路由组 - 不需要认证
	public := router.Group("/api")
	{
		public.GET("/hello", HelloHandler)
		public.POST("/login", LoginHandler)
		public.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusFound, "/api/hello")
		})
	}

	// 5.2 受保护路由组 - 需要认证
	protected := router.Group("/api")
	protected.Use(AuthMiddleware(), VersionMiddleware())
	{
		protected.GET("/profile", ProfileHandler)
		protected.POST("/update", UpdateHandler)
		protected.DELETE("/delete", DeleteHandler)
	}

	// 5.3 管理路由组 - 嵌套分组示例
	admin := protected.Group("/admin")
	{
		admin.GET("/users", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "管理员用户列表"})
		})
		admin.POST("/users", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "创建用户"})
		})
	}

	// 5.4 注册404和405处理器
	router.NoRoute(Handle404)
	router.NoMethod(Handle405)

	// ============================= 6. 自定义日志配置 ====================

	// 6.1 创建日志文件
	logFile, err := os.Create("gin.log")
	if err != nil {
		log.Fatal("创建日志文件失败:", err)
	}

	// 6.2 配置日志输出到文件和控制台
	gin.DefaultWriter = logFile

	// 6.3 自定义路由调试日志
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Printf("🚀 注册路由: %-6s %-25s --> %s (%d handlers)\n",
			httpMethod, absolutePath, handlerName, nuHandlers)
	}

	// ============================= 7. 服务器配置和启动 ====================

	server := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	fmt.Println("🎯 Gin Web服务器启动成功!")
	fmt.Println("📍 访问地址: http://localhost:8080")
	fmt.Println("")
	fmt.Println("📚 可用路由列表:")
	fmt.Println("  公开路由:")
	fmt.Println("    GET  /api/hello")
	fmt.Println("    POST /api/login")
	fmt.Println("    GET  /api/")
	fmt.Println("  受保护路由 (需要Authorization头):")
	fmt.Println("    GET  /api/profile")
	fmt.Println("    POST /api/update")
	fmt.Println("    DELETE /api/delete")
	fmt.Println("    GET  /api/admin/users")
	fmt.Println("    POST /api/admin/users")
	fmt.Println("  静态文件:")
	fmt.Println("    GET  /static/*filepath")
	fmt.Println("    GET  /favicon.ico")
	fmt.Println("")
	fmt.Println("💡 测试提示:")
	fmt.Println("  - 登录: POST /api/login (username=admin, password=123456)")
	fmt.Println("  - 查看个人信息: GET /api/profile (需要设置Authorization头)")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ 服务器启动失败: %v", err)
	}
}

// ============================= 总结知识点 ====================
/*
1. 路由管理:
   - 路由分组: Group() 方法创建逻辑相关的路由组
   - 嵌套分组: 支持多级路由分组
   - 404处理: NoRoute() 处理不存在的路由
   - 405处理: NoMethod() + HandleMethodNotAllowed = true

2. 中间件系统:
   - 全局中间件: Use() 注册，所有请求都会经过
   - 路由组中间件: 在Group()中注册，组内路由使用
   - 单路由中间件: 在具体路由中注册
   - 执行顺序: 按照注册顺序执行，c.Next()控制流程

3. 会话控制:
   - Session中间件: gin-contrib/sessions
   - Cookie存储: cookie.NewStore()
   - 分布式存储: 支持Redis等(需额外配置)

4. 静态文件服务:
   - Static(): 静态文件夹映射
   - StaticFile(): 单个静态文件映射
   - StaticFS(): 自定义文件系统映射

5. 服务配置:
   - 自定义Server: 配置超时、头部大小等
   - 生产环境: 建议配置合理的超时时间

6. 日志管理:
   - 文件日志: 将日志输出到文件
   - 自定义格式: LoggerWithFormatter
   - 路由调试: DebugPrintRouteFunc 自定义路由注册日志

7. 跨域处理:
   - CORS中间件: 处理跨域请求
   - OPTIONS预检: 自动处理OPTIONS请求

8. 最佳实践:
   - 使用路由组组织相关功能
   - 中间件按功能划分(认证、日志、跨域等)
   - 生产环境关闭控制台颜色
   - 合理配置静态文件服务路径
   - 使用结构化的错误响应

9. 重要提醒:
   - 中间件数量不要超过63个(abortIndex限制)
   - 异步处理要使用c.Copy()副本
   - 文件上传要配置MaxMultipartMemory
   - Session密钥要足够复杂
*/
