package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/heybox-mcp/configs"
	"github.com/yourusername/heybox-mcp/cookies"
)

func main() {
	var (
		headless bool
		binPath  string
		port     string
	)
	flag.BoolVar(&headless, "headless", true, "是否无头模式")
	flag.StringVar(&binPath, "bin", "", "浏览器二进制文件路径")
	flag.StringVar(&port, "port", ":18060", "服务端口")
	flag.Parse()

	if len(binPath) == 0 {
		binPath = os.Getenv("ROD_BROWSER_BIN")
	}

	configs.InitHeadless(headless)
	configs.SetBinPath(binPath)
	cookies.InitCookieDir("./cookies")

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	logrus.Info("🎮 heybox-mcp 启动中...")
	logrus.Infof("配置: headless=%v, port=%s", headless, port)

	// 初始化服务
	heyboxService := NewHeyboxService()

	// 创建应用服务器
	appServer := NewAppServer(heyboxService)

	// 处理优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		logrus.Info("正在关闭服务器...")
		if err := appServer.Shutdown(nil); err != nil {
			logrus.Errorf("关闭服务器失败: %v", err)
		}
		logrus.Info("服务器已关闭")
		os.Exit(0)
	}()

	// 启动服务器
	if err := appServer.Start(port); err != nil {
		logrus.Fatalf("服务器启动失败: %v", err)
	}
}
