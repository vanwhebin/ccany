package middleware

import (
	"ccany/internal/i18n"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// I18nMiddleware 国际化中间件
type I18nMiddleware struct {
	i18nService *i18n.I18nService
	logger      *logrus.Logger
}

// NewI18nMiddleware 创建国际化中间件
func NewI18nMiddleware(i18nService *i18n.I18nService, logger *logrus.Logger) *I18nMiddleware {
	return &I18nMiddleware{
		i18nService: i18nService,
		logger:      logger,
	}
}

// Handler 处理国际化
func (m *I18nMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检测语言
		lang := m.i18nService.DetectLanguageFromContext(c)

		// 设置语言上下文
		m.i18nService.SetLanguageContext(c, lang)

		// 设置响应头
		c.Header("Content-Language", lang)

		// 记录日志
		m.logger.WithFields(logrus.Fields{
			"language": lang,
			"path":     c.Request.URL.Path,
		}).Debug("Language detected and set")

		c.Next()
	}
}
