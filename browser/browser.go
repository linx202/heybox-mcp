package browser

import (
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"
)

// BrowserOption 浏览器配置选项函数类型
type BrowserOption func(*launcher.Launcher)

// WithBinPath 设置浏览器二进制文件路径
func WithBinPath(binPath string) BrowserOption {
	return func(l *launcher.Launcher) {
		if binPath != "" {
			l.Bin(binPath)
		}
	}
}

// WithSlowMo 设置慢动作模式（便于调试）
func WithSlowMo(duration time.Duration) BrowserOption {
	return func(l *launcher.Launcher) {
		// SlowMotion 需要在 Browser 对象上设置
		// 这里先忽略
	}
}

// WithTrace 开启追踪模式
func WithTrace(enable bool) BrowserOption {
	return func(l *launcher.Launcher) {
		if enable {
			l.Devtools(true)
		}
	}
}

// NewBrowser 创建新的浏览器实例
func NewBrowser(headless bool, opts ...BrowserOption) *rod.Browser {
	// 创建 launcher 并应用选项
	l := launcher.New()
	for _, opt := range opts {
		opt(l)
	}
	l.Headless(headless)

	// 添加反检测参数
	l.Set("disable-blink-features", "AutomationControlled")
	l.Set("disable-infobars")
	l.Set("start-maximized")

	url := l.MustLaunch()

	// 创建浏览器实例
	browser := rod.New().ControlURL(url)

	// 连接浏览器
	if err := browser.Connect(); err != nil {
		logrus.Fatalf("连接浏览器失败: %v", err)
	}

	logrus.Infof("浏览器已启动 (headless=%v)", headless)
	return browser
}

// NewBrowserWithLauncher 使用自定义 launcher 创建浏览器
func NewBrowserWithLauncher(launcherURL string) *rod.Browser {
	browser := rod.New().ControlURL(launcherURL)

	if err := browser.Connect(); err != nil {
		logrus.Fatalf("连接浏览器失败: %v", err)
	}

	return browser
}

// MustNewPage 创建新页面（简化操作）
func MustNewPage(browser *rod.Browser) *rod.Page {
	page, err := browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		page = browser.MustPage("")
	}
	return page
}
