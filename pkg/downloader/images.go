package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// ImageDownloadResult 图片下载结果
type ImageDownloadResult struct {
	URL      string // 原始 URL
	LocalPath string // 本地保存路径
	Error    error  // 错误信息
}

// ImageDownloader 图片下载器
type ImageDownloader struct {
	client    *http.Client
	saveDir   string
	userAgent string
}

// DownloaderOption 下载器配置选项
type DownloaderOption func(*ImageDownloader)

// WithSaveDir 设置保存目录
func WithSaveDir(dir string) DownloaderOption {
	return func(d *ImageDownloader) {
		d.saveDir = dir
	}
}

// WithUserAgent 设置 User-Agent
func WithUserAgent(ua string) DownloaderOption {
	return func(d *ImageDownloader) {
		d.userAgent = ua
	}
}

// NewImageDownloader 创建图片下载器
func NewImageDownloader(opts ...DownloaderOption) *ImageDownloader {
	d := &ImageDownloader{
		client: &http.Client{},
		saveDir: "./images",
		userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

// Download 下载单张图片
func (d *ImageDownloader) Download(ctx context.Context, imgURL string) (string, error) {
	// 解析 URL
	parsedURL, err := url.Parse(imgURL)
	if err != nil {
		return "", fmt.Errorf("解析 URL 失败: %w", err)
	}

	// 生成本地文件名
	filename := d.generateFilename(parsedURL)
	localPath := filepath.Join(d.saveDir, filename)

	// 创建目录
	if err := os.MkdirAll(d.saveDir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imgURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置 headers
	req.Header.Set("User-Agent", d.userAgent)
	req.Header.Set("Accept", "image/webp,image/apng,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	// 发送请求
	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 创建文件
	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 写入文件
	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	logrus.Infof("图片下载成功: %s -> %s", imgURL, localPath)
	return localPath, nil
}

// DownloadBatch 批量下载图片
func (d *ImageDownloader) DownloadBatch(ctx context.Context, urls []string) []ImageDownloadResult {
	results := make([]ImageDownloadResult, len(urls))
	var wg sync.WaitGroup

	for i, imgURL := range urls {
		wg.Add(1)
		go func(idx int, url string) {
			defer wg.Done()

			localPath, err := d.Download(ctx, url)
			results[idx] = ImageDownloadResult{
				URL:      url,
				LocalPath: localPath,
				Error:    err,
			}
		}(i, imgURL)
	}

	wg.Wait()
	return results
}

// generateFilename 根据 URL 生成文件名
func (d *ImageDownloader) generateFilename(u *url.URL) string {
	// 获取路径的最后一个部分
	path := u.Path
	if path == "" || path == "/" {
		// 使用 query 参数中的文件名
		query := u.Query()
		if filename := query.Get("filename"); filename != "" {
			return sanitizeFilename(filename)
		}
		// 生成随机文件名
		return fmt.Sprintf("image_%d.jpg", generateRandomID())
	}

	// 获取文件名
	filename := filepath.Base(path)
	if filename == "" || filename == "." {
		return fmt.Sprintf("image_%d.jpg", generateRandomID())
	}

	return sanitizeFilename(filename)
}

// sanitizeFilename 清理文件名
func sanitizeFilename(name string) string {
	// 移除或替换不安全的字符
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")

	// 限制长度
	if len(name) > 200 {
		name = name[:200]
	}

	return name
}

// generateRandomID 生成随机 ID
func generateRandomID() int {
	return int(uint32(^uint32(0)) >> 1)
}
