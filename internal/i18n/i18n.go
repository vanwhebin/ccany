package i18n

import (
	"encoding/json"
	"strings"

	"ccany/internal/webfs"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
)

// I18nService 国际化服务
type I18nService struct {
	bundle             *i18n.Bundle
	logger             *logrus.Logger
	supportedLanguages []string
}

// NewI18nService 创建国际化服务
func NewI18nService(logger *logrus.Logger) *I18nService {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	service := &I18nService{
		bundle: bundle,
		logger: logger,
	}

	// 加载语言文件
	if err := service.loadLocales(); err != nil {
		logger.WithError(err).Error("Failed to load locales")
	}

	return service
}

// loadLocales 加载语言文件
func (s *I18nService) loadLocales() error {
	// 从embedded文件系统获取可用的语言
	availableLocales, err := webfs.GetAvailableLocales()
	if err != nil {
		s.logger.WithError(err).Error("Failed to get available locales")
		// 使用默认支持的语言
		availableLocales = []string{"en-US", "zh-CN"}
	}

	// 如果没有找到任何语言文件，使用默认支持的语言
	if len(availableLocales) == 0 {
		s.logger.Warn("No locale files found in embedded filesystem")
		availableLocales = []string{"en-US", "zh-CN"}
	}

	var loadedLanguages []string
	for _, lang := range availableLocales {
		// 加载到bundle中
		_, err := s.bundle.LoadMessageFileFS(webfs.GetLocalesFS(), lang+".json")
		if err != nil {
			s.logger.WithError(err).WithField("language", lang).Error("Failed to load language file")
			continue
		}

		loadedLanguages = append(loadedLanguages, lang)
		s.logger.WithField("language", lang).Info("Loaded language file")
	}

	// 保存支持的语言列表
	s.supportedLanguages = loadedLanguages
	if len(s.supportedLanguages) == 0 {
		s.supportedLanguages = []string{"en-US", "zh-CN"}
	}

	return nil
}

// GetLocalizer 获取本地化器
func (s *I18nService) GetLocalizer(lang string) *i18n.Localizer {
	// 解析语言标签
	acceptLanguage := s.parseAcceptLanguage(lang)
	return i18n.NewLocalizer(s.bundle, acceptLanguage...)
}

// parseAcceptLanguage 解析Accept-Language头
func (s *I18nService) parseAcceptLanguage(acceptLanguage string) []string {
	if acceptLanguage == "" {
		return []string{"en-US"} // 默认语言
	}

	// 简单解析Accept-Language头
	languages := strings.Split(acceptLanguage, ",")
	result := make([]string, 0, len(languages))

	for _, lang := range languages {
		// 移除权重值 (如 "en-US;q=0.8")
		lang = strings.TrimSpace(strings.Split(lang, ";")[0])
		if lang != "" {
			result = append(result, lang)
		}
	}

	// 如果没有找到有效语言，返回默认语言
	if len(result) == 0 {
		result = append(result, "en-US")
	}

	return result
}

// Translate 翻译文本
func (s *I18nService) Translate(localizer *i18n.Localizer, messageID string, templateData map[string]interface{}) string {
	message, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
	if err != nil {
		s.logger.WithError(err).WithField("messageID", messageID).Warn("Failed to translate message")
		return messageID // 返回原始ID作为后备
	}
	return message
}

// GetAllMessages 获取指定语言的所有消息
func (s *I18nService) GetAllMessages(lang string) (map[string]interface{}, error) {
	// 从embedded文件系统读取语言文件
	data, err := webfs.GetLocaleFile(lang)
	if err != nil {
		return nil, err
	}

	// 解析JSON
	var messages map[string]interface{}
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// DetectLanguageFromContext 从上下文检测语言
func (s *I18nService) DetectLanguageFromContext(c *gin.Context) string {
	// 1. 首先从查询参数获取语言设置
	if lang := c.Query("lang"); lang != "" {
		return lang
	}

	// 2. 从Cookie获取语言设置
	if lang, err := c.Cookie("language"); err == nil && lang != "" {
		return lang
	}

	// 3. 从Accept-Language头获取
	acceptLanguage := c.GetHeader("Accept-Language")
	if acceptLanguage != "" {
		languages := s.parseAcceptLanguage(acceptLanguage)
		if len(languages) > 0 {
			return languages[0]
		}
	}

	// 4. 默认语言
	return "en-US"
}

// GetSupportedLanguages 获取支持的语言列表
func (s *I18nService) GetSupportedLanguages() []string {
	return s.supportedLanguages
}

// SetLanguageContext 设置语言上下文
func (s *I18nService) SetLanguageContext(c *gin.Context, lang string) {
	c.Set("language", lang)
	c.Set("localizer", s.GetLocalizer(lang))
}

// GetLocalizerFromContext 从上下文获取本地化器
func GetLocalizerFromContext(c *gin.Context) *i18n.Localizer {
	if localizer, exists := c.Get("localizer"); exists {
		if l, ok := localizer.(*i18n.Localizer); ok {
			return l
		}
	}
	return nil
}

// GetLanguageFromContext 从上下文获取语言
func GetLanguageFromContext(c *gin.Context) string {
	if lang, exists := c.Get("language"); exists {
		if l, ok := lang.(string); ok {
			return l
		}
	}
	return "en-US"
}

// T 翻译函数 - 从上下文获取本地化器进行翻译
func T(c *gin.Context, messageID string, templateData ...map[string]interface{}) string {
	localizer := GetLocalizerFromContext(c)
	if localizer == nil {
		return messageID
	}

	var data map[string]interface{}
	if len(templateData) > 0 {
		data = templateData[0]
	}

	message, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	})
	if err != nil {
		return messageID
	}
	return message
}
