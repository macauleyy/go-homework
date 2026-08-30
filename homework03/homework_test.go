package main

import (
	"fmt"
	"homework03/testutil"
	"os"
	"path/filepath"
	"testing"

	"gorm.io/gorm"
)

func setupDB(t *testing.T) *gorm.DB {
	t.Helper()

	// 创建测试数据库连接
	db := testutil.NewTestDB(t, "test.db")

	// 题目1：AutoMigrate 自动创建 User/Post/Comment 对应的数据库表
	if err := db.AutoMigrate(&User{}, &Post{}, &Comment{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	// 重置测试数据，确保每次测试都从干净的状态开始
	if err := resetData(t, db); err != nil {
		t.Fatalf("reset data: %v", err)
	}

	// 注册清理函数，测试结束后自动清理（类似 Java 的 @After）
	t.Cleanup(func() {
		// SQLite 测试会在 db/ 目录生成库文件，测试结束后删除，避免残留
		if db.Dialector.Name() == "sqlite" {
			_ = os.Remove(filepath.Join("db", "test_sqlite.db"))
		}
	})
	return db
}

// 题目3：验证 AfterDelete 钩子 —— 删除最后一条评论后，文章评论状态更新为"无评论"
func TestAfterDelete(t *testing.T) {
	db := setupDB(t)

	// 文章1 初始有3条评论，先删2条，状态应保持"有评论"
	for _, id := range []int{1, 2} {
		var c Comment
		if err := db.First(&c, id).Error; err != nil {
			t.Fatalf("first comment %d: %v", id, err)
		}
		if err := db.Delete(&c).Error; err != nil {
			t.Fatalf("delete comment %d: %v", id, err)
		}
	}
	var p Post
	if err := db.First(&p, 1).Error; err != nil {
		t.Fatalf("first post: %v", err)
	}
	if p.CommentStatus != "有评论" {
		t.Fatalf("还有评论时状态应保持 有评论, 实际 %q", p.CommentStatus)
	}

	// 删最后一条，状态应变"无评论"
	var last Comment
	if err := db.First(&last, 3).Error; err != nil {
		t.Fatalf("first comment 3: %v", err)
	}
	if err := db.Delete(&last).Error; err != nil {
		t.Fatalf("delete comment 3: %v", err)
	}
	if err := db.First(&p, 1).Error; err != nil {
		t.Fatalf("first post: %v", err)
	}
	if p.CommentStatus != "无评论" {
		t.Fatalf("删除全部评论后状态应为 无评论, 实际 %q", p.CommentStatus)
	}
}

// 题目2：关联查询 —— 查询用户发布的所有文章及其评论
func TestQueryPostWithComment(t *testing.T) {
	db := setupDB(t)
	userID := 1

	type PostWithComment struct {
		UserID      int
		Name        string
		PostID      int
		CommentBody string
	}

	var results []PostWithComment

	if err := db.Table("users").
		Select(`
			users.id AS user_id,
			users.name AS name,
			posts.id AS post_id,
			comments.body AS comment_body
		`).
		Joins("LEFT JOIN posts ON users.id = posts.user_id").
		Joins("LEFT JOIN comments ON posts.id = comments.post_id").
		Where("users.id = ?", userID).
		Scan(&results).Error; err != nil {
		t.Fatalf("query posts with comments: %v", err)
	}

	for _, r := range results {
		t.Logf("用户ID:%d 用户名:%s 文章ID:%d 评论:%s",
			r.UserID, r.Name, r.PostID, r.CommentBody)
	}

	// 用户1有4篇文章（ID 1~4），其中文章1有3条评论，其余文章无评论
	// LEFT JOIN 会为每条评论产生一行（文章1的3条），并为无评论的文章产生 NULL 行（3篇），共 6 行
	if len(results) != 6 {
		t.Fatalf("查询结果行数期望 6, 实际 %d", len(results))
	}

	// 所有行都应属于用户1（名字与种子数据一致）
	for _, r := range results {
		if r.UserID != userID || r.Name != "Alice" {
			t.Fatalf("查询结果包含非用户1的行: %+v", r)
		}
	}

	// 应包含文章1的第一条评论（Body 为 "test"）
	found := false
	for _, r := range results {
		if r.PostID == 1 && r.CommentBody == "test" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("查询结果中未找到文章1的评论 test")
	}
}

// 题目2：关联查询 —— 查询评论数量最多的文章
func TestQueryMostCommentedPost(t *testing.T) {
	db := setupDB(t)

	type PostCommentCount struct {
		Post
		CommentCount int
	}

	var result PostCommentCount

	if err := db.Table("posts").
		Select("posts.*, COUNT(comments.id) AS comment_count").
		Joins("LEFT JOIN comments ON posts.id = comments.post_id").
		Group("posts.id").
		Order("comment_count DESC").
		Limit(1).
		Scan(&result).Error; err != nil {
		t.Fatalf("query most commented post: %v", err)
	}

	t.Logf("评论最多的文章ID:%d 评论数量:%d", result.ID, result.CommentCount)

	// 种子数据中只有文章1有评论（3条），应被选为评论最多的文章
	if result.ID != 1 || result.CommentCount != 3 {
		t.Fatalf("期望文章1评论数3, 实际 文章ID:%d 评论数:%d", result.ID, result.CommentCount)
	}
}

// 题目2：关联预加载 —— 使用 Preload 查询用户的所有文章及其评论
func TestPreloadUserPostsWithComments(t *testing.T) {
	db := setupDB(t)

	// Preload 嵌套预加载：先加载用户，再加载其文章及每篇文章的评论
	var user User
	if err := db.Preload("Posts").Preload("Posts.Comments").First(&user, 1).Error; err != nil {
		t.Fatalf("preload user posts comments: %v", err)
	}

	t.Logf("用户ID:%d 用户名:%s 文章数:%d", user.ID, user.Name, len(user.Posts))
	for _, p := range user.Posts {
		t.Logf(" 文章ID:%d 评论数:%d", p.ID, len(p.Comments))
		for _, c := range p.Comments {
			t.Logf("  - 评论ID:%d 内容:%s", c.ID, c.Body)
		}
	}

	// 断言：用户1有4篇文章（ID 1~4）
	if len(user.Posts) != 4 {
		t.Fatalf("用户1文章数期望 4, 实际 %d", len(user.Posts))
	}

	// 断言：文章1有3条评论
	var post1 *Post
	for i := range user.Posts {
		if user.Posts[i].ID == 1 {
			post1 = &user.Posts[i]
			break
		}
	}
	if post1 == nil {
		t.Fatalf("未找到文章1")
	}
	if len(post1.Comments) != 3 {
		t.Fatalf("文章1评论数期望 3, 实际 %d", len(post1.Comments))
	}
}

// 题目2：关联预加载 —— 使用 Association 统计评论数量最多的文章
func TestAssociationMostCommentedPost(t *testing.T) {
	db := setupDB(t)

	var posts []Post
	if err := db.Find(&posts).Error; err != nil {
		t.Fatalf("find posts: %v", err)
	}

	// 遍历文章，用 Association 统计每篇的评论数，找出最多的一篇
	var most Post
	var maxCount int64
	for i := range posts {
		// Association 的 Count 只返回数量，错误记录在 assoc.Error 中
		assoc := db.Model(&posts[i]).Association("Comments")
		count := assoc.Count()
		if assoc.Error != nil {
			t.Fatalf("count comments of post %d: %v", posts[i].ID, assoc.Error)
		}
		t.Logf("文章ID:%d 评论数:%d", posts[i].ID, count)
		if count > maxCount {
			maxCount = count
			most = posts[i]
		}
	}

	t.Logf("评论最多的文章ID:%d 评论数量:%d", most.ID, maxCount)

	// 断言：文章1评论数最多（3条）
	if most.ID != 1 || maxCount != 3 {
		t.Fatalf("期望文章1评论数3, 实际 文章ID:%d 评论数:%d", most.ID, maxCount)
	}
}

// 题目3：验证 AfterCreate 钩子 —— 创建文章后用户 post_counter 自增
func TestAfterCreate(t *testing.T) {
	db := setupDB(t)
	post := Post{UserID: 2}
	if err := db.Create(&post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}

	var user User
	if err := db.First(&user, 2).Error; err != nil {
		t.Fatalf("get user 2: %v", err)
	}
	if user.PostCounter != 1 {
		t.Fatalf("AfterCreate 未更新 post_counter: 期望 1, 实际 %d", user.PostCounter)
	}
}

// 题目3：验证评论 AfterCreate 钩子 —— 删光评论变"无评论"后，再新增评论状态恢复"有评论"
func TestAfterCreateComment(t *testing.T) {
	db := setupDB(t)

	// 先删光文章1的所有评论，使状态变为"无评论"
	for _, id := range []int{1, 2, 3} {
		var c Comment
		if err := db.First(&c, id).Error; err != nil {
			t.Fatalf("first comment %d: %v", id, err)
		}
		if err := db.Delete(&c).Error; err != nil {
			t.Fatalf("delete comment %d: %v", id, err)
		}
	}
	var p Post
	if err := db.First(&p, 1).Error; err != nil {
		t.Fatalf("first post: %v", err)
	}
	if p.CommentStatus != "无评论" {
		t.Fatalf("删光评论后状态应为 无评论, 实际 %q", p.CommentStatus)
	}

	// 再新增一条评论，状态应恢复为"有评论"
	if err := db.Create(&Comment{PostID: 1, Body: "new"}).Error; err != nil {
		t.Fatalf("create comment: %v", err)
	}
	if err := db.First(&p, 1).Error; err != nil {
		t.Fatalf("first post: %v", err)
	}
	if p.CommentStatus != "有评论" {
		t.Fatalf("新增评论后状态应为 有评论, 实际 %q", p.CommentStatus)
	}
}

// 验证按 ID 删除时，钩子可以自行补齐关联字段并维护统计信息
func TestDeleteWithOnlyID(t *testing.T) {
	db := setupDB(t)

	post := Post{UserID: 2}
	if err := db.Create(&post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}
	if err := db.Delete(&Post{ID: post.ID}).Error; err != nil {
		t.Fatalf("delete post by id: %v", err)
	}
	var user User
	if err := db.First(&user, 2).Error; err != nil {
		t.Fatalf("get user: %v", err)
	}
	if user.PostCounter != 0 {
		t.Fatalf("删除文章后 post_counter 应为 0, 实际 %d", user.PostCounter)
	}

	if err := db.Delete(&Comment{ID: 1}).Error; err != nil {
		t.Fatalf("delete comment by id: %v", err)
	}
	var postWithStatus Post
	if err := db.First(&postWithStatus, 1).Error; err != nil {
		t.Fatalf("get post: %v", err)
	}
	if postWithStatus.CommentStatus != HasComments {
		t.Fatalf("删除一条评论后状态应保持有评论, 实际 %q", postWithStatus.CommentStatus)
	}
}

// resetData 重置测试数据，确保每次测试都从干净的状态开始
// 必须先清空所有表再重新播种：
//   - 清空顺序按外键依赖的逆序（comments → posts → users），
//     否则在开启外键约束的 MySQL/Postgres 上会触发外键冲突；
//   - 播种顺序按外键依赖的正序（users → posts → comments），保证引用有效
func resetData(t *testing.T, db *gorm.DB) error {
	t.Helper()

	// 1. 清空所有表：先删子表，再删父表
	for _, table := range []string{"comments", "posts", "users"} {
		if err := db.Exec("DELETE FROM " + table).Error; err != nil {
			return fmt.Errorf("clear table %s: %w", table, err)
		}
	}

	// 2. 重置 SQLite 的 AUTOINCREMENT 序列（确保 ID 从 1 开始）
	//    该语句是 SQLite 专有，其他数据库（MySQL/Postgres）跳过；
	//    序列记录可能尚不存在，错误忽略即可
	if db.Dialector.Name() == "sqlite" {
		for _, table := range []string{"comments", "posts", "users"} {
			_ = db.Exec("DELETE FROM sqlite_sequence WHERE name='" + table + "'").Error
		}
	}

	// 3. 播种用户，显式设置 ID 为 1~3，确保每次测试都使用相同的 ID
	users := []User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Cooper"},
	}
	if err := db.Create(&users).Error; err != nil {
		return fmt.Errorf("seed users: %w", err)
	}

	// 4. 播种文章，显式设置 ID 为 1~6
	//    注意：种子数据用 SkipHooks 跳过 AfterCreate 钩子，所以 post_counter
	//    保持 0、comment_status 由显式值给定——这些反规范化字段不代表真实统计，
	//    真实数据由钩子维护（这也是各测试从确定状态起步的前提）
	posts := []Post{
		{ID: 1, UserID: 1, CommentStatus: "有评论"},
		{ID: 2, UserID: 1, CommentStatus: "无评论"},
		{ID: 3, UserID: 1, CommentStatus: "无评论"},
		{ID: 4, UserID: 1, CommentStatus: "无评论"},
		{ID: 5, UserID: 2, CommentStatus: "无评论"},
		{ID: 6, UserID: 3, CommentStatus: "无评论"},
	}
	if err := db.Session(&gorm.Session{SkipHooks: true}).Create(&posts).Error; err != nil {
		return fmt.Errorf("seed posts: %w", err)
	}

	// 5. 播种评论，显式设置 ID 为 1~3，全部属于文章1
	//    同样用 SkipHooks，避免 AfterCreate 钩子改动种子状态
	comments := []Comment{
		{ID: 1, PostID: 1, Body: "test"},
		{ID: 2, PostID: 1, Body: ""},
		{ID: 3, PostID: 1, Body: ""},
	}
	if err := db.Session(&gorm.Session{SkipHooks: true}).Create(&comments).Error; err != nil {
		return fmt.Errorf("seed comments: %w", err)
	}
	return nil
}
