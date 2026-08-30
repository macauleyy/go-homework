package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"homework04/internal/database"
	"homework04/internal/middleware"
	"homework04/internal/models"
)

type postInput struct {
	// postInput 用于创建和更新文章，两个字段都不能为空
	Title   string `json:"title" binding:"required,max=200"`
	Content string `json:"content" binding:"required"`
}

// CreatePost 创建当前认证用户的文章
func CreatePost(c *gin.Context) {
	userID, _ := middleware.CurrentUserID(c)
	var input postInput
	if err := c.ShouldBindJSON(&input); err != nil || strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Content) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title and content are required"})
		return
	}
	post := models.Post{Title: strings.TrimSpace(input.Title), Content: input.Content, UserID: userID}
	if err := database.DB.Create(&post).Error; err != nil {
		log.Printf("create post: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create post"})
		return
	}
	database.DB.Preload("Author").First(&post, post.ID)
	c.JSON(http.StatusCreated, post)
}

func GetPosts(c *gin.Context) {
	// 按创建时间倒序返回文章列表，并预加载作者信息
	var posts []models.Post
	if err := database.DB.Preload("Author").Order("created_at DESC").Find(&posts).Error; err != nil {
		log.Printf("list posts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch posts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

func GetPost(c *gin.Context) {
	// 获取文章详情，同时加载作者和按时间排序的评论
	id, ok := pathID(c)
	if !ok {
		return
	}
	var post models.Post
	err := database.DB.Preload("Author").Preload("Comments", func(db *gorm.DB) *gorm.DB { return db.Order("created_at ASC") }).Preload("Comments.Author").First(&post, id).Error
	if err != nil {
		postLookupError(c, err)
		return
	}
	c.JSON(http.StatusOK, post)
}

func UpdatePost(c *gin.Context) {
	// 更新前必须确认文章存在，并且当前用户就是文章作者
	id, ok := pathID(c)
	if !ok {
		return
	}
	userID, _ := middleware.CurrentUserID(c)
	var post models.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		postLookupError(c, err)
		return
	}
	if post.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the author can update this post"})
		return
	}
	var input postInput
	if err := c.ShouldBindJSON(&input); err != nil || strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Content) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title and content are required"})
		return
	}
	post.Title, post.Content = strings.TrimSpace(input.Title), input.Content
	if err := database.DB.Save(&post).Error; err != nil {
		log.Printf("update post: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update post"})
		return
	}
	database.DB.Preload("Author").First(&post, post.ID)
	c.JSON(http.StatusOK, post)
}

func DeletePost(c *gin.Context) {
	// 只有文章作者可以删除文章；删除使用 GORM 软删除
	id, ok := pathID(c)
	if !ok {
		return
	}
	userID, _ := middleware.CurrentUserID(c)
	var post models.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		postLookupError(c, err)
		return
	}
	if post.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the author can delete this post"})
		return
	}
	if err := database.DB.Delete(&post).Error; err != nil {
		log.Printf("delete post: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete post"})
		return
	}
	c.Status(http.StatusNoContent)
}

func pathID(c *gin.Context) (uint, bool) {
	// 统一解析路由中的文章 ID，并将非法 ID 转换为 400 响应
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return uint(id), true
}

func postLookupError(c *gin.Context, err error) {
	// 将 GORM 的记录不存在错误映射为 404，其余错误记录日志并返回 500
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}
	log.Printf("post lookup: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch post"})
}
