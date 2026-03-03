package downloader

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateFilename(t *testing.T) {
	d := NewImageDownloader()

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "简单 URL",
			url:      "https://example.com/image.jpg",
			expected: "image.jpg",
		},
		{
			name:     "带路径的 URL",
			url:      "https://example.com/path/to/photo.png",
			expected: "photo.png",
		},
		{
			name:     "带查询参数的 URL",
			url:      "https://example.com/image?width=200",
			expected: "image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.url)
			if err != nil {
				t.Fatalf("Parse URL failed: %v", err)
			}
			result := d.generateFilename(u)
			if result != tt.expected {
				t.Errorf("generateFilename() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal.jpg", "normal.jpg"},
		{"file\\name.jpg", "file_name.jpg"},
		{"file:name.jpg", "file_name.jpg"},
		{"file*name?.jpg", "file_name_.jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateImage(t *testing.T) {
	p := NewImageProcessor()

	// 测试不存在的文件
	err := p.ValidateImage("/nonexistent/file.jpg")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}

	// 测试创建一个临时图片文件
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("not an image"), 0644)

	err = p.ValidateImage(tmpFile)
	if err == nil {
		t.Error("Expected error for invalid image file")
	}
}

func TestGetImageInfo(t *testing.T) {
	// 测试不存在的文件
	_, _, _, err := GetImageInfo("/nonexistent/file.jpg")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}
