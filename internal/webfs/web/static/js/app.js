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
            success: `<svg viewBox="0 0 24 24" width="14" height="14">
                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" fill="currentColor"/>
            </svg>`,
            error: `<svg viewBox="0 0 24 24" width="14" height="14">
                <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z" fill="currentColor"/>
            </svg>`,
            warning: `<svg viewBox="0 0 24 24" width="14" height="14">
                <path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z" fill="currentColor"/>
            </svg>`,
            info: `<svg viewBox="0 0 24 24" width="14" height="14">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z" fill="currentColor"/>
            </svg>`
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
        // Minimal fallback translations - most translations should come from backend
        const fallbackTranslations = {
            'login': {
                'title': this.currentLanguage === 'zh-CN' ? '管理员登录' : 'Admin Login',
                'login_button': this.currentLanguage === 'zh-CN' ? '登录' : 'Login',
                'network_error': this.currentLanguage === 'zh-CN' ? '网络错误，请重试' : 'Network error, please try again'
            },
            'common': {
                'loading': this.currentLanguage === 'zh-CN' ? '加载中...' : 'Loading...',
                'error': this.currentLanguage === 'zh-CN' ? '错误' : 'Error'
            }
        };
        console.log('Using minimal fallback translations for language:', this.currentLanguage);
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

        // 刷新按钮和搜索功能
        const refreshBtn = document.getElementById('refreshLogs');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => this.loadRequestLogs());
        }
        
        // 搜索功能
        const searchBtn = document.getElementById('searchBtn');
        const searchInput = document.getElementById('searchInput');
        const startTimeInput = document.getElementById('startTime');
        const endTimeInput = document.getElementById('endTime');
        
        if (searchBtn) {
            searchBtn.addEventListener('click', () => this.performSearch());
        }
        
        if (searchInput) {
            searchInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.performSearch();
                }
            });
        }
        
        if (startTimeInput || endTimeInput) {
            [startTimeInput, endTimeInput].forEach(input => {
                if (input) {
                    input.addEventListener('change', () => this.performSearch());
                }
            });
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
                    const originalHtml = generateJwtBtn.innerHTML;
                    generateJwtBtn.innerHTML = `
                        <svg viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg" width="16" height="16">
                            <path d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64z m193.5 301.7l-210.6 292c-12.7 17.7-39 17.7-51.7 0L318.5 484.9c-3.8-5.3 0-12.7 6.5-12.7h46.9c10.2 0 19.9 4.9 25.9 13.3l71.2 98.8 157.2-218c6-8.3 15.6-13.3 25.9-13.3H699c6.5 0 10.3 7.4 6.5 12.7z" fill="currentColor"></path>
                        </svg>
                        已生成
                    `;
                    generateJwtBtn.style.background = 'var(--success-color)';
                    generateJwtBtn.style.color = 'white';
                    
                    setTimeout(() => {
                        generateJwtBtn.innerHTML = originalHtml;
                        generateJwtBtn.style.background = '';
                        generateJwtBtn.style.color = '';
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

        // 模型测试按钮
        document.querySelectorAll('.test-model-icon').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const modelType = btn.dataset.model; // 'big' or 'small'
                this.testModel(modelType);
            });
        });

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
        const requestSearchInput = document.getElementById('searchInput');
        if (requestSearchInput) {
            requestSearchInput.addEventListener('input', (e) => {
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

        // 请求详情模态框控制
        const closeRequestDetailsModal = document.getElementById('closeRequestDetailsModal');
        if (closeRequestDetailsModal) {
            closeRequestDetailsModal.addEventListener('click', () => this.hideRequestDetailsModal());
        }

        // 请求详情模态框点击外部关闭
        const requestDetailsModal = document.getElementById('requestDetailsModal');
        if (requestDetailsModal) {
            requestDetailsModal.addEventListener('click', (e) => {
                if (e.target === requestDetailsModal) {
                    this.hideRequestDetailsModal();
                }
            });
        }

        // OpenAI Base URL 输入框变化监听
        const openaiBaseUrlInput = document.getElementById('openaiBaseUrl');
        if (openaiBaseUrlInput) {
            openaiBaseUrlInput.addEventListener('input', (e) => {
                this.updateFinalEndpointURL(e.target.value);
            });
        }

        // 代理配置事件监听
        const proxyEnabledCheckbox = document.getElementById('proxyEnabled');
        if (proxyEnabledCheckbox) {
            proxyEnabledCheckbox.addEventListener('change', (e) => {
                this.toggleProxyConfig(e.target.checked);
            });
        }

        const proxyTypeSelect = document.getElementById('proxyType');
        if (proxyTypeSelect) {
            proxyTypeSelect.addEventListener('change', (e) => {
                this.switchProxyType(e.target.value);
            });
        }

        // 代理测试按钮
        const testProxyBtn = document.getElementById('testProxyBtn');
        if (testProxyBtn) {
            testProxyBtn.addEventListener('click', () => this.testProxy());
        }

        // 语言选择器将在showMainApp()中设置
    }

    // Request details modal methods
    hideRequestDetailsModal() {
        const modal = document.getElementById('requestDetailsModal');
        if (modal) {
            modal.classList.remove('show');
        }
    }

    renderRequestDetails(requestData) {
        console.log('renderRequestDetails called with data:', requestData);
        const content = document.getElementById('requestDetailsContent');
        if (!content) {
            console.error('requestDetailsContent element not found');
            return;
        }

        const formatJson = (obj) => {
            if (!obj) return 'N/A';
            if (typeof obj === 'string') {
                try {
                    return JSON.stringify(JSON.parse(obj), null, 2);
                } catch (e) {
                    return obj;
                }
            }
            return JSON.stringify(obj, null, 2);
        };

        const formatTimestamp = (timestamp) => {
            if (!timestamp) return 'N/A';
            return new Date(timestamp).toLocaleString();
        };

        // Map backend data structure to display fields
        const displayData = {
            id: requestData.id || 'N/A',
            timestamp: requestData.created_at || requestData.timestamp,
            model: requestData.claude_model || requestData.openai_model || requestData.model || 'N/A',
            status_code: requestData.status_code || requestData.status || 200,
            duration: requestData.duration_ms ? `${requestData.duration_ms}ms` : (requestData.duration || 'N/A'),
            input_tokens: requestData.input_tokens || 0,
            output_tokens: requestData.output_tokens || 0,
            request_body: requestData.request_body || requestData.request,
            response_body: requestData.response_body || requestData.response,
            error_message: requestData.error_message || requestData.error
        };

        console.log('Mapped display data:', displayData);

        const html = `
            <div class="request-details-sections">
                <!-- 基本信息 -->
                <div class="details-section">
                    <h4 class="section-title">${this.t('requests.basic_info') || '基本信息'}</h4>
                    <div class="details-grid">
                        <div class="detail-item">
                            <label>${this.t('requests.request_id') || '请求ID'}:</label>
                            <span class="detail-value">${displayData.id}</span>
                        </div>
                        <div class="detail-item">
                            <label>${this.t('requests.timestamp') || '时间戳'}:</label>
                            <span class="detail-value">${formatTimestamp(displayData.timestamp)}</span>
                        </div>
                        <div class="detail-item">
                            <label>${this.t('requests.model') || '模型'}:</label>
                            <span class="detail-value">${displayData.model}</span>
                        </div>
                        <div class="detail-item">
                            <label>${this.t('requests.status') || '状态'}:</label>
                            <span class="detail-value status-badge ${this.getStatusClass(displayData.status_code)}">
                                ${this.getStatusText(displayData.status_code)}
                            </span>
                        </div>
                        <div class="detail-item">
                            <label>${this.t('requests.duration') || '响应时间'}:</label>
                            <span class="detail-value">${displayData.duration}</span>
                        </div>
                        <div class="detail-item">
                            <label>${this.t('requests.tokens') || 'Token使用'}:</label>
                            <span class="detail-value">
                                ${this.t('requests.input') || '输入'}: ${displayData.input_tokens} / 
                                ${this.t('requests.output') || '输出'}: ${displayData.output_tokens}
                            </span>
                        </div>
                    </div>
                </div>

                <!-- 请求内容 -->
                <div class="details-section">
                    <h4 class="section-title">${this.t('requests.request_content') || '请求内容'}</h4>
                    <div class="code-block">
                        <pre><code class="json">${formatJson(displayData.request_body)}</code></pre>
                    </div>
                </div>

                <!-- 响应内容 -->
                <div class="details-section">
                    <h4 class="section-title">${this.t('requests.response_content') || '响应内容'}</h4>
                    <div class="code-block">
                        <pre><code class="json">${formatJson(displayData.response_body)}</code></pre>
                    </div>
                </div>

                <!-- 错误信息 (如果有) -->
                ${displayData.error_message ? `
                <div class="details-section error-section">
                    <h4 class="section-title">${this.t('requests.error_info') || '错误信息'}</h4>
                    <div class="error-content">
                        <pre><code class="error">${displayData.error_message}</code></pre>
                    </div>
                </div>
                ` : ''}
            </div>
        `;

        console.log('Setting innerHTML with HTML:', html);
        content.innerHTML = html;
        console.log('Content after setting innerHTML:', content.innerHTML);
    }

    generateMockRequestDetails(requestId) {
        const models = ['Claude 3.5 Sonnet', 'Claude 3 Haiku', 'Claude 3 Opus'];
        const isSuccess = Math.random() > 0.3; // 70% success rate
        
        const mockRequest = {
            model: 'claude-3-5-sonnet-20241022',
            max_tokens: 1000,
            messages: [
                {
                    role: 'user',
                    content: 'Hello, can you help me write a Python function to calculate fibonacci numbers?'
                }
            ]
        };

        const mockResponse = isSuccess ? {
            id: requestId,
            type: 'message',
            role: 'assistant',
            content: [
                {
                    type: 'text',
                    text: 'I\'d be happy to help you write a Python function to calculate Fibonacci numbers. Here are a couple of different approaches:\n\n```python\ndef fibonacci_iterative(n):\n    """Calculate the nth Fibonacci number using iteration"""\n    if n <= 1:\n        return n\n    \n    a, b = 0, 1\n    for _ in range(2, n + 1):\n        a, b = b, a + b\n    \n    return b\n```'
                }
            ],
            model: 'claude-3-5-sonnet-20241022',
            usage: {
                input_tokens: 25,
                output_tokens: 150
            }
        } : {
            error: {
                type: 'rate_limit_error',
                message: 'Rate limit exceeded. Please try again later.'
            }
        };

        return {
            id: requestId,
            created_at: new Date().toISOString(),
            claude_model: models[Math.floor(Math.random() * models.length)],
            status_code: isSuccess ? 200 : 429,
            duration_ms: isSuccess ? Math.floor(Math.random() * 2000) + 500 : 0,
            input_tokens: isSuccess ? 25 : 0,
            output_tokens: isSuccess ? 150 : 0,
            request_body: mockRequest,
            response_body: mockResponse,
            error_message: isSuccess ? null : 'Rate limit exceeded. Please try again later.'
        };
    }

    // 代理配置相关方法
    toggleProxyConfig(enabled) {
        const proxyConfigSection = document.getElementById('proxyConfigSection');
        const proxyTestSection = document.getElementById('proxyTestSection');
        
        if (proxyConfigSection) {
            proxyConfigSection.style.display = enabled ? 'block' : 'none';
        }
        
        if (proxyTestSection) {
            proxyTestSection.style.display = enabled ? 'block' : 'none';
        }
    }

    switchProxyType(type) {
        const httpProxyConfig = document.getElementById('httpProxyConfig');
        const socks5ProxyConfig = document.getElementById('socks5ProxyConfig');
        
        if (httpProxyConfig && socks5ProxyConfig) {
            if (type === 'http') {
                httpProxyConfig.style.display = 'block';
                socks5ProxyConfig.style.display = 'none';
            } else if (type === 'socks5') {
                httpProxyConfig.style.display = 'none';
                socks5ProxyConfig.style.display = 'block';
            }
        }
    }

    // 测试代理连接
    async testProxy() {
        const proxyEnabled = document.getElementById('proxyEnabled').checked;
        if (!proxyEnabled) {
            this.showProxyTestResult('请先启用代理', 'error');
            return;
        }

        const proxyType = document.getElementById('proxyType').value;
        const testUrl = document.getElementById('proxyTestUrl').value || 'https://httpbin.org/ip';
        const ignoreSSL = document.getElementById('ignoreSSL').checked;

        let proxyConfig = null;
        if (proxyType === 'http') {
            const httpProxy = document.getElementById('httpProxy').value;
            if (!httpProxy) {
                this.showProxyTestResult('请填写HTTP代理地址', 'error');
                return;
            }
            proxyConfig = {
                type: 'http',
                address: httpProxy
            };
        } else if (proxyType === 'socks5') {
            const socks5Proxy = document.getElementById('socks5Proxy').value;
            if (!socks5Proxy) {
                this.showProxyTestResult('请填写SOCKS5代理地址', 'error');
                return;
            }
            proxyConfig = {
                type: 'socks5',
                address: socks5Proxy,
                username: document.getElementById('socks5Username').value || '',
                password: document.getElementById('socks5Password').value || ''
            };
        }

        const testBtn = document.getElementById('testProxyBtn');
        const originalHtml = testBtn.innerHTML;
        
        // 显示加载状态
        testBtn.innerHTML = `
            <svg viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg" width="20" height="20" class="spinning">
                <path d="M512 1024C229.222 1024 0 794.778 0 512S229.222 0 512 0s512 229.222 512 512-229.222 512-512 512z m259.149-568.883h-290.74a25.293 25.293 0 0 0-25.292 25.293l-0.026 63.206c0 13.952 11.315 25.293 25.267 25.293h290.74c13.952 0 25.293-11.341 25.293-25.293V480.41c0-13.952-11.341-25.293-25.293-25.293z" fill="currentColor"/>
            </svg>
        `;
        testBtn.style.pointerEvents = 'none';

        try {
            const response = await this.apiCall('/admin/test-proxy', {
                method: 'POST',
                body: JSON.stringify({
                    proxy_config: proxyConfig,
                    test_url: testUrl,
                    ignore_ssl: ignoreSSL
                })
            });

            const data = await response.json();
            
            if (response.ok) {
                this.showProxyTestResult(`✅ 代理测试成功！\n响应时间: ${data.duration}ms\n代理IP: ${data.ip || '获取失败'}`, 'success');
            } else {
                this.showProxyTestResult(`❌ 代理测试失败: ${data.error}`, 'error');
            }
        } catch (error) {
            console.error('代理测试失败:', error);
            this.showProxyTestResult(`❌ 代理测试失败: ${error.message}`, 'error');
        } finally {
            // 恢复按钮状态
            testBtn.innerHTML = originalHtml;
            testBtn.style.pointerEvents = '';
        }
    }

    showProxyTestResult(message, type) {
        const resultElement = document.getElementById('proxyTestResult');
        if (resultElement) {
            resultElement.style.display = 'block';
            resultElement.className = `proxy-test-result ${type}`;
            resultElement.textContent = message;
            
            // 5秒后自动隐藏
            setTimeout(() => {
                resultElement.style.display = 'none';
            }, 5000);
        }
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
        const loginBtn = document.querySelector('.login-form .btn-primary');

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
        const userRoleElement = document.getElementById('userRole');
        if (userNameElement && this.currentUser) {
            userNameElement.textContent = this.currentUser.username;
            // Generate and set user avatar
            this.generateUserAvatar(this.currentUser.username);
        }
        
        if (userRoleElement && this.currentUser) {
            const roleText = this.currentUser.role === 'admin' ? 
                (this.t('users.admin') || '管理员') : 
                (this.t('users.user') || '用户');
            userRoleElement.textContent = roleText;
        }

        // Setup role-based UI permissions
        this.setupRoleBasedUI();

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

    // Setup role-based UI permissions
    setupRoleBasedUI() {
        if (!this.currentUser) return;

        const isAdmin = this.currentUser.role === 'admin';
        console.log('Setting up role-based UI for user:', this.currentUser.username, 'role:', this.currentUser.role, 'isAdmin:', isAdmin);

        // Admin-only menu items
        const adminOnlyTabs = ['config', 'users'];
        
        adminOnlyTabs.forEach(tabName => {
            const menuItem = document.querySelector(`[data-tab="${tabName}"]`);
            if (menuItem) {
                if (isAdmin) {
                    menuItem.style.display = '';
                    menuItem.classList.remove('hidden');
                } else {
                    menuItem.style.display = 'none';
                    menuItem.classList.add('hidden');
                }
            }
        });

        // If current user is not admin and currently on an admin tab, redirect to dashboard
        if (!isAdmin && adminOnlyTabs.includes(this.currentTab)) {
            this.switchTab('dashboard');
        }

        // Store user permission level for later checks
        this.userPermissions = {
            isAdmin: isAdmin,
            canAccessConfig: isAdmin,
            canAccessUsers: isAdmin,
            canAccessDashboard: true,
            canAccessRequests: true,
            canAccessTest: true,
            canAccessHelp: true
        };
    }

    // Check if user has permission to access a specific tab
    hasPermission(tabName) {
        if (!this.userPermissions) return false;
        
        switch (tabName) {
            case 'config':
                return this.userPermissions.canAccessConfig;
            case 'users':
                return this.userPermissions.canAccessUsers;
            case 'dashboard':
                return this.userPermissions.canAccessDashboard;
            case 'requests':
                return this.userPermissions.canAccessRequests;
            case 'test':
                return this.userPermissions.canAccessTest;
            case 'help':
                return this.userPermissions.canAccessHelp;
            default:
                return false;
        }
    }

    switchTab(tabName) {
        // Check permissions before switching tabs
        if (!this.hasPermission(tabName)) {
            console.warn(`Access denied to tab: ${tabName}`);
            this.showNotification(this.t('errors.access_denied_page'), 'error');
            return;
        }

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
                    if (this.hasPermission('config')) {
                        this.loadConfig();
                    }
                    break;
                case 'users':
                    if (this.hasPermission('users')) {
                        this.loadUsers();
                    }
                    break;
                case 'help':
                    this.loadHelpPage();
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
            'encryptAlgo': this.config.encrypt_algorithm,
            // 代理配置
            'proxyEnabled': this.config.proxy_enabled,
            'proxyType': this.config.proxy_type,
            'httpProxy': this.config.http_proxy,
            'socks5Proxy': this.config.socks5_proxy,
            'socks5Username': this.config.socks5_proxy_user,
            'socks5Password': this.config.socks5_proxy_password,
            'ignoreSSL': this.config.ignore_ssl_verification
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
        
        // 初始化代理配置显示状态
        this.toggleProxyConfig(this.config.proxy_enabled === 'true' || this.config.proxy_enabled === true);
        this.switchProxyType(this.config.proxy_type || 'http');
        
        // 初始化Anthropic Base URL为当前页面URL
        this.initializeAnthropicBaseUrl();

        // 更新最终端点URL显示
        this.updateFinalEndpointURL(this.config.openai_base_url || '');

        // 更新仪表板中的当前模型
        const currentModelElement = document.getElementById('currentModel');
        if (currentModelElement) {
            currentModelElement.textContent = `${this.config.big_model} / ${this.config.small_model}`;
        }
    }

    async saveConfig() {
        // Check admin permission
        if (!this.userPermissions?.isAdmin) {
            this.showNotification(this.t('errors.access_denied_config'), 'error');
            return;
        }

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
            encrypt_algorithm: document.getElementById('encryptAlgo').value,
            // 代理配置
            proxy_enabled: document.getElementById('proxyEnabled').checked,
            proxy_type: document.getElementById('proxyType').value,
            http_proxy: document.getElementById('httpProxy').value,
            socks5_proxy: document.getElementById('socks5Proxy').value,
            socks5_proxy_user: document.getElementById('socks5Username').value,
            socks5_proxy_password: document.getElementById('socks5Password').value,
            ignore_ssl_verification: document.getElementById('ignoreSSL').checked
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
                
                // Reload the configuration to show updated values
                setTimeout(() => {
                    this.loadConfig();
                }, 1000);
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
        // Check admin permission
        if (!this.userPermissions?.isAdmin) {
            this.showNotification(this.t('errors.access_denied_config_test'), 'error');
            return;
        }

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
        // Check admin permission
        if (!this.userPermissions?.isAdmin) {
            this.showNotification(this.t('errors.access_denied_users'), 'error');
            return;
        }

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
        // Check admin permission
        if (!this.userPermissions?.isAdmin) {
            this.showNotification(this.t('errors.access_denied_user_save'), 'error');
            return;
        }

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
        // Check admin permission
        if (!this.userPermissions?.isAdmin) {
            this.showNotification(this.t('errors.access_denied_user_edit'), 'error');
            return;
        }

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
        // Check admin permission
        if (!this.userPermissions?.isAdmin) {
            this.showNotification(this.t('errors.access_denied_user_delete'), 'error');
            return;
        }

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

    performSearch() {
        const searchInput = document.getElementById('searchInput');
        const startTimeInput = document.getElementById('startTime');
        const endTimeInput = document.getElementById('endTime');
        
        const filters = {
            keyword: searchInput ? searchInput.value.trim() : '',
            startTime: startTimeInput ? startTimeInput.value : '',
            endTime: endTimeInput ? endTimeInput.value : ''
        };
        
        this.loadRequestLogs(filters);
    }

    async loadRequestLogs(filters = {}) {
        const tableBody = document.getElementById('requestsTableBody');
        if (!tableBody) return;

        tableBody.innerHTML = `<tr><td colspan="6" class="loading">${this.t('requests.loading')}</td></tr>`;

        try {
            // 构建查询参数
            const params = new URLSearchParams();
            if (filters.keyword) {
                params.append('search', filters.keyword);
            }
            if (filters.startTime) {
                params.append('start_time', new Date(filters.startTime).toISOString());
            }
            if (filters.endTime) {
                params.append('end_time', new Date(filters.endTime).toISOString());
            }
            
            const url = `/admin/logs${params.toString() ? '?' + params.toString() : ''}`;
            const response = await this.apiCall(url);
            
            if (response.ok) {
                const data = await response.json();
                this.renderRequestLogs(data.logs || []);
            } else {
                throw new Error('Failed to load request logs');
            }
        } catch (error) {
            console.error('Error loading request logs:', error);
            // 如果API失败，显示模拟数据
            setTimeout(() => {
                const mockLogs = this.generateMockLogs(20, filters);
                this.renderRequestLogs(mockLogs);
            }, 500);
        }
    }

    renderRequestLogs(logs) {
        const tableBody = document.getElementById('requestsTableBody');
        if (!tableBody) return;

        if (logs.length === 0) {
            tableBody.innerHTML = `<tr><td colspan="6" class="no-data">${this.t('requests.no_data')}</td></tr>`;
            return;
        }

        const html = logs.map(log => `
            <tr>
                <td>${this.formatTime(log.created_at || log.time)}</td>
                <td>${log.claude_model || log.model}</td>
                <td><span class="status-badge ${this.getStatusClass(log.status_code || log.status)}">${this.getStatusText(log.status_code || log.status)}</span></td>
                <td>${log.duration_ms ? log.duration_ms + 'ms' : (log.responseTime || '-')}</td>
                <td>${log.input_tokens || log.tokens || 0}</td>
                <td>
                    <button class="action-btn" onclick="app.viewRequestDetails('${log.id}')">${this.t('requests.details')}</button>
                </td>
            </tr>
        `).join('');
        
        tableBody.innerHTML = html;
    }

    getStatusClass(status) {
        if (typeof status === 'number') {
            return status >= 200 && status < 300 ? 'success' : 'error';
        }
        return status || 'error';
    }

    getStatusText(status) {
        if (typeof status === 'number') {
            return status >= 200 && status < 300 ? this.t('requests.success') : this.t('requests.failed');
        }
        if (status === 'success') return this.t('requests.success');
        if (status === 'error') return this.t('requests.failed');
        if (status === 'warning') return this.t('requests.warning');
        return this.t('requests.failed');
    }

    formatTime(timeStr) {
        if (!timeStr) return '-';
        const date = new Date(timeStr);
        return date.toLocaleString();
    }

    generateMockLogs(count, filters = {}) {
        const models = ['Claude 3.5 Sonnet', 'Claude 3 Haiku', 'Claude 3 Opus'];
        const statuses = [
            { status: 'success', text: this.t('requests.success') },
            { status: 'error', text: this.t('requests.failed') },
            { status: 'warning', text: this.t('requests.warning') }
        ];

        let mockLogs = Array.from({ length: count }, (_, i) => {
            const status = statuses[Math.floor(Math.random() * statuses.length)];
            const isSuccess = status.status === 'success';
            const now = new Date();
            const time = new Date(now.getTime() - Math.random() * 7 * 24 * 60 * 60 * 1000); // 随机过去7天内的时间
            
            return {
                id: `req_${Date.now()}_${i}`,
                time: time.toISOString(),
                created_at: time.toISOString(),
                model: models[Math.floor(Math.random() * models.length)],
                claude_model: models[Math.floor(Math.random() * models.length)],
                status: status.status,
                status_code: isSuccess ? 200 : 500,
                statusText: status.text,
                responseTime: isSuccess ? `${Math.floor(Math.random() * 3000) + 200}ms` : '-',
                duration_ms: isSuccess ? Math.floor(Math.random() * 3000) + 200 : 0,
                tokens: isSuccess ? Math.floor(Math.random() * 2000) + 100 : 0,
                input_tokens: isSuccess ? Math.floor(Math.random() * 2000) + 100 : 0
            };
        });

        // 应用过滤器
        if (filters.keyword) {
            const keyword = filters.keyword.toLowerCase();
            mockLogs = mockLogs.filter(log => 
                log.model.toLowerCase().includes(keyword) ||
                log.statusText.toLowerCase().includes(keyword)
            );
        }

        if (filters.startTime) {
            const startTime = new Date(filters.startTime);
            mockLogs = mockLogs.filter(log => new Date(log.time) >= startTime);
        }

        if (filters.endTime) {
            const endTime = new Date(filters.endTime);
            mockLogs = mockLogs.filter(log => new Date(log.time) <= endTime);
        }

        return mockLogs.sort((a, b) => new Date(b.time) - new Date(a.time));
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
                    resultElement.innerHTML = `
                        <strong>${this.t('test.stream_test_success')}</strong><br>
                        ${this.t('test.status_code')}: ${response.status}<br>
                        ${this.t('test.stream_response_initiated')}
                    `;
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

    async testModel(modelType) {
        const modelInputId = modelType === 'big' ? 'bigModel' : 'smallModel';
        const modelInput = document.getElementById(modelInputId);
        const testIcon = document.querySelector(`.test-model-icon[data-model="${modelType}"]`);
        
        if (!modelInput || !modelInput.value.trim()) {
            this.uiController.showEnhancedNotification(this.t('config.please_enter_model') || '请输入模型名称', 'error');
            return;
        }

        const model = modelInput.value.trim();
        const originalHtml = testIcon.innerHTML;
        
        // Show loading state - change icon to spinner
        testIcon.style.pointerEvents = 'none';
        testIcon.innerHTML = `
            <svg viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg" width="24" height="24" style="animation: spin 1s linear infinite;">
                <path d="M512 1024c-69.1 0-136.2-13.5-199.3-40.2C251.7 958 194 911 151.7 849.3S80 704.4 80 640c0-29.9 13.7-56.9 37.2-75.5 23.5-18.6 54.4-24.7 83.1-16.4 28.7 8.3 52.4 28.5 63.7 54.4 11.3 25.9 9.3 55.4-5.4 79.2-5.2 8.4-12.2 15.5-20.4 20.8-16.4 10.6-35.6 14.6-54.4 11.1l13.2-44.7c5.2-1.1 10.3-2.9 14.9-5.4 4.6-2.5 8.6-5.7 11.9-9.5 8.6-9.8 12.2-23.1 9.6-35.8s-11.1-23.4-22.1-29.4c-11.1-6-24.3-6.9-35.6-2.4-11.3 4.5-20.3 13.1-24.4 23.6-4.1 10.5-3.5 22.1 1.6 31.8 5.1 9.7 13.4 17.7 23.1 22.2 9.7 4.5 20.6 5.4 30.8 2.6 10.2-2.8 19.1-8.9 25.1-17.1 6-8.2 9.1-18.1 8.7-28.1-0.4-10-4.5-19.5-11.4-26.9s-16.2-12.1-26.2-13.1c-10-1-20.3 1.4-28.6 6.8-8.3 5.4-14.1 13.4-16.1 22.4l-48.4-14.3c7.2-31.8 26.4-59.6 53.7-77.7 27.3-18.1 61.2-24.5 94.2-17.8 33 6.7 62.5 25.7 82.1 53.1 19.6 27.4 28.2 61.8 24 95.3-4.2 33.5-20.8 64.1-46.3 85.4-25.5 21.3-57.8 31.8-90.1 29.3-32.3-2.5-62.7-16.3-84.9-38.5-22.2-22.2-36-51.6-38.5-82.1-2.5-30.5 4.8-61.4 20.4-86.4 15.6-25 38.7-44.7 64.6-55.1 25.9-10.4 54.4-11.6 79.7-3.4 25.3 8.2 47.1 24.4 61 45.2 13.9 20.8 19.3 45.4 15.1 68.8-4.2 23.4-16.3 44.9-33.8 60.1-17.5 15.2-39.8 23.4-62.4 22.9-22.6-0.5-44.4-9.7-60.9-25.7-16.5-16-26.8-37.5-28.7-60.1-1.9-22.6 4.1-45.2 16.7-63.2 12.6-18 30.5-31.2 50.1-36.9 19.6-5.7 40.4-3.8 58.2 5.3 17.8 9.1 32.1 24.4 39.9 42.7 7.8 18.3 8.5 38.6 1.9 56.7-6.6 18.1-19.4 33.3-35.8 42.4-16.4 9.1-35.4 10.7-53.1 4.5l18.7-43.1c6.2 2.2 13.1 2.1 19.4-0.3 6.3-2.4 11.4-6.8 14.3-12.4 2.9-5.6 3.4-12.1 1.4-18.2-2-6.1-6.2-11.2-11.8-14.3-5.6-3.1-12.1-4-18.3-2.5-6.2 1.5-11.5 5.2-14.9 10.3-3.4 5.1-4.6 11.3-3.4 17.3 1.2 6 4.5 11.2 9.2 14.6 4.7 3.4 10.4 5 16.1 4.5 5.7-0.5 11-2.9 14.8-6.7 3.8-3.8 6.2-8.7 6.7-14.1 0.5-5.4-0.8-10.8-3.6-15.3-2.8-4.5-7-8-12-9.7-5-1.7-10.4-1.5-15.2 0.6-4.8 2.1-8.7 5.6-10.8 10.1-2.1 4.5-2.2 9.7-0.3 14.3 1.9 4.6 5.3 8.4 9.6 10.6 4.3 2.2 9.2 2.7 13.8 1.4 4.6-1.3 8.6-4.1 11.1-7.9 2.5-3.8 3.4-8.3 2.5-12.7-0.9-4.4-3.4-8.3-7-10.9-3.6-2.6-8.1-3.8-12.6-3.3-4.5 0.5-8.6 2.5-11.5 5.6-2.9 3.1-4.4 7.1-4.2 11.2 0.2 4.1 2 7.9 5 10.6 3 2.7 6.9 4.1 10.9 3.9 4-0.2 7.7-1.9 10.3-4.7 2.6-2.8 4-6.4 3.9-10.1-0.1-3.7-1.6-7.2-4.2-9.8-2.6-2.6-6.1-4.1-9.8-4.2-3.7-0.1-7.3 1.3-10.1 3.9s-4.5 6.3-4.7 10.3c-0.2 4 1.2 7.9 3.9 10.9 2.7 3 6.5 4.8 10.6 5 4.1 0.2 8.1-1.2 11.2-4.2 3.1-3 4.9-7.1 5.1-11.5 0.2-4.4-1.1-8.7-3.8-12.1-2.7-3.4-6.6-5.7-10.9-6.5-4.3-0.8-8.7 0.1-12.4 2.5-3.7 2.4-6.4 6-7.6 10.1-1.2 4.1-0.8 8.5 1.1 12.2 1.9 3.7 5.1 6.5 9 7.9 3.9 1.4 8.1 1.2 11.8-0.6 3.7-1.8 6.6-4.8 8.1-8.4 1.5-3.6 1.4-7.6 0-11.3-1.4-3.7-4.1-6.8-7.6-8.7-3.5-1.9-7.5-2.4-11.3-1.4-3.8 1-7.1 3.4-9.3 6.7-2.2 3.3-3.1 7.3-2.5 11.2 0.6 3.9 2.6 7.4 5.6 9.9 3 2.5 6.8 3.8 10.7 3.6 3.9-0.2 7.5-1.9 10.2-4.8 2.7-2.9 4.2-6.7 4.2-10.7 0-4-1.5-7.8-4.2-10.7-2.7-2.9-6.3-4.6-10.2-4.8-3.9-0.2-7.7 1.1-10.7 3.6-3 2.5-5 6-5.6 9.9-0.6 3.9 0.3 7.9 2.5 11.2 2.2 3.3 5.5 5.7 9.3 6.7 3.8 1 7.8 0.5 11.3-1.4 3.5-1.9 6.2-5 7.6-8.7 1.4-3.7 1.5-7.7 0-11.3-1.5-3.6-4.4-6.6-8.1-8.4-3.7-1.8-7.9-2-11.8-0.6-3.9 1.4-7.1 4.2-9 7.9-1.9 3.7-2.3 8.1-1.1 12.2 1.2 4.1 3.9 7.7 7.6 10.1 3.7 2.4 8.1 3.3 12.4 2.5 4.3-0.8 8.2-3.1 10.9-6.5 2.7-3.4 3.8-7.7 3.1-11.8-0.7-4.1-3.1-7.8-6.8-10.3-3.7-2.5-8.2-3.7-12.7-3.3-4.5 0.4-8.6 2.6-11.5 6.2-2.9 3.6-4.4 8.2-4.2 12.8 0.2 4.6 2.1 9 5.3 12.4 3.2 3.4 7.5 5.5 12.1 5.9 4.6 0.4 9.2-0.8 12.9-3.4 3.7-2.6 6.3-6.4 7.3-10.6 1-4.2 0.3-8.6-1.9-12.4-2.2-3.8-5.8-6.7-10.1-8.1-4.3-1.4-9-1.1-13.1 0.8-4.1 1.9-7.4 5.1-9.2 9.1-1.8 4-2.1 8.5-0.8 12.7 1.3 4.2 4.1 7.8 7.8 10.1 3.7 2.3 8.1 3.2 12.5 2.5 4.4-0.7 8.4-3 11.3-6.4 2.9-3.4 4.5-7.8 4.5-12.3 0-4.5-1.6-8.9-4.5-12.3-2.9-3.4-6.9-5.7-11.3-6.4-4.4-0.7-8.8 0.2-12.5 2.5-3.7 2.3-6.5 5.9-7.8 10.1-1.3 4.2-1 8.7 0.8 12.7 1.8 4 5.1 7.2 9.2 9.1 4.1 1.9 8.8 2.2 13.1 0.8 4.3-1.4 7.9-4.3 10.1-8.1 2.2-3.8 2.9-8.2 1.9-12.4-1-4.2-3.6-8-7.3-10.6-3.7-2.6-8.3-3.8-12.9-3.4-4.6 0.4-8.9 2.5-12.1 5.9-3.2 3.4-5.1 7.8-5.3 12.4-0.2 4.6 1.3 9.2 4.2 12.8 2.9 3.6 7 5.8 11.5 6.2 4.5 0.4 9-0.8 12.7-3.3 3.7-2.5 6.1-6.2 6.8-10.3 0.7-4.1-0.4-8.4-3.1-11.8-2.7-3.4-6.6-5.7-10.9-6.5-4.3-0.8-8.7 0.1-12.4 2.5-3.7 2.4-6.4 6-7.6 10.1-1.2 4.1-0.8 8.5 1.1 12.2 1.9 3.7 5.1 6.5 9 7.9 3.9 1.4 8.1 1.2 11.8-0.6 3.7-1.8 6.6-4.8 8.1-8.4 1.5-3.6 1.4-7.6 0-11.3-1.4-3.7-4.1-6.8-7.6-8.7-3.5-1.9-7.5-2.4-11.3-1.4-3.8 1-7.1 3.4-9.3 6.7-2.2 3.3-3.1 7.3-2.5 11.2 0.6 3.9 2.6 7.4 5.6 9.9 3 2.5 6.8 3.8 10.7 3.6 3.9-0.2 7.5-1.9 10.2-4.8 2.7-2.9 4.2-6.7 4.2-10.7 0-4-1.5-7.8-4.2-10.7-2.7-2.9-6.3-4.6-10.2-4.8-3.9-0.2-7.7 1.1-10.7 3.6-3 2.5-5 6-5.6 9.9-0.6 3.9 0.3 7.9 2.5 11.2 2.2 3.3 5.5 5.7 9.3 6.7 3.8 1 7.8 0.5 11.3-1.4 3.5-1.9 6.2-5 7.6-8.7 1.4-3.7 1.5-7.7 0-11.3-1.5-3.6-4.4-6.6-8.1-8.4-3.7-1.8-7.9-2-11.8-0.6-3.9 1.4-7.1 4.2-9 7.9-1.9 3.7-2.3 8.1-1.1 12.2 1.2 4.1 3.9 7.7 7.6 10.1 3.7 2.4 8.1 3.3 12.4 2.5 4.3-0.8 8.2-3.1 10.9-6.5 2.7-3.4 3.8-7.7 3.1-11.8-0.7-4.1-3.1-7.8-6.8-10.3-3.7-2.5-8.2-3.7-12.7-3.3-4.5 0.4-8.6 2.6-11.5 6.2-2.9 3.6-4.4 8.2-4.2 12.8 0.2 4.6 2.1 9 5.3 12.4 3.2 3.4 7.5 5.5 12.1 5.9 4.6 0.4 9.2-0.8 12.9-3.4 3.7-2.6 6.3-6.4 7.3-10.6 1-4.2 0.3-8.6-1.9-12.4-2.2-3.8-5.8-6.7-10.1-8.1-4.3-1.4-9-1.1-13.1 0.8-4.1 1.9-7.4 5.1-9.2 9.1-1.8 4-2.1 8.5-0.8 12.7 1.3 4.2 4.1 7.8 7.8 10.1 3.7 2.3 8.1 3.2 12.5 2.5 4.4-0.7 8.4-3 11.3-6.4 2.9-3.4 4.5-7.8 4.5-12.3 0-4.5-1.6-8.9-4.5-12.3-2.9-3.4-6.9-5.7-11.3-6.4-4.4-0.7-8.8 0.2-12.5 2.5z" fill="currentColor"></path>
            </svg>
        `;

        try {
            // Call the direct model test API
            const response = await this.apiCall('/v1/test/model', {
                method: 'POST',
                body: JSON.stringify({
                    model: model
                })
            });

            const data = await response.json();

            if (response.ok && data.status === 'success') {
                this.uiController.showEnhancedNotification(
                    `${this.t('config.model_test_success') || '模型测试成功'}: ${model}\n` +
                    `${this.t('test.duration') || '耗时'}: ${data.duration}\n` +
                    `${this.t('test.input_tokens') || '输入Token'}: ${data.input_tokens}\n` +
                    `${this.t('test.output_tokens') || '输出Token'}: ${data.output_tokens}`, 
                    'success'
                );
            } else {
                this.uiController.showEnhancedNotification(
                    `${this.t('config.model_test_failed') || '模型测试失败'}: ${data.message || data.error?.message || 'Unknown error'}`, 
                    'error'
                );
            }
        } catch (error) {
            console.error('Model test failed:', error);
            this.uiController.showEnhancedNotification(
                `${this.t('config.model_test_failed') || '模型测试失败'}: ${error.message}`, 
                'error'
            );
        } finally {
            testIcon.style.pointerEvents = '';
            testIcon.innerHTML = originalHtml;
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

    async viewRequestDetails(requestId) {
        console.log('viewRequestDetails called with requestId:', requestId);
        const modal = document.getElementById('requestDetailsModal');
        const title = document.getElementById('requestDetailsTitle');
        const content = document.getElementById('requestDetailsContent');
        
        console.log('Modal element:', modal);
        console.log('Content element:', content);
        
        if (!modal || !content) {
            console.error('Request details modal elements not found');
            return;
        }

        // Show modal
        modal.classList.add('show');
        console.log('Modal show class added');
        
        // Show loading state
        content.innerHTML = `
            <div class="loading" style="text-align: center; padding: 40px;">
                <div class="loading-spinner-container">
                    <div class="loading-spinner enhanced"></div>
                </div>
                <p>${this.t('requests.loading_details') || '加载详情中...'}</p>
            </div>
        `;
        console.log('Loading state set');

        try {
            // Try to fetch real request details from backend
            console.log('Attempting to fetch from backend...');
            const response = await this.apiCall(`/admin/logs/${requestId}`);
            
            if (response.ok) {
                console.log('Backend response OK');
                const data = await response.json();
                console.log('Backend data received:', data);
                // Handle the correct data structure - backend returns {log: {...}}
                const logData = data.log || data.request || data;
                console.log('Extracted log data:', logData);
                this.renderRequestDetails(logData);
            } else {
                console.log('Backend response not OK, using mock data');
                // If API call fails, show mock data
                const mockData = this.generateMockRequestDetails(requestId);
                console.log('Generated mock data:', mockData);
                this.renderRequestDetails(mockData);
            }
        } catch (error) {
            console.error('Failed to load request details:', error);
            // Show mock data on error
            console.log('Using mock data due to error');
            const mockData = this.generateMockRequestDetails(requestId);
            console.log('Generated mock data:', mockData);
            this.renderRequestDetails(mockData);
        }
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
                            <div class="logo-icon animate-pulse">
                                <svg viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg" width="48" height="48">
                                    <path d="M512 512m-512 0a512 512 0 1 0 1024 0 512 512 0 1 0-1024 0Z" fill="#D97757"></path>
                                    <path d="M278.698667 638.592l151.04-84.736 2.56-7.381333-2.56-4.053334H422.4l-25.258667-1.578666-86.314666-2.304-74.88-3.114667-72.533334-3.882667-18.261333-3.882666L128 505.088l1.749333-11.264 15.36-10.282667 21.973334 1.92 48.64 3.328 72.874666 5.034667 52.906667 3.114667 78.336 8.149333h12.458667l1.706666-5.034667-4.266666-3.114666-3.285334-3.114667-75.434666-51.072L269.354667 388.693333 226.56 357.632l-23.125333-15.744-11.648-14.762667-5.077334-32.256 20.992-23.082666 28.202667 1.92 7.210667 1.962666 28.586666 21.930667 61.013334 47.232 79.744 58.666667 11.648 9.728 4.693333-3.328 0.554667-2.346667-5.248-8.704-43.349334-78.293333L334.506667 240.896l-20.608-33.066667-5.418667-19.797333a95.018667 95.018667 0 0 1-3.328-23.296l23.893333-32.426667 13.226667-4.309333 31.914667 4.266667 13.44 11.648 19.797333 45.269333 32.085333 71.296 49.792 96.981333 14.592 28.757334 7.765334 26.581333 2.901333 8.192h5.077333v-4.693333l4.053334-54.613334 7.594666-66.986666 7.381334-86.272 2.56-24.277334 12.032-29.141333 23.893333-15.744 18.688 8.96 15.36 21.930667-2.133333 14.208-9.130667 59.221333-17.92 92.885333-11.648 62.165334h6.826667l7.765333-7.765334 31.488-41.813333 52.906667-66.005333 23.296-26.24 27.221333-28.928 17.493333-13.824h33.066667l24.32 36.138666-10.88 37.290667-34.048 43.136-28.16 36.522667-40.448 54.4-25.301333 43.52 2.346666 3.498666 6.016-0.554666 91.392-19.456 49.365334-8.96 58.922666-10.069334 26.624 12.416 2.944 12.629334-10.496 25.856-63.018666 15.530666-73.856 14.762667-110.08 26.026667-1.365334 0.981333 1.578667 1.962667 49.578667 4.693333 21.205333 1.109333h51.882667l96.64 7.210667 25.301333 16.725333 15.146667 20.394667-2.56 15.530667-38.826667 19.797333-52.522667-12.416-122.453333-29.141333-42.026667-10.496h-5.845333v3.498666l34.986667 34.176 64.170666 57.898667 80.298667 74.624 4.096 18.474667-10.325333 14.549333-10.88-1.536-70.570667-53.034667-27.221333-23.893333-61.653334-51.882667h-4.053333v5.418667l14.165333 20.778667 75.093334 112.682666 3.84 34.56-5.418667 11.306667-19.456 6.826667-21.376-3.925334-43.946667-61.568-45.312-69.376-36.522666-62.165333-4.48 2.56-21.589334 232.106667-10.112 11.904-23.338666 8.917333-19.456-14.762667-10.282667-23.893333 10.282667-47.232 12.458666-61.568 10.112-48.981333 9.130667-60.8 5.461333-20.181334-0.426666-1.365333-4.437334 0.554667-45.866666 62.976-69.802667 94.208-55.253333 59.050666-13.226667 5.248-22.912-11.818666 2.133333-21.205334 12.8-18.816 76.458667-97.152 46.08-60.245333 29.738667-34.773333-0.213334-5.034667h-1.706666L260.181333 721.92l-36.181333 4.693333-15.530667-14.592 1.92-23.893333 7.381334-7.765333 61.056-41.984-0.170667 0.213333z" fill="#FFFFFF"></path>
                                </svg>
                            </div>
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

    // Construct final endpoint URL using universal logic
    constructFinalURL(baseURL) {
        if (!baseURL || baseURL.trim() === '') {
            return 'https://api.openai.com/v1/chat/completions';
        }

        // Handle trailing slash - always remove it
        if (baseURL.endsWith('/')) {
            baseURL = baseURL.slice(0, -1);
        }

        // Check if URL already contains /v1 - don't add another one
        if (baseURL.includes('/v1')) {
            return baseURL + '/chat/completions';
        }

        // Parse the URL to analyze its structure
        if (this.shouldAppendV1(baseURL)) {
            return baseURL + '/v1/chat/completions';
        }
        
        return baseURL + '/chat/completions';
    }

    // Determine if /v1 should be appended based on URL structure
    shouldAppendV1(baseURL) {
        // For standard OpenAI API format (api.openai.com), we should append /v1
        if (baseURL.toLowerCase().includes('api.openai.com')) {
            return true;
        }
        
        // Extract the path part after the domain
        const parts = baseURL.split('/');
        if (parts.length < 4) { // protocol://domain only, no path
            return true; // Simple domain, likely needs /v1
        }
        
        // Count meaningful path segments (ignore empty strings)
        let pathSegments = 0;
        for (let i = 3; i < parts.length; i++) { // Start from index 3 to skip protocol://domain
            if (parts[i] !== '') {
                pathSegments++;
            }
        }
        
        // If URL has 2+ path segments (like /api/openrouter), it's likely a proxy service
        // that has its own routing and doesn't need /v1
        if (pathSegments >= 2) {
            return false;
        }
        
        // Single path segment or simple domain - likely needs /v1
        return true;
    }

    // Update the final endpoint URL display
    updateFinalEndpointURL(baseURL) {
        const endpointUrlElement = document.getElementById('openaiEndpointUrl');
        if (endpointUrlElement) {
            const finalURL = this.constructFinalURL(baseURL);
            endpointUrlElement.textContent = finalURL;
        }
    }

    // Load help page and setup functionality
    loadHelpPage() {
        // Update current domain in help content
        this.updateHelpPageDomains();
        // Setup copy button functionality
        this.setupCopyButtons();
    }

    // Update domain references in help page
    updateHelpPageDomains() {
        const currentDomain = window.location.origin;
        
        // Update domain display
        const domainElement = document.getElementById('currentDomain');
        if (domainElement) {
            domainElement.textContent = currentDomain;
        }

        // Get real token from config if available
        let realToken = 'sk-...';
        if (this.config && this.config.claude_api_key) {
            realToken = this.config.claude_api_key;
        } else if (this.config && this.config.anthropic_auth_token) {
            realToken = this.config.anthropic_auth_token;
        }

        // Update usage commands with current domain and real token
        const usageCommands = document.getElementById('usageCommands');
        if (usageCommands) {
            usageCommands.textContent = `export ANTHROPIC_AUTH_TOKEN=${realToken} 
export ANTHROPIC_BASE_URL=${currentDomain}
claude`;
        }

        // Update usage commands copy button
        const usageCommandsCopy = document.getElementById('usageCommandsCopy');
        if (usageCommandsCopy) {
            usageCommandsCopy.dataset.copy = `export ANTHROPIC_AUTH_TOKEN=${realToken} 
export ANTHROPIC_BASE_URL=${currentDomain}
claude`;
        }

        // Update environment commands with current domain and real token
        const envCommands = document.getElementById('envCommands');
        if (envCommands) {
            envCommands.textContent = `echo -e '\\n export ANTHROPIC_AUTH_TOKEN=${realToken}' >> ~/.bash_profile
echo -e '\\n export ANTHROPIC_BASE_URL=${currentDomain}' >> ~/.bash_profile
echo -e '\\n export ANTHROPIC_AUTH_TOKEN=${realToken}' >> ~/.bashrc
echo -e '\\n export ANTHROPIC_BASE_URL=${currentDomain}' >> ~/.bashrc
echo -e '\\n export ANTHROPIC_AUTH_TOKEN=${realToken}' >> ~/.zshrc
echo -e '\\n export ANTHROPIC_BASE_URL=${currentDomain}' >> ~/.zshrc`;
        }

        // Update environment commands copy button
        const envCommandsCopy = document.getElementById('envCommandsCopy');
        if (envCommandsCopy) {
            envCommandsCopy.dataset.copy = `echo -e '\\n export ANTHROPIC_AUTH_TOKEN=${realToken}' >> ~/.bash_profile
echo -e '\\n export ANTHROPIC_BASE_URL=${currentDomain}' >> ~/.bash_profile
echo -e '\\n export ANTHROPIC_AUTH_TOKEN=${realToken}' >> ~/.bashrc
echo -e '\\n export ANTHROPIC_BASE_URL=${currentDomain}' >> ~/.bashrc
echo -e '\\n export ANTHROPIC_AUTH_TOKEN=${realToken}' >> ~/.zshrc
echo -e '\\n export ANTHROPIC_BASE_URL=${currentDomain}' >> ~/.zshrc`;
        }
    }

    // Setup copy button functionality
    setupCopyButtons() {
        document.querySelectorAll('.copy-btn').forEach(btn => {
            // Remove existing event listeners
            btn.replaceWith(btn.cloneNode(true));
        });

        // Re-attach event listeners
        document.querySelectorAll('.copy-btn').forEach(btn => {
            btn.addEventListener('click', async () => {
                const textToCopy = btn.dataset.copy;
                if (textToCopy) {
                    try {
                        await navigator.clipboard.writeText(textToCopy);
                        
                        // Show success feedback
                        const originalContent = btn.innerHTML;
                        btn.innerHTML = `
                            <svg viewBox="0 0 24 24" width="16" height="16">
                                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" fill="currentColor"/>
                            </svg>
                        `;
                        btn.classList.add('copied');
                        
                        setTimeout(() => {
                            btn.innerHTML = originalContent;
                            btn.classList.remove('copied');
                        }, 2000);
                        
                        // Show notification
                        this.uiController.showEnhancedNotification(
                            this.t('help.copied_success') || '已复制到剪贴板', 
                            'success'
                        );
                    } catch (err) {
                        console.error('Failed to copy text: ', err);
                        this.uiController.showEnhancedNotification(
                            this.t('help.copy_failed') || '复制失败，请手动选择复制', 
                            'error'
                        );
                    }
                }
            });
        });
    }

    // Generate user avatar based on username
    generateUserAvatar(username) {
        const avatarElement = document.getElementById('userAvatar');
        if (!avatarElement || !username) return;

        // Generate initials from username (take first 2 characters)
        const initials = username.slice(0, 2).toUpperCase();
        
        // Generate a consistent color based on username hash
        const colors = [
            '#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7',
            '#DDA0DD', '#98D8C8', '#F7DC6F', '#BB8FCE', '#85C1E9',
            '#F8C471', '#82E0AA', '#F1948A', '#85C1E9', '#D7DBDD'
        ];
        
        // Simple hash function to get consistent color
        let hash = 0;
        for (let i = 0; i < username.length; i++) {
            hash = username.charCodeAt(i) + ((hash << 5) - hash);
        }
        const colorIndex = Math.abs(hash) % colors.length;
        const backgroundColor = colors[colorIndex];
        
        // Set avatar content and style
        avatarElement.textContent = initials;
        avatarElement.style.backgroundColor = backgroundColor;
        
        // Add a subtle glow effect matching the background color
        avatarElement.style.boxShadow = `0 2px 8px ${backgroundColor}30, 0 0 0 2px ${backgroundColor}20`;
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
        background: var(--claude-primary);
        color: var(--text-inverse);
        border: none;
        border-radius: 4px;
        padding: 4px 8px;
        font-size: 12px;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .action-btn:hover {
        background: var(--claude-primary-hover);
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