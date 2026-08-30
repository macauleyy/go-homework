package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	// Username 和 Email 建立唯一索引，防止重复注册
	Username string `gorm:"uniqueIndex;not null;size:64" json:"username"`
	// json:"-" 防止密码哈希被接口响应返回
	Password string `gorm:"not null;size:255" json:"-"`
	Email    string `gorm:"uniqueIndex;not null;size:255" json:"email"`
	// Posts 和 Comments 表示用户拥有的文章及发表的评论
	Posts    []Post    `json:"-"`
	Comments []Comment `json:"-"`
}
