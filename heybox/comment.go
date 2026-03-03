package heybox

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"
)

// CommentAction 评论操作
type CommentAction struct {
	page      *rod.Page
	navigator *Navigator
}

// NewCommentAction 创建评论操作实例
func NewCommentAction(page *rod.Page) *CommentAction {
	return &CommentAction{
		page:      page,
		navigator: NewNavigator(page),
	}
}

// PostComment 发表评论
func (c *CommentAction) PostComment(ctx context.Context, feedID, content string) error {
	pp := c.page.Context(ctx).Timeout(60 * time.Second)

	// 导航到动态详情页
	detailURL := feedID
	if len(feedID) < 10 || feedID[:4] != "http" {
		detailURL = fmt.Sprintf("https://www.xiaoheihe.cn/app/bbs/link/%s", feedID)
	}

	if err := c.navigator.NavigateToFeed(ctx, detailURL); err != nil {
		return fmt.Errorf("导航失败: %w", err)
	}

	time.Sleep(2 * time.Second)

	// 查找评论输入框
	commentInputSelectors := []string{
		"textarea[placeholder*='评论']",
		"textarea[placeholder*='回复']",
		".comment-input textarea",
		"[class*='comment'] textarea",
	}

	var commentInput *rod.Element
	var err error
	for _, sel := range commentInputSelectors {
		commentInput, err = pp.Element(sel)
		if err == nil && commentInput != nil {
			logrus.Infof("找到评论输入框: %s", sel)
			break
		}
	}

	if commentInput == nil {
		return fmt.Errorf("未找到评论输入框")
	}

	// 输入评论内容
	if err := commentInput.Input(content); err != nil {
		return fmt.Errorf("输入评论失败: %w", err)
	}

	logrus.Infof("已输入评论: %s", content)
	time.Sleep(500 * time.Millisecond)

	// 查找并点击发送按钮
	submitSelectors := []string{
		"button[type='submit']",
		".submit-btn",
		"[class*='submit']",
		"[class*='send']",
	}

	for _, sel := range submitSelectors {
		submitBtn, err := pp.Element(sel)
		if err == nil && submitBtn != nil {
			if err := submitBtn.Click(proto.InputMouseButtonLeft, 1); err == nil {
				logrus.Info("评论发送成功")
				time.Sleep(1 * time.Second)
				return nil
			}
		}
	}

	// 尝试按回车发送
	commentInput.MustFocus()
	pp.Keyboard.Type(input.Enter)
	logrus.Info("评论发送成功（回车）")
	time.Sleep(1 * time.Second)

	return nil
}

// ReplyComment 回复评论
func (c *CommentAction) ReplyComment(ctx context.Context, feedID, commentID, content string) error {
	pp := c.page.Context(ctx).Timeout(60 * time.Second)

	// 导航到动态详情页
	detailURL := feedID
	if len(feedID) < 10 || feedID[:4] != "http" {
		detailURL = fmt.Sprintf("https://www.xiaoheihe.cn/app/bbs/link/%s", feedID)
	}

	if err := c.navigator.NavigateToFeed(ctx, detailURL); err != nil {
		return fmt.Errorf("导航失败: %w", err)
	}

	time.Sleep(2 * time.Second)

	// 查找回复按钮
	replyBtnSelectors := []string{
		"[class*='reply']",
		".reply-btn",
		"button:contains('回复')",
	}

	for _, sel := range replyBtnSelectors {
		replyBtn, err := pp.Element(sel)
		if err == nil && replyBtn != nil {
			if err := replyBtn.Click(proto.InputMouseButtonLeft, 1); err == nil {
				logrus.Info("点击回复按钮成功")
				time.Sleep(500 * time.Millisecond)
				break
			}
		}
	}

	// 查找评论输入框并输入
	commentInputSelectors := []string{
		"textarea[placeholder*='回复']",
		"textarea[placeholder*='评论']",
		".comment-input textarea",
	}

	var commentInput *rod.Element
	for _, sel := range commentInputSelectors {
		var err error
		commentInput, err = pp.Element(sel)
		if err == nil && commentInput != nil {
			break
		}
	}

	if commentInput == nil {
		return fmt.Errorf("未找到评论输入框")
	}

	// 输入回复内容
	if err := commentInput.Input(content); err != nil {
		return fmt.Errorf("输入回复失败: %w", err)
	}

	logrus.Infof("已输入回复: %s", content)
	time.Sleep(500 * time.Millisecond)

	// 发送
	commentInput.MustFocus()
	pp.Keyboard.Type(input.Enter)
	logrus.Info("回复发送成功")
	time.Sleep(1 * time.Second)

	return nil
}

// GetComments 获取评论列表
func (c *CommentAction) GetComments(ctx context.Context, feedID string, limit int) ([]Comment, error) {
	pp := c.page.Context(ctx).Timeout(60 * time.Second)

	// 导航到动态详情页
	detailURL := feedID
	if len(feedID) < 10 || feedID[:4] != "http" {
		detailURL = fmt.Sprintf("https://www.xiaoheihe.cn/app/bbs/link/%s", feedID)
	}

	if err := c.navigator.NavigateToFeed(ctx, detailURL); err != nil {
		return nil, fmt.Errorf("导航失败: %w", err)
	}

	time.Sleep(2 * time.Second)

	// 使用 JavaScript 提取评论
	result, err := pp.Eval(`() => {
		const comments = [];
		const commentSelectors = ['.comment-item', '[class*="comment-item"]', '.comment'];

		let commentEls = [];
		for (const sel of commentSelectors) {
			const found = document.querySelectorAll(sel);
			if (found.length > commentEls.length) {
				commentEls = found;
			}
		}

		commentEls.forEach((el, i) => {
			if (i >= 50) return;
			const comment = {
				comment_id: 'comment_' + i,
				content: '',
				author: { nickname: '' },
				like_count: 0,
				created_at: ''
			};

			const contentEl = el.querySelector('.content, [class*="content"], p');
			if (contentEl) comment.content = contentEl.textContent.trim();

			const authorEl = el.querySelector('.author, .user-name, [class*="author"]');
			if (authorEl) comment.author.nickname = authorEl.textContent.trim();

			if (comment.content) {
				comments.push(comment);
			}
		});

		return comments;
	}`)

	if err != nil {
		return nil, fmt.Errorf("提取评论失败: %w", err)
	}

	var comments []Comment
	if err := result.Value.Unmarshal(&comments); err != nil {
		return nil, fmt.Errorf("解析评论失败: %w", err)
	}

	if limit > 0 && len(comments) > limit {
		comments = comments[:limit]
	}

	logrus.Infof("获取到 %d 条评论", len(comments))
	return comments, nil
}
