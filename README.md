# heybox-mcp

> 小黑盒 MCP 服务器 - 为 AI 助手提供小黑盒平台交互能力

---

## 🎯 项目概述

**heybox-mcp** 是一个基于 Model Context Protocol (MCP) 的服务器，为 AI 助手（Claude、Cursor、VS Code 等）提供小黑盒平台的交互能力。

### 核心功能

| 功能 | 状态 | 说明 |
|------|------|------|
| 登录管理 | ✅ | 二维码扫码登录 |
| 发布动态 | ✅ | 标题+正文，支持 API 响应截取 |
| 获取动态列表 | ✅ | 返回完整 JSON 数据 |
| 搜索内容 | ✅ | 搜索动态和用户 |
| 评论/回复 | ✅ | 发表评论和回复 |
| 点赞/收藏 | ✅ | 动态点赞和收藏 |
| 用户信息 | ✅ | 获取用户主页信息 |
| 发布视频 | ⏳ | 待实现 |

---

## 📁 项目结构

```
heybox-mcp/
├── main.go                      # 主入口
├── mcp_server.go                # MCP 服务器
├── service.go                   # 服务层
├── go.mod / go.sum              # Go 依赖
├── Makefile                     # 构建脚本
│
├── configs/                     # 配置管理
├── cookies/                     # Cookie 管理
├── browser/                     # 浏览器封装
├── heybox/                      # 业务逻辑
│   ├── login.go                 # 登录功能
│   ├── publish.go               # 发布动态
│   ├── feeds.go                 # 动态列表
│   ├── search.go                # 搜索功能
│   ├── comment.go               # 评论功能
│   ├── like.go                  # 点赞收藏
│   ├── user_profile.go          # 用户主页
│   └── navigate.go              # 页面导航
│
├── pkg/downloader/              # 图片下载器
├── errors/                      # 错误处理
├── cmd/                         # 命令行工具
│   ├── login/                   # 登录工具
│   └── extract_selectors/       # 选择器提取工具
│
├── test_publish.go              # 发布测试脚本
├── test_feeds.go                # 动态列表测试
├── README.md                    # 项目说明
└── PROGRESS.md                  # 开发进度
```

---

## 🚀 快速开始

### 前置要求

- Go 1.21+
- Chrome/Chromium 浏览器

### 安装

```bash
# 克隆项目
git clone https://github.com/linx202/heybox-mcp.git
cd heybox-mcp

# 安装依赖
go mod download

# 编译
go build -o heybox-mcp .
```

### 使用

```bash
# 启动 MCP 服务（显示浏览器）
./heybox-mcp -headless=false

# 启动 MCP 服务（无头模式）
./heybox-mcp -headless=true

# 测试发布功能
go run test_publish.go

# 测试动态列表
go run test_feeds.go
```

### MCP Inspector 测试

```bash
npx @modelcontextprotocol/inspector --transport http --server-url http://localhost:18060/mcp
```

---

## 🔧 MCP 工具列表

| 工具名称 | 说明 |
|----------|------|
| `check_login_status` | 检查登录状态 |
| `get_login_qrcode` | 获取登录二维码 |
| `delete_cookies` | 删除 Cookie（退出登录） |
| `publish_content` | 发布动态内容 |
| `list_feeds` | 获取推荐动态列表 |
| `search_feeds` | 搜索内容 |
| `get_feed_detail` | 获取动态详情 |
| `get_user_profile` | 获取用户信息 |
| `post_comment_to_feed` | 发表评论 |
| `reply_comment_in_feed` | 回复评论 |
| `like_feed` | 点赞动态 |
| `favorite_feed` | 收藏动态 |

---

## 🔌 Claude Code 接入

```bash
# 添加 HTTP MCP 服务器
claude mcp add --transport http heybox-mcp http://localhost:18060/mcp

# 检查状态
claude mcp list
```

---

## 📝 发布功能说明

### 发布页面结构

- **URL**: `https://www.xiaoheihe.cn/creator/editor/draft/image_text`
- **标题**: `.editor-title__container [contenteditable='true']`
- **正文**: `.image-text__edit-content--inner [contenteditable='true']`
- **发布按钮**: `button.editor-publish__btn.main-btn`

### API 响应截取

发布时会自动监听 `api.xiaoheihe.cn/bbs/app/api/link/post` 接口响应：

```json
{
  "status": "ok|failed",
  "msg": "响应消息",
  "result": {}
}
```

---

## 🛠️ Make 命令

```bash
make build        # 编译主程序
make run          # 运行主程序
make test         # 运行测试
make clean        # 清理构建
```

---

## ⚠️ 注意事项

1. **选择器适配**：小黑盒 DOM 结构可能变化，需要定期更新选择器
2. **API 接口**：部分功能通过 API 实现而非浏览器自动化
3. **反爬虫**：注意请求频率和 Cookie 管理
4. **账号要求**：发布内容需要绑定手机号

---

## 📅 开发进度

- [x] 项目初始化
- [x] 基础设施层
- [x] 登录功能
- [x] 发布动态（标题+正文）
- [x] 动态列表
- [x] 搜索功能
- [x] 评论互动
- [x] 用户主页
- [ ] 图片上传
- [ ] 关联社区/话题
- [ ] 发布视频
- [ ] 完善文档

详细进度请查看 [PROGRESS.md](./PROGRESS.md)

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

## 📄 许可

MIT License
