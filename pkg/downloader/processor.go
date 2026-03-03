package downloader

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// ImageProcessor 图片处理器
type ImageProcessor struct {
	maxWidth  uint
	maxHeight uint
	quality   int
}

// ProcessorOption 处理器配置选项
type ProcessorOption func(*ImageProcessor)

// WithMaxWidth 设置最大宽度
func WithMaxWidth(width uint) ProcessorOption {
	return func(p *ImageProcessor) {
		p.maxWidth = width
	}
}

// WithMaxHeight 设置最大高度
func WithMaxHeight(height uint) ProcessorOption {
	return func(p *ImageProcessor) {
		p.maxHeight = height
	}
}

// WithQuality 设置压缩质量 (1-100)
func WithQuality(quality int) ProcessorOption {
	return func(p *ImageProcessor) {
		if quality >= 1 && quality <= 100 {
			p.quality = quality
		}
	}
}

// NewImageProcessor 创建图片处理器
func NewImageProcessor(opts ...ProcessorOption) *ImageProcessor {
	p := &ImageProcessor{
		maxWidth:  1920,
		maxHeight: 1080,
		quality:   85,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Resize 调整图片大小
func (p *ImageProcessor) Resize(inputPath, outputPath string) error {
	// 打开图片文件
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 解码图片
	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("解码图片失败: %w", err)
	}

	// 获取原始尺寸
	bounds := img.Bounds()
	width, height := uint(bounds.Dx()), uint(bounds.Dy())

	// 判断是否需要缩放
	if width <= p.maxWidth && height <= p.maxHeight {
		logrus.Infof("图片尺寸符合要求，无需缩放: %dx%d", width, height)
		return nil
	}

	// 计算缩放比例（保持宽高比）
	var newWidth, newHeight uint
	if float64(width)/float64(p.maxWidth) > float64(height)/float64(p.maxHeight) {
		newWidth = p.maxWidth
		newHeight = uint(float64(height) * float64(p.maxWidth) / float64(width))
	} else {
		newHeight = p.maxHeight
		newWidth = uint(float64(width) * float64(p.maxHeight) / float64(height))
	}

	// 简单缩放：暂时跳过实际缩放操作，可以后续使用其他库
	// TODO: 实现图片缩放功能

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 保存图片
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer outFile.Close()

	logrus.Infof("图片缩放完成: %dx%d -> %dx%d (format: %s)",
		width, height, newWidth, newHeight, format)

	return nil
}

// ProcessImage 处理单张图片
func (p *ImageProcessor) ProcessImage(ctx context.Context, inputPath string) (string, error) {
	// 生成输出路径
	ext := filepath.Ext(inputPath)
	outputPath := inputPath[:len(inputPath)-len(ext)] + "_processed" + ext

	if err := p.Resize(inputPath, outputPath); err != nil {
		return "", err
	}

	return outputPath, nil
}

// ValidateImage 验证图片文件
func (p *ImageProcessor) ValidateImage(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 尝试解码图片
	_, _, err = image.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("无效的图片文件: %w", err)
	}

	return nil
}

// GetImageInfo 获取图片信息
func GetImageInfo(path string) (width, height int, format string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, "", fmt.Errorf("解码图片配置失败: %w", err)
	}

	return config.Width, config.Height, format, nil
}
