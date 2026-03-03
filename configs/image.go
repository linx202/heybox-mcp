package configs

import (
	"path/filepath"
	"sync"
)

var (
	// imageSaveDir 图片保存目录
	imageSaveDir string

	// imageDirMux 保护 imageSaveDir 变量的互斥锁
	imageDirMux sync.RWMutex
)

// InitImageDir 初始化图片保存目录
func InitImageDir(dir string) {
	imageDirMux.Lock()
	defer imageDirMux.Unlock()
	imageSaveDir = dir
}

// GetImageSaveDir 获取图片保存目录
func GetImageSaveDir() string {
	imageDirMux.RLock()
	defer imageDirMux.RUnlock()
	if imageSaveDir == "" {
		// 默认保存到当前目录的 images 文件夹
		return filepath.Join(".", "images")
	}
	return imageSaveDir
}
