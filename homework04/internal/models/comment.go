package models

import "gorm.io/gorm"

type Comment struct {
	gorm.Model
	// Content 保存评论正文
	Content string `gorm:"not null" json:"content"`
	// UserID 和 PostID 分别关联评论作者与被评论的文章
	UserID uint `gorm:"not null;index" json:"user_id"`
	PostID uint `gorm:"not null;index" json:"post_id"`
	Author User `gorm:"foreignKey:UserID" json:"author,omitempty"`
}
