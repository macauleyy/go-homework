package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"homework04/internal/models"
	"os"
	"path/filepath"
)

// DB 是应用程序共享的 GORM 数据库连接
var DB *gorm.DB

// Init 初始化 SQLite 连接，并根据模型自动迁移数据表结构
func Init() error {
	// 允许通过环境变量切换数据库文件，默认使用项目下的 data/blog.db
	path := os.Getenv("BLOG_DB_PATH")
	if path == "" {
		path = "data/blog.db"
	}
	if dir := filepath.Dir(path); dir != "." {
		// 数据库目录不存在时自动创建，避免首次启动失败
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	// 开启 SQLite 外键约束，保证文章、评论与用户之间的关联有效
	db, err := gorm.Open(sqlite.Open(path+"?_foreign_keys=on"), &gorm.Config{})
	if err != nil {
		return err
	}
	if err := db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{}); err != nil {
		return err
	}
	DB = db
	return nil
}
