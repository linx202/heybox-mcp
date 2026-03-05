package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

func main() {
	// 启动浏览器
	fmt.Println("启动浏览器...")
	u := launcher.New().
		Headless(false).
		MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage("")
	page.MustSetViewport(1920, 1080, 1, false)

	// 1. 加载已保存的 cookies
	cookieFile := "./cookies/cookies.json"
	loadCookies(page, cookieFile)

	// 2. 访问发布图文页面
	fmt.Println("\n访问发布图文页面...")
	page.MustNavigate("https://www.xiaoheihe.cn/creator/editor/draft/image_text").MustWaitLoad()
	time.Sleep(8 * time.Second) // SPA 页面需要更长等待时间

	// 等待页面空闲
	page.Timeout(30 * time.Second).MustWaitIdle()
	time.Sleep(3 * time.Second)

	// 3. 检查登录状态
	fmt.Println("\n检查登录状态...")
	for !checkAndEnsureLogin(page) {
		fmt.Println("等待登录...")
		time.Sleep(2 * time.Second)
	}

	// 4. 登录成功，保存 cookies
	fmt.Println("\n登录成功，保存 cookies...")
	saveCookies(page, cookieFile)

	// 5. 抓取选择器
	fmt.Println("\n\n========== 抓取页面选择器 ==========\n")
	scrapeAllClasses(page, "发布图文页面")

	fmt.Println("\n\n========== 分析完成 ==========")
	fmt.Println("按 Ctrl+C 退出...")
	select {}
}

func checkAndEnsureLogin(page *rod.Page) bool {
	// 检查页面是否显示未登录提示
	result, _ := page.Eval(`() => {
		// 检查是否有未登录提示
		const unauthTips = document.querySelector('.unauth-tips') ||
						   document.querySelector('[class*="unauth"]') ||
						   document.querySelector('[class*="login-tip"]');

		// 检查是否有用户信息（已登录）
		const userInfo = document.querySelector('.user-box__avatar') ||
						 document.querySelector('[class*="avatar"]') ||
						 document.querySelector('.user-info');

		// 检查是否有快捷登录按钮
		const quickLoginBtn = document.querySelector('.user-box__login');

		// 检查是否有登录弹窗
		const loginModal = document.querySelector('.login-box') ||
						   document.querySelector('[class*="login-modal"]');

		return {
			needLogin: !!unauthTips || (!userInfo && !loginModal),
			hasQuickLogin: !!quickLoginBtn,
			hasLoginModal: !!loginModal,
			hasUserInfo: !!userInfo
		};
	}`)

	var status struct {
		NeedLogin     bool `json:"needLogin"`
		HasQuickLogin bool `json:"hasQuickLogin"`
		HasLoginModal bool `json:"hasLoginModal"`
		HasUserInfo   bool `json:"hasUserInfo"`
	}
	result.Value.Unmarshal(&status)

	fmt.Printf("状态: needLogin=%v, hasQuickLogin=%v, hasLoginModal=%v, hasUserInfo=%v\n",
		status.NeedLogin, status.HasQuickLogin, status.HasLoginModal, status.HasUserInfo)

	// 如果已有用户信息，说明已登录
	if status.HasUserInfo {
		fmt.Println("✅ 已登录")
		return true
	}

	// 如果有登录弹窗，等待扫码
	if status.HasLoginModal {
		fmt.Println("检测到登录弹窗，等待扫码...")
		return waitForScanLogin(page)
	}

	// 如果需要登录且有快捷登录按钮
	if status.NeedLogin && status.HasQuickLogin {
		fmt.Println("点击快捷登录按钮...")
		clickResult, _ := page.Eval(`() => {
			const btn = document.querySelector('.user-box__login');
			if (btn) {
				btn.click();
				return { clicked: true };
			}
			return { clicked: false };
		}`)
		var cr struct{ Clicked bool }
		clickResult.Value.Unmarshal(&cr)

		if cr.Clicked {
			time.Sleep(2 * time.Second)
			// 重新检查
			return checkAndEnsureLogin(page)
		}
	}

	// 如果需要登录但没有快捷登录按钮，尝试点击普通登录按钮
	if status.NeedLogin {
		fmt.Println("尝试点击登录按钮...")
		doLoginFlow(page)
		return false
	}

	return false
}

func doLoginFlow(page *rod.Page) {
	// 等待页面稳定
	time.Sleep(2 * time.Second)

	// 点击登录按钮
	result, _ := page.Eval(`() => {
		// 查找登录按钮
		const selectors = [
			'.user-box__login',
			'[class*="login-btn"]',
			'button'
		];

		for (const sel of selectors) {
			const btns = document.querySelectorAll(sel);
			for (const btn of btns) {
				if (btn.textContent?.includes('登录') || btn.className?.includes('login')) {
					btn.click();
					return { clicked: true, selector: sel };
				}
			}
		}
		return { clicked: false };
	}`)

	var cr struct {
		Clicked  bool   `json:"clicked"`
		Selector string `json:"selector"`
	}
	result.Value.Unmarshal(&cr)

	if !cr.Clicked {
		fmt.Println("未找到登录按钮")
		return
	}

	fmt.Printf("已点击登录按钮: %s\n", cr.Selector)

	// 等待登录弹窗动画
	time.Sleep(3 * time.Second)

	// 等待并获取二维码
	waitForQrcodeAndShow(page)
}

func waitForQrcodeAndShow(page *rod.Page) {
	// 等待二维码出现（SPA 需要更长时间）
	fmt.Println("等待二维码加载...")

	// 多次尝试获取二维码
	for i := 0; i < 5; i++ {
		time.Sleep(2 * time.Second)

		result, _ := page.Eval(`() => {
			// 查找二维码 canvas
			const canvas = document.querySelector('.login-qrcode canvas') ||
						   document.querySelector('canvas[class*="qr"]') ||
						   document.querySelector('.login-box canvas') ||
						   document.querySelector('#login-qrcode');

			if (canvas && canvas.tagName === 'CANVAS') {
				return {
					found: true,
					type: 'canvas',
					data: canvas.toDataURL('image/png'),
					className: canvas.className
				};
			}

			// 查找二维码图片
			const img = document.querySelector('.login-qrcode img') ||
						document.querySelector('img[class*="qr"]');
			if (img) {
				return {
					found: true,
					type: 'img',
					data: img.src,
					className: img.className
				};
			}

			return { found: false };
		}`)

		var qr struct {
			Found     bool   `json:"found"`
			Type      string `json:"type"`
			Data      string `json:"data"`
			ClassName string `json:"className"`
		}
		result.Value.Unmarshal(&qr)

		if qr.Found {
			fmt.Printf("找到二维码: type=%s class=%s\n", qr.Type, qr.ClassName)

			// 保存并显示二维码
			if qr.Type == "canvas" && strings.HasPrefix(qr.Data, "data:image/") {
				base64Index := strings.Index(qr.Data, ",")
				if base64Index > -1 {
					imgData, err := base64.StdEncoding.DecodeString(qr.Data[base64Index+1:])
					if err == nil {
						// 带日期后缀
						timestamp := time.Now().Format("20060102_150405")
						qrcodeFile := fmt.Sprintf("./qrcode_%s.png", timestamp)
						os.WriteFile(qrcodeFile, imgData, 0644)
						fmt.Println("\n╔══════════════════════════════════════════════════════════════╗")
						fmt.Println("║                        请扫描二维码                           ║")
						fmt.Printf("║  二维码已保存到: %s              ║\n", qrcodeFile)
						fmt.Println("╚══════════════════════════════════════════════════════════════╝")
						exec.Command("open", qrcodeFile).Start()
						return
					}
				}
			}
		}

		fmt.Printf("第 %d 次尝试未找到二维码，继续等待...\n", i+1)
	}

	fmt.Println("未找到二维码")
}

func waitForScanLogin(page *rod.Page) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("登录超时")
			return false
		case <-ticker.C:
			result, _ := page.Eval(`() => {
				const userInfo = document.querySelector('.user-box__avatar') ||
								 document.querySelector('[class*="avatar"]');
				const loginModal = document.querySelector('.login-box') ||
								   document.querySelector('[class*="login-modal"]');
				return {
					hasUserInfo: !!userInfo,
					hasLoginModal: !!loginModal
				};
			}`)

			var s struct {
				HasUserInfo   bool `json:"hasUserInfo"`
				HasLoginModal bool `json:"hasLoginModal"`
			}
			result.Value.Unmarshal(&s)

			if s.HasUserInfo && !s.HasLoginModal {
				return true
			}
		}
	}
}

func loadCookies(page *rod.Page, file string) {
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("没有已保存的 cookies")
		return
	}

	var cookies []map[string]interface{}
	if err := json.Unmarshal(data, &cookies); err != nil {
		return
	}

	fmt.Printf("加载 %d 个 cookies...\n", len(cookies))
	for _, c := range cookies {
		cookie := &proto.NetworkCookieParam{
			Name:     fmt.Sprintf("%v", c["name"]),
			Value:    fmt.Sprintf("%v", c["value"]),
			Domain:   fmt.Sprintf("%v", c["domain"]),
			Path:     getString(c, "path", "/"),
			Secure:   getBool(c, "secure"),
			HTTPOnly: getBool(c, "httpOnly"),
		}
		page.MustSetCookies(cookie)
	}
}

func saveCookies(page *rod.Page, file string) {
	result, err := proto.NetworkGetCookies{}.Call(page)
	if err != nil {
		fmt.Printf("获取 cookies 失败: %v\n", err)
		return
	}

	cookiesData := make([]map[string]interface{}, 0, len(result.Cookies))
	for _, cookie := range result.Cookies {
		cookieMap := map[string]interface{}{
			"name":     cookie.Name,
			"value":    cookie.Value,
			"domain":   cookie.Domain,
			"path":     cookie.Path,
			"secure":   cookie.Secure,
			"httpOnly": cookie.HTTPOnly,
		}
		cookiesData = append(cookiesData, cookieMap)
	}

	data, _ := json.MarshalIndent(cookiesData, "", "  ")
	os.MkdirAll("./cookies", 0755)
	os.WriteFile(file, data, 0644)
	fmt.Printf("✅ 已保存 %d 个 cookies\n", len(cookiesData))
}

func scrapeAllClasses(page *rod.Page, pageName string) {
	result, _ := page.Eval(`() => {
		const results = {};
		results.url = location.href;
		results.title = document.title;

		// 获取所有 class
		const allElements = document.querySelectorAll('*');
		const classSet = new Set();

		allElements.forEach(el => {
			if (el.className && typeof el.className === 'string') {
				el.className.split(' ').forEach(c => {
					if (c && c.length > 2) classSet.add(c);
				});
			}
		});

		results.classes = Array.from(classSet).sort();

		// 关键元素
		results.elements = [];

		document.querySelectorAll('input, textarea, [contenteditable="true"], button').forEach(el => {
			results.elements.push({
				tag: el.tagName,
				className: el.className,
				id: el.id,
				type: el.type || el.getAttribute('contenteditable'),
				placeholder: el.placeholder,
				text: el.textContent?.trim().substring(0, 30)
			});
		});

		return JSON.stringify(results);
	}`)

	// 解析 JSON
	jsonStr := result.Value.String()
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		fmt.Printf("解析失败: %v, raw: %s\n", err, jsonStr)
		return
	}

	fmt.Printf("页面: %s\n", pageName)
	fmt.Printf("URL: %v\n", data["url"])
	fmt.Printf("Title: %v\n", data["title"])

	fmt.Println("\n=== 相关 Class ===")
	if classes, ok := data["classes"].([]interface{}); ok {
		keywords := []string{"login", "avatar", "btn", "input", "editor", "title", "upload", "content", "user", "publish", "submit"}
		for _, c := range classes {
			cls := strings.ToLower(fmt.Sprintf("%v", c))
			for _, kw := range keywords {
				if strings.Contains(cls, kw) {
					fmt.Printf("  .%s\n", c)
					break
				}
			}
		}
	}

	fmt.Println("\n=== 关键元素 ===")
	if elements, ok := data["elements"].([]interface{}); ok {
		for _, el := range elements {
			if m, ok := el.(map[string]interface{}); ok {
				fmt.Printf("  %v", m["tag"])
				if m["className"] != "" && m["className"] != nil {
					fmt.Printf(" .%v", m["className"])
				}
				if m["placeholder"] != "" && m["placeholder"] != nil {
					fmt.Printf(" placeholder='%v'", m["placeholder"])
				}
				if m["text"] != "" && m["text"] != nil {
					fmt.Printf(" text='%v'", m["text"])
				}
				fmt.Println()
			}
		}
	}
}

func getString(m map[string]interface{}, key string, def string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
