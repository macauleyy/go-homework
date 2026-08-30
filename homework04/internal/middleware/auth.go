package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// jwtSecret 用于签名和验证 JWT，生产环境应通过 JWT_SECRET 配置
var jwtSecret = []byte(getSecret())

// getSecret 返回环境变量中的密钥，未配置时使用开发环境默认值
func getSecret() string {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}
	return "change-this-development-secret"
}

type Claims struct {
	// Claims 保存 JWT 中的用户标识和标准过期时间
	ID       uint   `json:"id"`
	Username string `json:"username"`
	jwt.StandardClaims
}

// GenerateToken 为登录成功的用户生成 24 小时有效的 JWT
func GenerateToken(id uint, username string) (string, error) {
	claims := Claims{ID: id, Username: username, StandardClaims: jwt.StandardClaims{ExpiresAt: time.Now().Add(24 * time.Hour).Unix()}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}

func AuthMiddleware() gin.HandlerFunc {
	// AuthMiddleware 校验 Authorization: Bearer <token> 请求头
	return func(c *gin.Context) {
		parts := strings.Fields(strings.TrimSpace(c.GetHeader("Authorization")))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "a Bearer token is required"})
			return
		}
		claims, err := ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		// 将解析出的用户信息放入上下文，供文章和评论处理器使用
		c.Set("userID", claims.ID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func ParseToken(tokenString string) (*Claims, error) {
	// ParseToken 限制使用 HS256 算法，并检查签名、有效期和用户 ID
	if tokenString == "" {
		return nil, errors.New("empty token")
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid || claims.ID == 0 {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

func CurrentUserID(c *gin.Context) (uint, bool) {
	// CurrentUserID 从 Gin 上下文中安全读取认证用户 ID
	value, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	id, ok := value.(uint)
	return id, ok && id != 0
}
