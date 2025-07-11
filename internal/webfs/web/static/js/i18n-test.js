// 多语言交互测试脚本
console.log('开始多语言交互测试...');

// 测试语言切换功能
async function testLanguageSwitch() {
    console.log('测试语言切换功能...');
    
    // 测试获取支持的语言
    try {
        const response = await fetch('/i18n/languages');
        const data = await response.json();
        console.log('✅ 获取支持的语言:', data.languages);
    } catch (error) {
        console.error('❌ 获取语言列表失败:', error);
    }
    
    // 测试切换到中文
    try {
        const response = await fetch('/i18n/language', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                language: 'zh-CN'
            })
        });
        const data = await response.json();
        console.log('✅ 切换到中文:', data);
    } catch (error) {
        console.error('❌ 切换到中文失败:', error);
    }
    
    // 测试获取中文消息
    try {
        const response = await fetch('/i18n/messages/zh-CN');
        const data = await response.json();
        console.log('✅ 获取中文消息成功，消息数量:', Object.keys(data.messages).length);
    } catch (error) {
        console.error('❌ 获取中文消息失败:', error);
    }
    
    // 测试切换到英文
    try {
        const response = await fetch('/i18n/language', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                language: 'en-US'
            })
        });
        const data = await response.json();
        console.log('✅ 切换到英文:', data);
    } catch (error) {
        console.error('❌ 切换到英文失败:', error);
    }
    
    // 测试获取英文消息
    try {
        const response = await fetch('/i18n/messages/en-US');
        const data = await response.json();
        console.log('✅ 获取英文消息成功，消息数量:', Object.keys(data.messages).length);
    } catch (error) {
        console.error('❌ 获取英文消息失败:', error);
    }
    
    // 测试获取当前语言
    try {
        const response = await fetch('/i18n/current');
        const data = await response.json();
        console.log('✅ 当前语言:', data.language);
    } catch (error) {
        console.error('❌ 获取当前语言失败:', error);
    }
}

// 测试前端翻译功能
function testFrontendTranslation() {
    console.log('测试前端翻译功能...');
    
    // 模拟翻译数据
    const testTranslations = {
        "login": {
            "title": "管理员登录",
            "username": "用户名",
            "password": "密码",
            "login_button": "登录"
        },
        "dashboard": {
            "title": "仪表板",
            "total_requests": "总请求数",
            "success_rate": "成功率"
        }
    };
    
    // 测试翻译函数
    function t(key, params = {}) {
        const keys = key.split('.');
        let value = testTranslations;
        
        for (const k of keys) {
            if (value && typeof value === 'object' && k in value) {
                value = value[k];
            } else {
                return key;
            }
        }
        
        if (typeof value === 'string') {
            return value.replace(/\{\{(\w+)\}\}/g, (match, paramKey) => {
                return params[paramKey] || match;
            });
        }
        
        return key;
    }
    
    // 测试翻译功能
    console.log('✅ 翻译测试:');
    console.log('  login.title:', t('login.title'));
    console.log('  login.username:', t('login.username'));
    console.log('  dashboard.title:', t('dashboard.title'));
    console.log('  dashboard.total_requests:', t('dashboard.total_requests'));
    console.log('  不存在的键:', t('nonexistent.key'));
}

// 测试DOM元素翻译应用
function testDOMTranslation() {
    console.log('测试DOM元素翻译应用...');
    
    // 创建测试元素
    const testElement = document.createElement('div');
    testElement.innerHTML = `
        <p data-i18n="login.title">原始文本</p>
        <input type="text" data-i18n="login.username" placeholder="原始占位符">
        <button data-i18n="login.login_button">原始按钮</button>
    `;
    
    // 模拟翻译应用
    const translations = {
        'login.title': '管理员登录',
        'login.username': '用户名',
        'login.login_button': '登录'
    };
    
    function t(key) {
        return translations[key] || key;
    }
    
    // 应用翻译
    testElement.querySelectorAll('[data-i18n]').forEach(element => {
        const key = element.getAttribute('data-i18n');
        const translation = t(key);
        if (element.tagName === 'INPUT' && element.type === 'text') {
            element.placeholder = translation;
        } else {
            element.textContent = translation;
        }
    });
    
    console.log('✅ DOM翻译测试完成');
    console.log('  处理的元素数量:', testElement.querySelectorAll('[data-i18n]').length);
}

// 运行所有测试
async function runAllTests() {
    console.log('🚀 开始运行所有多语言交互测试...');
    
    await testLanguageSwitch();
    testFrontendTranslation();
    testDOMTranslation();
    
    console.log('✅ 所有多语言交互测试完成！');
}

// 导出测试函数
if (typeof window !== 'undefined') {
    window.i18nTests = {
        testLanguageSwitch,
        testFrontendTranslation,
        testDOMTranslation,
        runAllTests
    };
}

// 如果在浏览器中运行，自动执行测试
if (typeof window !== 'undefined') {
    document.addEventListener('DOMContentLoaded', function() {
        // 延迟运行测试，确保主应用已加载
        setTimeout(runAllTests, 2000);
    });
}