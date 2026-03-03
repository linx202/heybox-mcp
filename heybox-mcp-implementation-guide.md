# heybox-mcp 实现指南

> 基于 xiaohongshu-mcp 架构的小黑盒 MCP 服务器实现方案
> 创建时间：2025-02-27

---

## 目录

1. [项目概述](#1-项目概述)
2. [平台差异分析](#2-平台差异分析)
3. [迁移方案](#3-迁移方案)
4. [实现步骤](#4-实现步骤)
5. [核心代码模板](#5-核心代码模板)
6. [部署指南](#6-部署指南)
7. [注意事项](#7-注意事项)

---

## 1. 项目概述

### 1.1 小黑盒平台特点

**小黑盒（Heybox）** 是一个游戏社区平台，主要功能包括：

| 功能模块 | 说明 |
|---------|------|
| 游戏资讯 | 游戏新闻、评测、攻略 |
| 用户动态 | 类似微博的动态发布 |
| 评论互动 | 对动态和资讯的评论、点赞 |
| 游戏库 | 游戏数据库和评分 |
| 用户系统 | 关注、私信、个人主页 |

### 1.2 目标功能

基于小红书 MCP 的架构，实现以下功能：

- ✅ 登录管理（二维码/账号密码）
- ✅ 发布动态（文字+图片）
- ✅ 获取动态列表
- ✅ 搜索内容
- ✅ 评论/回复
- ✅ 点赞/收藏
- ✅ 获取用户信息
- ⭐ 游戏库查询（特色功能）

---

## 2. 平台差异分析

### 2.1 网站结构对比

| 特性 | 小红书 | 小黑盒 |
|-----|-------|-------|
| 域名 | xiaohongshu.com | heybox.cn |
| 登录方式 | 二维码扫码 | 二维码/密码 |
| 发布入口 | creator.xiaohongshu.com | /动态发布 |
| 数据获取方式 | `__INITIAL_STATE__` | API 接口 |
| 图片上传 | 单次多图 | 单次多图 |
| 标签系统 | # 话题标签 | @ 提及 + # 话题 |
| 视频支持 | 是 | 是 |

### 2.2 技术实现差异

```javascript
// 小红书：页面内置数据
window.__INITIAL_STATE__.feed.feeds._value

// 小黑盒：可能使用 API 接口
fetch("https://apiheybox.cn/Bbs/GetRecommendPostList")
```

### 2.3 DOM 选择器差异

| 功能 | 小红书选择器 | 小黑盒选择器（待确认） |
|-----|-------------|---------------------|
| 登录按钮 | `.login-container .qrcode-img` | `.login-qrcode img` |
| 发布输入框 | `div.ql-editor` | `.publish-textarea` |
| 图片上传 | `.upload-input` | `.image-upload input` |
| 发布按钮 | `.publish-page-publish-btn button` | `.publish-btn` |

---

## 3. 迁移方案

### 3.1 架构复用

保持与 xiaohongshu-mcp 相同的架构：

```
┌─────────────────────────────────────┐
│         客户端层                     │
│  Claude Code / Cursor / VSCode      │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│       传输层 (Transport)             │
│   MCP 协议 + HTTP REST API          │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│      协议层 (Protocol)               │
│      MCP Server (13 tools)          │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│    处理器层 (Handlers)               │
│   MCP Handlers + HTTP Handlers      │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│     服务层 (Service)                 │
│      HeyboxService                  │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│   业务逻辑层 (Domain Actions)        │
│  Login / Publish / Feeds / Comments │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│   基础设施层 (Infrastructure)        │
│  Browser / Cookies / Downloader     │
└─────────────────────────────────────┘
```

### 3.2 代码结构映射

| xiaohongshu-mcp | heybox-mcp | 说明 |
|----------------|-----------|------|
| `xiaohongshu/` | `heybox/` | 业务逻辑目录 |
| `xiaohongshu/login.go` | `heybox/login.go` | 登录逻辑 |
| `xiaohongshu/publish.go` | `heybox/publish.go` | 发布动态 |
| `xiaohongshu/feeds.go` | `heybox/feeds.go` | 动态列表 |
| `XiaohongshuService` | `HeyboxService` | 服务层 |
| `configs/username.go` | `configs/username.go` | 用户配置 |

### 3.3 工具列表调整

| 序号 | 工具名 | 小红书 | 小黑盒 | 差异说明 |
|-----|-------|-------|-------|---------|
| 1 | `check_login_status` | ✅ | ✅ | 选择器需调整 |
| 2 | `get_login_qrcode` | ✅ | ✅ | 选择器需调整 |
| 3 | `delete_cookies` | ✅ | ✅ | 相同 |
| 4 | `publish_content` | ✅ | ✅ | 发布入口不同 |
| 5 | `publish_with_video` | ✅ | ✅ | 视频上传逻辑 |
| 6 | `list_feeds` | ✅ | ✅ | 推荐→动态列表 |
| 7 | `search_feeds` | ✅ | ✅ | 搜索接口 |
| 8 | `get_feed_detail` | ✅ | ✅ | 帖子详情 |
| 9 | `user_profile` | ✅ | ✅ | 用户主页 |
| 10 | `post_comment_to_feed` | ✅ | ✅ | 评论接口 |
| 11 | `reply_comment_in_feed` | ✅ | ✅ | 回复接口 |
| 12 | `like_feed` | ✅ | ✅ | 点赞接口 |
| 13 | `favorite_feed` | ✅ | ⭐ | 小黑盒可能无收藏，需确认 |

---

## 4. 实现步骤

### Phase 1: 项目初始化

```bash
# 1. 创建项目目录
mkdir heybox-mcp
cd heybox-mcp

# 2. 初始化 Go 模块
go mod init github.com/yourusername/heybox-mcp

# 3. 复制基础文件
# 从 xiaohongshu-mcp 复制以下文件（修改命名）：
# - main.go → main.go
# - app_server.go → app_server.go
# - service.go → service.go
# - types.go → types.go
# - mcp_server.go → mcp_server.go
# - mcp_handlers.go → mcp_handlers.go
# - handlers_api.go → handlers_api.go
# - routes.go → routes.go
# - middleware.go → middleware.go

# 4. 复制支持目录
cp -r xiaohongshu-mcp/configs .
cp -r xiaohongshu-mcp/cookies .
cp -r xiaohongshu-mcp/browser .
cp -r xiaohongshu-mcp/pkg .
cp -r xiaohongshu-mcp/errors .
```

### Phase 2: 业务逻辑实现

```bash
# 1. 创建 heybox 目录
mkdir heybox

# 2. 实现核心模块
touch heybox/login.go           # 登录
touch heybox/publish.go         # 发布
touch heybox/feeds.go           # 动态列表
touch heybox/search.go          # 搜索
touch heybox/feed_detail.go     # 详情
touch heybox/comment.go         # 评论
touch heybox/like.go            # 点赞
touch heybox/user_profile.go    # 用户主页
touch heybox/types.go           # 数据结构
touch heybox/navigate.go        # 页面导航
```

### Phase 3: 适配修改

1. **全局替换**
   - `xiaohongshu` → `heybox`
   - `小红书` → `小黑盒`
   - `XiaohongshuService` → `HeyboxService`

2. **URL 修改**
   ```go
   // 小红书
   const urlOfPublic = `https://creator.xiaohongshu.com/publish/publish`

   // 小黑盒
   const urlOfPublish = `https://www.heybox.cn/publish`
   ```

3. **选择器调整**
   - 使用浏览器开发者工具查看小黑盒页面
   - 更新所有 CSS 选择器

### Phase 4: 测试验证

```bash
# 1. 构建项目
go build -o heybox-mcp .

# 2. 运行登录工具
go run cmd/login/main.go

# 3. 启动 MCP 服务
./heybox-mcp

# 4. 测试 MCP 连接
npx @modelcontextprotocol/inspector
```

---

## 5. 核心代码模板

### 5.1 main.go

```go
package main

import (
    "flag"
    "os"

    "github.com/sirupsen/logrus"
    "github.com/yourusername/heybox-mcp/configs"
)

func main() {
    var (
        headless bool
        binPath  string
        port     string
    )
    flag.BoolVar(&headless, "headless", true, "是否无头模式")
    flag.StringVar(&binPath, "bin", "", "浏览器二进制文件路径")
    flag.StringVar(&port, "port", ":18060", "端口")
    flag.Parse()

    if len(binPath) == 0 {
        binPath = os.Getenv("ROD_BROWSER_BIN")
    }

    configs.InitHeadless(headless)
    configs.SetBinPath(binPath)

    // 初始化服务
    heyboxService := NewHeyboxService()

    // 创建并启动应用服务器
    appServer := NewAppServer(heyboxService)
    if err := appServer.Start(port); err != nil {
        logrus.Fatalf("failed to run server: %v", err)
    }
}
```

### 5.2 service.go

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/go-rod/rod"
    "github.com/yourusername/headless_browser"
    "github.com/yourusername/heybox-mcp/browser"
    "github.com/yourusername/heybox-mcp/configs"
    "github.com/yourusername/heybox-mcp/cookies"
    "github.com/yourusername/heybox-mcp/heybox"
)

// HeyboxService 小黑盒业务服务
type HeyboxService struct{}

// NewHeyboxService 创建小黑盒服务实例
func NewHeyboxService() *HeyboxService {
    return &HeyboxService{}
}

// PublishRequest 发布请求
type PublishRequest struct {
    Title      string   `json:"title" binding:"required"`
    Content    string   `json:"content" binding:"required"`
    Images     []string `json:"images" binding:"required,min=1"`
    Tags       []string `json:"tags,omitempty"`
}

// LoginStatusResponse 登录状态响应
type LoginStatusResponse struct {
    IsLoggedIn bool   `json:"is_logged_in"`
    Username   string `json:"username,omitempty"`
}

// CheckLoginStatus 检查登录状态
func (s *HeyboxService) CheckLoginStatus(ctx context.Context) (*LoginStatusResponse, error) {
    b := newBrowser()
    defer b.Close()

    page := b.NewPage()
    defer page.Close()

    loginAction := heybox.NewLogin(page)

    isLoggedIn, err := loginAction.CheckLoginStatus(ctx)
    if err != nil {
        return nil, err
    }

    return &LoginStatusResponse{
        IsLoggedIn: isLoggedIn,
        Username:   configs.Username,
    }, nil
}

// PublishContent 发布内容
func (s *HeyboxService) PublishContent(ctx context.Context, req *PublishRequest) (*PublishResponse, error) {
    // 参数校验
    if len(req.Title) == 0 {
        return nil, fmt.Errorf("标题不能为空")
    }

    // 处理图片
    imagePaths, err := s.processImages(req.Images)
    if err != nil {
        return nil, err
    }

    // 构建发布内容
    content := heybox.PublishContent{
        Title:      req.Title,
        Content:    req.Content,
        Tags:       req.Tags,
        ImagePaths: imagePaths,
    }

    // 执行发布
    if err := s.publishContent(ctx, content); err != nil {
        return nil, err
    }

    return &PublishResponse{
        Title:   req.Title,
        Content: req.Content,
        Images:  len(imagePaths),
        Status:  "发布完成",
    }, nil
}

func newBrowser() *headless_browser.Browser {
    return browser.NewBrowser(configs.IsHeadless(), browser.WithBinPath(configs.GetBinPath()))
}

func saveCookies(page *rod.Page) error {
    cks, err := page.Browser().GetCookies()
    if err != nil {
        return err
    }

    data, err := json.Marshal(cks)
    if err != nil {
        return err
    }

    cookieLoader := cookies.NewLoadCookie(cookies.GetCookiesFilePath())
    return cookieLoader.SaveCookies(data)
}
```

### 5.3 heybox/login.go

```go
package heybox

import (
    "context"
    "time"

    "github.com/go-rod/rod"
    "github.com/pkg/errors"
)

type LoginAction struct {
    page *rod.Page
}

const (
    heyboxExploreURL = "https://www.heybox.cn/"
)

func NewLogin(page *rod.Page) *LoginAction {
    return &LoginAction{page: page}
}

// CheckLoginStatus 检查登录状态
func (a *LoginAction) CheckLoginStatus(ctx context.Context) (bool, error) {
    pp := a.page.Context(ctx)
    pp.MustNavigate(heyboxExploreURL).MustWaitLoad()

    time.Sleep(1 * time.Second)

    // ⚠️ 需要根据实际页面调整选择器
    // 示例：检查用户头像是否存在
    exists, _, err := pp.Has(".user-avatar")
    if err != nil {
        return false, errors.Wrap(err, "check login status failed")
    }

    return exists, nil
}

// FetchQrcodeImage 获取登录二维码
func (a *LoginAction) FetchQrcodeImage(ctx context.Context) (string, bool, error) {
    pp := a.page.Context(ctx)
    pp.MustNavigate(heyboxExploreURL).MustWaitLoad()
    time.Sleep(2 * time.Second)

    // 检查是否已登录
    if exists, _, _ := pp.Has(".user-avatar"); exists {
        return "", true, nil
    }

    // 点击登录按钮
    loginBtn, err := pp.Element(".login-btn")
    if err != nil {
        return "", false, errors.Wrap(err, "find login button failed")
    }
    loginBtn.MustClick()

    time.Sleep(1 * time.Second)

    // ⚠️ 需要根据实际页面调整二维码选择器
    src, err := pp.MustElement(".qrcode-container img").Attribute("src")
    if err != nil {
        return "", false, errors.Wrap(err, "get qrcode src failed")
    }

    if src == nil || len(*src) == 0 {
        return "", false, errors.New("qrcode src is empty")
    }

    return *src, false, nil
}

// WaitForLogin 等待登录完成
func (a *LoginAction) WaitForLogin(ctx context.Context) bool {
    pp := a.page.Context(ctx)
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return false
        case <-ticker.C:
            el, err := pp.Element(".user-avatar")
            if err == nil && el != nil {
                return true
            }
        }
    }
}
```

### 5.4 heybox/publish.go

```go
package heybox

import (
    "context"
    "log/slog"
    "os"
    "time"

    "github.com/go-rod/rod"
    "github.com/go-rod/rod/lib/proto"
    "github.com/pkg/errors"
    "github.com/sirupsen/logrus"
)

type PublishContent struct {
    Title      string
    Content    string
    Tags       []string
    ImagePaths []string
}

type PublishAction struct {
    page *rod.Page
}

const (
    heyboxPublishURL = "https://www.heybox.cn/publish"
)

func NewPublishAction(page *rod.Page) (*PublishAction, error) {
    pp := page.Timeout(300 * time.Second)

    if err := pp.Navigate(heyboxPublishURL); err != nil {
        return nil, errors.Wrap(err, "导航到发布页面失败")
    }

    pp.WaitLoad()
    time.Sleep(2 * time.Second)

    return &PublishAction{page: pp}, nil
}

func (p *PublishAction) Publish(ctx context.Context, content PublishContent) error {
    if len(content.ImagePaths) == 0 {
        return errors.New("图片不能为空")
    }

    page := p.page.Context(ctx)

    // 1. 上传图片
    if err := uploadImages(page, content.ImagePaths); err != nil {
        return errors.Wrap(err, "上传图片失败")
    }

    logrus.Infof("发布内容: title=%s, images=%v", content.Title, len(content.ImagePaths))

    // 2. 输入标题
    titleElem := page.MustElement(".title-input")
    titleElem.MustInput(content.Title)
    time.Sleep(500 * time.Millisecond)

    // 3. 输入正文
    contentElem := page.MustElement(".content-textarea")
    contentElem.MustInput(content.Content)
    time.Sleep(500 * time.Millisecond)

    // 4. 添加标签（可选）
    if len(content.Tags) > 0 {
        for _, tag := range content.Tags {
            contentElem.MustInput("#" + tag + " ")
            time.Sleep(200 * time.Millisecond)
        }
    }

    // 5. 点击发布
    publishBtn := page.MustElement(".publish-btn")
    publishBtn.MustClick(proto.InputMouseButtonLeft, 1)

    time.Sleep(3 * time.Second)

    return nil
}

func uploadImages(page *rod.Page, imagesPaths []string) error {
    pp := page.Timeout(30 * time.Second)

    // 验证文件
    validPaths := make([]string, 0, len(imagesPaths))
    for _, path := range imagesPaths {
        if _, err := os.Stat(path); os.IsNotExist(err) {
            logrus.Warnf("图片文件不存在: %s", path)
            continue
        }
        validPaths = append(validPaths, path)
    }

    // ⚠️ 需要根据实际页面调整上传输入框选择器
    uploadInput := pp.MustElement(".image-upload-input")
    uploadInput.MustSetFiles(validPaths...)

    // 等待上传完成
    time.Sleep(3 * time.Second)

    return nil
}
```

### 5.5 mcp_server.go（工具注册）

```go
package main

import (
    "context"
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/sirupsen/logrus"
)

// InitMCPServer 初始化 MCP Server
func InitMCPServer(appServer *AppServer) *mcp.Server {
    server := mcp.NewServer(&mcp.Implementation{
        Name:    "heybox-mcp",
        Version: "1.0.0",
    }, nil)

    registerTools(server, appServer)

    logrus.Info("MCP Server initialized")
    return server
}

func registerTools(server *mcp.Server, appServer *AppServer) {
    // 工具 1: 检查登录状态
    mcp.AddTool(server, &mcp.Tool{
        Name:        "check_login_status",
        Description: "检查小黑盒登录状态",
    }, func(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
        result := appServer.handleCheckLoginStatus(ctx)
        return convertToMCPResult(result), nil, nil
    })

    // 工具 2: 发布内容
    mcp.AddTool(server, &mcp.Tool{
        Name:        "publish_content",
        Description: "发布小黑盒动态",
    }, func(ctx context.Context, req *mcp.CallToolRequest, args PublishContentArgs) (*mcp.CallToolResult, any, error) {
        argsMap := map[string]interface{}{
            "title":    args.Title,
            "content":  args.Content,
            "images":   convertStringsToInterfaces(args.Images),
            "tags":     convertStringsToInterfaces(args.Tags),
        }
        result := appServer.handlePublishContent(ctx, argsMap)
        return convertToMCPResult(result), nil, nil
    })

    // ... 注册其他工具

    logrus.Infof("Registered MCP tools")
}
```

---

## 6. 部署指南

### 6.1 本地开发

```bash
# 克隆项目
git clone https://github.com/yourusername/heybox-mcp.git
cd heybox-mcp

# 安装依赖
go mod download

# 运行登录工具
go run cmd/login/main.go

# 启动 MCP 服务（无头模式）
go run .

# 启动 MCP 服务（有界面，便于调试）
go run . -headless=false

# 验证 MCP
npx @modelcontextprotocol/inspector
```

### 6.2 Claude Code 接入

```bash
# 添加 HTTP MCP 服务器
claude mcp add --transport http heybox-mcp http://localhost:18060/mcp

# 检查状态
claude mcp list
```

### 6.3 Docker 部署

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o heybox-mcp .

FROM alpine:latest
RUN apk --no-cache add chromium
WORKDIR /app
COPY --from=builder /app/heybox-mcp .
ENV ROD_BROWSER_BIN=/usr/bin/chromium-browser
EXPOSE 18060
CMD ["./heybox-mcp"]
```

```yaml
# docker-compose.yml
version: '3.8'
services:
  heybox-mcp:
    build: .
    ports:
      - "18060:18060"
    volumes:
      - ./data:/app/data
      - ./cookies.json:/app/cookies.json
```

---

## 7. 注意事项

### 7.1 选择器适配

**重要**：小黑盒的 DOM 结构可能与小红书完全不同，需要逐个适配：

1. 打开小黑盒网站
2. 使用开发者工具（F12）检查元素
3. 找到对应的 CSS 选择器
4. 更新代码中的选择器

```javascript
// 示例：查找登录二维码
document.querySelector('.qrcode-container img')

// 示例：查找发布输入框
document.querySelector('.publish-textarea')

// 示例：查找图片上传按钮
document.querySelector('.image-upload-input')
```

### 7.2 API 接口分析

小黑盒可能使用 REST API 而非页面内置数据：

```javascript
// 打开浏览器开发者工具 → Network 标签
// 观察数据请求，找到 API 端点

// 可能的 API 端点（待确认）
https://apiheybox.cn/Bbs/GetRecommendPostList
https://apiheybox.cn/Bbs/GetPostDetail
https://apiheybox.cn/Bbs/PublishPost
```

如果使用 API，可以直接使用 `http.Client` 而非浏览器自动化：

```go
// 替代方案：直接调用 API
func (s *HeyboxService) ListFeeds(ctx context.Context) (*FeedsListResponse, error) {
    resp, err := http.Get("https://apiheybox.cn/Bbs/GetRecommendPostList")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result FeedsListResponse
    json.NewDecoder(resp.Body).Decode(&result)
    return &result, nil
}
```

### 7.3 反爬虫机制

- 小黑盒可能有更严格的反爬虫措施
- 注意请求频率控制
- 使用真实 User-Agent
- Cookie 有效期管理

### 7.4 功能差异

**需要确认的功能**：

| 功能 | 小红书 | 小黑盒 | 状态 |
|-----|-------|-------|------|
| 收藏功能 | ✅ | ❓ | 需确认 |
| 视频发布 | ✅ | ❓ | 需确认 |
| 定时发布 | ✅ | ❓ | 需确认 |
| 话题标签 | ✅ | ❓ | 需确认 |
| 私信功能 | ❌ | ❓ | 待开发 |

### 7.5 测试策略

```bash
# 1. 单元测试
go test ./...

# 2. 集成测试（需要真实登录）
go test -tags=integration ./...

# 3. MCP 测试
npx @modelcontextprotocol/inspector

# 4. 压力测试
heybox-mcp-stress-test
```

---

## 8. 开发路线图

### Milestone 1: 核心功能（1-2周）

- [x] 项目初始化
- [ ] 登录功能
- [ ] 发布动态（图文）
- [ ] 获取动态列表
- [ ] 基础测试

### Milestone 2: 交互功能（1周）

- [ ] 搜索功能
- [ ] 评论/回复
- [ ] 点赞
- [ ] 用户主页

### Milestone 3: 高级功能（1-2周）

- [ ] 视频发布
- [ ] API 直接调用（非浏览器）
- [ ] 性能优化

### Milestone 4: 特色功能（可选）

- [ ] 游戏库查询
- [ ] 游戏评分
- [ ] 游戏资讯获取

---

## 9. 常见问题

### Q1: 如何快速找到页面选择器？

```javascript
// 在浏览器控制台运行
// 1. 查找二维码
document.querySelectorAll('img')

// 2. 查找输入框
document.querySelectorAll('input, textarea')

// 3. 查找按钮
document.querySelectorAll('button')

// 4. 使用特定类名
document.querySelector('[class*="login"]')
document.querySelector('[class*="publish"]')
```

### Q2: Cookie 失效如何处理？

```go
// 在每个请求前检查登录状态
func (s *HeyboxService) ensureLoggedIn(ctx context.Context) error {
    status, err := s.CheckLoginStatus(ctx)
    if err != nil {
        return err
    }

    if !status.IsLoggedIn {
        return errors.New("未登录，请先运行登录工具")
    }

    return nil
}
```

### Q3: 如何调试浏览器操作？

```bash
# 使用非无头模式
./heybox-mcp -headless=false

# 或设置环境变量
ROD_BROWSER_BIN=/path/to/chrome ./heybox-mcp -headless=false
```

---

## 10. 参考资源

- [xiaohongshu-mcp 源码](https://github.com/xpzouying/xiaohongshu-mcp)
- [go-rod 文档](https://go-rod.github.io/)
- [MCP 协议规范](https://modelcontextprotocol.io/)
- [小黑盒网站](https://www.heybox.cn/)

---

## 总结

基于 xiaohongshu-mcp 的成熟架构，实现 heybox-mcp 的关键在于：

1. **保持架构一致**：复用分层设计和代码结构
2. **适配平台差异**：重点调整 DOM 选择器和 API 接口
3. **渐进式开发**：从核心功能开始，逐步完善
4. **充分测试**：使用真实环境验证功能

预计开发周期：**3-4 周**（核心功能）

祝你开发顺利！🎮
