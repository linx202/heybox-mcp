package heybox

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/sirupsen/logrus"
)

// SearchRequest 搜索请求
type SearchRequest struct {
	Query     string `json:"query"`
	Limit     int    `json:"limit,omitempty"`
	Cursor    string `json:"cursor,omitempty"`
	SortType  string `json:"sort_type,omitempty"` // "hot" "热门" 等
	SortTime  string `json:"sort_time,omitempty"` // "综合" 综合排序（按热度）
	SortOrder string `json:"sort_order,omitempty"` // "综合" 综合排序（按时间)
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Query       string     `json:"query"`
	Total       int        `json:"total"`
	Feeds       []FeedItem `json:"feeds"`
	Users       []UserInfo `json:"users"`
	HasMore     bool       `json:"has_more"`
	NextCursor  string     `json:"next_cursor"`
	Took        float64    `json:"took"`
	SortByScore []string   `json:"sort_by_score"`
}

// SearchAction 搜索操作
type SearchAction struct {
	page      *rod.Page
	navigator *Navigator
}

// NewSearchAction 创建搜索操作实例
func NewSearchAction(page *rod.Page) *SearchAction {
	return &SearchAction{
		page:      page,
		navigator: NewNavigator(page),
	}
}

// SearchFeeds 搜索内容
func (s *SearchAction) SearchFeeds(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	startTime := time.Now()

	response := &SearchResponse{
		Query:   req.Query,
		Feeds:   []FeedItem{},
		Users:   []UserInfo{},
		HasMore: false,
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}

	pp := s.page.Context(ctx).Timeout(60 * time.Second)

	// 导航到首页
	if err := s.navigator.NavigateToHome(ctx); err != nil {
		return nil, fmt.Errorf("导航到搜索页面失败: %w", err)
	}

	// 等待页面加载
	time.Sleep(2 * time.Second)

	// 尝试找到搜索框
	searchInputSelectors := []string{
		"input[type='search']",
		"input[placeholder*='搜索']",
		".search-input",
		"[class*='search'] input",
	}

	var searchInput *rod.Element
	var err error
	for _, selector := range searchInputSelectors {
		searchInput, err = pp.Element(selector)
		if err == nil && searchInput != nil {
			logrus.Infof("找到搜索输入框: %s", selector)
			break
		}
	}

	if searchInput == nil {
		logrus.Warn("未找到搜索输入框，尝试通过其他方式")
		// 尝试通过 JavaScript 查找
		searchInput, err = pp.ElementByJS(rod.Eval(`() => {
			const inputs = document.querySelectorAll('input');
			for (const input of inputs) {
				if (input.placeholder && input.placeholder.includes('搜索')) {
					return input;
				}
			}
			return null;
		}`))
		if err != nil || searchInput == nil {
			return response, fmt.Errorf("未找到搜索输入框")
		}
	}

	// 输入搜索关键词
	if err := searchInput.Input(req.Query); err != nil {
		return nil, fmt.Errorf("输入搜索关键词失败: %w", err)
	}

	logrus.Infof("已输入搜索关键词: %s", req.Query)
	time.Sleep(1 * time.Second)

	// 按回车搜索
	searchInput.MustFocus()
	time.Sleep(100 * time.Millisecond)
	if err := pp.Keyboard.Type(input.Enter); err != nil {
		return nil, fmt.Errorf("按下回车失败: %w", err)
	}

	logrus.Info("已按下回车，等待搜索结果...")
	time.Sleep(3 * time.Second)

	// 检查是否有搜索结果
	searchResultSelectors := []string{
		".search-result-item",
		".search-results-list",
		"[class*='search-result']",
		".search-no-result",
		".result-item",
		"[class*='ResultItem']",
	}

	var foundResults bool
	for _, selector := range searchResultSelectors {
		results, err := pp.Elements(selector)
		if err == nil && len(results) > 0 {
			logrus.Infof("找到搜索结果元素: %s, 数量: %d", selector, len(results))
			foundResults = true

			for i, result := range results {
				if i >= req.Limit {
					break
				}

				item := FeedItem{}

				// 尝试获取标题
				titleEl, _ := result.Element("h1, h2, h3, .title, [class*='title']")
				if titleEl != nil {
					item.Title, _ = titleEl.Text()
				}

				// 尝试获取内容
				contentEl, _ := result.Element(".content, .desc, [class*='content'], p")
				if contentEl != nil {
					item.Content, _ = contentEl.Text()
				}

				// 尝试获取作者
				authorEl, _ := result.Element(".author, .user-name, [class*='author']")
				if authorEl != nil {
					authorName, _ := authorEl.Text()
					item.Author = UserInfo{
						Nickname: authorName,
					}
				}

				// 尝试获取链接/ID
				linkEl, _ := result.Element("a[href*='/feed/'], a[href*='/post/']")
				if linkEl != nil {
					href, _ := linkEl.Attribute("href")
					if href != nil {
						item.ItemID = *href
					}
				}

				if item.Title != "" || item.Content != "" {
					response.Feeds = append(response.Feeds, item)
				}
			}
			break
		}
	}

	if !foundResults {
		logrus.Warn("未找到搜索结果元素，可能没有结果或选择器需要调整")
	}

	// 检查是否有更多
	loadMoreSelectors := []string{
		".load-more",
		"[class*='loadMore']",
		".pagination-next",
	}

	for _, selector := range loadMoreSelectors {
		_, err := pp.Element(selector)
		if err == nil {
			response.HasMore = true
			break
		}
	}

	response.Total = len(response.Feeds)
	response.Took = time.Since(startTime).Seconds()

	logrus.Infof("搜索完成，找到 %d 条结果，耗时 %.2f 秒", response.Total, response.Took)
	return response, nil
}
