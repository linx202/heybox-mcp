package heybox

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	stderrors "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apperrors "github.com/yourusername/heybox-mcp/errors"
)

// PublishContent 发布内容结构
type PublishContent struct {
	Title      string   // 标题
	Content    string   // 正文内容
	Tags       []string // 话题标签
	ImagePaths []string // 本地图片路径列表
}

// PublishActionResult 发布操作结果
type PublishActionResult struct {
	NeedLogin   bool   `json:"need_login"`    // 是否需要登录
	QrcodeURL   string `json:"qrcode_url"`    // 二维码 URL（需要登录时）
	QrcodeSaved bool   `json:"qrcode_saved"`  // 二维码是否已保存到文件
}

// PublishAPIResponse 发布 API 响应
type PublishAPIResponse struct {
	Status  string                 `json:"status"`       // "ok" 或 "failed"
	Msg     string                 `json:"msg"`          // 响应消息
	Version string                 `json:"version"`      // 版本号
	Result  map[string]interface{} `json:"result"`       // 结果数据
}

// PublishAction 发布操作
type PublishAction struct {
	page         *rod.Page
	navigator    *Navigator
	apiResponse  *PublishAPIResponse
	apiMutex     sync.Mutex
}

// NewPublishAction 创建发布操作实例
func NewPublishAction(page *rod.Page) (*PublishAction, error) {
	pp := page.Timeout(300 * time.Second) // 5分钟超时

	navigator := NewNavigator(page)

	// 导航到发布页面
	if err := navigator.NavigateToPublish(context.Background()); err != nil {
		return nil, stderrors.Wrap(err, "导航到发布页面失败")
	}

	// 等待页面加载完成
	time.Sleep(2 * time.Second)

	return &PublishAction{
		page:      pp,
		navigator: navigator,
	}, nil
}

// CheckLoginAndFetchQrcode 检查登录状态，如果未登录则获取二维码
func (p *PublishAction) CheckLoginAndFetchQrcode(ctx context.Context) (*PublishActionResult, error) {
	pp := p.page.Context(ctx).Timeout(30 * time.Second)

	// 检查是否有登录弹框
	loginBoxSelectors := []string{
		".login-box",
		".login-modal",
		".modal-login",
		"[class*='login'][class*='modal']",
		"[class*='login'][class*='dialog']",
	}

	var loginBox *rod.Element
	for _, selector := range loginBoxSelectors {
		has, elem, _ := pp.Has(selector)
		if has && elem != nil {
			loginBox = elem
			logrus.Infof("检测到登录弹框: %s", selector)
			break
		}
	}

	// 如果没有登录弹框，检查是否有发布相关元素（表示已登录）
	if loginBox == nil {
		// 检查是否有发布页面的特征元素
		publishElements := []string{
			"input[type='file']",
			".title-input",
			"textarea",
			"[contenteditable='true']",
		}

		for _, selector := range publishElements {
			has, _, _ := pp.Has(selector)
			if has {
				logrus.Info("已检测到发布页面元素，用户已登录")
				return &PublishActionResult{NeedLogin: false}, nil
			}
		}

		// 如果也没有发布元素，可能页面加载异常
		logrus.Warn("未检测到登录弹框或发布页面元素")
	}

	// 有登录弹框，需要获取二维码
	logrus.Info("检测到需要登录，正在获取二维码...")

	// 点击二维码图标触发二维码请求
	qrcodeIconSelectors := []string{
		".login-box .title img",
		".login-box img",
		".title img",
		"[class*='qrcode']",
	}

	for _, selector := range qrcodeIconSelectors {
		qrcodeIcon, err := pp.Element(selector)
		if err == nil && qrcodeIcon != nil {
			logrus.Infof("找到二维码图标: %s", selector)
			// 使用 JavaScript 点击
			_, _ = qrcodeIcon.Eval(`() => this.click()`)
			break
		}
	}

	// 等待二维码生成
	time.Sleep(2 * time.Second)

	// 获取二维码 Canvas
	qrcodeCanvasSelectors := []string{
		"#login-qrcode",
		"canvas#login-qrcode",
		".login-box canvas",
		"canvas[class*='qrcode']",
	}

	var qrcodeCanvas *rod.Element
	for _, selector := range qrcodeCanvasSelectors {
		canvas, err := pp.Element(selector)
		if err == nil && canvas != nil {
			qrcodeCanvas = canvas
			logrus.Infof("找到二维码 Canvas: %s", selector)
			break
		}
	}

	if qrcodeCanvas == nil {
		return &PublishActionResult{
			NeedLogin: true,
		}, stderrors.New("未找到二维码元素，请手动刷新页面")
	}

	// 获取二维码 Base64 数据
	base64Data, err := qrcodeCanvas.Eval(`() => this.toDataURL('image/png')`)
	if err != nil {
		return &PublishActionResult{
			NeedLogin: true,
		}, stderrors.Wrap(err, "获取二维码数据失败")
	}

	qrcodeURL := base64Data.Value.String()

	// 保存二维码到文件
	qrcodeSaved := false
	if qrcodeURL != "" {
		if err := saveQrcodeToFile(qrcodeURL); err == nil {
			qrcodeSaved = true
			logrus.Info("✅ 二维码已保存到 ./qrcode.png")
		}
	}

	return &PublishActionResult{
		NeedLogin:   true,
		QrcodeURL:   qrcodeURL,
		QrcodeSaved: qrcodeSaved,
	}, nil
}

// saveQrcodeToFile 保存二维码到文件
func saveQrcodeToFile(dataURL string) error {
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
	return os.WriteFile("./qrcode.png", imgData, 0644)
}

// Publish 发布动态
func (p *PublishAction) Publish(ctx context.Context, content PublishContent) error {
	if len(content.ImagePaths) == 0 && len(content.Content) == 0 {
		return apperrors.ErrInvalidParameter("图片和内容不能同时为空")
	}

	page := p.page.Context(ctx)

	// 启动网络请求监听，截取发布 API 响应
	go p.listenPublishAPI(ctx)

	// 1. 上传图片
	if len(content.ImagePaths) > 0 {
		if err := p.uploadImages(page, content.ImagePaths); err != nil {
			return apperrors.ErrUploadFailed("上传图片失败", err)
		}
		logrus.Infof("已上传 %d 张图片", len(content.ImagePaths))
	}

	// 2. 输入标题
	if content.Title != "" {
		if err := p.inputTitle(page, content.Title); err != nil {
			return stderrors.Wrap(err, "输入标题失败")
		}
	}

	// 3. 输入正文
	if content.Content != "" {
		if err := p.inputContent(page, content.Content, content.Tags); err != nil {
			return stderrors.Wrap(err, "输入正文失败")
		}
	}

	// 4. 点击发布按钮
	if err := p.clickPublishButton(page); err != nil {
		return apperrors.ErrPublishFailed("点击发布按钮失败", err)
	}

	logrus.Info("发布完成，等待服务器响应...")
	time.Sleep(3 * time.Second)

	return nil
}

// listenPublishAPI 监听发布 API 响应
func (p *PublishAction) listenPublishAPI(ctx context.Context) {
	// 监听网络响应
	go func() {
		_ = rod.Try(func() {
			p.page.EachEvent(func(e *proto.NetworkResponseReceived) {
				// 只处理发布相关的 API
				if strings.Contains(e.Response.URL, "api.xiaoheihe.cn") &&
				   (strings.Contains(e.Response.URL, "/post") || strings.Contains(e.Response.URL, "/link/post")) {
					logrus.Infof("捕获到发布 API 响应: %s", e.Response.URL)

					// 获取响应体
					body, err := proto.NetworkGetResponseBody{RequestID: e.RequestID}.Call(p.page)
					if err != nil {
						logrus.Warnf("获取响应体失败: %v", err)
						return
					}

					logrus.Infof("API 响应内容: %s", body.Body)

					// 解析响应
					var resp PublishAPIResponse
					if err := json.Unmarshal([]byte(body.Body), &resp); err == nil {
						p.apiMutex.Lock()
						p.apiResponse = &resp
						p.apiMutex.Unlock()

						// 打印解析结果
						logrus.Infof("API 响应解析: status=%d, msg=%s, errno=%d, is_login=%d",
							resp.Status, resp.Msg, resp.Errno, resp.IsLogin)
					}
				}
			})()
		})
	}()
}

// GetAPIResponse 获取 API 响应
func (p *PublishAction) GetAPIResponse() *PublishAPIResponse {
	p.apiMutex.Lock()
	defer p.apiMutex.Unlock()
	return p.apiResponse
}

// uploadImages 上传图片
func (p *PublishAction) uploadImages(page *rod.Page, imagePaths []string) error {
	pp := page.Timeout(60 * time.Second)

	// 验证文件存在
	validPaths := make([]string, 0, len(imagePaths))
	for _, path := range imagePaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			logrus.Warnf("图片文件不存在: %s", path)
			continue
		}
		validPaths = append(validPaths, path)
	}

	if len(validPaths) == 0 {
		return stderrors.New("没有有效的图片文件")
	}

	// ⚠️ 需要根据实际页面调整选择器
	uploadSelectors := []string{
		"input[type='file']",
		".image-upload-input",
		"[class*='upload'] input[type='file']",
		"input[accept='image*']",
	}

	var uploadInput *rod.Element
	var err error
	for _, selector := range uploadSelectors {
		uploadInput, err = pp.Element(selector)
		if err == nil && uploadInput != nil {
			logrus.Infof("找到上传输入框: %s", selector)
			break
		}
	}

	if uploadInput == nil {
		return apperrors.ErrElementNotFound("图片上传输入框", err)
	}

	// 设置文件
	if err := uploadInput.SetFiles(validPaths); err != nil {
		return stderrors.Wrap(err, "设置上传文件失败")
	}

	// 等待上传完成
	time.Sleep(3 * time.Second)

	return nil
}

// inputTitle 输入标题
func (p *PublishAction) inputTitle(page *rod.Page, title string) error {
	// 标题输入框在 editor-title__container 下
	titleSelectors := []string{
		".editor-title__container [contenteditable='true']",
		".editor-title__container .ProseMirror",
		".hb-cpt__editor-title [contenteditable='true']",
	}

	var titleInput *rod.Element
	var err error
	for _, selector := range titleSelectors {
		titleInput, err = page.Element(selector)
		if err == nil && titleInput != nil {
			logrus.Infof("找到标题输入框: %s", selector)
			break
		}
	}

	if titleInput == nil {
		return apperrors.ErrElementNotFound("标题输入框", err)
	}

	// 使用 JavaScript 点击并输入
	_, _ = titleInput.Eval(`() => this.focus()`)
	time.Sleep(200 * time.Millisecond)

	// 清空现有内容
	_, _ = titleInput.Eval(`() => { this.innerHTML = ''; }`)

	// 输入标题
	_, err = titleInput.Eval(`(title) => { this.innerHTML = title; }`, title)
	if err != nil {
		return stderrors.Wrap(err, "输入标题失败")
	}

	// 触发 input 和 blur 事件
	_, _ = titleInput.Eval(`() => {
		this.dispatchEvent(new Event('input', { bubbles: true }));
		this.dispatchEvent(new Event('change', { bubbles: true }));
	}`)

	time.Sleep(500 * time.Millisecond)
	logrus.Infof("已输入标题: %s", title)
	return nil
}

// inputContent 输入正文
func (p *PublishAction) inputContent(page *rod.Page, content string, tags []string) error {
	// 正文输入框在 image-text__edit-content--inner 下
	contentSelectors := []string{
		".image-text__edit-content--inner [contenteditable='true']",
		".image-text__edit-content--inner .ProseMirror",
		".image-text__edit-content [contenteditable='true']",
	}

	var contentInput *rod.Element
	var err error
	for _, selector := range contentSelectors {
		contentInput, err = page.Element(selector)
		if err == nil && contentInput != nil {
			logrus.Infof("找到正文输入框: %s", selector)
			break
		}
	}

	if contentInput == nil {
		return apperrors.ErrElementNotFound("正文输入框", err)
	}

	// 使用 JavaScript 点击并输入
	_, _ = contentInput.Eval(`() => this.focus()`)
	time.Sleep(200 * time.Millisecond)

	// 清空现有内容
	_, _ = contentInput.Eval(`() => { this.innerHTML = ''; }`)

	// 构建完整内容（包含标签）
	fullContent := content
	for _, tag := range tags {
		fullContent += " #" + tag
	}

	// 输入正文
	_, err = contentInput.Eval(`(content) => { this.innerHTML = content; }`, fullContent)
	if err != nil {
		return stderrors.Wrap(err, "输入正文失败")
	}

	// 触发 input 和 change 事件
	_, _ = contentInput.Eval(`() => {
		this.dispatchEvent(new Event('input', { bubbles: true }));
		this.dispatchEvent(new Event('change', { bubbles: true }));
	}`)

	time.Sleep(500 * time.Millisecond)
	logrus.Infof("已输入正文: %s", fullContent[:min(50, len(fullContent))])
	return nil
}

// clickPublishButton 点击发布按钮（真正的发布，不是保存草稿）
func (p *PublishAction) clickPublishButton(page *rod.Page) error {
	// 精确匹配发布按钮（class 包含 main-btn）
	publishSelectors := []string{
		"button.editor-publish__btn.main-btn",
		".editor-publish__btn.main-btn",
		"button[class*='main-btn']",
	}

	var publishBtn *rod.Element
	var err error
	for _, selector := range publishSelectors {
		publishBtn, err = page.Element(selector)
		if err == nil && publishBtn != nil {
			// 验证按钮文本是"发布"
			text, _ := publishBtn.Text()
			if text == "发布" {
				logrus.Infof("找到发布按钮: %s", selector)
				break
			}
		}
	}

	if publishBtn == nil {
		// 尝试通过文本精确查找
		elements, _ := page.Elements("button")
		for _, btn := range elements {
			text, _ := btn.Text()
			if text == "发布" {
				publishBtn = btn
				logrus.Info("找到发布按钮: 通过文本匹配")
				break
			}
		}
	}

	if publishBtn == nil {
		return apperrors.ErrElementNotFound("发布按钮", err)
	}

	// 使用 JavaScript 点击（更可靠）
	_, err = publishBtn.Eval(`() => this.click()`)
	if err != nil {
		return stderrors.Wrap(err, "点击发布按钮失败")
	}

	logrus.Info("✅ 已点击发布按钮")
	return nil
}

// ValidatePublish 验证发布结果
func (p *PublishAction) ValidatePublish(ctx context.Context) (bool, string, error) {
	pp := p.page.Context(ctx)

	// 检查是否有成功提示
	successSelectors := []string{
		".success-message",
		"[class*='success']",
		".publish-success",
	}

	for _, selector := range successSelectors {
		if elem, err := pp.Element(selector); err == nil && elem != nil {
			if text, err := elem.Text(); err == nil {
				return true, text, nil
			}
		}
	}

	// 检查是否有错误提示
	errorSelectors := []string{
		".error-message",
		"[class*='error']",
		".publish-error",
	}

	for _, selector := range errorSelectors {
		if elem, err := pp.Element(selector); err == nil && elem != nil {
			if text, err := elem.Text(); err == nil {
				return false, text, apperrors.ErrPublishFailed(text, nil)
			}
		}
	}

	// 如果没有明确的提示，检查 URL 是否跳转
	currentURL := p.page.MustInfo().URL
	if currentURL != heyboxPublishURL {
		return true, "发布成功（页面已跳转）", nil
	}

	return false, "发布状态未知", nil
}
