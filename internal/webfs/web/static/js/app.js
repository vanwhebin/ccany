// Claude Code Proxy Backend Management Application - Enhanced with Claude UI
console.log('Enhanced JavaScript file loaded successfully');

// Test JavaScript execution by making a simple request
fetch('/health').then(response => {
    console.log('JavaScript test request successful:', response.status);
}).catch(error => {
    console.error('JavaScript test request failed:', error);
});

// Enhanced Theme Controller
class ThemeController {
    constructor() {
        this.currentTheme = this.getStoredTheme() || this.getSystemTheme();
        this.init();
    }

    init() {
        this.applyTheme(this.currentTheme);
        this.setupThemeToggle();
        this.setupSystemThemeListener();
    }

    getStoredTheme() {
        return localStorage.getItem('theme');
    }

    getSystemTheme() {
        return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }

    setStoredTheme(theme) {
        localStorage.setItem('theme', theme);
    }

    applyTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        this.currentTheme = theme;
        this.setStoredTheme(theme);
        this.updateToggleButton();
    }

    toggleTheme() {
        const newTheme = this.currentTheme === 'light' ? 'dark' : 'light';
        this.applyTheme(newTheme);
        
        // Add smooth transition animation
        document.body.style.transition = 'background 0.3s ease';
        setTimeout(() => {
            document.body.style.transition = '';
        }, 300);
    }

    setupThemeToggle() {
        const toggleButton = document.getElementById('themeToggle');
        if (toggleButton) {
            toggleButton.addEventListener('click', () => {
                this.toggleTheme();
                // Add click animation
                toggleButton.style.transform = 'scale(0.9)';
                setTimeout(() => {
                    toggleButton.style.transform = '';
                }, 150);
            });
        }
    }

    updateToggleButton() {
        const toggleButton = document.getElementById('themeToggle');
        if (toggleButton) {
            const sunIcon = toggleButton.querySelector('.sun');
            const moonIcon = toggleButton.querySelector('.moon');
            
            if (this.currentTheme === 'dark') {
                toggleButton.title = '切换到明亮模式';
                if (sunIcon) sunIcon.style.display = 'block';
                if (moonIcon) moonIcon.style.display = 'none';
            } else {
                toggleButton.title = '切换到暗黑模式';
                if (sunIcon) sunIcon.style.display = 'none';
                if (moonIcon) moonIcon.style.display = 'block';
            }
        }
    }

    setupSystemThemeListener() {
        const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
        mediaQuery.addEventListener('change', (e) => {
            // Only auto-switch if user hasn't manually set a theme
            if (!this.getStoredTheme()) {
                this.applyTheme(e.matches ? 'dark' : 'light');
            }
        });
    }
}
class UIAnimationController {
    constructor() {
        this.animationQueue = [];
        this.isAnimating = false;
        this.observers = new Map();
        this.init();
    }

    init() {
        this.setupIntersectionObserver();
        this.setupScrollAnimations();
        this.setupHoverEffects();
        this.setupClickAnimations();
    }

    setupIntersectionObserver() {
        const options = {
            root: null,
            rootMargin: '0px',
            threshold: 0.1
        };

        const callback = (entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    this.animateElementIn(entry.target);
                }
            });
        };

        this.observer = new IntersectionObserver(callback, options);
        
        // Observe elements that should animate on scroll
        document.querySelectorAll('.card, .stat-card, .table-container, .page-header').forEach(el => {
            this.observer.observe(el);
        });
    }

    animateElementIn(element) {
        // Determine animation type based on element type
        if (element.classList.contains('stat-card')) {
            element.classList.add('animate-slide-in-bottom');
        } else if (element.classList.contains('card')) {
            element.classList.add('animate-fade-in');
        } else if (element.classList.contains('table-container')) {
            element.classList.add('animate-slide-in-bottom');
        } else if (element.classList.contains('page-header')) {
            element.classList.add('animate-slide-in-top');
        }

        // Add stagger effect for multiple elements
        const siblings = Array.from(element.parentNode.children).filter(child => 
            child.classList.contains(element.classList[0])
        );
        const index = siblings.indexOf(element);
        element.style.animationDelay = `${index * 0.1}s`;
    }

    setupScrollAnimations() {
        let ticking = false;

        const updateScrollEffects = () => {
            const scrolled = window.pageYOffset;
            const rate = scrolled * -0.5;
            
            // Parallax effect for login background
            const loginContainer = document.querySelector('.login-container');
            if (loginContainer) {
                loginContainer.style.transform = `translateY(${rate}px)`;
            }

            ticking = false;
        };

        const requestTick = () => {
            if (!ticking) {
                requestAnimationFrame(updateScrollEffects);
                ticking = true;
            }
        };

        window.addEventListener('scroll', requestTick, { passive: true });
    }

    setupHoverEffects() {
        // Enhanced button hover effects
        document.addEventListener('mouseover', (e) => {
            if (e.target.classList.contains('btn')) {
                this.addHoverGlow(e.target);
            }
        });

        document.addEventListener('mouseout', (e) => {
            if (e.target.classList.contains('btn')) {
                this.removeHoverGlow(e.target);
            }
        });
    }

    addHoverGlow(element) {
        if (!element.querySelector('.hover-glow')) {
            const glow = document.createElement('div');
            glow.className = 'hover-glow';
            glow.style.cssText = `
                position: absolute;
                top: -2px;
                left: -2px;
                right: -2px;
                bottom: -2px;
                background: linear-gradient(45deg, var(--claude-orange), var(--claude-blue));
                border-radius: inherit;
                z-index: -1;
                opacity: 0;
                transition: opacity 0.3s ease;
                pointer-events: none;
            `;
            element.style.position = 'relative';
            element.appendChild(glow);
            
            // Animate glow in
            setTimeout(() => {
                glow.style.opacity = '0.3';
            }, 10);
        }
    }

    removeHoverGlow(element) {
        const glow = element.querySelector('.hover-glow');
        if (glow) {
            glow.style.opacity = '0';
            setTimeout(() => {
                if (glow.parentNode) {
                    glow.parentNode.removeChild(glow);
                }
            }, 300);
        }
    }

    setupClickAnimations() {
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('btn') || e.target.classList.contains('menu-item')) {
                this.createRippleEffect(e.target, e);
            }
        });
    }

    createRippleEffect(element, event) {
        const ripple = document.createElement('div');
        const rect = element.getBoundingClientRect();
        const size = Math.max(rect.width, rect.height);
        const x = event.clientX - rect.left - size / 2;
        const y = event.clientY - rect.top - size / 2;
        
        ripple.style.cssText = `
            position: absolute;
            width: ${size}px;
            height: ${size}px;
            left: ${x}px;
            top: ${y}px;
            background: rgba(255, 255, 255, 0.4);
            border-radius: 50%;
            transform: scale(0);
            animation: ripple-animation 0.6s ease-out;
            pointer-events: none;
        `;
        
        element.style.position = 'relative';
        element.style.overflow = 'hidden';
        element.appendChild(ripple);
        
        // Add CSS animation if not already present
        if (!document.querySelector('#ripple-animation-styles')) {
            const style = document.createElement('style');
            style.id = 'ripple-animation-styles';
            style.textContent = `
                @keyframes ripple-animation {
                    to {
                        transform: scale(2);
                        opacity: 0;
                    }
                }
            `;
            document.head.appendChild(style);
        }
        
        setTimeout(() => {
            if (ripple.parentNode) {
                ripple.parentNode.removeChild(ripple);
            }
        }, 600);
    }

    // Enhanced loading animations
    showLoadingState(element, text = 'Loading...') {
        if (element.dataset.originalContent) return; // Already in loading state
        
        element.dataset.originalContent = element.innerHTML;
        element.innerHTML = `
            <div class="loading-content">
                <div class="loading-spinner"></div>
                <span>${text}</span>
            </div>
        `;
        element.classList.add('btn-loading');
        element.disabled = true;
    }

    hideLoadingState(element) {
        if (element.dataset.originalContent) {
            element.innerHTML = element.dataset.originalContent;
            delete element.dataset.originalContent;
            element.classList.remove('btn-loading');
            element.disabled = false;
        }
    }

    // Smooth transitions between states
    transitionToState(element, newState, duration = 300) {
        element.style.opacity = '0';
        element.style.transform = 'translateY(10px)';
        
        setTimeout(() => {
            element.className = newState;
            element.style.opacity = '1';
            element.style.transform = 'translateY(0)';
        }, duration / 2);
    }

    // Enhanced notification system
    showEnhancedNotification(message, type = 'info', duration = 4000) {
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        
        const icon = this.getNotificationIcon(type);
        notification.innerHTML = `
            <div class="notification-icon">${icon}</div>
            <div class="notification-content">
                <div class="notification-message">${message}</div>
            </div>
            <button class="notification-close">×</button>
        `;
        
        document.body.appendChild(notification);
        
        // Animate in
        setTimeout(() => {
            notification.classList.add('show');
        }, 10);
        
        // Close button functionality
        const closeBtn = notification.querySelector('.notification-close');
        closeBtn.addEventListener('click', () => {
            this.hideNotification(notification);
        });
        
        // Auto-hide
        setTimeout(() => {
            this.hideNotification(notification);
        }, duration);
    }

    getNotificationIcon(type) {
        const icons = {
            success: '✅',
            error: '❌',
            warning: '⚠️',
            info: 'ℹ️'
        };
        return icons[type] || icons.info;
    }

    hideNotification(notification) {
        notification.classList.remove('show');
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 300);
    }

    // Performance optimized counter animations
    animateCounter(element, start, end, duration = 1000) {
        const range = end - start;
        const increment = end > start ? 1 : -1;
        const stepTime = Math.abs(Math.floor(duration / range));
        let current = start;
        
        const timer = setInterval(() => {
            current += increment;
            element.textContent = current;
            
            if (current === end) {
                clearInterval(timer);
            }
        }, stepTime);
    }

    // Enhanced form validation with animations
    validateFormField(field, isValid) {
        const parent = field.closest('.form-group');
        if (!parent) return;
        
        parent.classList.remove('field-valid', 'field-invalid');
        
        if (isValid) {
            parent.classList.add('field-valid');
            this.showFieldSuccess(field);
        } else {
            parent.classList.add('field-invalid');
            this.showFieldError(field);
        }
    }

    showFieldSuccess(field) {
        field.style.borderColor = 'var(--success-color)';
        field.style.boxShadow = '0 0 0 3px rgba(34, 197, 94, 0.1)';
    }

    showFieldError(field) {
        field.style.borderColor = 'var(--error-color)';
        field.style.boxShadow = '0 0 0 3px rgba(239, 68, 68, 0.1)';
        field.classList.add('animate-shake');
        
        setTimeout(() => {
            field.classList.remove('animate-shake');
        }, 500);
    }

    // Cleanup method
    destroy() {
        if (this.observer) {
            this.observer.disconnect();
        }
        this.observers.clear();
    }
}

class ClaudeProxyApp {
    constructor() {
        this.currentTab = 'dashboard';
        this.config = {};
        this.stats = {};
        this.currentUser = null;
        this.isAuthenticated = false;
        this.currentLanguage = 'zh-CN';
        this.translations = {};
        this.uiController = new UIAnimationController();
        this.themeController = new ThemeController();
        this.debounceTimers = new Map();
        this.performanceMetrics = {
            startTime: Date.now(),
            apiCalls: 0,
            errors: 0
        };
        // Don't call init() here, it will be called from DOMContentLoaded
    }

    async init() {
        console.log('ClaudeProxyApp.init() called');
        
        // 显示加载指示器
        this.showLoadingIndicator();
        
        try {
            await this.initializeI18n();
            console.log('i18n initialized');
            
            this.setupEventListeners();
            console.log('event listeners setup');
            
            // 隐藏加载指示器
            this.hideLoadingIndicator();
            
            this.checkAuthentication();
            console.log('authentication checked');
            
            this.startPeriodicUpdates();
            console.log('periodic updates started');
        } catch (error) {
            console.error('App initialization failed:', error);
            this.hideLoadingIndicator();
            // 显示错误信息
            this.showInitializationError();
        }
    }

    // Initialize i18n
    async initializeI18n() {
        // Detect language preference
        await this.detectLanguage();
        // Load translation files
        await this.loadTranslations();
        // Apply translations
        this.applyTranslations();
    }

    // Detect language preference
    async detectLanguage() {
        // 1. Get saved language preference from localStorage
        const savedLanguage = localStorage.getItem('preferred_language');
        if (savedLanguage) {
            this.currentLanguage = savedLanguage;
            return;
        }

        // 2. Get from cookie
        const cookieLanguage = this.getCookie('language');
        if (cookieLanguage) {
            this.currentLanguage = cookieLanguage;
            return;
        }

        // 3. Detect from browser language settings
        const browserLanguage = navigator.language || navigator.userLanguage;
        if (browserLanguage.startsWith('zh')) {
            this.currentLanguage = 'zh-CN';
        } else {
            this.currentLanguage = 'en-US';
        }

        // 4. Get current language from backend
        try {
            const response = await fetch('/i18n/current');
            if (response.ok) {
                const data = await response.json();
                this.currentLanguage = data.language;
            }
        } catch (error) {
            console.log('Unable to get backend language settings, using default language');
        }
    }

    // Load translation files
    async loadTranslations() {
        try {
            console.log('Loading translations for language:', this.currentLanguage);
            const response = await fetch(`/i18n/messages/${this.currentLanguage}`);
            console.log('Translation API response status:', response.status);
            
            if (response.ok) {
                const data = await response.json();
                console.log('Translation data received:', data);
                this.translations = data.messages;
                console.log('Translations loaded successfully:', Object.keys(this.translations));
            } else {
                console.error('Failed to load translation files, status:', response.status);
                console.error('Response text:', await response.text());
                
                // If loading fails, try to load default language
                if (this.currentLanguage !== 'zh-CN') {
                    console.log('Trying to load fallback language: zh-CN');
                    const fallbackResponse = await fetch('/i18n/messages/zh-CN');
                    if (fallbackResponse.ok) {
                        const fallbackData = await fallbackResponse.json();
                        this.translations = fallbackData.messages;
                        console.log('Fallback translations loaded successfully');
                    } else {
                        console.error('Failed to load fallback translations');
                        // Use hardcoded fallback for critical login elements
                        this.translations = this.getFallbackTranslations();
                    }
                } else {
                    // Use hardcoded fallback for critical login elements
                    this.translations = this.getFallbackTranslations();
                }
            }
        } catch (error) {
            console.error('Error occurred while loading translation files:', error);
            // Use hardcoded fallback for critical login elements
            this.translations = this.getFallbackTranslations();
        }
    }

    // Get fallback translations for critical UI elements
    getFallbackTranslations() {
        const fallbackTranslations = {
            'login': {
                'title': this.currentLanguage === 'zh-CN' ? '管理员登录' : 'Admin Login',
                'username': this.currentLanguage === 'zh-CN' ? '用户名' : 'Username',
                'password': this.currentLanguage === 'zh-CN' ? '密码' : 'Password',
                'login_button': this.currentLanguage === 'zh-CN' ? '登录' : 'Login',
                'logging_in': this.currentLanguage === 'zh-CN' ? '登录中...' : 'Logging in...',
                'logout': this.currentLanguage === 'zh-CN' ? '退出' : 'Logout',
                'login_error': this.currentLanguage === 'zh-CN' ? '登录失败' : 'Login failed',
                'login_success': this.currentLanguage === 'zh-CN' ? '登录成功' : 'Login successful',
                'network_error': this.currentLanguage === 'zh-CN' ? '网络错误，请重试' : 'Network error, please try again'
            },
            'menu': {
                'dashboard': this.currentLanguage === 'zh-CN' ? '仪表板' : 'Dashboard',
                'requests': this.currentLanguage === 'zh-CN' ? '请求日志' : 'Request Logs',
                'config': this.currentLanguage === 'zh-CN' ? '配置管理' : 'Configuration',
                'users': this.currentLanguage === 'zh-CN' ? '用户管理' : 'User Management',
                'test': this.currentLanguage === 'zh-CN' ? 'API测试' : 'API Testing'
            },
            'dashboard': {
                'title': this.currentLanguage === 'zh-CN' ? '仪表板' : 'Dashboard',
                'subtitle': this.currentLanguage === 'zh-CN' ? 'Claude Code Proxy 运行状态概览' : 'Claude Code Proxy Status Overview',
                'total_requests': this.currentLanguage === 'zh-CN' ? '总请求数' : 'Total Requests',
                'avg_response_time': this.currentLanguage === 'zh-CN' ? '平均响应时间' : 'Average Response Time',
                'success_rate': this.currentLanguage === 'zh-CN' ? '成功率' : 'Success Rate',
                'tokens_used': this.currentLanguage === 'zh-CN' ? 'Token使用量' : 'Tokens Used',
                'system_info': this.currentLanguage === 'zh-CN' ? '系统信息' : 'System Information',
                'version': this.currentLanguage === 'zh-CN' ? '版本' : 'Version',
                'uptime': this.currentLanguage === 'zh-CN' ? '运行时间' : 'Uptime',
                'current_model': this.currentLanguage === 'zh-CN' ? '当前模型' : 'Current Model',
                'recent_requests': this.currentLanguage === 'zh-CN' ? '最近请求' : 'Recent Requests',
                'loading': this.currentLanguage === 'zh-CN' ? '加载中...' : 'Loading...',
                'service_normal': this.currentLanguage === 'zh-CN' ? '服务正常' : 'Service Normal',
                'service_error': this.currentLanguage === 'zh-CN' ? '服务异常' : 'Service Error',
                'connection_failed': this.currentLanguage === 'zh-CN' ? '连接失败' : 'Connection Failed'
            },
            'requests': {
                'title': this.currentLanguage === 'zh-CN' ? '请求日志' : 'Request Logs',
                'subtitle': this.currentLanguage === 'zh-CN' ? '查看所有API请求的详细记录' : 'View detailed records of all API requests',
                'search_placeholder': this.currentLanguage === 'zh-CN' ? '搜索请求...' : 'Search requests...',
                'refresh': this.currentLanguage === 'zh-CN' ? '刷新' : 'Refresh',
                'time': this.currentLanguage === 'zh-CN' ? '时间' : 'Time',
                'model': this.currentLanguage === 'zh-CN' ? '模型' : 'Model',
                'status': this.currentLanguage === 'zh-CN' ? '状态' : 'Status',
                'response_time': this.currentLanguage === 'zh-CN' ? '响应时间' : 'Response Time',
                'token': this.currentLanguage === 'zh-CN' ? 'Token' : 'Token',
                'actions': this.currentLanguage === 'zh-CN' ? '操作' : 'Actions',
                'details': this.currentLanguage === 'zh-CN' ? '详情' : 'Details',
                'loading': this.currentLanguage === 'zh-CN' ? '加载中...' : 'Loading...',
                'success': this.currentLanguage === 'zh-CN' ? '成功' : 'Success',
                'failed': this.currentLanguage === 'zh-CN' ? '失败' : 'Failed',
                'warning': this.currentLanguage === 'zh-CN' ? '警告' : 'Warning'
            },
            'config': {
                'title': this.currentLanguage === 'zh-CN' ? '配置管理' : 'Configuration',
                'subtitle': this.currentLanguage === 'zh-CN' ? '管理代理服务器的配置参数' : 'Manage proxy server configuration parameters',
                'api_config': this.currentLanguage === 'zh-CN' ? 'API配置' : 'API Configuration',
                'openai_api_key': this.currentLanguage === 'zh-CN' ? 'OpenAI API Key' : 'OpenAI API Key',
                'claude_api_key': this.currentLanguage === 'zh-CN' ? 'Claude API Key' : 'Claude API Key',
                'openai_base_url': this.currentLanguage === 'zh-CN' ? 'OpenAI Base URL' : 'OpenAI Base URL',
                'claude_base_url': this.currentLanguage === 'zh-CN' ? 'Claude Base URL' : 'Claude Base URL',
                'openai_api_key_placeholder': this.currentLanguage === 'zh-CN' ? '输入OpenAI API Key' : 'Enter OpenAI API Key',
                'claude_api_key_placeholder': this.currentLanguage === 'zh-CN' ? '输入Claude API Key' : 'Enter Claude API Key',
                'openai_base_url_placeholder': this.currentLanguage === 'zh-CN' ? 'https://api.openai.com/v1' : 'https://api.openai.com/v1',
                'claude_base_url_placeholder': this.currentLanguage === 'zh-CN' ? 'https://api.anthropic.com' : 'https://api.anthropic.com',
                'test_api_key': this.currentLanguage === 'zh-CN' ? '🔍 测试API密钥' : '🔍 Test API Keys',
                'model_config': this.currentLanguage === 'zh-CN' ? '模型配置' : 'Model Configuration',
                'big_model': this.currentLanguage === 'zh-CN' ? '大模型 (Sonnet/Opus)' : 'Large Model (Sonnet/Opus)',
                'small_model': this.currentLanguage === 'zh-CN' ? '小模型 (Haiku)' : 'Small Model (Haiku)',
                'big_model_placeholder': this.currentLanguage === 'zh-CN' ? 'claude-3-5-sonnet-20241022' : 'claude-3-5-sonnet-20241022',
                'small_model_placeholder': this.currentLanguage === 'zh-CN' ? 'claude-3-haiku-20240307' : 'claude-3-haiku-20240307',
                'max_tokens': this.currentLanguage === 'zh-CN' ? '最大Token限制' : 'Max Tokens Limit',
                'max_tokens_placeholder': this.currentLanguage === 'zh-CN' ? '4096' : '4096',
                'request_timeout': this.currentLanguage === 'zh-CN' ? '请求超时 (秒)' : 'Request Timeout (seconds)',
                'request_timeout_placeholder': this.currentLanguage === 'zh-CN' ? '90' : '90',
                'server_config': this.currentLanguage === 'zh-CN' ? '服务器配置' : 'Server Configuration',
                'server_host': this.currentLanguage === 'zh-CN' ? '服务器地址' : 'Server Host',
                'server_port': this.currentLanguage === 'zh-CN' ? '服务器端口' : 'Server Port',
                'server_host_placeholder': this.currentLanguage === 'zh-CN' ? '0.0.0.0' : '0.0.0.0',
                'server_port_placeholder': this.currentLanguage === 'zh-CN' ? '8080' : '8080',
                'log_level': this.currentLanguage === 'zh-CN' ? '日志级别' : 'Log Level',
                'stream_enabled': this.currentLanguage === 'zh-CN' ? '启用流式响应' : 'Enable Streaming Response',
                'security_config': this.currentLanguage === 'zh-CN' ? '安全配置' : 'Security Configuration',
                'jwt_secret': this.currentLanguage === 'zh-CN' ? 'JWT密钥' : 'JWT Secret',
                'db_encrypt_key': this.currentLanguage === 'zh-CN' ? '数据库加密密钥' : 'Database Encryption Key',
                'jwt_secret_placeholder': this.currentLanguage === 'zh-CN' ? '输入JWT密钥' : 'Enter JWT Secret',
                'db_encrypt_key_placeholder': this.currentLanguage === 'zh-CN' ? '输入数据库加密密钥' : 'Enter Database Encryption Key',
                'encrypt_algo': this.currentLanguage === 'zh-CN' ? '配置加密算法' : 'Configuration Encryption Algorithm',
                'test_config': this.currentLanguage === 'zh-CN' ? '测试配置' : 'Test Configuration',
                'save_config': this.currentLanguage === 'zh-CN' ? '保存配置' : 'Save Configuration',
                'saving': this.currentLanguage === 'zh-CN' ? '保存中...' : 'Saving...',
                'config_saved': this.currentLanguage === 'zh-CN' ? '配置已保存' : 'Configuration Saved',
                'save_failed': this.currentLanguage === 'zh-CN' ? '保存失败' : 'Save Failed',
                'testing': this.currentLanguage === 'zh-CN' ? '测试中...' : 'Testing...',
                'testing_config': this.currentLanguage === 'zh-CN' ? '正在测试配置...' : 'Testing configuration...',
                'config_test_success': this.currentLanguage === 'zh-CN' ? '配置测试成功' : 'Configuration Test Successful',
                'config_test_failed': this.currentLanguage === 'zh-CN' ? '配置测试失败' : 'Configuration Test Failed',
                'testing_api_key': this.currentLanguage === 'zh-CN' ? '正在测试API密钥...' : 'Testing API keys...',
                'api_key_test_success': this.currentLanguage === 'zh-CN' ? 'API密钥测试成功' : 'API Key Test Successful',
                'api_key_test_failed': this.currentLanguage === 'zh-CN' ? 'API密钥测试失败' : 'API Key Test Failed',
                'network_error': this.currentLanguage === 'zh-CN' ? '网络错误' : 'Network Error'
            },
            'users': {
                'title': this.currentLanguage === 'zh-CN' ? '用户管理' : 'User Management',
                'subtitle': this.currentLanguage === 'zh-CN' ? '管理系统用户和权限' : 'Manage system users and permissions',
                'search_placeholder': this.currentLanguage === 'zh-CN' ? '搜索用户...' : 'Search users...',
                'add_user': this.currentLanguage === 'zh-CN' ? '添加用户' : 'Add User',
                'username': this.currentLanguage === 'zh-CN' ? '用户名' : 'Username',
                'email': this.currentLanguage === 'zh-CN' ? '邮箱' : 'Email',
                'password': this.currentLanguage === 'zh-CN' ? '密码' : 'Password',
                'role': this.currentLanguage === 'zh-CN' ? '角色' : 'Role',
                'status': this.currentLanguage === 'zh-CN' ? '状态' : 'Status',
                'last_login': this.currentLanguage === 'zh-CN' ? '最后登录' : 'Last Login',
                'actions': this.currentLanguage === 'zh-CN' ? '操作' : 'Actions',
                'edit': this.currentLanguage === 'zh-CN' ? '编辑' : 'Edit',
                'delete': this.currentLanguage === 'zh-CN' ? '删除' : 'Delete',
                'loading': this.currentLanguage === 'zh-CN' ? '加载中...' : 'Loading...',
                'admin': this.currentLanguage === 'zh-CN' ? '管理员' : 'Admin',
                'user': this.currentLanguage === 'zh-CN' ? '用户' : 'User',
                'active': this.currentLanguage === 'zh-CN' ? '活跃' : 'Active',
                'inactive': this.currentLanguage === 'zh-CN' ? '禁用' : 'Inactive',
                'never_logged_in': this.currentLanguage === 'zh-CN' ? '从未登录' : 'Never logged in',
                'role_admin': this.currentLanguage === 'zh-CN' ? '管理员' : 'Administrator',
                'role_user': this.currentLanguage === 'zh-CN' ? '用户' : 'User',
                'enable_user': this.currentLanguage === 'zh-CN' ? '启用用户' : 'Enable User',
                'add_user_modal': this.currentLanguage === 'zh-CN' ? '添加用户' : 'Add User',
                'edit_user_modal': this.currentLanguage === 'zh-CN' ? '编辑用户' : 'Edit User',
                'cancel': this.currentLanguage === 'zh-CN' ? '取消' : 'Cancel',
                'save': this.currentLanguage === 'zh-CN' ? '保存' : 'Save',
                'user_saved': this.currentLanguage === 'zh-CN' ? '用户已保存' : 'User Saved',
                'user_deleted': this.currentLanguage === 'zh-CN' ? '用户已删除' : 'User Deleted',
                'confirm_delete': this.currentLanguage === 'zh-CN' ? '确定要删除此用户吗？' : 'Are you sure you want to delete this user?'
            },
            'test': {
                'title': this.currentLanguage === 'zh-CN' ? 'API测试' : 'API Testing',
                'subtitle': this.currentLanguage === 'zh-CN' ? '测试Claude API的功能和性能' : 'Test Claude API functionality and performance',
                'connection_test': this.currentLanguage === 'zh-CN' ? '连接测试' : 'Connection Test',
                'connection_test_desc': this.currentLanguage === 'zh-CN' ? '测试与OpenAI API的连接状态' : 'Test connection status with OpenAI API',
                'test_connection': this.currentLanguage === 'zh-CN' ? '测试连接' : 'Test Connection',
                'message_test': this.currentLanguage === 'zh-CN' ? '消息测试' : 'Message Test',
                'select_model': this.currentLanguage === 'zh-CN' ? '选择模型' : 'Select Model',
                'test_message': this.currentLanguage === 'zh-CN' ? '测试消息' : 'Test Message',
                'test_message_placeholder': this.currentLanguage === 'zh-CN' ? '输入要测试的消息...' : 'Enter test message...',
                'stream_response': this.currentLanguage === 'zh-CN' ? '流式响应' : 'Stream Response',
                'send_test': this.currentLanguage === 'zh-CN' ? '发送测试' : 'Send Test',
                'testing': this.currentLanguage === 'zh-CN' ? '测试中...' : 'Testing...',
                'sending': this.currentLanguage === 'zh-CN' ? '发送中...' : 'Sending...',
                'connection_success': this.currentLanguage === 'zh-CN' ? '连接测试成功' : 'Connection Test Successful',
                'connection_failed': this.currentLanguage === 'zh-CN' ? '连接测试失败' : 'Connection Test Failed',
                'message_sent': this.currentLanguage === 'zh-CN' ? '消息发送成功' : 'Message Sent Successfully',
                'request_failed': this.currentLanguage === 'zh-CN' ? '请求失败' : 'Request Failed',
                'stream_test_success': this.currentLanguage === 'zh-CN' ? '流式响应测试成功' : 'Stream Response Test Successful',
                'model_used': this.currentLanguage === 'zh-CN' ? '使用模型' : 'Model Used',
                'duration': this.currentLanguage === 'zh-CN' ? '耗时' : 'Duration',
                'response_id': this.currentLanguage === 'zh-CN' ? '响应ID' : 'Response ID',
                'response': this.currentLanguage === 'zh-CN' ? '响应' : 'Response',
                'input_tokens': this.currentLanguage === 'zh-CN' ? '输入Token' : 'Input Tokens',
                'output_tokens': this.currentLanguage === 'zh-CN' ? '输出Token' : 'Output Tokens',
                'status_code': this.currentLanguage === 'zh-CN' ? '状态码' : 'Status Code',
                'error': this.currentLanguage === 'zh-CN' ? '错误' : 'Error',
                'check_network': this.currentLanguage === 'zh-CN' ? '请检查网络连接' : 'Please check network connection',
                'check_config': this.currentLanguage === 'zh-CN' ? '请检查配置' : 'Please check configuration',
                'suggestions': this.currentLanguage === 'zh-CN' ? '建议' : 'Suggestions'
            },
            'common': {
                'loading': this.currentLanguage === 'zh-CN' ? '加载中...' : 'Loading...',
                'error': this.currentLanguage === 'zh-CN' ? '错误' : 'Error',
                'success': this.currentLanguage === 'zh-CN' ? '成功' : 'Success',
                'network_error': this.currentLanguage === 'zh-CN' ? '网络错误，请重试' : 'Network error, please try again',
                'loading_failed': this.currentLanguage === 'zh-CN' ? '加载失败' : 'Loading Failed',
                'operation_failed': this.currentLanguage === 'zh-CN' ? '操作失败' : 'Operation Failed',
                'unknown_error': this.currentLanguage === 'zh-CN' ? '未知错误' : 'Unknown Error'
            },
            'setup': {
                'welcome': this.currentLanguage === 'zh-CN' ? '欢迎使用 CCany！让我们完成首次设置' : 'Welcome to CCany! Let\'s complete the initial setup',
                'version': this.currentLanguage === 'zh-CN' ? '版本:' : 'Version:',
                'build_time': this.currentLanguage === 'zh-CN' ? '构建时间:' : 'Build Time:',
                'create_admin': this.currentLanguage === 'zh-CN' ? '创建管理员账户' : 'Create Administrator Account',
                'admin_username': this.currentLanguage === 'zh-CN' ? '管理员用户名' : 'Administrator Username',
                'admin_password': this.currentLanguage === 'zh-CN' ? '管理员密码' : 'Administrator Password',
                'confirm_password': this.currentLanguage === 'zh-CN' ? '确认密码' : 'Confirm Password',
                'password_requirements': this.currentLanguage === 'zh-CN' ? '密码至少6个字符，建议包含大小写字母、数字和特殊字符' : 'Password must be at least 6 characters, recommend including uppercase letters, lowercase letters, numbers and special characters',
                'create_account': this.currentLanguage === 'zh-CN' ? '创建管理员账户' : 'Create Administrator Account',
                'creating_account': this.currentLanguage === 'zh-CN' ? '正在创建管理员账户...' : 'Creating administrator account...',
                'setup_complete': this.currentLanguage === 'zh-CN' ? '✅ 设置完成！' : '✅ Setup Complete!',
                'account_created': this.currentLanguage === 'zh-CN' ? '管理员账户已创建成功。您现在可以登录到系统。' : 'Administrator account has been created successfully. You can now log in to the system.',
                'go_to_login': this.currentLanguage === 'zh-CN' ? '前往登录' : 'Go to Login'
            }
        };
        console.log('Using fallback translations for language:', this.currentLanguage);
        return fallbackTranslations;
    }

    // Translate text
    t(key, params = {}) {
        const keys = key.split('.');
        let value = this.translations;
        
        for (const k of keys) {
            if (value && typeof value === 'object' && k in value) {
                value = value[k];
            } else {
                return key; // Return original key if translation not found
            }
        }
        
        if (typeof value === 'string') {
            // Simple parameter replacement
            return value.replace(/\{\{(\w+)\}\}/g, (match, paramKey) => {
                return params[paramKey] || match;
            });
        }
        
        return key;
    }

    // Apply translations to page
    applyTranslations() {
        // Update all elements with data-i18n attributes
        document.querySelectorAll('[data-i18n]').forEach(element => {
            const key = element.getAttribute('data-i18n');
            const translation = this.t(key);
            if (element.tagName === 'INPUT' && (element.type === 'text' || element.type === 'search')) {
                element.placeholder = translation;
            } else if (element.tagName === 'TEXTAREA') {
                element.placeholder = translation;
            } else {
                element.textContent = translation;
            }
        });
        
        // Update all elements with data-i18n-placeholder attributes
        document.querySelectorAll('[data-i18n-placeholder]').forEach(element => {
            const key = element.getAttribute('data-i18n-placeholder');
            const translation = this.t(key);
            if (element.tagName === 'INPUT' || element.tagName === 'TEXTAREA') {
                element.placeholder = translation;
            }
        });
        
        // Update all elements with data-i18n-title attributes
        document.querySelectorAll('[data-i18n-title]').forEach(element => {
            const key = element.getAttribute('data-i18n-title');
            const translation = this.t(key);
            element.title = translation;
        });
        
        // Update all elements with data-i18n-value attributes
        document.querySelectorAll('[data-i18n-value]').forEach(element => {
            const key = element.getAttribute('data-i18n-value');
            const translation = this.t(key);
            if (element.tagName === 'INPUT' && element.type === 'button') {
                element.value = translation;
            }
        });
        
        // After applying translations, ensure language selectors are set correctly
        this.setupLanguageSelector();
    }

    // Switch language
    async changeLanguage(language) {
        console.log('changeLanguage called with:', language);
        
        // 防止重复切换相同语言
        if (this.currentLanguage === language) {
            console.log('Language already set to:', language);
            return;
        }
        
        // 显示切换中的状态
        this.showLanguageChanging();
        
        const oldLanguage = this.currentLanguage;
        this.currentLanguage = language;
        
        try {
            // Save to localStorage
            localStorage.setItem('preferred_language', language);
            
            // Send language preference to backend
            try {
                console.log('Sending language preference to backend:', language);
                const response = await fetch('/i18n/language', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        language: language
                    })
                });
                console.log('Backend response:', response.status);
                
                if (!response.ok) {
                    console.warn('Backend language setting failed, but continuing with client-side change');
                }
            } catch (error) {
                console.error('Failed to set language preference on backend:', error);
            }
            
            // Reload translations and apply
            console.log('Loading translations for:', language);
            await this.loadTranslations();
            console.log('Applying translations...');
            this.applyTranslations();
            
            // Update language selector values
            this.updateLanguageSelectors();
            
            // Refresh current page data
            this.refreshCurrentPage();
            console.log('Language change completed successfully');
            
        } catch (error) {
            console.error('Language change failed:', error);
            // 回滚到原来的语言
            this.currentLanguage = oldLanguage;
            this.showNotification('语言切换失败，请重试', 'error');
        } finally {
            this.hideLanguageChanging();
        }
    }

    // Refresh current page data
    refreshCurrentPage() {
        switch (this.currentTab) {
            case 'dashboard':
                this.loadDashboardData();
                break;
            case 'requests':
                this.loadRequestLogs();
                break;
            case 'config':
                this.loadConfig();
                break;
            case 'users':
                this.loadUsers();
                break;
        }
    }

    // Get Cookie
    getCookie(name) {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; ${name}=`);
        if (parts.length === 2) return parts.pop().split(';').shift();
        return null;
    }

    // Setup language selector event listener
    setupLanguageSelector() {
        console.log('Setting up language selectors, current language:', this.currentLanguage);
        
        // Setup main app language selector
        const languageSelect = document.getElementById('languageSelect');
        if (languageSelect) {
            console.log('Found main app language selector');
            // Ensure the correct option is selected
            languageSelect.value = this.currentLanguage;
            // Force update if value didn't set properly
            if (languageSelect.value !== this.currentLanguage) {
                const option = languageSelect.querySelector(`option[value="${this.currentLanguage}"]`);
                if (option) {
                    option.selected = true;
                }
            }
            
            // 移除现有的事件监听器（如果存在）
            if (languageSelect._changeHandler) {
                languageSelect.removeEventListener('change', languageSelect._changeHandler);
                console.log('Removed existing main app language selector event listener');
            }
            
            // 创建新的事件处理器
            languageSelect._changeHandler = (e) => {
                console.log('Main app language selector changed to:', e.target.value);
                console.log('About to call changeLanguage with:', e.target.value);
                this.changeLanguage(e.target.value);
            };
            
            // 绑定新的事件监听器
            languageSelect.addEventListener('change', languageSelect._changeHandler);
            console.log('Main app language selector event listener attached, current value:', languageSelect.value);
        } else {
            console.log('Main app language selector not found');
        }
        
        // Setup login page language selector
        const loginLanguageSelect = document.getElementById('loginLanguageSelect');
        if (loginLanguageSelect) {
            console.log('Found login language selector');
            // Ensure the correct option is selected
            loginLanguageSelect.value = this.currentLanguage;
            // Force update if value didn't set properly
            if (loginLanguageSelect.value !== this.currentLanguage) {
                const option = loginLanguageSelect.querySelector(`option[value="${this.currentLanguage}"]`);
                if (option) {
                    option.selected = true;
                }
            }
            
            // 移除现有的事件监听器（如果存在）
            if (loginLanguageSelect._changeHandler) {
                loginLanguageSelect.removeEventListener('change', loginLanguageSelect._changeHandler);
                console.log('Removed existing login language selector event listener');
            }
            
            // 创建新的事件处理器
            loginLanguageSelect._changeHandler = (e) => {
                console.log('Login language selector changed to:', e.target.value);
                console.log('About to call changeLanguage with:', e.target.value);
                this.changeLanguage(e.target.value);
            };
            
            // 绑定新的事件监听器
            loginLanguageSelect.addEventListener('change', loginLanguageSelect._changeHandler);
            console.log('Login language selector event listener attached, current value:', loginLanguageSelect.value);
            
            // 添加额外的调试信息
            console.log('Login language selector element:', loginLanguageSelect);
            console.log('Login language selector options:', Array.from(loginLanguageSelect.options).map(opt => ({value: opt.value, text: opt.text})));
            
            // 测试事件监听器是否正常工作
            console.log('Testing event listener by simulating change...');
            loginLanguageSelect.dispatchEvent(new Event('change', { bubbles: true }));
        } else {
            console.log('Login language selector not found');
        }
    }

    setupEventListeners() {
        // 登录表单
        const loginForm = document.getElementById('loginForm');
        if (loginForm) {
            loginForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.handleLogin();
            });
        }

        // 登出按钮
        const logoutBtn = document.getElementById('logoutBtn');
        if (logoutBtn) {
            logoutBtn.addEventListener('click', () => this.handleLogout());
        }

        // 标签页切换
        document.querySelectorAll('.menu-item').forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const tab = item.dataset.tab;
                this.switchTab(tab);
            });
        });

        // 刷新按钮
        const refreshBtn = document.getElementById('refreshLogs');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => this.loadRequestLogs());
        }

        // 测试按钮
        const testConnectionBtn = document.getElementById('testConnection');
        if (testConnectionBtn) {
            testConnectionBtn.addEventListener('click', () => this.testConnection());
        }

        const testMessageBtn = document.getElementById('testMessageBtn');
        if (testMessageBtn) {
            testMessageBtn.addEventListener('click', () => this.testMessage());
        }

        // 配置管理按钮
        const testConfigBtn = document.getElementById('testConfigBtn');
        if (testConfigBtn) {
            testConfigBtn.addEventListener('click', () => this.testConfig());
        }

        const saveConfigBtn = document.getElementById('saveConfigBtn');
        if (saveConfigBtn) {
            saveConfigBtn.addEventListener('click', () => this.saveConfig());
        }

        // JWT密钥生成按钮
        const generateJwtBtn = document.getElementById('generateJwtBtn');
        if (generateJwtBtn) {
            generateJwtBtn.addEventListener('click', () => {
                const jwtSecretInput = document.getElementById('jwtSecret');
                if (jwtSecretInput) {
                    // Generate a random UUID for JWT secret
                    const uuid = this.generateUUID();
                    jwtSecretInput.value = uuid;
                    
                    // Show a brief success indication
                    const originalText = generateJwtBtn.textContent;
                    generateJwtBtn.textContent = '✅ 已生成';
                    generateJwtBtn.style.background = 'var(--success-color)';
                    
                    setTimeout(() => {
                        generateJwtBtn.textContent = originalText;
                        generateJwtBtn.style.background = '';
                    }, 2000);
                }
            });
        }

        // 初始化Anthropic Base URL
        this.initializeAnthropicBaseUrl();

        // API密钥测试按钮
        const testApiKeyBtn = document.getElementById('testApiKeyBtn');
        if (testApiKeyBtn) {
            testApiKeyBtn.addEventListener('click', () => this.testApiKey());
        }

        // 用户管理按钮
        const addUserBtn = document.getElementById('addUserBtn');
        if (addUserBtn) {
            addUserBtn.addEventListener('click', () => this.showUserModal());
        }

        // 用户表单
        const userForm = document.getElementById('userForm');
        if (userForm) {
            userForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.handleUserSave();
            });
        }

        // 模态框控制
        const closeUserModal = document.getElementById('closeUserModal');
        if (closeUserModal) {
            closeUserModal.addEventListener('click', () => this.hideUserModal());
        }

        const cancelUserBtn = document.getElementById('cancelUserBtn');
        if (cancelUserBtn) {
            cancelUserBtn.addEventListener('click', () => this.hideUserModal());
        }

        // 搜索功能
        const searchInput = document.getElementById('searchInput');
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                this.filterRequests(e.target.value);
            });
        }

        const searchUsers = document.getElementById('searchUsers');
        if (searchUsers) {
            searchUsers.addEventListener('input', (e) => {
                this.filterUsers(e.target.value);
            });
        }

        // 模态框点击外部关闭
        const userModal = document.getElementById('userModal');
        if (userModal) {
            userModal.addEventListener('click', (e) => {
                if (e.target === userModal) {
                    this.hideUserModal();
                }
            });
        }

        // 语言选择器将在showMainApp()中设置
    }

    // 身份验证相关
    async checkAuthentication() {
        const token = localStorage.getItem('auth_token');
        if (!token) {
            this.showLoginPage();
            return;
        }

        try {
            const response = await fetch('/auth/me', {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            if (response.ok) {
                const data = await response.json();
                this.currentUser = data.user;
                this.isAuthenticated = true;
                this.showMainApp();
                this.loadInitialData();
            } else {
                localStorage.removeItem('auth_token');
                this.showLoginPage();
            }
        } catch (error) {
            console.error('Authentication check failed:', error);
            this.showLoginPage();
        }
    }

    async handleLogin() {
        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;
        const loginError = document.getElementById('loginError');
        const loginBtn = document.querySelector('.login-btn');

        // Enhanced loading state
        this.uiController.showLoadingState(loginBtn, this.t('login.logging_in'));
        loginError.classList.remove('show');

        // Add form validation animations
        const usernameField = document.getElementById('username');
        const passwordField = document.getElementById('password');
        
        this.uiController.validateFormField(usernameField, username.length > 0);
        this.uiController.validateFormField(passwordField, password.length > 0);

        try {
            const response = await fetch('/auth/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username, password })
            });

            const data = await response.json();

            if (response.ok) {
                localStorage.setItem('auth_token', data.token);
                this.currentUser = data.user;
                this.isAuthenticated = true;
                
                // Show success animation
                this.uiController.showEnhancedNotification(this.t('login.login_success') || '登录成功', 'success');
                
                // Smooth transition to main app
                setTimeout(() => {
                    this.showMainApp();
                    this.loadInitialData();
                }, 500);
            } else {
                loginError.textContent = data.error || this.t('login.login_error');
                loginError.classList.add('show');
                
                // Shake animation for login form
                const loginCard = document.querySelector('.login-card');
                if (loginCard) {
                    loginCard.classList.add('animate-shake');
                    setTimeout(() => {
                        loginCard.classList.remove('animate-shake');
                    }, 500);
                }
            }
        } catch (error) {
            console.error('Login failed:', error);
            loginError.textContent = this.t('login.network_error');
            loginError.classList.add('show');
        } finally {
            this.uiController.hideLoadingState(loginBtn);
        }
    }

    async handleLogout() {
        try {
            await fetch('/auth/logout', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
                }
            });
        } catch (error) {
            console.error('Logout failed:', error);
        }

        localStorage.removeItem('auth_token');
        this.currentUser = null;
        this.isAuthenticated = false;
        this.showLoginPage();
    }

    showLoginPage() {
        console.log('showLoginPage called');
        document.getElementById('loginContainer').style.display = 'flex';
        document.getElementById('mainApp').style.display = 'none';
        
        // Setup language selector for login page
        console.log('About to call setupLanguageSelector from showLoginPage');
        this.setupLanguageSelector();
    }

    showMainApp() {
        document.getElementById('loginContainer').style.display = 'none';
        document.getElementById('mainApp').style.display = 'flex';
        
        // Update user info display
        const userNameElement = document.getElementById('userName');
        if (userNameElement && this.currentUser) {
            userNameElement.textContent = this.currentUser.username;
        }

        // Setup language selector after switching to main app
        this.setupLanguageSelector();
    }

    // API调用辅助函数
    async apiCall(url, options = {}) {
        const token = localStorage.getItem('auth_token');
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            }
        };

        const mergedOptions = {
            ...defaultOptions,
            ...options,
            headers: {
                ...defaultOptions.headers,
                ...(options.headers || {})
            }
        };

        const response = await fetch(url, mergedOptions);
        
        if (response.status === 401) {
            localStorage.removeItem('auth_token');
            this.showLoginPage();
            throw new Error('Authentication required');
        }

        return response;
    }

    switchTab(tabName) {
        // Prevent rapid tab switching
        if (this.debounceTimers.has('tabSwitch')) {
            clearTimeout(this.debounceTimers.get('tabSwitch'));
        }
        
        this.debounceTimers.set('tabSwitch', setTimeout(() => {
            // 更新菜单状态
            document.querySelectorAll('.menu-item').forEach(item => {
                item.classList.remove('active');
            });
            document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');

            // Enhanced tab content switching with animations
            const currentContent = document.querySelector('.tab-content.active');
            const newContent = document.getElementById(tabName);
            
            if (currentContent && newContent && currentContent !== newContent) {
                // Fade out current content
                currentContent.style.opacity = '0';
                currentContent.style.transform = 'translateX(-20px)';
                
                setTimeout(() => {
                    currentContent.classList.remove('active');
                    newContent.classList.add('active');
                    
                    // Fade in new content
                    newContent.style.opacity = '0';
                    newContent.style.transform = 'translateX(20px)';
                    
                    setTimeout(() => {
                        newContent.style.opacity = '1';
                        newContent.style.transform = 'translateX(0)';
                    }, 50);
                }, 150);
            } else if (newContent) {
                document.querySelectorAll('.tab-content').forEach(content => {
                    content.classList.remove('active');
                });
                newContent.classList.add('active');
                newContent.classList.add('animate-fade-in');
            }

            this.currentTab = tabName;

            // 加载对应数据
            switch (tabName) {
                case 'dashboard':
                    this.loadDashboardData();
                    break;
                case 'requests':
                    this.loadRequestLogs();
                    break;
                case 'config':
                    this.loadConfig();
                    break;
                case 'users':
                    this.loadUsers();
                    break;
            }
        }, 100));
    }

    async loadInitialData() {
        await this.checkHealth();
        await this.loadConfig();
        await this.loadDashboardData();
    }

    async checkHealth() {
        try {
            const response = await fetch('/health');
            const data = await response.json();
            
            const statusDot = document.getElementById('statusDot');
            const statusText = document.getElementById('statusText');
            
            if (data.status === 'healthy') {
                statusDot.className = 'status-dot';
                statusText.textContent = this.t('dashboard.service_normal');
            } else {
                statusDot.className = 'status-dot error';
                statusText.textContent = this.t('dashboard.service_error');
            }
        } catch (error) {
            console.error('健康检查失败:', error);
            const statusDot = document.getElementById('statusDot');
            const statusText = document.getElementById('statusText');
            statusDot.className = 'status-dot error';
            statusText.textContent = this.t('dashboard.connection_failed');
        }
    }

    async loadConfig() {
        try {
            console.log('Loading configuration from backend...');
            const response = await this.apiCall('/admin/config');
            console.log('Config response status:', response.status);
            const data = await response.json();
            console.log('Config response data:', data);
            this.config = data.config;
            console.log('Parsed config:', this.config);
            
            // 更新配置页面
            if (this.config) {
                console.log('Updating config display...');
                this.updateConfigDisplay();
            } else {
                console.warn('No config data received from backend');
            }
        } catch (error) {
            console.error('加载配置失败:', error);
        }
    }

    updateConfigDisplay() {
        const elements = {
            'openaiApiKey': this.config.openai_api_key,
            'anthropicAuthToken': this.config.claude_api_key,
            'openaiBaseUrl': this.config.openai_base_url,
            'anthropicBaseUrl': this.config.claude_base_url,
            'bigModel': this.config.big_model,
            'smallModel': this.config.small_model,
            'maxTokensLimit': this.config.max_tokens_limit,
            'requestTimeout': this.config.request_timeout,
            'serverHost': this.config.host,
            'serverPort': this.config.port,
            'logLevel': this.config.log_level,
            'jwtSecret': this.config.jwt_secret,
            'encryptAlgo': this.config.encrypt_algorithm
        };

        Object.entries(elements).forEach(([id, value]) => {
            const element = document.getElementById(id);
            if (element) {
                if (element.type === 'checkbox') {
                    element.checked = value === 'true' || value === true;
                } else {
                    element.value = value || '';
                }
            }
        });
        
        // 初始化Anthropic Base URL为当前页面URL
        this.initializeAnthropicBaseUrl();

        // 更新仪表板中的当前模型
        const currentModelElement = document.getElementById('currentModel');
        if (currentModelElement) {
            currentModelElement.textContent = `${this.config.big_model} / ${this.config.small_model}`;
        }
    }

    async saveConfig() {
        const configData = {
            openai_api_key: document.getElementById('openaiApiKey').value,
            claude_api_key: document.getElementById('anthropicAuthToken').value,
            openai_base_url: document.getElementById('openaiBaseUrl').value,
            claude_base_url: document.getElementById('anthropicBaseUrl').value,
            big_model: document.getElementById('bigModel').value,
            small_model: document.getElementById('smallModel').value,
            max_tokens_limit: parseInt(document.getElementById('maxTokensLimit').value) || 4096,
            request_timeout: parseInt(document.getElementById('requestTimeout').value) || 90,
            host: document.getElementById('serverHost').value,
            port: parseInt(document.getElementById('serverPort').value) || 8080,
            log_level: document.getElementById('logLevel').value,
            jwt_secret: document.getElementById('jwtSecret').value,
            encrypt_algorithm: document.getElementById('encryptAlgo').value
        };

        console.log('Saving configuration data:', configData);

        const resultElement = document.getElementById('configResult');
        const saveBtn = document.getElementById('saveConfigBtn');

        // Enhanced loading state
        this.uiController.showLoadingState(saveBtn, this.t('config.saving'));

        try {
            console.log('Sending PUT request to /admin/config');
            const response = await this.apiCall('/admin/config', {
                method: 'PUT',
                body: JSON.stringify(configData)
            });

            console.log('Save config response status:', response.status);
            const data = await response.json();
            console.log('Save config response data:', data);

            if (response.ok) {
                resultElement.className = 'config-result success';
                resultElement.textContent = this.t('config.config_saved');
                this.config = data.config;
                
                console.log('Configuration saved successfully, updated count:', data.updated);
                
                // Show success animation
                this.uiController.showEnhancedNotification(this.t('config.config_saved'), 'success');
                
                // Animate the result element
                resultElement.classList.add('animate-fade-in');
            } else {
                console.error('Save config failed with status:', response.status, 'data:', data);
                resultElement.className = 'config-result error';
                resultElement.textContent = data.error || this.t('config.save_failed');
                
                // Shake animation for error
                resultElement.classList.add('animate-shake');
                setTimeout(() => {
                    resultElement.classList.remove('animate-shake');
                }, 500);
            }
        } catch (error) {
            console.error('保存配置失败:', error);
            resultElement.className = 'config-result error';
            resultElement.textContent = this.t('common.network_error');
        } finally {
            this.uiController.hideLoadingState(saveBtn);
        }
    }

    async testConfig() {
        const resultElement = document.getElementById('configResult');
        const testBtn = document.getElementById('testConfigBtn');

        testBtn.disabled = true;
        testBtn.textContent = this.t('config.testing');
        resultElement.className = 'config-result loading';
        resultElement.textContent = this.t('config.testing_config');

        try {
            const response = await this.apiCall('/admin/config/test', {
                method: 'POST'
            });

            const data = await response.json();

            if (response.ok) {
                resultElement.className = 'config-result success';
                resultElement.innerHTML = `
                    <strong>${this.t('config.config_test_success')}</strong><br>
                    ${data.results.map(r => `${r.service}: ${r.status}`).join('<br>')}
                `;
            } else {
                resultElement.className = 'config-result error';
                resultElement.innerHTML = `
                    <strong>${this.t('config.config_test_failed')}</strong><br>
                    ${data.error || this.t('common.unknown_error')}
                `;
            }
        } catch (error) {
            console.error('测试配置失败:', error);
            resultElement.className = 'config-result error';
            resultElement.textContent = this.t('common.network_error');
        } finally {
            testBtn.disabled = false;
            testBtn.textContent = this.t('config.test_config');
        }
    }

    async testApiKey() {
        const resultElement = document.getElementById('apiTestResult');
        const testBtn = document.getElementById('testApiKeyBtn');

        testBtn.disabled = true;
        testBtn.textContent = this.t('config.testing');
        resultElement.className = 'api-test-result loading';
        resultElement.textContent = this.t('config.testing_api_key');

        try {
            const response = await this.apiCall('/admin/config/test-api-key', {
                method: 'POST'
            });

            const data = await response.json();

            if (response.ok && data.success) {
                resultElement.className = 'api-test-result success';
                let resultHtml = `<strong>${this.t('config.api_key_test_success')}</strong><br>`;
                
                if (data.results && data.results.length > 0) {
                    resultHtml += '<div class="api-test-details">';
                    data.results.forEach(result => {
                        if (result.success) {
                            resultHtml += `
                                <div class="api-test-item success">
                                    <strong>${result.service}:</strong> ✅ ${result.message}<br>
                                    ${result.model ? `${this.t('test.model_used')}: ${result.model}<br>` : ''}
                                    ${result.response_time ? `${this.t('test.duration')}: ${result.response_time}` : ''}
                                </div>
                            `;
                        } else {
                            resultHtml += `
                                <div class="api-test-item error">
                                    <strong>${result.service}:</strong> ❌ ${result.error || this.t('test.request_failed')}
                                </div>
                            `;
                        }
                    });
                    resultHtml += '</div>';
                }
                
                resultElement.innerHTML = resultHtml;
            } else {
                resultElement.className = 'api-test-result error';
                let errorHtml = `<strong>${this.t('config.api_key_test_failed')}</strong><br>`;
                
                if (data.results && data.results.length > 0) {
                    errorHtml += '<div class="api-test-details">';
                    data.results.forEach(result => {
                        if (!result.success) {
                            errorHtml += `
                                <div class="api-test-item error">
                                    <strong>${result.service}:</strong> ${result.error || this.t('test.request_failed')}
                                </div>
                            `;
                        }
                    });
                    errorHtml += '</div>';
                } else {
                    errorHtml += `${this.t('test.error')}: ${data.error || this.t('common.unknown_error')}`;
                }
                
                resultElement.innerHTML = errorHtml;
            }
        } catch (error) {
            console.error('API密钥测试失败:', error);
            resultElement.className = 'api-test-result error';
            resultElement.innerHTML = `
                <strong>${this.t('config.network_error')}</strong><br>
                ${error.message || this.t('test.check_network')}
            `;
        } finally {
            testBtn.disabled = false;
            testBtn.textContent = this.t('config.test_api_key');
        }
    }

    async loadUsers() {
        const tableBody = document.getElementById('usersTableBody');
        if (!tableBody) return;

        tableBody.innerHTML = `<tr><td colspan="6" class="loading">${this.t('users.loading')}</td></tr>`;

        try {
            const response = await this.apiCall('/admin/users');
            const data = await response.json();

            if (response.ok) {
                const users = data.users || [];
                const html = users.map(user => `
                    <tr>
                        <td>${user.username}</td>
                        <td>${user.email}</td>
                        <td><span class="role-${user.role}">${user.role === 'admin' ? this.t('users.admin') : this.t('users.user')}</span></td>
                        <td><span class="status-${user.is_active ? 'active' : 'inactive'}">${user.is_active ? this.t('users.active') : this.t('users.inactive')}</span></td>
                        <td>${user.last_login ? new Date(user.last_login).toLocaleString('zh-CN') : this.t('users.never_logged_in')}</td>
                        <td>
                            <div class="user-actions">
                                <button class="edit-btn" onclick="app.editUser('${user.id}')">${this.t('users.edit')}</button>
                                <button class="delete-btn" onclick="app.deleteUser('${user.id}')">${this.t('users.delete')}</button>
                            </div>
                        </td>
                    </tr>
                `).join('');
                
                tableBody.innerHTML = html;
            } else {
                tableBody.innerHTML = `<tr><td colspan="6" class="error">${this.t('common.loading_failed')}: ${data.error}</td></tr>`;
            }
        } catch (error) {
            console.error('加载用户失败:', error);
            tableBody.innerHTML = `<tr><td colspan="6" class="error">${this.t('common.network_error')}</td></tr>`;
        }
    }

    showUserModal(user = null) {
        const modal = document.getElementById('userModal');
        const title = document.getElementById('userModalTitle');
        const form = document.getElementById('userForm');
        
        if (user) {
            title.textContent = this.t('users.edit_user_modal');
            document.getElementById('userId').value = user.id;
            document.getElementById('userUsername').value = user.username;
            document.getElementById('userEmail').value = user.email;
            document.getElementById('userPassword').required = false;
            document.getElementById('userRole').value = user.role;
            document.getElementById('userActive').checked = user.is_active;
        } else {
            title.textContent = this.t('users.add_user_modal');
            form.reset();
            document.getElementById('userId').value = '';
            document.getElementById('userPassword').required = true;
        }
        
        modal.classList.add('show');
    }

    hideUserModal() {
        const modal = document.getElementById('userModal');
        modal.classList.remove('show');
    }

    async handleUserSave() {
        const userId = document.getElementById('userId').value;
        const userData = {
            username: document.getElementById('userUsername').value,
            email: document.getElementById('userEmail').value,
            password: document.getElementById('userPassword').value,
            role: document.getElementById('userRole').value,
            is_active: document.getElementById('userActive').checked
        };

        try {
            const url = userId ? `/admin/users/${userId}` : '/admin/users';
            const method = userId ? 'PUT' : 'POST';
            
            const response = await this.apiCall(url, {
                method: method,
                body: JSON.stringify(userData)
            });

            const data = await response.json();

            if (response.ok) {
                this.hideUserModal();
                this.loadUsers();
                this.showNotification(this.t('users.user_saved'), 'success');
            } else {
                this.showNotification(data.error || this.t('common.operation_failed'), 'error');
            }
        } catch (error) {
            console.error('保存用户失败:', error);
            this.showNotification(this.t('common.network_error'), 'error');
        }
    }

    async editUser(userId) {
        try {
            const response = await this.apiCall(`/admin/users/${userId}`);
            const data = await response.json();

            if (response.ok) {
                this.showUserModal(data.user);
            } else {
                this.showNotification(data.error || this.t('common.operation_failed'), 'error');
            }
        } catch (error) {
            console.error('获取用户失败:', error);
            this.showNotification(this.t('common.network_error'), 'error');
        }
    }

    async deleteUser(userId) {
        if (!confirm(this.t('users.confirm_delete'))) {
            return;
        }

        try {
            const response = await this.apiCall(`/admin/users/${userId}`, {
                method: 'DELETE'
            });

            const data = await response.json();

            if (response.ok) {
                this.loadUsers();
                this.showNotification(this.t('users.user_deleted'), 'success');
            } else {
                this.showNotification(data.error || this.t('common.operation_failed'), 'error');
            }
        } catch (error) {
            console.error('删除用户失败:', error);
            this.showNotification(this.t('common.network_error'), 'error');
        }
    }

    showNotification(message, type = 'info') {
        // Use enhanced notification system
        this.uiController.showEnhancedNotification(message, type);
    }

    async loadDashboardData() {
        try {
            // 获取真实的系统信息
            const response = await this.apiCall('/admin/monitoring/info');
            const data = await response.json();
            
            // 解析真实数据
            const stats = {
                totalRequests: this.formatNumber(data.performance.request_count || 0),
                avgResponseTime: data.performance.avg_response_time ? `${data.performance.avg_response_time.toFixed(0)}ms` : '0ms',
                successRate: data.performance.request_count > 0 ? 
                    `${(((data.performance.request_count - data.performance.error_count) / data.performance.request_count) * 100).toFixed(1)}%` : '100%',
                tokensUsed: this.formatNumber(data.performance.token_count || 0)
            };

            // 更新仪表板显示
            Object.entries(stats).forEach(([key, value]) => {
                const element = document.getElementById(key);
                if (element) {
                    // 为数值类型添加动画效果
                    if (key === 'totalRequests') {
                        const numericValue = parseInt(value);
                        const currentValue = parseInt(element.textContent) || 0;
                        if (numericValue !== currentValue) {
                            this.uiController.animateCounter(element, currentValue, numericValue);
                        }
                    } else {
                        // 其他值使用淡入淡出效果
                        element.style.opacity = '0.5';
                        setTimeout(() => {
                            element.textContent = value;
                            element.style.opacity = '1';
                        }, 150);
                    }
                }
            });

            // 更新系统信息
            const versionElement = document.getElementById('version');
            if (versionElement) {
                versionElement.textContent = data.application.version || 'Unknown';
            }

            const uptimeElement = document.getElementById('uptime');
            if (uptimeElement) {
                uptimeElement.textContent = data.application.uptime || 'Unknown';
            }

        } catch (error) {
            console.error('Failed to load dashboard data:', error);
            
            // 如果API调用失败，显示错误信息
            const errorStats = {
                totalRequests: 'N/A',
                avgResponseTime: 'N/A',
                successRate: 'N/A',
                tokensUsed: 'N/A'
            };

            Object.entries(errorStats).forEach(([key, value]) => {
                const element = document.getElementById(key);
                if (element) {
                    element.textContent = value;
                }
            });
        }

        // 加载最近请求
        this.loadRecentRequests();
    }

    async loadRecentRequests() {
        const recentRequestsElement = document.getElementById('recentRequests');
        if (!recentRequestsElement) return;

        try {
            // 获取真实的最近请求数据
            const response = await this.apiCall('/admin/request-logs?limit=5');
            const data = await response.json();
            
            if (data.requests && data.requests.length > 0) {
                const html = data.requests.map(req => {
                    const time = new Date(req.created_at).toLocaleString();
                    const statusClass = req.status === 'success' ? 'success' : req.status === 'error' ? 'error' : 'warning';
                    const statusText = req.status === 'success' ? this.t('requests.success') : 
                                     req.status === 'error' ? this.t('requests.failed') : this.t('requests.warning');
                    
                    return `
                        <div class="recent-request-item">
                            <div class="request-info">
                                <span class="request-time">${time}</span>
                                <span class="request-model">${req.model || 'Unknown'}</span>
                            </div>
                            <div class="request-status">
                                <span class="status-badge ${statusClass}">${statusText}</span>
                                <span class="request-tokens">${req.input_tokens || 0} / ${req.output_tokens || 0} tokens</span>
                            </div>
                        </div>
                    `;
                }).join('');

                recentRequestsElement.innerHTML = html;
            } else {
                recentRequestsElement.innerHTML = '<p class="no-data">No recent requests</p>';
            }
        } catch (error) {
            console.error('Failed to load recent requests:', error);
            recentRequestsElement.innerHTML = '<p class="error-message">Failed to load recent requests</p>';
        }
    }

    async loadRequestLogs() {
        const tableBody = document.getElementById('requestsTableBody');
        if (!tableBody) return;

        tableBody.innerHTML = `<tr><td colspan="6" class="loading">${this.t('requests.loading')}</td></tr>`;

        // 模拟请求日志数据
        setTimeout(() => {
            const mockLogs = this.generateMockLogs(20);
            const html = mockLogs.map(log => `
                <tr>
                    <td>${log.time}</td>
                    <td>${log.model}</td>
                    <td><span class="status-badge ${log.status}">${log.statusText}</span></td>
                    <td>${log.responseTime}</td>
                    <td>${log.tokens}</td>
                    <td>
                        <button class="action-btn" onclick="app.viewRequestDetails('${log.id}')">${this.t('requests.details')}</button>
                    </td>
                </tr>
            `).join('');
            
            tableBody.innerHTML = html;
        }, 500);
    }

    generateMockLogs(count) {
        const models = ['Claude 3.5 Sonnet', 'Claude 3 Haiku', 'Claude 3 Opus'];
        const statuses = [
            { status: 'success', text: this.t('requests.success') },
            { status: 'error', text: this.t('requests.failed') },
            { status: 'warning', text: this.t('requests.warning') }
        ];

        return Array.from({ length: count }, (_, i) => {
            const status = statuses[Math.floor(Math.random() * statuses.length)];
            const isSuccess = status.status === 'success';
            
            return {
                id: `req_${Date.now()}_${i}`,
                time: this.getRandomTime(),
                model: models[Math.floor(Math.random() * models.length)],
                status: status.status,
                statusText: status.text,
                responseTime: isSuccess ? `${Math.floor(Math.random() * 3000) + 200}ms` : '-',
                tokens: isSuccess ? Math.floor(Math.random() * 2000) + 100 : 0
            };
        });
    }

    async testConnection() {
        const resultElement = document.getElementById('connectionResult');
        const button = document.getElementById('testConnection');
        
        button.disabled = true;
        button.textContent = this.t('test.testing');
        resultElement.className = 'test-result loading';
        resultElement.textContent = this.t('test.testing');

        try {
            const response = await fetch('/test-connection');
            const data = await response.json();
            
            if (response.ok && data.status === 'success') {
                resultElement.className = 'test-result success';
                resultElement.innerHTML = `
                    <strong>${this.t('test.connection_success')}</strong><br>
                    ${this.t('test.model_used')}: ${data.model_used}<br>
                    ${this.t('test.duration')}: ${data.duration}<br>
                    ${this.t('test.response_id')}: ${data.response_id}
                `;
            } else {
                resultElement.className = 'test-result error';
                resultElement.innerHTML = `
                    <strong>${this.t('test.connection_failed')}</strong><br>
                    ${this.t('test.error')}: ${data.message || this.t('common.unknown_error')}<br>
                    ${this.t('test.suggestions')}: ${data.suggestions ? data.suggestions.join(', ') : this.t('test.check_config')}
                `;
            }
        } catch (error) {
            resultElement.className = 'test-result error';
            resultElement.innerHTML = `
                <strong>${this.t('test.connection_failed')}</strong><br>
                ${this.t('test.error')}: ${error.message}
            `;
        } finally {
            button.disabled = false;
            button.textContent = this.t('test.test_connection');
        }
    }

    async testMessage() {
        const resultElement = document.getElementById('messageResult');
        const button = document.getElementById('testMessageBtn');
        const model = document.getElementById('testModel').value;
        const message = document.getElementById('testMessage').value;
        const isStream = document.getElementById('testStream').checked;
        
        if (!message.trim()) {
            resultElement.className = 'test-result error';
            resultElement.textContent = this.t('test.test_message_placeholder');
            return;
        }

        button.disabled = true;
        button.textContent = this.t('test.sending');
        resultElement.className = 'test-result loading';
        resultElement.textContent = this.t('test.sending');

        try {
            const requestBody = {
                model: model,
                max_tokens: 100,
                messages: [
                    { role: 'user', content: message }
                ],
                stream: isStream
            };

            const response = await fetch('/v1/messages', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(requestBody)
            });

            if (response.ok) {
                if (isStream) {
                    resultElement.className = 'test-result success';
                    resultElement.innerHTML = `<strong>${this.t('test.stream_test_success')}</strong><br>${this.t('test.check_network')}`;
                } else {
                    const data = await response.json();
                    resultElement.className = 'test-result success';
                    resultElement.innerHTML = `
                        <strong>${this.t('test.message_sent')}</strong><br>
                        ${this.t('test.response')}: ${data.content && data.content[0] ? data.content[0].text.substring(0, 200) + '...' : this.t('common.unknown_error')}<br>
                        ${this.t('test.input_tokens')}: ${data.usage.input_tokens}<br>
                        ${this.t('test.output_tokens')}: ${data.usage.output_tokens}
                    `;
                }
            } else {
                const errorData = await response.json();
                resultElement.className = 'test-result error';
                resultElement.innerHTML = `
                    <strong>${this.t('test.request_failed')}</strong><br>
                    ${this.t('test.status_code')}: ${response.status}<br>
                    ${this.t('test.error')}: ${errorData.error ? errorData.error.message : this.t('common.unknown_error')}
                `;
            }
        } catch (error) {
            resultElement.className = 'test-result error';
            resultElement.innerHTML = `
                <strong>${this.t('test.request_failed')}</strong><br>
                ${this.t('test.error')}: ${error.message}
            `;
        } finally {
            button.disabled = false;
            button.textContent = this.t('test.send_test');
        }
    }

    filterRequests(searchTerm) {
        const tableBody = document.getElementById('requestsTableBody');
        if (!tableBody) return;

        const rows = tableBody.querySelectorAll('tr');
        rows.forEach(row => {
            const text = row.textContent.toLowerCase();
            const shouldShow = text.includes(searchTerm.toLowerCase());
            row.style.display = shouldShow ? '' : 'none';
        });
    }

    filterUsers(searchTerm) {
        const tableBody = document.getElementById('usersTableBody');
        if (!tableBody) return;

        const rows = tableBody.querySelectorAll('tr');
        rows.forEach(row => {
            const text = row.textContent.toLowerCase();
            const shouldShow = text.includes(searchTerm.toLowerCase());
            row.style.display = shouldShow ? '' : 'none';
        });
    }

    viewRequestDetails(requestId) {
        // 模拟显示请求详情
        alert(`查看请求详情: ${requestId}\n\n这里将显示完整的请求和响应信息。`);
    }

    startPeriodicUpdates() {
        // 每30秒更新一次健康状态
        setInterval(() => {
            if (this.isAuthenticated) {
                this.checkHealth();
            }
        }, 30000);

        // 每60秒更新一次仪表板数据
        setInterval(() => {
            if (this.isAuthenticated && this.currentTab === 'dashboard') {
                this.loadDashboardData();
            }
        }, 60000);
    }

    // 工具函数
    formatNumber(num) {
        if (num >= 1000000) {
            return (num / 1000000).toFixed(1) + 'M';
        } else if (num >= 1000) {
            return (num / 1000).toFixed(1) + 'K';
        }
        return num.toString();
    }

    calculateUptime() {
        // 模拟运行时间计算
        const hours = Math.floor(Math.random() * 72) + 1;
        const minutes = Math.floor(Math.random() * 60);
        return `${hours}小时${minutes}分钟`;
    }

    getRandomTime() {
        const now = new Date();
        const randomMinutes = Math.floor(Math.random() * 1440); // 24小时内
        const time = new Date(now.getTime() - randomMinutes * 60000);
        return time.toLocaleString('zh-CN');
    }

    // Enhanced loading indicator with Claude branding
    showLoadingIndicator() {
        // 创建加载指示器如果不存在
        let loadingIndicator = document.getElementById('appLoadingIndicator');
        if (!loadingIndicator) {
            loadingIndicator = document.createElement('div');
            loadingIndicator.id = 'appLoadingIndicator';
            loadingIndicator.innerHTML = `
                <div class="loading-overlay">
                    <div class="loading-container">
                        <div class="loading-logo">
                            <div class="logo-icon animate-pulse">🤖</div>
                            <h2 class="logo-text">CCany</h2>
                        </div>
                        <div class="loading-spinner-container">
                            <div class="loading-spinner enhanced"></div>
                        </div>
                        <p class="loading-text">正在加载...</p>
                        <div class="loading-progress">
                            <div class="progress-bar"></div>
                        </div>
                    </div>
                </div>
                <style>
                    .loading-overlay {
                        position: fixed;
                        top: 0;
                        left: 0;
                        right: 0;
                        bottom: 0;
                        background: linear-gradient(135deg, rgba(255, 107, 53, 0.1), rgba(74, 144, 226, 0.1));
                        backdrop-filter: blur(10px);
                        display: flex;
                        align-items: center;
                        justify-content: center;
                        z-index: 9999;
                        animation: fadeIn 0.3s ease;
                    }
                    .loading-container {
                        text-align: center;
                        padding: 40px;
                        background: rgba(255, 255, 255, 0.9);
                        border-radius: 24px;
                        box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
                        backdrop-filter: blur(20px);
                        border: 1px solid rgba(255, 255, 255, 0.2);
                    }
                    .loading-logo {
                        margin-bottom: 30px;
                    }
                    .logo-icon {
                        font-size: 48px;
                        margin-bottom: 15px;
                        background: linear-gradient(135deg, var(--claude-primary), var(--claude-secondary));
                        -webkit-background-clip: text;
                        -webkit-text-fill-color: transparent;
                        background-clip: text;
                    }
                    .logo-text {
                        font-size: 24px;
                        font-weight: 600;
                        background: linear-gradient(135deg, var(--claude-primary), var(--claude-secondary));
                        -webkit-background-clip: text;
                        -webkit-text-fill-color: transparent;
                        background-clip: text;
                        margin: 0;
                    }
                    .loading-spinner-container {
                        margin: 30px 0;
                    }
                    .loading-spinner.enhanced {
                        width: 60px;
                        height: 60px;
                        border: 4px solid transparent;
                        border-top: 4px solid var(--claude-primary);
                        border-right: 4px solid var(--claude-secondary);
                        border-radius: 50%;
                        animation: spin 1s linear infinite;
                        margin: 0 auto;
                    }
                    .loading-text {
                        font-size: 16px;
                        color: var(--text-secondary);
                        margin: 20px 0;
                        font-weight: 500;
                    }
                    .loading-progress {
                        width: 200px;
                        height: 4px;
                        background: rgba(0, 0, 0, 0.1);
                        border-radius: 2px;
                        overflow: hidden;
                        margin: 0 auto;
                    }
                    .progress-bar {
                        height: 100%;
                        background: linear-gradient(135deg, var(--claude-primary), var(--claude-secondary));
                        animation: progress 2s ease-in-out infinite;
                    }
                    @keyframes progress {
                        0% { width: 0%; }
                        50% { width: 70%; }
                        100% { width: 100%; }
                    }
                    @keyframes spin {
                        0% { transform: rotate(0deg); }
                        100% { transform: rotate(360deg); }
                    }
                </style>
            `;
            document.body.appendChild(loadingIndicator);
        }
        loadingIndicator.style.display = 'flex';
    }

    // 隐藏加载指示器
    hideLoadingIndicator() {
        const loadingIndicator = document.getElementById('appLoadingIndicator');
        if (loadingIndicator) {
            loadingIndicator.style.display = 'none';
        }
    }

    // 显示初始化错误
    showInitializationError() {
        const errorHtml = `
            <div style="position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(255, 255, 255, 0.95);
                        display: flex; flex-direction: column; align-items: center; justify-content: center; z-index: 9999;">
                <div style="max-width: 400px; text-align: center; padding: 40px; background: white; border-radius: 8px;
                           box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);">
                    <h2 style="color: #FF3B30; margin-bottom: 20px;">初始化失败</h2>
                    <p style="color: #666; margin-bottom: 30px;">应用程序初始化失败，请检查网络连接或刷新页面重试。</p>
                    <button onclick="window.location.reload()"
                            style="background: #007AFF; color: white; border: none; padding: 12px 24px;
                                   border-radius: 6px; cursor: pointer; font-size: 16px;">
                        刷新页面
                    </button>
                </div>
            </div>
        `;
        document.body.insertAdjacentHTML('beforeend', errorHtml);
    }

    // 显示语言切换中状态
    showLanguageChanging() {
        const languageSelects = document.querySelectorAll('#languageSelect, #loginLanguageSelect');
        languageSelects.forEach(select => {
            if (select) {
                select.disabled = true;
                select.style.opacity = '0.6';
            }
        });
    }

    // 隐藏语言切换状态
    hideLanguageChanging() {
        const languageSelects = document.querySelectorAll('#languageSelect, #loginLanguageSelect');
        languageSelects.forEach(select => {
            if (select) {
                select.disabled = false;
                select.style.opacity = '1';
            }
        });
    }

    // 更新语言选择器
    updateLanguageSelectors() {
        const languageSelect = document.getElementById('languageSelect');
        if (languageSelect) {
            languageSelect.value = this.currentLanguage;
        }
        
        const loginLanguageSelect = document.getElementById('loginLanguageSelect');
        if (loginLanguageSelect) {
            loginLanguageSelect.value = this.currentLanguage;
        }
        
        console.log('Language selectors updated to:', this.currentLanguage);
    }

    // Utility method to generate UUID
    generateUUID() {
        // Generate a random UUID v4
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
            const r = Math.random() * 16 | 0;
            const v = c == 'x' ? r : (r & 0x3 | 0x8);
            return v.toString(16);
        });
    }

    // Initialize Anthropic Base URL with current page URL
    initializeAnthropicBaseUrl() {
        const anthropicBaseUrlInput = document.getElementById('anthropicBaseUrl');
        if (anthropicBaseUrlInput) {
            // Get current page's base URL
            const protocol = window.location.protocol; // http: or https:
            const host = window.location.host; // hostname:port
            const baseUrl = `${protocol}//${host}`;
            
            // Set the value if it's empty or not already set
            if (!anthropicBaseUrlInput.value || anthropicBaseUrlInput.value === 'https://api.anthropic.com') {
                anthropicBaseUrlInput.value = baseUrl;
            }
        }
    }
}

// 等待DOM加载完成后再初始化应用
document.addEventListener('DOMContentLoaded', async function() {
    console.log('DOMContentLoaded event fired');
    
    // 添加一些CSS样式用于最近请求
    const additionalStyles = `
    .recent-request-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 12px 0;
        border-bottom: 1px solid var(--gray-200);
    }

    .recent-request-item:last-child {
        border-bottom: none;
    }

    .request-info {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .request-time {
        font-size: 12px;
        color: var(--text-secondary);
    }

    .request-model {
        font-size: 14px;
        font-weight: 500;
        color: var(--text-primary);
    }

    .request-status {
        display: flex;
        flex-direction: column;
        align-items: flex-end;
        gap: 4px;
    }

    .request-tokens {
        font-size: 12px;
        color: var(--text-secondary);
    }

    .action-btn {
        background: var(--primary-color);
        color: white;
        border: none;
        border-radius: 4px;
        padding: 4px 8px;
        font-size: 12px;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .action-btn:hover {
        background: #0056CC;
    }

    .notification {
        animation: slideIn 0.3s ease;
    }

    @keyframes slideIn {
        from {
            opacity: 0;
            transform: translateY(-20px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }
    `;

    // 添加样式到页面
    const styleSheet = document.createElement('style');
    styleSheet.textContent = additionalStyles + `
        /* Enhanced animations for better UX */
        .animate-shake {
            animation: shake 0.5s ease-in-out;
        }
        
        @keyframes shake {
            0%, 100% { transform: translateX(0); }
            25% { transform: translateX(-5px); }
            75% { transform: translateX(5px); }
        }
        
        .field-valid input, .field-valid select, .field-valid textarea {
            border-color: var(--success-color) !important;
            box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.1) !important;
        }
        
        .field-invalid input, .field-invalid select, .field-invalid textarea {
            border-color: var(--error-color) !important;
            box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1) !important;
        }
        
        .loading-content {
            display: flex;
            align-items: center;
            gap: 8px;
            justify-content: center;
        }
        
        .tab-content {
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
        }
        
        .notification {
            display: flex;
            align-items: center;
            gap: 12px;
            min-width: 300px;
            max-width: 500px;
        }
        
        .notification-content {
            flex: 1;
        }
        
        .notification-message {
            font-size: 14px;
            line-height: 1.4;
        }
    `;
    document.head.appendChild(styleSheet);
    
    const app = new ClaudeProxyApp();
    console.log('ClaudeProxyApp instance created');
    // 将app实例挂载到全局作用域，以便在HTML中使用
    window.app = app;
    console.log('App instance attached to window');
    // 等待应用初始化完成
    await app.init();
    console.log('App initialization completed');
});