package main

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"
)

// loggingMiddleware 日志中间件
func (s *AppServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 创建响应写入器包装
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		logrus.WithFields(logrus.Fields{
			"method":   r.Method,
			"path":     r.URL.Path,
			"status":   wrapped.statusCode,
			"duration": duration.Milliseconds(),
			"remote":   r.RemoteAddr,
		}).Info("HTTP 请求")
	})
}

// recoveryMiddleware 错误恢复中间件
func (s *AppServer) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logrus.WithFields(logrus.Fields{
					"error":  err,
					"stack":  string(debug.Stack()),
					"method": r.Method,
					"path":   r.URL.Path,
				}).Error("服务器内部错误")

				http.Error(w, `{"error":"内部服务器错误"}`, http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// responseWriter 响应写入器包装
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader 写入状态码
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
