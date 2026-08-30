package main

import (
	"fmt"

	"gorm.io/gorm"
)

// CommentStatus 表示文章当前是否存在评论
type CommentStatus string

const (
	// HasComments 表示文章至少有一条评论
	HasComments CommentStatus = "有评论"
	// NoComments 表示文章没有评论
	NoComments CommentStatus = "无评论"
)

// User 用户模型，一个用户可以发布多篇文章
type User struct {
	ID          int `gorm:"primaryKey"`
	Name        string
	Posts       []Post `gorm:"constraint:OnDelete:CASCADE;"`
	PostCounter int
}

// Post 文章模型，维护作者、评论和文章评论状态
type Post struct {
	ID            int `gorm:"primaryKey"`
	UserID        int
	Comments      []Comment `gorm:"constraint:OnDelete:CASCADE;"`
	CommentStatus CommentStatus
	Body          string
}

// Comment 评论模型，关联所属文章
type Comment struct {
	ID     int `gorm:"primaryKey"`
	PostID int
	Body   string
}

// AfterCreate 创建文章后，原子地增加作者的文章计数
func (p *Post) AfterCreate(tx *gorm.DB) error {
	if p.UserID == 0 {
		return fmt.Errorf("after create post: UserID is empty")
	}
	result := tx.Model(&User{}).
		Where("id = ?", p.UserID).
		Update("post_counter", gorm.Expr("post_counter+?", 1))
	if result.Error != nil {
		return fmt.Errorf("increment user %d post_counter: %w", p.UserID, result.Error)
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("increment user %d post_counter: user not found", p.UserID)
	}
	return nil
}

// BeforeDelete 确保通过主键删除文章时也能取得作者 ID
func (p *Post) BeforeDelete(tx *gorm.DB) error {
	if p.UserID != 0 || p.ID == 0 {
		return nil
	}
	return tx.Model(&Post{}).Select("user_id").First(p, p.ID).Error
}

// AfterDelete 删除文章后同步减少作者的文章计数
func (p *Post) AfterDelete(tx *gorm.DB) error {
	if p.UserID == 0 {
		return fmt.Errorf("after delete post: UserID is empty")
	}
	result := tx.Model(&User{}).
		Where("id = ?", p.UserID).
		Update("post_counter", gorm.Expr("CASE WHEN post_counter > 0 THEN post_counter - 1 ELSE 0 END"))
	if result.Error != nil {
		return fmt.Errorf("decrement user %d post_counter: %w", p.UserID, result.Error)
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("decrement user %d post_counter: user not found", p.UserID)
	}
	return nil
}

// AfterCreate 创建评论后，将所属文章标记为“有评论”
func (c *Comment) AfterCreate(tx *gorm.DB) error {
	if c.PostID == 0 {
		return fmt.Errorf("after create comment: PostID is empty")
	}
	result := tx.Table("posts").
		Where("id = ?", c.PostID).
		Update("comment_status", HasComments)
	if result.Error != nil {
		return fmt.Errorf("update post %d comment status: %w", c.PostID, result.Error)
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("update post %d comment status: post not found", c.PostID)
	}
	return nil
}

// BeforeDelete 删除评论前补齐 PostID，使按主键删除也能正确更新文章状态
func (c *Comment) BeforeDelete(tx *gorm.DB) error {
	if c.PostID != 0 || c.ID == 0 {
		return nil
	}
	return tx.Model(&Comment{}).Select("post_id").First(c, c.ID).Error
}

// AfterDelete 根据剩余评论数量同步文章的评论状态
func (c *Comment) AfterDelete(tx *gorm.DB) error {
	if c.PostID == 0 {
		return fmt.Errorf("after delete comment: PostID is empty")
	}
	var count int64
	if err := tx.Model(&Comment{}).Where("post_id = ?", c.PostID).Count(&count).Error; err != nil {
		return fmt.Errorf("count comments of post %d: %w", c.PostID, err)
	}
	status := HasComments
	if count == 0 {
		status = NoComments
	}
	result := tx.Table("posts").Where("id = ?", c.PostID).Update("comment_status", status)
	if result.Error != nil {
		return fmt.Errorf("update post %d comment status: %w", c.PostID, result.Error)
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("update post %d comment status: post not found", c.PostID)
	}
	return nil
}
