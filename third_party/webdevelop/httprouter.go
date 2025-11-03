package main

import (
	"context"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/julienschmidt/httprouter"
)

// ============================= 1. 基本路由设置 ====================
func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "<h1>Welcome!</h1>\n")
}

func Hello(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "<h1>Hello World!</h1>")
}

// ============================= 2. 命名参数路由 ====================
// :name 是命名参数，匹配单个路径段 [citation:1]
// 示例：/user/john 匹配，/user/john/profile 不匹配
func HelloWithName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("name")
	fmt.Fprintf(w, "Hello, %s!", name)
}

// ============================= 3. 文件路径参数示例 ====================
func FileInfo(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	filename := ps.ByName("filename")
	fmt.Fprintf(w, "文件名是：%s\n", filename)
}

// ============================= 4. 捕获全部参数 ====================
// *filepath 捕获剩余所有路径段，必须放在模式末尾 [citation:1]
// 示例：/files/、/files/a/b/c 都匹配
func FileServer(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	filepath := ps.ByName("filepath")
	fmt.Fprintf(w, "文件路径是：%s\n", filepath)
}

// ============================= 5. 使用标准 Handler 接口 ====================
// 演示如何将标准 http.Handler 与 httprouter 结合使用
func StandardHello(w http.ResponseWriter, r *http.Request) {
	// 从上下文获取路由参数 [citation:1]
	params := httprouter.ParamsFromContext(r.Context())
	name := params.ByName("name")
	fmt.Fprintf(w, "Standard handler says: Hello, %s!", name)
}

// 包装器函数，将标准 Handler 适配到 httprouter
func WrapHandler(h http.HandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// 将参数存入上下文
		ctx := context.WithValue(r.Context(), httprouter.ParamsKey, ps)
		h(w, r.WithContext(ctx))
	}
}

// ============================= 6. 图书 API 示例 ====================
type Book struct {
	ISDN   string `json:"isdn"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Pages  int    `json:"pages"`
}

// 模拟数据库
var bookstore = make(map[string]*Book)

// GET /books - 获取所有图书
func BookIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	books := make([]*Book, 0, len(bookstore))
	for _, book := range bookstore {
		books = append(books, book)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%+v", books) // 简化输出，实际应用可使用 json.Marshal
}

// GET /books/:isdn - 获取特定图书
func BookShow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	isdn := ps.ByName("isdn")
	book, exists := bookstore[isdn]

	w.Header().Set("Content-Type", "application/json")
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"error": "Book not found"}`)
		return
	}

	fmt.Fprintf(w, "%+v", book)
}

// ============================= 7. 静态文件服务 ====================
// 提供静态文件服务，支持正确的 MIME 类型 [citation:5]
// 更好的实现方式
func ServeStaticFiles(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	filePath := "static" + ps.ByName("filepath")

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// 根据文件扩展名设置正确的 MIME 类型
	ext := filepath.Ext(filePath)
	if mimeType := mime.TypeByExtension(ext); mimeType != "" {
		w.Header().Set("Content-Type", mimeType)
	} else {
		// 未知类型默认为二进制流
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	http.ServeFile(w, r, filePath)
}

// ============================= 8. 基本认证中间件 ====================
// BasicAuth 创建一个需要基本认证的中间件 [citation:1]
func BasicAuth(h httprouter.Handle, requiredUser, requiredPassword string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// 获取 Basic Auth 凭据
		user, password, hasAuth := r.BasicAuth()

		if hasAuth && user == requiredUser && password == requiredPassword {
			// 认证成功，调用原始处理器
			h(w, r, ps)
		} else {
			// 认证失败，要求身份验证
			w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
	}
}

func ProtectedContent(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Protected content accessed successfully!")
}

// ============================= 9. 自定义 NotFound 处理器 ====================
func CustomNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>Custom 404 - Page not found</h1>")
}

// ============================= 10. Panic 处理 ====================
func PanicDemo(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	panic("demo panic")
}

func PanicHandler(w http.ResponseWriter, r *http.Request, err interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "Recovered from panic: %v", err)
}

// ============================= 11. 全局 OPTIONS 处理器 ====================
// 用于 CORS 预检请求 [citation:2]
func GlobalOPTIONSHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Access-Control-Request-Method") != "" {
		// 设置 CORS 响应头
		header := w.Header()
		header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	}
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	router := httprouter.New()

	// ============================= 路由器配置 ====================
	// 启用自动重定向尾随斜杠 [citation:1]
	router.RedirectTrailingSlash = true
	// 启用路径自动校正
	router.RedirectFixedPath = true
	// 启用方法不允许处理
	router.HandleMethodNotAllowed = true
	// 启用 OPTIONS 请求自动处理
	router.HandleOPTIONS = true

	// ============================= 基本路由注册 ====================
	router.GET("/", Index)
	router.GET("/hello", Hello)

	// ============================= 命名参数路由 ====================
	router.GET("/hello/:name", HelloWithName)
	router.GET("/src/:filename", FileInfo)

	// ============================= 捕获全部参数路由 ====================
	router.GET("/files/*filepath", FileServer)
	router.GET("/static/*filepath", ServeStaticFiles)

	// ============================= 标准 Handler 适配 ====================
	router.GET("/std/:name", WrapHandler(StandardHello))

	// ============================= RESTful API 路由 ====================
	router.GET("/books", BookIndex)
	router.GET("/books/:isdn", BookShow)

	// ============================= 特殊处理器配置 ====================
	// 自定义 404 处理器 [citation:3]
	router.NotFound = http.HandlerFunc(CustomNotFound)
	// 全局 OPTIONS 处理器
	router.GlobalOPTIONS = http.HandlerFunc(GlobalOPTIONSHandler)
	// Panic 处理器
	router.PanicHandler = PanicHandler

	// ============================= 中间件使用示例 ====================
	user := "admin"
	pass := "secret"
	router.GET("/public", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fmt.Fprint(w, "Public content - no auth required")
	})
	router.GET("/protected", BasicAuth(ProtectedContent, user, pass))

	// ============================= Panic 演示路由 ====================
	router.GET("/panic", PanicDemo)

	// ============================= 初始化数据 ====================
	// 初始化一些示例图书数据 [citation:9]
	bookstore["123"] = &Book{
		ISDN:   "123",
		Title:  "Silence of the Lambs",
		Author: "Thomas Harris",
		Pages:  367,
	}
	bookstore["124"] = &Book{
		ISDN:   "124",
		Title:  "To Kill a Mocking Bird",
		Author: "Harper Lee",
		Pages:  320,
	}

	// ============================= 启动服务器 ====================
	fmt.Println("HttpRouter 学习服务器启动在 :8080")
	fmt.Println("可用路由:")
	fmt.Println("  GET  /")
	fmt.Println("  GET  /hello")
	fmt.Println("  GET  /hello/:name")
	fmt.Println("  GET  /src/:filename")
	fmt.Println("  GET  /files/*filepath")
	fmt.Println("  GET  /static/*filepath")
	fmt.Println("  GET  /std/:name")
	fmt.Println("  GET  /books")
	fmt.Println("  GET  /books/:isdn")
	fmt.Println("  GET  /public")
	fmt.Println("  GET  /protected (需要基本认证: admin/secret)")
	fmt.Println("  GET  /panic (演示 Panic 处理)")

	// 创建自定义服务器配置
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}

// ============================= 核心知识点总结 ====================

// 1. 路由特性与参数类型
//    - 一对一匹配原则：每个请求只能匹配一个路由
//    - 命名参数 (:param)：匹配单个路径段，如 /user/:id
//    - 通配参数 (*param)：匹配剩余路径，必须放在末尾，如 /files/*filepath
//    - 自动路径校正：处理尾部斜杠和大小写问题

// 2. 参数获取方式
//    - 在 httprouter.Handle 中：通过函数参数 ps httprouter.Params
//    - 在标准 http.Handler 中：通过上下文 httprouter.ParamsFromContext(r.Context())
//    - 使用 ByName() 方法获取特定参数值

// 3. 高级功能与中间件
//    - NotFound 处理器：自定义 404 页面
//    - PanicHandler：自动恢复 panic，防止服务崩溃
//    - GlobalOPTIONS：处理 CORS 预检请求
//    - 中间件模式：通过包装函数实现认证、日志等功能
//    - RESTful API 支持：清晰的资源路由映射
//    - 高性能：基于基数树实现，零垃圾内存分配
