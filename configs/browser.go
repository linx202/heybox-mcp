package configs

import (
	"sync"
)

var (
	// headless 是否无头模式
	headless bool

	// binPath 浏览器二进制文件路径
	binPath string

	// headlessMux 保护 headless 变量的互斥锁
	headlessMux sync.RWMutex

	// binPathMux 保护 binPath 变量的互斥锁
	binPathMux sync.RWMutex
)

// InitHeadless 初始化无头模式配置
func InitHeadless(h bool) {
	headlessMux.Lock()
	defer headlessMux.Unlock()
	headless = h
}

// IsHeadless 返回是否无头模式
func IsHeadless() bool {
	headlessMux.RLock()
	defer headlessMux.RUnlock()
	return headless
}

// SetBinPath 设置浏览器二进制文件路径
func SetBinPath(path string) {
	binPathMux.Lock()
	defer binPathMux.Unlock()
	binPath = path
}

// GetBinPath 获取浏览器二进制文件路径
func GetBinPath() string {
	binPathMux.RLock()
	defer binPathMux.RUnlock()
	return binPath
}
