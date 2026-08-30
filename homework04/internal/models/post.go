package models

import "gorm.io/gorm"

type Post struct {
	gorm.Model
	// Title 和 Content 是文章的主要内容，标题限制为 200 个字符
	Title   string `gorm:"not null;size:200" json:"title"`
	Content string `gorm:"not null" json:"content"`
	// UserID 是文章作者的外键；Author 用于查询作者信息
	UserID uint `gorm:"not null;index" json:"user_id"`
	Author User `gorm:"foreignKey:UserID" json:"author,omitempty"`
	// 详情接口可以预加载该文章的全部评论
	Comments []Comment `json:"comments,omitempty"`
}
