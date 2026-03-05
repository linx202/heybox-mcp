// +build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/yourusername/heybox-mcp/browser"
	"github.com/yourusername/heybox-mcp/configs"
	"github.com/yourusername/heybox-mcp/cookies"
	"github.com/yourusername/heybox-mcp/heybox"
)

func main() {
	fmt.Println("=== 验证登录状态 ===")
	fmt.Println()

	// 创建浏览器（非 headless 模式，方便观察）
	b := browser.NewBrowser(
		false, // 显示浏览器窗口
		browser.WithBinPath(configs.GetBinPath()),
	)
	defer b.Close()

	page := b.MustPage("")
	defer page.Close()

	// 加载 Cookies
	loginAction := heybox.NewLogin(page)
	if err := loginAction.LoadCookies(); err != nil {
		fmt.Printf("⚠️ 加载 Cookies 失败: %v\n", err)
	} else {
		fmt.Println("✅ Cookies 已加载到页面")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second) // 5分钟超时
	defer cancel()

	// 先导航到首页检查登录状态
	fmt.Println("\n正在导航到首页...")
	navigator := heybox.NewNavigator(page)
	if err := navigator.NavigateToHome(ctx); err != nil {
		fmt.Printf("❌ 导航失败: %v\n", err)
		return
	}

	// 等待页面加载
	time.Sleep(3 * time.Second)

	// 检查登录状态
	fmt.Println("\n检查登录状态...")
	isLoggedIn, err := loginAction.CheckLoginStatusOnCurrentPage(ctx)
	if err != nil {
		fmt.Printf("❌ 检查登录状态失败: %v\n", err)
		return
	}

	if !isLoggedIn {
		fmt.Println("❌ 未登录，正在获取二维码...")

		qrcodeURL, alreadyLoggedIn, err := loginAction.FetchQrcodeImage(ctx)
		if err != nil {
			fmt.Printf("❌ 获取二维码失败: %v\n", err)
			return
		}

		if alreadyLoggedIn {
			fmt.Println("✅ 已登录")
		} else {
			// 保存并显示二维码
			if err := loginAction.PrintQRCodeInTerminal(qrcodeURL); err != nil {
				fmt.Printf("❌ 保存二维码失败: %v\n", err)
				return
			}

			fmt.Println("\n请扫描二维码登录...")
			fmt.Println("登录成功后会自动保存 Cookies")

			// 等待登录完成
			if loginAction.WaitForLogin(ctx) {
				fmt.Println("✅ 登录成功！")

				// 保存 Cookies
				if err := loginAction.SaveCookies(); err != nil {
					fmt.Printf("⚠️ 保存 Cookies 失败: %v\n", err)
				}

				// 刷新页面以获取用户信息
				fmt.Println("\n刷新页面获取用户信息...")
				page.MustReload()
				time.Sleep(3 * time.Second)
			} else {
				fmt.Println("❌ 登录超时")
				return
			}
		}
	}

	// 获取用户信息
	fmt.Println("\n=== 获取用户信息 ===")
	userInfo := getUserInfo(page)

	jsonData, _ := json.MarshalIndent(userInfo, "", "  ")
	fmt.Println(string(jsonData))

	// 保存最新的 Cookies
	fmt.Println("\n保存 Cookies...")
	if err := loginAction.SaveCookies(); err != nil {
		fmt.Printf("⚠️ 保存 Cookies 失败: %v\n", err)
	} else {
		fmt.Println("✅ Cookies 已保存")
	}

	// 显示当前 Cookies 数量
	cookieLoader := cookies.NewLoadCookie(cookies.GetCookiesFilePath())
	cookieData, _ := cookieLoader.LoadCookies()
	var cookiesList []map[string]interface{}
	json.Unmarshal(cookieData, &cookiesList)
	fmt.Printf("✅ 共保存 %d 个 Cookies\n", len(cookiesList))

	// 测试发布动态
	fmt.Println("\n=== 测试发布动态 ===")
	testPublish(ctx, page)
}

func testPublish(ctx context.Context, page *rod.Page) {
	// 导航到发布页面
	navigator := heybox.NewNavigator(page)
	fmt.Println("正在导航到发布页面...")
	if err := navigator.NavigateToPublish(ctx); err != nil {
		fmt.Printf("❌ 导航失败: %v\n", err)
		return
	}

	time.Sleep(3 * time.Second)

	// 创建发布操作
	publishAction, err := heybox.NewPublishAction(page)
	if err != nil {
		fmt.Printf("❌ 创建发布操作失败: %v\n", err)
		return
	}

	// 检查登录状态
	fmt.Println("检查发布页面登录状态...")
	loginResult, err := publishAction.CheckLoginAndFetchQrcode(ctx)
	if err != nil {
		fmt.Printf("❌ 检查登录状态失败: %v\n", err)
		return
	}

	if loginResult.NeedLogin {
		fmt.Println("❌ 发布页面需要重新登录")
		return
	}

	fmt.Println("✅ 已登录，开始发布...")

	// 执行发布
	content := heybox.PublishContent{
		Title:      "测试发布动态 - " + time.Now().Format("15:04:05"),
		Content:    "这是一条通过 heybox-mcp 测试发布的动态内容。\n\n这是正文内容，测试时间: " + time.Now().Format("2006-01-02 15:04:05"),
		Tags:       []string{},
		ImagePaths: []string{},
	}

	if err := publishAction.Publish(ctx, content); err != nil {
		fmt.Printf("❌ 发布失败: %v\n", err)
		return
	}

	// 验证发布结果
	time.Sleep(3 * time.Second)
	success, msg, err := publishAction.ValidatePublish(ctx)

	// 获取 API 响应
	apiResp := publishAction.GetAPIResponse()

	result := map[string]interface{}{
		"success": success,
		"message": msg,
	}

	if err != nil {
		result["error"] = err.Error()
	}

	// 添加 API 响应信息
	if apiResp != nil {
		result["api_response"] = map[string]interface{}{
			"status":  apiResp.Status,
			"msg":     apiResp.Msg,
			"version": apiResp.Version,
			"result":  apiResp.Result,
		}
	}

	jsonData, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println()
	fmt.Println("发布结果:")
	fmt.Println(string(jsonData))
}

func getUserInfo(page *rod.Page) map[string]interface{} {
	info := map[string]interface{}{
		"logged_in": false,
	}

	// 使用 JavaScript 获取用户信息
	result, err := page.Eval(`() => {
		const info = {
			logged_in: false,
			user_id: '',
			nickname: '',
			avatar: '',
			debug: {}
		};

		// 打印所有可能包含用户信息的元素
		info.debug.potentialAvatars = [];
		document.querySelectorAll('img').forEach(img => {
			if (img.className && (img.className.includes('avatar') || img.className.includes('user'))) {
				info.debug.potentialAvatars.push({class: img.className, src: img.src});
			}
		});

		info.debug.potentialNames = [];
		document.querySelectorAll('[class*="user"], [class*="name"], [class*="nick"]').forEach(el => {
			info.debug.potentialNames.push({tag: el.tagName, class: el.className, text: el.textContent?.substring(0, 50)});
		});

		// 检查 __NEXT_DATA__ 或类似的全局变量
		info.debug.hasNextData = !!window.__NEXT_DATA__;
		info.debug.hasInitialState = !!window.__INITIAL_STATE__;

		if (window.__NEXT_DATA__) {
			info.debug.nextData = JSON.stringify(window.__NEXT_DATA__).substring(0, 500);
		}

		// 检查 localStorage
		try {
			const localStorageKeys = Object.keys(localStorage);
			info.debug.localStorageKeys = localStorageKeys.filter(k =>
				k.includes('user') || k.includes('token') || k.includes('auth')
			);
		} catch(e) {
			info.debug.localStorageError = e.message;
		}

		// 尝试从页面获取用户信息
		// 检查是否有用户头像
		const avatarEl = document.querySelector('.user-avatar img') ||
						 document.querySelector('.avatar-img') ||
						 document.querySelector('img[class*="avatar"]') ||
						 document.querySelector('.user-info img');

		if (avatarEl) {
			info.logged_in = true;
			info.avatar = avatarEl.src || '';
		}

		// 尝试获取昵称
		const nicknameEl = document.querySelector('.user-name') ||
						   document.querySelector('.nickname') ||
						   document.querySelector('.username') ||
						   document.querySelector('[class*="nickname"]');

		if (nicknameEl) {
			info.nickname = (nicknameEl.textContent || nicknameEl.innerText || '').trim();
		}

		// 尝试从页面数据中获取用户 ID
		try {
			if (window.__INITIAL_STATE__) {
				const state = window.__INITIAL_STATE__;
				// 优先从 userinfo 获取
				if (state.userinfo) {
					info.user_id = state.userinfo.heybox_id || state.userinfo.id || '';
					info.nickname = state.userinfo.nickname || state.userinfo.name || '';
				}
				// 如果没有，尝试其他字段
				if (!info.user_id) {
					info.user_id = state.userId || state.user_id || (state.user && state.user.id) || '';
				}
				if (!info.nickname) {
					info.nickname = (state.user && state.user.nickname) || '';
				}
				info.debug.initialState = JSON.stringify(state).substring(0, 500);
			}
		} catch(e) {
			info.debug.stateError = e.message;
		}

		// 尝试从 __NEXT_DATA__ 获取
		try {
			if (window.__NEXT_DATA__ && window.__NEXT_DATA__.props && window.__NEXT_DATA__.props.pageProps) {
				const pageProps = window.__NEXT_DATA__.props.pageProps;
				if (pageProps.userInfo) {
					info.user_id = pageProps.userInfo.userId || pageProps.userInfo.id || info.user_id;
					info.nickname = pageProps.userInfo.nickname || pageProps.userInfo.name || info.nickname;
				}
			}
		} catch(e) {}

		return info;
	}`)

	if err != nil {
		info["error"] = err.Error()
		return info
	}

	_ = result.Value.Unmarshal(&info)
	return info
}
