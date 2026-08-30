package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"homework04/internal/database"
	"homework04/internal/middleware"
	"homework04/internal/models"
)

type credentials struct {
	// credentials 同时用于注册和登录请求体
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=72"`
	Email    string `json:"email"`
}

// Register 校验注册信息、哈希密码并创建用户
func Register(c *gin.Context) {
	var input credentials
	err := c.ShouldBindJSON(&input)
	input.Username, input.Email = strings.TrimSpace(input.Username), strings.TrimSpace(input.Email)
	if err != nil || input.Username == "" || input.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username, password and email are required"})
		return
	}
	// 只保存 bcrypt 哈希，绝不保存明文密码
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}
	user := models.User{Username: input.Username, Email: input.Email, Password: string(hash)}
	if err := database.DB.Create(&user).Error; err != nil {
		if isConflict(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "username or email already exists"})
			return
		}
		log.Printf("create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": user})
}

func Login(c *gin.Context) {
	// 登录时查询用户并比较密码哈希，成功后签发 JWT
	var input credentials
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}
	var user models.User
	if err := database.DB.Where("username = ?", strings.TrimSpace(input.Username)).First(&user).Error; err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}
	token, err := middleware.GenerateToken(user.ID, user.Username)
	if err != nil {
		log.Printf("sign token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func isConflict(err error) bool {
	// SQLite 和 GORM 对唯一约束错误的返回形式可能不同，统一转换为冲突响应
	return errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(strings.ToLower(err.Error()), "unique constraint")
}
