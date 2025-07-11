package handlers

import (
	"net/http"

	"ccany/internal/i18n"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// I18nHandler 多语言处理器
type I18nHandler struct {
	i18nService *i18n.I18nService
	logger      *logrus.Logger
}

// NewI18nHandler 创建多语言处理器
func NewI18nHandler(i18nService *i18n.I18nService, logger *logrus.Logger) *I18nHandler {
	return &I18nHandler{
		i18nService: i18nService,
		logger:      logger,
	}
}

// GetLanguages 获取支持的语言列表
func (h *I18nHandler) GetLanguages(c *gin.Context) {
	languages := h.i18nService.GetSupportedLanguages()

	c.JSON(http.StatusOK, gin.H{
		"languages": languages,
	})
}

// GetMessages 获取指定语言的消息
func (h *I18nHandler) GetMessages(c *gin.Context) {
	lang := c.Param("lang")
	if lang == "" {
		lang = h.i18nService.DetectLanguageFromContext(c)
	}

	// 验证语言是否支持
	supportedLanguages := h.i18nService.GetSupportedLanguages()
	isSupported := false
	for _, supportedLang := range supportedLanguages {
		if supportedLang == lang {
			isSupported = true
			break
		}
	}

	if !isSupported {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":               "Unsupported language",
			"supported_languages": supportedLanguages,
		})
		return
	}

	// 获取消息
	messages, err := h.i18nService.GetAllMessages(lang)
	if err != nil {
		h.logger.WithError(err).WithField("language", lang).Error("Failed to get messages")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get messages",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"language": lang,
		"messages": messages,
	})
}

// GetCurrentLanguage 获取当前语言
func (h *I18nHandler) GetCurrentLanguage(c *gin.Context) {
	lang := i18n.GetLanguageFromContext(c)

	c.JSON(http.StatusOK, gin.H{
		"language": lang,
	})
}

// SetLanguage 设置语言偏好
func (h *I18nHandler) SetLanguage(c *gin.Context) {
	var req struct {
		Language string `json:"language" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 验证语言是否支持
	supportedLanguages := h.i18nService.GetSupportedLanguages()
	isSupported := false
	for _, supportedLang := range supportedLanguages {
		if supportedLang == req.Language {
			isSupported = true
			break
		}
	}

	if !isSupported {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":               "Unsupported language",
			"supported_languages": supportedLanguages,
		})
		return
	}

	// 设置Cookie
	c.SetCookie("language", req.Language, 365*24*3600, "/", "", false, false)

	// 更新上下文
	h.i18nService.SetLanguageContext(c, req.Language)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Language preference updated",
		"language": req.Language,
	})
}
