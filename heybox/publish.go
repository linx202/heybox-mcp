package heybox

import (
	"context"
	"os"
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

// PublishAction 发布操作
type PublishAction struct {
	page      *rod.Page
	navigator *Navigator
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

// Publish 发布动态
func (p *PublishAction) Publish(ctx context.Context, content PublishContent) error {
	if len(content.ImagePaths) == 0 && len(content.Content) == 0 {
		return apperrors.ErrInvalidParameter("图片和内容不能同时为空")
	}

	page := p.page.Context(ctx)

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
	// ⚠️ 需要根据实际页面调整选择器
	titleSelectors := []string{
		".title-input",
		"input[placeholder*='标题']",
		"[class*='title'] input",
		"input[name='title']",
	}

	var titleInput *rod.Element
	var err error
	for _, selector := range titleSelectors {
		titleInput, err = page.Element(selector)
		if err == nil && titleInput != nil {
			break
		}
	}

	if titleInput == nil {
		return apperrors.ErrElementNotFound("标题输入框", err)
	}

	// 清空并输入
	_ = titleInput.MustText()
	titleInput.MustInput(title)
	titleInput.MustInput("\n") // 触发输入事件

	time.Sleep(500 * time.Millisecond)
	return nil
}

// inputContent 输入正文
func (p *PublishAction) inputContent(page *rod.Page, content string, tags []string) error {
	// ⚠️ 需要根据实际页面调整选择器
	contentSelectors := []string{
		".content-textarea",
		"textarea[placeholder*='内容']",
		"[class*='content'] textarea",
		"div[contenteditable='true']",
		".ql-editor",
	}

	var contentInput *rod.Element
	var err error
	for _, selector := range contentSelectors {
		contentInput, err = page.Element(selector)
		if err == nil && contentInput != nil {
			break
		}
	}

	if contentInput == nil {
		return apperrors.ErrElementNotFound("正文输入框", err)
	}

	// 输入正文
	contentInput.MustInput(content)
	time.Sleep(500 * time.Millisecond)

	// 添加标签
	for _, tag := range tags {
		contentInput.MustInput("#" + tag + " ")
		time.Sleep(200 * time.Millisecond)
	}

	return nil
}

// clickPublishButton 点击发布按钮
func (p *PublishAction) clickPublishButton(page *rod.Page) error {
	// ⚠️ 需要根据实际页面调整选择器
	publishSelectors := []string{
		".publish-btn",
		"button[class*='publish']",
		"[class*='PublishBtn']",
		"button:has-text('发布')",
	}

	var publishBtn *rod.Element
	var err error
	for _, selector := range publishSelectors {
		publishBtn, err = page.Element(selector)
		if err == nil && publishBtn != nil {
			break
		}
	}

	if publishBtn == nil {
		return apperrors.ErrElementNotFound("发布按钮", err)
	}

	// 点击发布
	if err := publishBtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return stderrors.Wrap(err, "点击发布按钮失败")
	}

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
