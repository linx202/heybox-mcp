package heybox

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	stderrors "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/heybox-mcp/configs"
	"github.com/yourusername/heybox-mcp/cookies"
	apperrors "github.com/yourusername/heybox-mcp/errors"
)

// LoginAction 登录操作
type LoginAction struct {
	page      *rod.Page
	navigator *Navigator
}

// NewLogin 创建登录操作实例
func NewLogin(page *rod.Page) *LoginAction {
	return &LoginAction{
		page:      page,
		navigator: NewNavigator(page),
	}
}

// CheckLoginStatus 检查登录状态（会导航到首页）
func (a *LoginAction) CheckLoginStatus(ctx context.Context) (bool, error) {
	return a.checkLoginStatus(ctx, true)
}

// CheckLoginStatusOnCurrentPage 在当前页面检查登录状态（不导航）
func (a *LoginAction) CheckLoginStatusOnCurrentPage(ctx context.Context) (bool, error) {
	return a.checkLoginStatus(ctx, false)
}

// checkLoginStatus 内部方法
func (a *LoginAction) checkLoginStatus(ctx context.Context, navigate bool) (bool, error) {
	pp := a.page.Context(ctx).Timeout(30 * time.Second)

	if navigate {
		if err := a.navigator.NavigateToHome(ctx); err != nil {
			return false, err
		}
	}

	// ⚠️ 需要根据实际页面调整选择器
	// 检查用户头像是否存在（表示已登录）
	exists, _, err := pp.Has(".user-avatar")
	if err != nil {
		return false, stderrors.Wrap(err, "检查登录状态失败")
	}

	// 如果第一个选择器不存在，尝试其他可能的选择器
	if !exists {
		exists, _, _ = pp.Has(".avatar-img")
		if !exists {
			exists, _, _ = pp.Has("[class*='avatar']")
		}
	}

	if exists {
		// 尝试获取用户名
		if usernameElem, err := pp.Element(".user-name"); err == nil {
			if username, err := usernameElem.Text(); err == nil {
				configs.SetUsername(username)
				logrus.Infof("当前登录用户: %s", username)
			}
		}
	}

	return exists, nil
}

// FetchQrcodeImage 获取登录二维码
func (a *LoginAction) FetchQrcodeImage(ctx context.Context) (string, bool, error) {
	pp := a.page.Context(ctx).Timeout(60 * time.Second)

	// 导航到首页
	if err := a.navigator.NavigateToHome(ctx); err != nil {
		return "", false, err
	}

	// 等待页面加载
	time.Sleep(2 * time.Second)

	// 检查是否已登录
	if isLoggedIn, _ := a.CheckLoginStatus(ctx); isLoggedIn {
		return "", true, nil
	}

	// 步骤1: 点击登录按钮
	loginSelectors := []string{
		".login-btn",
		"[class*='login']",
		"button[class*='Login']",
		".header-login-btn",
	}

	var loginBtn *rod.Element
	var err error
	for _, selector := range loginSelectors {
		loginBtn, err = pp.Element(selector)
		if err == nil && loginBtn != nil {
			logrus.Infof("找到登录按钮: %s", selector)
			break
		}
	}

	if loginBtn == nil {
		// 尝试通过文本查找
		loginBtn, err = pp.ElementR("button", "登录")
		if err != nil {
			return "", false, apperrors.ErrElementNotFound("登录按钮", err)
		}
	}

	// 点击登录按钮
	if err := loginBtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return "", false, stderrors.Wrap(err, "点击登录按钮失败")
	}

	logrus.Info("已点击登录按钮，等待弹窗...")
	time.Sleep(2 * time.Second)

	// 步骤2: 点击二维码图标触发请求
	qrcodeIconSelectors := []string{
		".login-box .title img",
		".login-box img",
		".title img",
	}

	var qrcodeIcon *rod.Element
	for _, selector := range qrcodeIconSelectors {
		qrcodeIcon, err = pp.Element(selector)
		if err == nil && qrcodeIcon != nil {
			logrus.Infof("找到二维码图标: %s", selector)
			break
		}
	}

	if qrcodeIcon != nil {
		// 使用 JavaScript 点击，绕过 pointer-events: none
		_, err := qrcodeIcon.Eval(`() => this.click()`)
		if err != nil {
			// 尝试点击父元素
			logrus.Warnf("点击图标失败，尝试点击父元素: %v", err)
			parentBtn, parentErr := pp.Element(".login-box .title")
			if parentErr == nil && parentBtn != nil {
				if clickErr := parentBtn.Click(proto.InputMouseButtonLeft, 1); clickErr != nil {
					logrus.Warnf("点击父元素也失败: %v", clickErr)
				} else {
					logrus.Info("已点击父元素，等待二维码生成...")
				}
			}
		} else {
			logrus.Info("已点击二维码图标，等待二维码生成...")
		}
	} else {
		logrus.Warn("未找到二维码图标，尝试直接获取二维码")
	}

	// 步骤3: 等待二维码 canvas 出现
	time.Sleep(2 * time.Second)

	qrcodeCanvasSelectors := []string{
		"#login-qrcode",
		"canvas#login-qrcode",
		".login-box canvas",
		"canvas[class*='qrcode']",
	}

	var qrcodeCanvas *rod.Element
	for _, selector := range qrcodeCanvasSelectors {
		qrcodeCanvas, err = pp.Element(selector)
		if err == nil && qrcodeCanvas != nil {
			logrus.Infof("找到二维码 Canvas: %s", selector)
			break
		}
	}

	if qrcodeCanvas == nil {
		return "", false, stderrors.New("未找到二维码 Canvas 元素")
	}

	// 步骤4: 从 Canvas 获取 Base64 图片数据
	// 使用 JavaScript 获取 canvas 的 base64 数据
	base64Data, err := qrcodeCanvas.Eval(`() => this.toDataURL('image/png')`)
	if err != nil {
		return "", false, stderrors.Wrap(err, "获取二维码 Canvas 数据失败")
	}

	qrcodeDataURL := base64Data.Value.String()
	if qrcodeDataURL == "" {
		return "", false, stderrors.New("二维码 Canvas 数据为空")
	}

	logrus.Info("✅ 成功获取二维码")
	return qrcodeDataURL, false, nil
}

// WaitForLogin 等待登录完成
func (a *LoginAction) WaitForLogin(ctx context.Context) bool {
	ticker := time.NewTicker(2 * time.Second) // 每2秒检查一次
	defer ticker.Stop()

	timeout := time.NewTimer(5 * time.Minute) // 5分钟超时
	defer timeout.Stop()

	logrus.Info("等待扫码登录...")

	for {
		select {
		case <-ctx.Done():
			logrus.Error("上下文已取消")
			return false
		case <-timeout.C:
			logrus.Error("等待登录超时")
			return false
		case <-ticker.C:
			// 使用 JavaScript 检查，避免触发页面交互导致失焦
			result, err := a.page.Eval(`() => {
				// 检查是否有用户头像（登录成功标志）
				const avatar = document.querySelector('.user-avatar') ||
							   document.querySelector('[class*="avatar"]');
				// 检查登录弹窗是否还存在
				const loginBox = document.querySelector('.login-box');

				return {
					hasAvatar: !!avatar,
					hasLoginBox: !!loginBox
				};
			}`)
			if err != nil {
				continue
			}

			var status struct {
				HasAvatar   bool `json:"hasAvatar"`
				HasLoginBox bool `json:"hasLoginBox"`
			}
			if err := result.Value.Unmarshal(&status); err != nil {
				continue
			}

			// 如果有头像且登录弹窗消失，说明登录成功
			if status.HasAvatar && !status.HasLoginBox {
				logrus.Info("检测到登录成功")
				return true
			}

			// 如果登录弹窗消失了但没有头像，刷新页面检查
			if !status.HasLoginBox && !status.HasAvatar {
				a.page.MustReload()
				time.Sleep(1 * time.Second)
				// 再次检查
				result2, err := a.page.Eval(`() => {
					const avatar = document.querySelector('.user-avatar') ||
								   document.querySelector('[class*="avatar"]');
					return !!avatar;
				}`)
				if err == nil {
					hasAvatar := result2.Value.Bool()
					if hasAvatar {
						logrus.Info("检测到登录成功")
						return true
					}
				}
			}
		}
	}
}

// Logout 退出登录
func (a *LoginAction) Logout(ctx context.Context) error {
	pp := a.page.Context(ctx).Timeout(30 * time.Second)

	// 点击用户头像
	avatarSelectors := []string{
		".user-avatar",
		".avatar-img",
		"[class*='avatar']",
	}

	var avatar *rod.Element
	var err error
	for _, selector := range avatarSelectors {
		avatar, err = pp.Element(selector)
		if err == nil && avatar != nil {
			break
		}
	}

	if avatar == nil {
		return apperrors.ErrElementNotFound("用户头像", err)
	}

	if err := avatar.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return stderrors.Wrap(err, "点击用户头像失败")
	}

	time.Sleep(1 * time.Second)

	// 点击退出登录按钮
	logoutSelectors := []string{
		".logout-btn",
		"[class*='logout']",
		".sign-out",
	}

	for _, selector := range logoutSelectors {
		logoutBtn, err := pp.Element(selector)
		if err == nil && logoutBtn != nil {
			if err := logoutBtn.Click(proto.InputMouseButtonLeft, 1); err == nil {
				logrus.Info("已退出登录")
				configs.SetUsername("")
				return nil
			}
		}
	}

	return apperrors.ErrElementNotFound("退出登录按钮", nil)
}

// SaveCookies 保存当前页面的 Cookies 到文件
func (a *LoginAction) SaveCookies() error {
	// 获取当前页面的所有 Cookies
	result, err := proto.NetworkGetCookies{}.Call(a.page)
	if err != nil {
		return stderrors.Wrap(err, "获取 Cookies 失败")
	}

	if len(result.Cookies) == 0 {
		logrus.Warn("没有找到任何 Cookies")
		return nil
	}

	// 转换为可序列化的格式
	cookiesData := make([]map[string]interface{}, 0, len(result.Cookies))
	for _, cookie := range result.Cookies {
		cookieMap := map[string]interface{}{
			"name":     cookie.Name,
			"value":    cookie.Value,
			"domain":   cookie.Domain,
			"path":     cookie.Path,
			"secure":   cookie.Secure,
			"httpOnly": cookie.HTTPOnly,
			"sameSite": string(cookie.SameSite),
			"expires":  float64(cookie.Expires),
		}
		cookiesData = append(cookiesData, cookieMap)
	}

	// 序列化为 JSON
	data, err := json.MarshalIndent(cookiesData, "", "  ")
	if err != nil {
		return stderrors.Wrap(err, "序列化 Cookies 失败")
	}

	// 保存到文件
	loader := cookies.NewLoadCookie(cookies.GetCookiesFilePath())
	if err := loader.SaveCookies(data); err != nil {
		return stderrors.Wrap(err, "保存 Cookies 到文件失败")
	}

	logrus.Infof("✅ 已保存 %d 个 Cookies 到 %s", len(cookiesData), cookies.GetCookiesFilePath())
	return nil
}

// LoadCookies 从文件加载 Cookies 到页面
func (a *LoginAction) LoadCookies() error {
	loader := cookies.NewLoadCookie(cookies.GetCookiesFilePath())
	data, err := loader.LoadCookies()
	if err != nil {
		return stderrors.Wrap(err, "加载 Cookies 文件失败")
	}

	if len(data) == 0 || string(data) == "[]" {
		logrus.Info("没有已保存的 Cookies")
		return nil
	}

	var cookiesData []map[string]interface{}
	if err := json.Unmarshal(data, &cookiesData); err != nil {
		return stderrors.Wrap(err, "解析 Cookies 数据失败")
	}

	// 转换为 proto.NetworkCookieParam 格式
	cookieParams := make([]*proto.NetworkCookieParam, 0, len(cookiesData))
	for _, c := range cookiesData {
		param := &proto.NetworkCookieParam{
			Name:     c["name"].(string),
			Value:    c["value"].(string),
			Domain:   c["domain"].(string),
			Path:     getStringOrDefault(c, "path", "/"),
			Secure:   getBoolOrDefault(c, "secure", false),
			HTTPOnly: getBoolOrDefault(c, "httpOnly", false),
		}

		if sameSite, ok := c["sameSite"].(string); ok {
			param.SameSite = proto.NetworkCookieSameSite(sameSite)
		}

		if expires, ok := c["expires"]; ok {
			switch v := expires.(type) {
			case float64:
				param.Expires = proto.TimeSinceEpoch(v)
			case int64:
				param.Expires = proto.TimeSinceEpoch(v)
			}
		}

		cookieParams = append(cookieParams, param)
	}

	// 设置 Cookies
	if err := a.page.SetCookies(cookieParams); err != nil {
		return stderrors.Wrap(err, "设置 Cookies 失败")
	}

	logrus.Infof("✅ 已加载 %d 个 Cookies", len(cookieParams))
	return nil
}

// getStringOrDefault 从 map 中获取字符串，如果不存在则返回默认值
func getStringOrDefault(m map[string]interface{}, key string, defaultValue string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultValue
}

// getBoolOrDefault 从 map 中获取布尔值，如果不存在则返回默认值
func getBoolOrDefault(m map[string]interface{}, key string, defaultValue bool) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// ClearCookies 清除保存的 Cookies
func (a *LoginAction) ClearCookies() error {
	loader := cookies.NewLoadCookie(cookies.GetCookiesFilePath())
	// 保存空数组来清除
	if err := loader.SaveCookies([]byte("[]")); err != nil {
		return stderrors.Wrap(err, "清除 Cookies 失败")
	}

	// 同时清除浏览器中的 Cookies
	err := proto.NetworkClearBrowserCookies{}.Call(a.page)
	if err != nil {
		logrus.Warnf("清除浏览器 Cookies 失败: %v", err)
	}

	logrus.Info("✅ 已清除 Cookies")
	return nil
}

// PrintQRCodeInTerminal 在终端打印二维码并保存图片文件
// dataURL 格式: data:image/png;base64,xxxxx
func (a *LoginAction) PrintQRCodeInTerminal(dataURL string) error {
	// 检查是否是 base64 图片数据
	if !strings.HasPrefix(dataURL, "data:image/") {
		return stderrors.New("不是有效的图片 data URL")
	}

	// 提取 base64 数据
	base64Index := strings.Index(dataURL, ",")
	if base64Index == -1 {
		return stderrors.New("无效的 data URL 格式")
	}
	base64Data := dataURL[base64Index+1:]

	// 解码 base64
	imgData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return stderrors.Wrap(err, "解码 base64 失败")
	}

	// 保存二维码图片文件
	qrcodeFile := "./qrcode.png"
	if err := os.WriteFile(qrcodeFile, imgData, 0644); err != nil {
		return stderrors.Wrap(err, "保存二维码图片失败")
	}

	logrus.Info("")
	logrus.Info("╔══════════════════════════════════════════════════════════════╗")
	logrus.Info("║                        请扫描二维码                           ║")
	logrus.Info("╠══════════════════════════════════════════════════════════════╣")
	logrus.Info("║  二维码已保存到: ./qrcode.png                                ║")
	logrus.Info("║  请用小黑盒 APP 扫描该图片                                   ║")
	logrus.Info("╚══════════════════════════════════════════════════════════════╝")
	logrus.Info("")

	// 尝试用系统默认程序打开图片
	go func() {
		// macOS 使用 open 命令
		_ = exec.Command("open", qrcodeFile).Start()
	}()

	return nil
}
