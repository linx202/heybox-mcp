package configs

import (
	"sync"
)

var (
	// username 当前登录用户名
	username string

	// usernameMux 保护 username 变量的互斥锁
	usernameMux sync.RWMutex
)

// SetUsername 设置当前登录用户名
func SetUsername(name string) {
	usernameMux.Lock()
	defer usernameMux.Unlock()
	username = name
}

// GetUsername 获取当前登录用户名
func GetUsername() string {
	usernameMux.RLock()
	defer usernameMux.RUnlock()
	return username
}

// Username 导出的变量，用于向后兼容
// Deprecated: 使用 GetUsername() 代替
var Username = ""
