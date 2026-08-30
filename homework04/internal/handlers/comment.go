package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"homework04/internal/database"
	"homework04/internal/middleware"
	"homework04/internal/models"
)

type commentInput struct {
	// commentInput 是发表评论时接收的请求体
	Content string `json:"content" binding:"required"`
}

// CreateComment 为指定文章创建当前用户的评论
func CreateComment(c *gin.Context) {
	postID, ok := pathID(c)
	if !ok {
		return
	}
	userID, _ := middleware.CurrentUserID(c)
	var post models.Post
	if err := database.DB.First(&post, postID).Error; err != nil {
		postLookupError(c, err)
		return
	}
	var input commentInput
	if err := c.ShouldBindJSON(&input); err != nil || strings.TrimSpace(input.Content) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}
	comment := models.Comment{Content: strings.TrimSpace(input.Content), UserID: userID, PostID: postID}
	if err := database.DB.Create(&comment).Error; err != nil {
		log.Printf("create comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create comment"})
		return
	}
	database.DB.Preload("Author").First(&comment, comment.ID)
	c.JSON(http.StatusCreated, comment)
}

func GetComments(c *gin.Context) {
	// 读取评论前先确认文章存在，避免对不存在的文章返回空列表
	postID, ok := pathID(c)
	if !ok {
		return
	}
	var post models.Post
	if err := database.DB.First(&post, postID).Error; err != nil {
		postLookupError(c, err)
		return
	}
	var comments []models.Comment
	if err := database.DB.Where("post_id = ?", postID).Preload("Author").Order("created_at ASC").Find(&comments).Error; err != nil {
		log.Printf("list comments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch comments"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comments": comments})
}
