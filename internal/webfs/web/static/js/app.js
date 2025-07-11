// Claude Code Proxy Backend Management Application
console.log('JavaScript file loaded successfully');

// Test JavaScript execution by making a simple request
fetch('/health').then(response => {
    console.log('JavaScript test request successful:', response.status);
}).catch(error => {
    console.error('JavaScript test request failed:', error);
});

class ClaudeProxyApp {
    constructor() {
        this.currentTab = 'dashboard';
        this.config = {};
        this.stats = {};
        this.currentUser = null;
        this.isAuthenticated = false;
        this.currentLanguage = 'zh-CN';
        this.translations = {};
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
                'network_error': this.currentLanguage === 'zh-CN' ? '网络错误，请重试' : 'Network error, please try again'
            },
            'menu': {
                'dashboard': this.currentLanguage === 'zh-CN' ? '仪表板' : 'Dashboard',
                'requests': this.currentLanguage === 'zh-CN' ? '请求日志' : 'Request Logs',
                'config': this.currentLanguage === 'zh-CN' ? '配置管理' : 'Configuration',
                'users': this.currentLanguage === 'zh-CN' ? '用户管理' : 'User Management',
                'test': this.currentLanguage === 'zh-CN' ? 'API测试' : 'API Testing'
            },
            'common': {
                'loading': this.currentLanguage === 'zh-CN' ? '加载中...' : 'Loading...',
                'error': this.currentLanguage === 'zh-CN' ? '错误' : 'Error',
                'success': this.currentLanguage === 'zh-CN' ? '成功' : 'Success'
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

        loginBtn.disabled = true;
        loginBtn.textContent = this.t('login.logging_in');
        loginError.classList.remove('show');

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
                this.showMainApp();
                this.loadInitialData();
            } else {
                loginError.textContent = data.error || this.t('login.login_error');
                loginError.classList.add('show');
            }
        } catch (error) {
            console.error('Login failed:', error);
            loginError.textContent = this.t('login.network_error');
            loginError.classList.add('show');
        } finally {
            loginBtn.disabled = false;
            loginBtn.textContent = this.t('login.login_button');
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
        // 更新菜单状态
        document.querySelectorAll('.menu-item').forEach(item => {
            item.classList.remove('active');
        });
        document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');

        // 更新内容区域
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.remove('active');
        });
        document.getElementById(tabName).classList.add('active');

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
            const response = await this.apiCall('/admin/config');
            const data = await response.json();
            this.config = data.config;
            
            // 更新配置页面
            if (this.config) {
                this.updateConfigDisplay();
            }
        } catch (error) {
            console.error('加载配置失败:', error);
        }
    }

    updateConfigDisplay() {
        const elements = {
            'openaiApiKey': this.config.openai_api_key,
            'claudeApiKey': this.config.claude_api_key,
            'openaiBaseUrl': this.config.openai_base_url,
            'claudeBaseUrl': this.config.claude_base_url,
            'bigModel': this.config.big_model,
            'smallModel': this.config.small_model,
            'maxTokens': this.config.max_tokens_limit,
            'requestTimeout': this.config.request_timeout,
            'serverHost': this.config.host,
            'serverPort': this.config.port,
            'logLevel': this.config.log_level,
            'jwtSecret': this.config.jwt_secret,
            'dbEncryptKey': this.config.db_encrypt_key,
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

        // 更新仪表板中的当前模型
        const currentModelElement = document.getElementById('currentModel');
        if (currentModelElement) {
            currentModelElement.textContent = `${this.config.big_model} / ${this.config.small_model}`;
        }
    }

    async saveConfig() {
        const configData = {
            openai_api_key: document.getElementById('openaiApiKey').value,
            claude_api_key: document.getElementById('claudeApiKey').value,
            openai_base_url: document.getElementById('openaiBaseUrl').value,
            claude_base_url: document.getElementById('claudeBaseUrl').value,
            big_model: document.getElementById('bigModel').value,
            small_model: document.getElementById('smallModel').value,
            max_tokens_limit: parseInt(document.getElementById('maxTokens').value) || 4096,
            request_timeout: parseInt(document.getElementById('requestTimeout').value) || 90,
            host: document.getElementById('serverHost').value,
            port: parseInt(document.getElementById('serverPort').value) || 8080,
            log_level: document.getElementById('logLevel').value,
            jwt_secret: document.getElementById('jwtSecret').value,
            db_encrypt_key: document.getElementById('dbEncryptKey').value,
            encrypt_algorithm: document.getElementById('encryptAlgo').value,
            stream_enabled: document.getElementById('streamEnabled').checked
        };

        const resultElement = document.getElementById('configResult');
        const saveBtn = document.getElementById('saveConfigBtn');

        saveBtn.disabled = true;
        saveBtn.textContent = this.t('config.saving');

        try {
            const response = await this.apiCall('/admin/config', {
                method: 'PUT',
                body: JSON.stringify(configData)
            });

            const data = await response.json();

            if (response.ok) {
                resultElement.className = 'config-result success';
                resultElement.textContent = this.t('config.config_saved');
                this.config = data.config;
            } else {
                resultElement.className = 'config-result error';
                resultElement.textContent = data.error || this.t('config.save_failed');
            }
        } catch (error) {
            console.error('保存配置失败:', error);
            resultElement.className = 'config-result error';
            resultElement.textContent = this.t('common.network_error');
        } finally {
            saveBtn.disabled = false;
            saveBtn.textContent = this.t('config.save_config');
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
        // 简单的通知实现
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.textContent = message;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 12px 20px;
            border-radius: 8px;
            color: white;
            font-weight: 500;
            z-index: 2000;
            opacity: 0;
            transform: translateY(-20px);
            transition: all 0.3s ease;
        `;

        if (type === 'success') {
            notification.style.backgroundColor = '#34C759';
        } else if (type === 'error') {
            notification.style.backgroundColor = '#FF3B30';
        } else {
            notification.style.backgroundColor = '#007AFF';
        }

        document.body.appendChild(notification);

        // 显示动画
        setTimeout(() => {
            notification.style.opacity = '1';
            notification.style.transform = 'translateY(0)';
        }, 100);

        // 自动消失
        setTimeout(() => {
            notification.style.opacity = '0';
            notification.style.transform = 'translateY(-20px)';
            setTimeout(() => {
                document.body.removeChild(notification);
            }, 300);
        }, 3000);
    }

    async loadDashboardData() {
        // 模拟统计数据（实际项目中应该从后端API获取）
        const mockStats = {
            totalRequests: Math.floor(Math.random() * 10000) + 1000,
            avgResponseTime: (Math.random() * 2000 + 500).toFixed(0) + 'ms',
            successRate: (95 + Math.random() * 4).toFixed(1) + '%',
            tokensUsed: this.formatNumber(Math.floor(Math.random() * 1000000) + 100000)
        };

        // 更新统计卡片
        Object.entries(mockStats).forEach(([key, value]) => {
            const element = document.getElementById(key);
            if (element) {
                element.textContent = value;
            }
        });

        // 更新运行时间
        const uptimeElement = document.getElementById('uptime');
        if (uptimeElement) {
            const uptime = this.calculateUptime();
            uptimeElement.textContent = uptime;
        }

        // 加载最近请求
        this.loadRecentRequests();
    }

    async loadRecentRequests() {
        const recentRequestsElement = document.getElementById('recentRequests');
        if (!recentRequestsElement) return;

        // 模拟最近请求数据
        const mockRequests = [
            { time: '2分钟前', model: 'Claude 3.5 Sonnet', status: 'success', tokens: 1250 },
            { time: '5分钟前', model: 'Claude 3 Haiku', status: 'success', tokens: 890 },
            { time: '8分钟前', model: 'Claude 3.5 Sonnet', status: 'error', tokens: 0 },
            { time: '12分钟前', model: 'Claude 3 Haiku', status: 'success', tokens: 456 }
        ];

        const html = mockRequests.map(req => `
            <div class="recent-request-item">
                <div class="request-info">
                    <span class="request-time">${req.time}</span>
                    <span class="request-model">${req.model}</span>
                </div>
                <div class="request-status">
                    <span class="status-badge ${req.status}">${req.status === 'success' ? this.t('requests.success') : this.t('requests.failed')}</span>
                    <span class="request-tokens">${req.tokens} tokens</span>
                </div>
            </div>
        `).join('');

        recentRequestsElement.innerHTML = html;
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

    // 显示加载指示器
    showLoadingIndicator() {
        // 创建加载指示器如果不存在
        let loadingIndicator = document.getElementById('appLoadingIndicator');
        if (!loadingIndicator) {
            loadingIndicator = document.createElement('div');
            loadingIndicator.id = 'appLoadingIndicator';
            loadingIndicator.innerHTML = `
                <div style="position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(255, 255, 255, 0.9);
                            display: flex; flex-direction: column; align-items: center; justify-content: center; z-index: 9999;">
                    <div style="width: 50px; height: 50px; border: 3px solid #f3f3f3; border-top: 3px solid #007AFF;
                               border-radius: 50%; animation: spin 1s linear infinite;"></div>
                    <p style="margin-top: 20px; font-size: 16px; color: #666;">正在加载...</p>
                </div>
                <style>
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
    styleSheet.textContent = additionalStyles;
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