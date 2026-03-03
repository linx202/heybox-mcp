# heybox-mcp

> 基于 xiaohongshu-mcp 架构的小黑盒 MCP 服务器实现

> 📚 **学习文档**：
> - [xiaohongshu-mcp-architecture.md](./xiaohongshu-mcp-architecture.md) - 原项目架构学习
> - [heybox-mcp-implementation-guide.md](./heybox-mcp-implementation-guide.md) - 实现指南

---

## 🎯 项目概述

**heybox-mcp** 是一个基于 Model Context Protocol (MCP) 的服务器，为 AI 助手（Claude、Cursor、VS Code 等）提供小黑盒平台的交互能力。

### 核心功能

- ✅ 登录管理（二维码）
- ✅ 发布动态（文字+图片）
- ✅ 获取动态列表
- ✅ 搜索内容
- ✅ 评论/回复
- ✅ 点赞/收藏
- ✅ 用户信息查询

### 开发状态

> 🚧 **项目初始化完成，正在开发中...**

---

## 📁 项目结构

```
heybox-mcp/
├── main.go                      # 主入口
├── go.mod                       # Go 依赖
├── Makefile                     # 构建脚本
├── .gitignore                   # Git 忽略
│
├── configs/                     # 配置管理
│   ├── browser.go               # 浏览器配置
│   ├── image.go                 # 图片配置
│   └── username.go              # 用户名配置
│
├── cookies/                     # Cookie 管理
│   └── cookies.go               # Cookie 存储/加载
│
├── browser/                     # 浏览器封装
│   └── browser.go               # 浏览器工厂
│
├── heybox/                      # 小黑盒业务逻辑
│   ├── login.go                 # 登录逻辑
│   ├── publish.go               # 发布动态
│   ├── feeds.go                 # 动态列表
│   ├── search.go                # 搜索功能
│   ├── feed_detail.go           # 帖子详情
│   ├── comment.go               # 评论功能
│   ├── like.go                  # 点赞收藏
│   ├── user_profile.go          # 用户主页
│   ├── types.go                 # 数据结构
│   └── navigate.go              # 页面导航
│
├── pkg/downloader/              # 图片下载器
│   ├── images.go
│   ├── processor.go
│   └── images_test.go
│
├── errors/                      # 错误处理
│   └── errors.go
│
├── cmd/login/                   # 登录工具
│   └── main.go
│
├── docs/                        # 文档（可选）
│   └── API.md
│
└── *.md                         # 学习文档
```

---

## 🚀 快速开始

### 前置要求

- Go 1.21+
- Chrome/Chromium 浏览器

### 安装

```bash
# 1. 克隆项目（如果从 GitHub）
git clone https://github.com/yourusername/heybox-mcp.git
cd heybox-mcp

# 2. 安装依赖
make deps

# 3. 编译
make build
```

### 使用

```bash
# 1. 首次登录
make run-login

# 2. 启动 MCP 服务
make run

# 3. 或者带界面模式（便于调试）
./heybox-mcp -headless=false
```

### 验证 MCP

```bash
npx @modelcontextprotocol/inspector
# 连接到: http://localhost:18060/mcp
```

---

## 🔧 开发指南

### 核心文件实现优先级

1. **Phase 1: 基础设施** ⭐
   - `configs/browser.go` - 浏览器配置
   - `cookies/cookies.go` - Cookie 管理
   - `browser/browser.go` - 浏览器封装

2. **Phase 2: 登录功能** 🔑
   - `heybox/login.go` - 登录逻辑
   - `heybox/types.go` - 数据结构
   - `cmd/login/main.go` - 登录工具

3. **Phase 3: 发布功能** 📝
   - `heybox/publish.go` - 发布动态
   - `service.go` - 业务服务层
   - `mcp_handlers.go` - MCP 处理器

4. **Phase 4: 其他功能** 📚
   - `heybox/feeds.go` - 动态列表
   - `heybox/search.go` - 搜索
   - `heybox/comment.go` - 评论
   - 等等...

### 参考实现

所有核心代码模板都在 **[heybox-mcp-implementation-guide.md](./heybox-mcp-implementation-guide.md)** 中：

- ✅ `main.go` 完整代码（第 5 节）
- ✅ `service.go` 完整代码（第 5 节）
- ✅ `heybox/login.go` 完整代码（第 5 节）
- ✅ `heybox/publish.go` 完整代码（第 5 节）
- ✅ `mcp_server.go` 完整代码（第 5 节）

---

## 📖 学习资源

| 文档 | 说明 |
|-----|------|
| [xiaohongshu-mcp-architecture.md](./xiaohongshu-mcp-architecture.md) | 完整的架构学习文档（10 章） |
| [heybox-mcp-implementation-guide.md](./heybox-mcp-implementation-guide.md) | 实现指南和代码模板（10 章） |
| [原项目](https://github.com/xpzouying/xiaohongshu-mcp) | xiaohongshu-mcp 源码 |
| [MCP 协议](https://modelcontextprotocol.io/) | MCP 官方文档 |
| [go-rod 文档](https://go-rod.github.io/) | 浏览器自动化库 |

---

## 🛠️ Make 命令

```bash
make build        # 编译主程序
make login        # 编译登录工具
make all          # 编译所有
make run          # 运行主程序
make run-login    # 运行登录工具
make test         # 运行测试
make clean        # 清理构建
make deps         # 安装依赖
make docker-build # Docker 构建
make docker-run   # Docker 运行
```

---

## 🔌 Claude Code 接入

```bash
# 添加 HTTP MCP 服务器
claude mcp add --transport http heybox-mcp http://localhost:18060/mcp

# 检查状态
claude mcp list
```

---

## ⚠️ 注意事项

1. **选择器适配**：小黑盒 DOM 结构需要实际分析，参考实现指南第 7 节
2. **API 接口**：可能需要直接调用 API 而非浏览器自动化
3. **反爬虫**：注意请求频率和 Cookie 管理
4. **功能确认**：部分功能（如收藏）需要先确认小黑盒是否支持

---

## 📅 开发计划

- [x] 项目初始化
- [ ] 基础设施层
- [ ] 登录功能
- [ ] 发布动态
- [ ] 动态列表
- [ ] 搜索功能
- [ ] 评论互动
- [ ] 用户主页
- [ ] 完善文档

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

## 📄 许可

MIT License

---

## 🙏 致谢

本项目基于 [xpzouying/xiaohongshu-mcp](https://github.com/xpzouying/xiaohongshu-mcp) 的架构设计。
