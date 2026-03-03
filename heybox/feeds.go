package heybox

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
)

// FeedsAction 动态列表操作
type FeedsAction struct {
	page      *rod.Page
	navigator *Navigator
}

// NewFeedsAction 创建动态列表操作实例
func NewFeedsAction(page *rod.Page) *FeedsAction {
	return &FeedsAction{
		page:      page,
		navigator: NewNavigator(page),
	}
}

// GetRecommendedFeeds 获取推荐动态列表
func (f *FeedsAction) GetRecommendedFeeds(ctx context.Context, cursor string, limit int) (*FeedListResponse, error) {
	response := &FeedListResponse{
		Feeds:   []FeedItem{},
		HasMore: false,
	}

	if limit <= 0 {
		limit = 20
	}

	pp := f.page.Context(ctx).Timeout(60 * time.Second)

	// 导航到社区页面（推荐动态所在页面）
	if err := f.navigator.NavigateToCommunity(ctx); err != nil {
		return nil, fmt.Errorf("导航到社区页面失败: %w", err)
	}

	// 等待页面加载 - 增加等待时间，因为内容是动态加载的
	logrus.Info("等待页面内容加载...")
	time.Sleep(5 * time.Second)

	// 等待内容区域出现
	waitSelectors := []string{
		"[class*='content']",
		"[class*='list']",
		"[class*='item']",
		".link-item",
		"a[href*='/news/']",
	}

	for _, selector := range waitSelectors {
		has, _, _ := pp.Has(selector)
		if has {
			logrus.Infof("检测到内容区域: %s", selector)
			break
		}
	}

	// 尝试多种选择器来获取动态列表
	feedSelectors := []string{
		".feed-item",
		".post-item",
		"[class*='feed-item']",
		"[class*='FeedItem']",
		"[class*='postItem']",
		".content-item",
		".article-item",
		"li[class*='item']",
		".card-item",
	}

	var feeds []*rod.Element
	var foundSelector string
	for _, selector := range feedSelectors {
		elements, err := pp.Elements(selector)
		if err == nil && len(elements) > 0 {
			feeds = elements
			foundSelector = selector
			logrus.Infof("找到动态列表元素: %s, 数量: %d", selector, len(elements))
			break
		}
	}

	if len(feeds) == 0 {
		logrus.Warn("未找到动态列表元素，尝试通过 JavaScript 提取")
		return f.extractFeedsViaJS(ctx, pp, limit)
	}

	// 解析动态列表
	count := 0
	for _, feed := range feeds {
		if count >= limit {
			break
		}

		item := f.parseFeedItem(feed)
		if item.Title != "" || item.Content != "" {
			response.Feeds = append(response.Feeds, item)
			count++
		}
	}

	// 检查是否有更多
	response.HasMore = f.checkHasMore(pp)

	logrus.Infof("从 %s 获取到 %d 条动态", foundSelector, len(response.Feeds))
	return response, nil
}

// extractFeedsViaJS 通过 JavaScript 提取动态列表
func (f *FeedsAction) extractFeedsViaJS(ctx context.Context, page *rod.Page, limit int) (*FeedListResponse, error) {
	response := &FeedListResponse{
		Feeds:   []FeedItem{},
		HasMore: false,
	}

	// 使用 JavaScript 提取页面上的动态数据
	result, err := page.Eval(`() => {
		const feeds = [];

		// 首先检查页面内容，用于调试
		const bodyClasses = document.body.className;
		const mainContent = document.querySelector('main, .main, #app, [class*="app"]');
		const debugInfo = {
			bodyClasses: bodyClasses,
			hasMain: !!mainContent,
			mainClasses: mainContent ? mainContent.className : '',
			allClasses: [],
			allLinks: [],
			pageHTML: document.body.innerHTML.substring(0, 3000)
		};

		// 收集页面上所有 class 名称
		document.querySelectorAll('[class]').forEach(el => {
			if (typeof el.className === 'string') {
				el.className.split(' ').forEach(c => {
					if (c && debugInfo.allClasses.length < 200) {
						debugInfo.allClasses.push(c);
					}
				});
			}
		});

		// 收集所有链接
		document.querySelectorAll('a').forEach(el => {
			if (el.href && debugInfo.allLinks.length < 50) {
				debugInfo.allLinks.push({
					href: el.href,
					text: el.textContent.trim().substring(0, 50)
				});
			}
		});

		// 尝试多种方式获取动态
		const selectors = [
			'.feed-item', '.post-item', '[class*="feed"]', '[class*="post"]',
			'.content-item', '.article-item', '.card-item', '.item',
			'.news-item', '.list-item', '[class*="card"]', '[class*="Card"]',
			'a[href*="/bbs/link/"]', 'a[href*="/news/"]', 'a[href*="/post/"]',
			'.link-item', '.content-list > *', '.list > *',
			'[class*="link-item"]', '[class*="LinkItem"]', '.bbs-item'
		];

		// 优先查找包含帖子链接的元素
		let linkElements = document.querySelectorAll('a[href*="/bbs/link/"]');
		if (linkElements.length > 0) {
			debugInfo.usedSelector = 'a[href*="/bbs/link/"]';
			debugInfo.elementCount = linkElements.length;
			linkElements.forEach((el, index) => {
				if (index >= 50) return;

				const feed = {
					title: '',
					content: '',
					author: '',
					likeCount: 0,
					commentCount: 0,
					itemId: el.href,
					images: [],
					html: el.outerHTML.substring(0, 500)
				};

				// 获取标题 - 链接文本通常就是标题
				feed.title = el.textContent.trim().substring(0, 200);

				// 尝试获取父容器中的更多信息
				let parent = el.closest('[class*="item"], [class*="card"], [class*="post"]');
				if (parent) {
					// 获取作者
					const authorEl = parent.querySelector('[class*="author"], [class*="user"], a[href*="/user/"]');
					if (authorEl && authorEl !== el) {
						feed.author = authorEl.textContent.trim();
					}

					// 获取图片
					const imgEls = parent.querySelectorAll('img');
					imgEls.forEach(img => {
						if (img.src && !img.src.includes('avatar') && !img.src.includes('icon')) {
							feed.images.push(img.src);
						}
					});
				}

				if (feed.title) {
					feeds.push(feed);
				}
			});

			const hasMore = !!document.querySelector('.load-more, [class*="loadMore"], .pagination-next, [class*="more"]');
			return { feeds, hasMore, debugInfo };
		}

		let elements = [];
		let usedSelector = '';
		for (const selector of selectors) {
			try {
				const found = document.querySelectorAll(selector);
				if (found.length > elements.length) {
					elements = found;
					usedSelector = selector;
				}
			} catch (e) {}
		}

		debugInfo.usedSelector = usedSelector;
		debugInfo.elementCount = elements.length;

		elements.forEach((el, index) => {
			if (index >= 50) return; // 限制最大数量

			const feed = {
				title: '',
				content: '',
				author: '',
				likeCount: 0,
				commentCount: 0,
				itemId: '',
				images: [],
				html: el.outerHTML.substring(0, 500) // 保存部分 HTML 用于调试
			};

			// 获取标题
			const titleEl = el.querySelector('h1, h2, h3, h4, .title, [class*="title"]');
			if (titleEl) feed.title = titleEl.textContent.trim();

			// 获取内容
			const contentEl = el.querySelector('.content, .desc, [class*="content"], p, [class*="desc"]');
			if (contentEl) feed.content = contentEl.textContent.trim();

			// 如果没有标题，尝试从链接文本获取
			if (!feed.title) {
				const linkEl = el.querySelector('a');
				if (linkEl) {
					feed.title = linkEl.textContent.trim().substring(0, 100);
				}
			}

			// 获取作者
			const authorEl = el.querySelector('.author, .user-name, [class*="author"], [class*="username"], [class*="user"]');
			if (authorEl) feed.author = authorEl.textContent.trim();

			// 获取点赞数
			const likeEl = el.querySelector('[class*="like"], [class*="zan"], [class*="vote"]');
			if (likeEl) {
				const num = parseInt(likeEl.textContent.trim());
				if (!isNaN(num)) feed.likeCount = num;
			}

			// 获取评论数
			const commentEl = el.querySelector('[class*="comment"]');
			if (commentEl) {
				const num = parseInt(commentEl.textContent.trim());
				if (!isNaN(num)) feed.commentCount = num;
			}

			// 获取链接/ID
			const linkEl = el.querySelector('a[href*="/feed/"], a[href*="/post/"], a[href*="/article/"], a[href*="/news/"]');
			if (linkEl) feed.itemId = linkEl.href;

			// 如果元素本身就是链接
			if (!feed.itemId && el.tagName === 'A' && el.href) {
				feed.itemId = el.href;
			}

			// 获取图片
			const imgEls = el.querySelectorAll('img');
			imgEls.forEach(img => {
				if (img.src && !img.src.includes('avatar') && !img.src.includes('icon')) {
					feed.images.push(img.src);
				}
			});

			if (feed.title || feed.content) {
				feeds.push(feed);
			}
		});

		// 检查是否有更多
		const hasMore = !!document.querySelector('.load-more, [class*="loadMore"], .pagination-next, [class*="more"]');

		return { feeds, hasMore, debugInfo };
	}`)

	if err != nil {
		logrus.Errorf("JavaScript 提取失败: %v", err)
		return response, fmt.Errorf("提取动态列表失败: %w", err)
	}

	var data struct {
		Feeds []struct {
			Title        string   `json:"title"`
			Content      string   `json:"content"`
			Author       string   `json:"author"`
			LikeCount    int64    `json:"likeCount"`
			CommentCount int64    `json:"commentCount"`
			ItemID       string   `json:"itemId"`
			Images       []string `json:"images"`
			HTML         string   `json:"html"`
		} `json:"feeds"`
		HasMore   bool `json:"hasMore"`
		DebugInfo struct {
			BodyClasses   string   `json:"bodyClasses"`
			HasMain       bool     `json:"hasMain"`
			MainClasses   string   `json:"mainClasses"`
			AllClasses    []string `json:"allClasses"`
			AllLinks      []struct {
				Href string `json:"href"`
				Text string `json:"text"`
			} `json:"allLinks"`
			PageHTML      string `json:"pageHTML"`
			UsedSelector  string `json:"usedSelector"`
			ElementCount  int     `json:"elementCount"`
		} `json:"debugInfo"`
	}

	if err := result.Value.Unmarshal(&data); err != nil {
		logrus.Errorf("解析 JavaScript 结果失败: %v", err)
		return response, fmt.Errorf("解析动态数据失败: %w", err)
	}

	// 输出调试信息
	logrus.Infof("页面调试信息: usedSelector=%s, elementCount=%d", data.DebugInfo.UsedSelector, data.DebugInfo.ElementCount)
	logrus.Infof("页面上找到的 class 名称 (前50个): %v", data.DebugInfo.AllClasses[:min(50, len(data.DebugInfo.AllClasses))])

	// 输出链接信息
	if len(data.DebugInfo.AllLinks) > 0 {
		logrus.Info("页面上找到的链接 (前10个):")
		for _, link := range data.DebugInfo.AllLinks[:min(10, len(data.DebugInfo.AllLinks))] {
			logrus.Infof("  - %s: %s", link.Text, link.Href)
		}
	}

	// 输出部分页面 HTML 用于调试
	if data.DebugInfo.PageHTML != "" {
		logrus.Debugf("页面 HTML (前500字符): %s", data.DebugInfo.PageHTML[:min(500, len(data.DebugInfo.PageHTML))])
	}

	// 转换为 FeedItem
	for i, feed := range data.Feeds {
		if i >= limit {
			break
		}
		response.Feeds = append(response.Feeds, FeedItem{
			ItemID:       feed.ItemID,
			Title:        feed.Title,
			Content:      feed.Content,
			Images:       feed.Images,
			Author:       UserInfo{Nickname: feed.Author},
			LikeCount:    feed.LikeCount,
			CommentCount: feed.CommentCount,
		})
	}

	response.HasMore = data.HasMore

	logrus.Infof("通过 JavaScript 提取到 %d 条动态", len(response.Feeds))
	return response, nil
}

// parseFeedItem 解析单个动态项
func (f *FeedsAction) parseFeedItem(feed *rod.Element) FeedItem {
	item := FeedItem{}

	// 获取标题
	titleSelectors := []string{"h1", "h2", "h3", ".title", "[class*='title']"}
	for _, selector := range titleSelectors {
		titleEl, err := feed.Element(selector)
		if err == nil && titleEl != nil {
			item.Title, _ = titleEl.Text()
			if item.Title != "" {
				break
			}
		}
	}

	// 获取内容
	contentSelectors := []string{".content", ".desc", "[class*='content']", "p"}
	for _, selector := range contentSelectors {
		contentEl, err := feed.Element(selector)
		if err == nil && contentEl != nil {
			item.Content, _ = contentEl.Text()
			if item.Content != "" {
				break
			}
		}
	}

	// 获取作者
	authorSelectors := []string{".author", ".user-name", "[class*='author']", "[class*='username']"}
	for _, selector := range authorSelectors {
		authorEl, err := feed.Element(selector)
		if err == nil && authorEl != nil {
			authorName, _ := authorEl.Text()
			item.Author = UserInfo{Nickname: authorName}
			break
		}
	}

	// 获取点赞数
	likeSelectors := []string{"[class*='like']", "[class*='zan']", ".like-count"}
	for _, selector := range likeSelectors {
		likeEl, err := feed.Element(selector)
		if err == nil && likeEl != nil {
			likeText, _ := likeEl.Text()
			item.LikeCount = parseCount(likeText)
			break
		}
	}

	// 获取评论数
	commentSelectors := []string{"[class*='comment']", ".comment-count"}
	for _, selector := range commentSelectors {
		commentEl, err := feed.Element(selector)
		if err == nil && commentEl != nil {
			commentText, _ := commentEl.Text()
			item.CommentCount = parseCount(commentText)
			break
		}
	}

	// 获取链接/ID
	linkEl, err := feed.Element("a[href*='/feed/'], a[href*='/post/'], a[href*='/article/']")
	if err == nil && linkEl != nil {
		href, _ := linkEl.Attribute("href")
		if href != nil {
			item.ItemID = *href
		}
	}

	// 获取图片
	imgEls, err := feed.Elements("img")
	if err == nil {
		for _, imgEl := range imgEls {
			src, _ := imgEl.Attribute("src")
			if src != nil && *src != "" {
				// 过滤掉头像和图标
				if !contains(*src, "avatar") && !contains(*src, "icon") {
					item.Images = append(item.Images, *src)
				}
			}
		}
	}

	return item
}

// checkHasMore 检查是否有更多内容
func (f *FeedsAction) checkHasMore(page *rod.Page) bool {
	moreSelectors := []string{
		".load-more",
		"[class*='loadMore']",
		".pagination-next",
		"[class*='next']",
	}

	for _, selector := range moreSelectors {
		has, _, _ := page.Has(selector)
		if has {
			return true
		}
	}

	return false
}

// parseCount 解析数字（支持 "1.2万" 这种格式）
func parseCount(s string) int64 {
	s = trim(s)
	if s == "" {
		return 0
	}

	// 处理中文单位
	if contains(s, "万") {
		numStr := replace(s, "万", "")
		if num, err := strconv.ParseFloat(numStr, 64); err == nil {
			return int64(num * 10000)
		}
	}

	if contains(s, "k") || contains(s, "K") {
		numStr := replace(replace(s, "k", ""), "K", "")
		if num, err := strconv.ParseFloat(numStr, 64); err == nil {
			return int64(num * 1000)
		}
	}

	// 普通数字
	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return num
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// trim 去除字符串首尾空白
func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && isWhitespace(rune(s[start])) {
		start++
	}
	for end > start && isWhitespace(rune(s[end-1])) {
		end--
	}
	return s[start:end]
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// replace 替换字符串
func replace(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); {
		if i <= len(s)-len(old) && s[i:i+len(old)] == old {
			result += new
			i += len(old)
		} else {
			result += string(s[i])
			i++
		}
	}
	return result
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
