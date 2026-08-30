package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"homework04/internal/database"
	"homework04/internal/handlers"
	"homework04/internal/middleware"
)

func setupRouter() *gin.Engine {
	// 使用 Gin 的日志和恢复中间件，确保请求可追踪且 panic 不会直接终止进程
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	// 注册和登录不需要认证；文章读取和评论读取允许公开访问
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)
	r.GET("/posts", handlers.GetPosts)
	r.GET("/posts/:id", handlers.GetPost)
	r.GET("/posts/:id/comments", handlers.GetComments)
	// 写操作统一放在认证路由组中，由 JWT 中间件验证当前用户身份
	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware())
	auth.POST("/posts", handlers.CreatePost)
	auth.PUT("/posts/:id", handlers.UpdatePost)
	auth.PATCH("/posts/:id", handlers.UpdatePost)
	auth.DELETE("/posts/:id", handlers.DeletePost)
	auth.POST("/posts/:id/comments", handlers.CreateComment)
	return r
}

func main() {
	// 启动时连接数据库并自动创建或更新数据表
	if err := database.Init(); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	addr := os.Getenv("BLOG_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	// BLOG_ADDR 可用于在本地测试时切换监听端口
	server := &http.Server{Addr: addr, Handler: setupRouter()}
	log.Printf("blog server listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to start server: %v", err)
	}
}
