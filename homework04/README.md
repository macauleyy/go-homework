# Homework 4 Blog API

一个基于 Gin、GORM 和 SQLite 的个人博客后端，包含用户认证、文章 CRUD 和评论功能。

## 运行

```bash
go run ./cmd/server
```

默认监听 `:8080`，数据库文件为 `data/blog.db`。可以通过环境变量覆盖配置：

```bash
BLOG_DB_PATH=data/blog.db JWT_SECRET='use-a-long-random-secret' BLOG_ADDR=':8080' go run ./cmd/server
```

## 接口

- `POST /register`：注册，JSON 参数 `username`、`password`、`email`
- `POST /login`：登录，返回 JWT；JSON 参数 `username`、`password`
- `GET /posts`：文章列表
- `GET /posts/:id`：文章详情（包含作者和评论）
- `POST /posts`：创建文章，需要 `Authorization: Bearer <token>`
- `PUT/PATCH /posts/:id`：作者更新文章
- `DELETE /posts/:id`：作者删除文章
- `GET /posts/:id/comments`：读取文章评论
- `POST /posts/:id/comments`：发表评论，需要 JWT

密码使用 bcrypt 哈希后存储。JWT 默认有效期为 24 小时；更新和删除文章会校验当前用户是否为作者。接口错误统一返回 JSON `{"error":"..."}`，数据库和服务错误会记录到标准日志。
