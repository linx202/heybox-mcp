# heybox-mcp 开发进度记录

> 更新时间：2026-03-03

---

## 总体进度

| 阶段 | 状态 | 进度 |
|------|------|------|
| Phase 1: 项目初始化 | ✅ | 100% |
| Phase 2: 基础设施层 | ✅ | 100% |
| Phase 2: 业务逻辑层 | ✅ | 100% |
| Phase 3: MCP 服务层 | ✅ | 100% |
| Phase 4: 业务功能扩展 | ✅ | 100% |
| Phase 5: 测试与优化 | 🚧 | 30% |

---

## 已完成功能

### MCP 工具 (12/13 完成)

| # | 工具名称 | 状态 | 说明 |
|---|----------|------|------|
| 1 | `check_login_status` | ✅ | 检查登录状态 |
| 2 | `get_login_qrcode` | ✅ | 获取登录二维码 |
| 3 | `delete_cookies` | ✅ | 删除 Cookie |
| 4 | `publish_content` | ✅ | 发布内容 |
| 5 | `list_feeds` | ✅ | 获取动态列表（返回完整JSON） |
| 6 | `search_feeds` | ✅ | 搜索内容（返回完整JSON) |
| 7 | `get_feed_detail` | ⚠️ | 动态详情（有验证码问题） |
| 8 | `get_user_profile` | ✅ | 获取用户信息 |
| 9 | `post_comment_to_feed` | ✅ | 发表评论 |
| 10 | `reply_comment_in_feed` | ✅ | 回复评论 |
| 11 | `like_feed` | ✅ | 点赞 |
| 12 | `favorite_feed` | ✅ | 收藏 |
| 13 | `publish_with_video` | ⏳ | 发布视频（未实现） |

### 业务文件

| 文件 | 状态 | 说明 |
|------|------|------|
| `heybox/feeds.go` | ✅ | 动态列表获取 |
| `heybox/search.go` | ✅ | 搜索功能 |
| `heybox/feed_detail.go` | ⚠️ | 动态详情（有验证码问题） |
| `heybox/comment.go` | ✅ | 评论/回复 |
| `heybox/like.go` | ✅ | 点赞/收藏 |
| `heybox/user_profile.go` | ✅ | 用户主页获取 |
| `heybox/publish.go` | ✅ | 内容发布 |
| `heybox/login.go` | ✅ | 登录功能 |
| `heybox/navigate.go` | ✅ | 页面导航 |
| `heybox/human.go` | ✅ | 人类行为模拟 |

---

## 待完成

### Phase 5: 测试与优化 (30%)

- [ ] 单元测试覆盖
- [ ] 集成测试
- [ ] MCP 协议测试
- [ ] 性能优化
- [ ] 文档完善

### 优先级 1 - 功能完善

| 功能 | 说明 | 复杂度 |
|------|------|--------|
| 视频发布 | `publish_with_video` 功能实现 | 中 |
| 获取评论列表 | `get_comments` 工具实现 | 低 |
| 验证码绕过 | 研究更好的反检测方案 | 高 |

### 优先级 2 - 体验优化

| 功能 | 说明 |
|------|------|
| 选择器优化 | 改进点赞数、评论数的获取 |
| 错误重试 | 请求失败自动重试机制 |
| 请求限流 | 避免触发风控的频率控制 |
| 日志完善 | 更详细的调试日志 |

---

## 需要注意的内容

### DOM 选择器适配

**所有 `heybox/` 包中的 CSS 选择器都需要根据小黑盒网站实际结构进行调整。**

当前使用的选择器（示例）：
```go
// 登录相关
".login-btn"
".qrcode-container img"
".user-avatar"

// 发布相关
".title-input"
".content-textarea"
".publish-btn"
".image-upload-input"
```

**调整方法**：
1. 打开小黑盒网站 (https://www.xiaoheihe.cn/home)
2. 使用浏览器开发者工具（F12）检查元素
3. 找到实际的 CSS 选择器
4. 更新代码中的选择器

---

## 目录结构

```
heybox-mcp/
├── main.go                      # 主入口
├── mcp_server.go                # MCP 服务器
├── app_server.go                # HTTP 服务
├── middleware.go                # 中间件
├── service.go                   # 服务层
├── go.mod / go.sum              # Go 依赖
├── Makefile                     # 构建脚本
│
├── configs/                     # 配置管理
├── cookies/                     # Cookie 管理
├── browser/                     # 浏览器封装
├── heybox/                      # 业务逻辑
├── pkg/downloader/              # 图片下载器
├── errors/                      # 错误处理
├── cmd/login/                   # 登录工具
│
├── README.md                    # 项目说明
└── PROGRESS.md                  # 本文件
```

---

## 编译和运行

```bash
# 编译主程序
go build -o heybox-mcp .

# 编译登录工具
go build -o heybox-login cmd/login/main.go

# 运行登录工具
./heybox-login

# 启动 MCP 服务
./heybox-mcp -headless=true

# 启动 MCP Inspector
npx @modelcontextprotocol/inspector --transport http --server-url http://localhost:18060/mcp

# 测试动态列表
go run test_feeds.go
```

---

## 最新更新 (2026-03-03)

1. **动态列表功能测试通过**
   - `list_feeds` 成功获取 10 条动态
   - Cookies 加载正常（30 个）
   - JavaScript 提取方式正常工作

2. **MCP Inspector 测试通过**
   - HTTP 传输模式正常
   - 工具调用正常响应

3. **返回数据优化**
   - `list_feeds` 现在返回完整 JSON 数据（包含标题、作者、链接、图片等）
   - `search_feeds` 同样返回完整数据

4. **添加随机延迟和行为模拟**
   - 导航后模拟页面交互

5. **修复 Cookie 功能** ✅

6. **依赖版本问题**
   - `github.com/nfnt/resize` 包存在版本问题，已暂时移除

---

**记录时间**: 2026-03-03
