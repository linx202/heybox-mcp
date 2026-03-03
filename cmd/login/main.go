package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/heybox-mcp/browser"
	"github.com/yourusername/heybox-mcp/configs"
	"github.com/yourusername/heybox-mcp/cookies"
	"github.com/yourusername/heybox-mcp/heybox"
)

func main() {
	// 命令行参数
	var clearCookies bool
	flag.BoolVar(&clearCookies, "clear", false, "清除已保存的 Cookies")
	flag.Parse()

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	logrus.Info("🔐 小黑盒登录工具启动")

	// 初始化配置
	configs.InitHeadless(false) // 登录时显示浏览器窗口
	cookies.InitCookieDir("./cookies")

	// 创建浏览器
	browserInstance := browser.NewBrowser(
		configs.IsHeadless(),
		browser.WithBinPath(configs.GetBinPath()),
	)
	defer browserInstance.Close()

	// 创建页面
	page := browserInstance.MustPage("")
	defer page.Close()

	ctx := context.Background()
	loginAction := heybox.NewLogin(page)

	// 如果指定了清除 Cookies
	if clearCookies {
		if err := loginAction.ClearCookies(); err != nil {
			logrus.Errorf("清除 Cookies 失败: %v", err)
		} else {
			logrus.Info("已清除 Cookies，请重新登录")
		}
		return
	}

	// 尝试加载已保存的 Cookies
	logrus.Info("尝试加载已保存的 Cookies...")
	if err := loginAction.LoadCookies(); err != nil {
		logrus.Warnf("加载 Cookies 失败: %v", err)
	} else {
		logrus.Info("已加载 Cookies，检查登录状态...")
	}

	// 导航到首页并检查登录状态
	logrus.Info("检查登录状态...")
	isLoggedIn, err := loginAction.CheckLoginStatus(ctx)
	if err != nil {
		logrus.Warnf("检查登录状态失败: %v", err)
	}

	if isLoggedIn {
		logrus.Info("✅ 已通过 Cookies 登录")
		return
	}

	// Cookies 无效或过期，需要扫码登录
	logrus.Info("Cookies 无效或已过期，需要扫码登录")

	// 获取登录二维码
	logrus.Info("获取登录二维码...")
	qrcodeURL, alreadyLoggedIn, err := loginAction.FetchQrcodeImage(ctx)
	if err != nil {
		logrus.Fatalf("获取二维码失败: %v", err)
	}

	if alreadyLoggedIn {
		logrus.Info("✅ 已登录")
		// 保存 Cookies
		if err := loginAction.SaveCookies(); err != nil {
			logrus.Warnf("保存 Cookies 失败: %v", err)
		}
		return
	}

	if qrcodeURL != "" {
		// 在终端显示二维码
		if err := loginAction.PrintQRCodeInTerminal(qrcodeURL); err != nil {
			logrus.Warnf("终端显示二维码失败: %v", err)
			logrus.Info("===========================================")
			logrus.Info("📱 请使用小黑盒 APP 扫描二维码登录")
			logrus.Infof("二维码链接: %s", qrcodeURL)
			logrus.Info("===========================================")
		}
	}

	// 等待登录
	logrus.Info("等待扫码登录...")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logrus.Info("用户中断")
		os.Exit(0)
	}()

	success := loginAction.WaitForLogin(ctx)
	if success {
		logrus.Info("✅ 登录成功！")

		// 保存 Cookies
		time.Sleep(2 * time.Second) // 等待页面完全加载
		if err := loginAction.SaveCookies(); err != nil {
			logrus.Warnf("保存 Cookies 失败: %v", err)
		} else {
			logrus.Info("Cookies 已保存，下次启动将自动登录")
		}
	} else {
		logrus.Error("❌ 登录失败或超时")
		os.Exit(1)
	}
}
