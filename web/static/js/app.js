// Основной JavaScript файл для EduBot

// Глобальные переменные
let currentUser = null;
let authToken = null;

// Инициализация приложения
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
    setupEventListeners();
    checkAuthStatus();
    initializeTelegramWebApp();
});

// Инициализация Telegram WebApp
function initializeTelegramWebApp() {
    if (window.Telegram && window.Telegram.WebApp) {
        console.log('Telegram WebApp detected');
        
        // Настраиваем WebApp
        window.Telegram.WebApp.ready();
        window.Telegram.WebApp.expand();
        
        // Настраиваем тему
        document.body.style.backgroundColor = window.Telegram.WebApp.backgroundColor;
        document.body.style.color = window.Telegram.WebApp.textColor;
        
        // Скрываем кнопку закрытия, если мы в Telegram
        const closeButtons = document.querySelectorAll('.telegram-hidden');
        closeButtons.forEach(btn => btn.style.display = 'none');
        
        // Настраиваем главную кнопку
        window.Telegram.WebApp.MainButton.setText('Записаться на пробное занятие');
        window.Telegram.WebApp.MainButton.onClick(function() {
            openTrialModal();
        });
        window.Telegram.WebApp.MainButton.show();
        
        console.log('Telegram WebApp initialized successfully');
    } else {
        console.log('Running outside Telegram WebApp');
    }
}

// Инициализация приложения
function initializeApp() {
    console.log('EduBot initialized');
    
    // Плавная прокрутка для навигации
    setupSmoothScrolling();
    
    // Анимации при скролле
    setupScrollAnimations();
}

// Настройка обработчиков событий
function setupEventListeners() {
    // Форма записи на пробное занятие
    const trialForm = document.getElementById('trialForm');
    if (trialForm) {
        trialForm.addEventListener('submit', handleTrialSubmission);
    }
    
    // Закрытие модального окна по клику вне его
    const modal = document.getElementById('trialModal');
    if (modal) {
        modal.addEventListener('click', function(e) {
            if (e.target === modal) {
                closeTrialModal();
            }
        });
        modal.addEventListener('click', function(e) {
            if (e.target.classList.contains('modal-close')) {
                closeTrialModal();
            }
        });
    }
    
    // Обработка клавиши Escape
    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape') {
            closeTrialModal();
        }
    });
}

// Плавная прокрутка для навигации
function setupSmoothScrolling() {
    const navLinks = document.querySelectorAll('.nav-link[href^="#"]');
    
    navLinks.forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            
            const targetId = this.getAttribute('href').substring(1);
            const targetElement = document.getElementById(targetId);
            
            if (targetElement) {
                const offsetTop = targetElement.offsetTop - 70; // Учитываем высоту навигации
                
                window.scrollTo({
                    top: offsetTop,
                    behavior: 'smooth'
                });
                
                // Закрываем мобильное меню при клике на ссылку
                closeMobileMenu();
            }
        });
    });
}

// Анимации при скролле
function setupScrollAnimations() {
    const observerOptions = {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    };
    
    const observer = new IntersectionObserver(function(entries) {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.style.opacity = '1';
                entry.target.style.transform = 'translateY(0)';
            }
        });
    }, observerOptions);
    
    // Наблюдаем за элементами для анимации
    const animatedElements = document.querySelectorAll('.subject-card, .achievement, .contact-item');
    animatedElements.forEach(el => {
        el.style.opacity = '0';
        el.style.transform = 'translateY(30px)';
        el.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
        observer.observe(el);
    });
}

// Проверка статуса авторизации
function checkAuthStatus() {
    const token = localStorage.getItem('authToken');
    if (token) {
        authToken = token;
        // TODO: Валидировать токен с сервером
        console.log('User is authenticated');
    }
}

// Открытие модального окна записи на пробное занятие
function openTrialModal() {
    const modal = document.getElementById('trialModal');
    if (modal) {
        modal.classList.add('active');
        document.body.style.overflow = 'hidden';
        
        // Фокус на первом поле формы
        const firstInput = modal.querySelector('input, select');
        if (firstInput) {
            setTimeout(() => firstInput.focus(), 100);
        }
    }
}

// Закрытие модального окна
function closeTrialModal() {
    const modal = document.getElementById('trialModal');
    if (modal) {
        modal.classList.remove('active');
        document.body.style.overflow = '';
        
        // Очистка формы
        const form = document.getElementById('trialForm');
        if (form) {
            form.reset();
        }
    }
}

// Обработка отправки формы записи на пробное занятие
async function handleTrialSubmission(e) {
    e.preventDefault();
    
    const form = e.target;
    const formData = new FormData(form);
    const rawData = Object.fromEntries(formData.entries());
    
    // Преобразуем типы данных
    const data = {
        name: rawData.name,
        grade: parseInt(rawData.grade),
        subject: rawData.subject,
        level: parseInt(rawData.level),
        phone: rawData.phone,
        comment: rawData.comment || ''
    };
    
    // Валидация данных
    if (!validateTrialForm(data)) {
        return;
    }
    
    // Показываем индикатор загрузки
    const submitBtn = form.querySelector('button[type="submit"]');
    const originalText = submitBtn.innerHTML;
    submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Отправка...';
    submitBtn.disabled = true;
    
    try {
        const headers = {
            'Content-Type': 'application/json'
        };
        
        if (authToken) {
            headers['Authorization'] = `Bearer ${authToken}`;
        }
        
        const response = await fetch('/api/public/trial-request', {
            method: 'POST',
            headers: headers,
            body: JSON.stringify(data)
        });
        
        if (response.ok) {
            showNotification('Заявка успешно отправлена! Мы свяжемся с вами в ближайшее время.', 'success');
            closeTrialModal();
        } else {
            const errorText = await response.text();
            console.error('Trial request error:', response.status, errorText);
            
            try {
                const error = JSON.parse(errorText);
                showNotification(error.error || 'Произошла ошибка при отправке заявки', 'error');
            } catch (e) {
                showNotification(`Ошибка сервера (${response.status}): ${errorText}`, 'error');
            }
        }
    } catch (error) {
        console.error('Error submitting trial request:', error);
        showNotification('Произошла ошибка при отправке заявки. Попробуйте еще раз.', 'error');
    } finally {
        // Восстанавливаем кнопку
        submitBtn.innerHTML = originalText;
        submitBtn.disabled = false;
    }
}

// Мобильное меню
function toggleMobileMenu() {
    const navMenu = document.getElementById('navMenu');
    if (navMenu) {
        navMenu.classList.toggle('active');
    }
}

function closeMobileMenu() {
    const navMenu = document.getElementById('navMenu');
    if (navMenu) {
        navMenu.classList.remove('active');
    }
}

// Закрытие мобильного меню при клике вне его
document.addEventListener('click', function(e) {
    const navMenu = document.getElementById('navMenu');
    const mobileBtn = document.querySelector('.mobile-menu-btn');
    
    if (navMenu && !navMenu.contains(e.target) && !mobileBtn.contains(e.target)) {
        closeMobileMenu();
    }
});

// Валидация формы записи на пробное занятие
function validateTrialForm(data) {
    const errors = [];
    
    if (!data.name || data.name.trim().length < 2) {
        errors.push('Имя должно содержать минимум 2 символа');
    }
    
    if (!data.grade || data.grade < 10 || data.grade > 11) {
        errors.push('Выберите корректный класс (10 или 11)');
    }
    
    if (!data.subject) {
        errors.push('Выберите предмет');
    }
    
    if (!data.level || data.level < 1 || data.level > 5) {
        errors.push('Выберите корректный уровень подготовки (1-5)');
    }
    
    if (!data.phone || !isValidPhone(data.phone)) {
        errors.push('Введите корректный номер телефона');
    }
    
    if (errors.length > 0) {
        showNotification(errors.join('<br>'), 'error');
        return false;
    }
    
    return true;
}

// Валидация номера телефона
function isValidPhone(phone) {
    const phoneRegex = /^[\+]?[0-9\s\-\(\)]{10,}$/;
    return phoneRegex.test(phone);
}

// Показ уведомлений
function showNotification(message, type = 'info') {
    // Удаляем существующие уведомления
    const existingNotifications = document.querySelectorAll('.notification');
    existingNotifications.forEach(notification => notification.remove());
    
    // Создаем новое уведомление
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.innerHTML = `
        <div class="notification-content">
            <i class="fas ${getNotificationIcon(type)}"></i>
            <span>${message}</span>
            <button class="notification-close" onclick="this.parentElement.parentElement.remove()">
                <i class="fas fa-times"></i>
            </button>
        </div>
    `;
    
    // Добавляем стили для уведомления
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        background: ${getNotificationColor(type)};
        color: white;
        padding: 15px 20px;
        border-radius: 8px;
        box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
        z-index: 3000;
        max-width: 400px;
        animation: slideInRight 0.3s ease;
    `;
    
    // Добавляем в DOM
    document.body.appendChild(notification);
    
    // Автоматическое удаление через 5 секунд
    setTimeout(() => {
        if (notification.parentElement) {
            notification.style.animation = 'slideOutRight 0.3s ease';
            setTimeout(() => notification.remove(), 300);
        }
    }, 5000);
}

// Получение иконки для уведомления
function getNotificationIcon(type) {
    const icons = {
        success: 'fa-check-circle',
        error: 'fa-exclamation-circle',
        warning: 'fa-exclamation-triangle',
        info: 'fa-info-circle'
    };
    return icons[type] || icons.info;
}

// Получение цвета для уведомления
function getNotificationColor(type) {
    const colors = {
        success: '#27AE60',
        error: '#E74C3C',
        warning: '#F39C12',
        info: '#3498DB'
    };
    return colors[type] || colors.info;
}

// Авторизация через Telegram
async function authenticateWithTelegram(telegramData) {
    try {
        const response = await fetch('/api/public/auth/telegram', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(telegramData)
        });
        
        if (response.ok) {
            const result = await response.json();
            authToken = result.token;
            currentUser = result.user;
            
            // Сохраняем токен в localStorage
            localStorage.setItem('authToken', authToken);
            
            // Обновляем интерфейс
            updateUIForUser(result.user);
            
            return result;
        } else {
            const error = await response.json();
            throw new Error(error.error);
        }
    } catch (error) {
        console.error('Authentication error:', error);
        showNotification('Ошибка авторизации: ' + error.message, 'error');
        throw error;
    }
}

// Обновление интерфейса для авторизованного пользователя
function updateUIForUser(user) {
    // Скрываем кнопку записи на пробное занятие для авторизованных пользователей
    const trialButtons = document.querySelectorAll('[onclick="openTrialModal()"]');
    trialButtons.forEach(btn => {
        if (user.role === 'student') {
            btn.textContent = 'Личный кабинет';
            btn.onclick = () => window.location.href = '/student/dashboard';
        } else if (user.role === 'teacher') {
            btn.textContent = 'Панель преподавателя';
            btn.onclick = () => window.location.href = '/teacher/dashboard';
        }
    });
}

// Выход из системы
function logout() {
    authToken = null;
    currentUser = null;
    localStorage.removeItem('authToken');
    
    // Обновляем интерфейс
    location.reload();
}

// Утилиты для работы с API
const api = {
    async request(url, options = {}) {
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
                ...(authToken && { 'Authorization': `Bearer ${authToken}` })
            }
        };
        
        const mergedOptions = { ...defaultOptions, ...options };
        
        const response = await fetch(url, mergedOptions);
        
        if (!response.ok) {
            const error = await response.json().catch(() => ({ error: 'Network error' }));
            throw new Error(error.error || 'Request failed');
        }
        
        return response.json();
    },
    
    get(url) {
        return this.request(url, { method: 'GET' });
    },
    
    post(url, data) {
        return this.request(url, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    },
    
    put(url, data) {
        return this.request(url, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    },
    
    delete(url) {
        return this.request(url, { method: 'DELETE' });
    }
};

// Добавляем CSS для анимаций уведомлений
const style = document.createElement('style');
style.textContent = `
    @keyframes slideInRight {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }
    
    @keyframes slideOutRight {
        from {
            transform: translateX(0);
            opacity: 1;
        }
        to {
            transform: translateX(100%);
            opacity: 0;
        }
    }
    
    .notification-content {
        display: flex;
        align-items: center;
        gap: 10px;
    }
    
    .notification-close {
        background: none;
        border: none;
        color: white;
        cursor: pointer;
        padding: 5px;
        border-radius: 50%;
        transition: background-color 0.2s;
    }
    
    .notification-close:hover {
        background: rgba(255, 255, 255, 0.2);
    }
`;
document.head.appendChild(style);

// Функции для работы с модальным окном учителя
function openTeacherLogin() {
    const modal = document.getElementById('teacherLoginModal');
    if (modal) {
        modal.style.display = 'flex';
        document.body.style.overflow = 'hidden';
    }
}

function closeTeacherLoginModal() {
    const modal = document.getElementById('teacherLoginModal');
    if (modal) {
        modal.style.display = 'none';
        document.body.style.overflow = 'auto';
        
        // Очищаем форму
        const form = document.getElementById('teacherLoginForm');
        if (form) {
            form.reset();
        }
    }
}

// Обработка отправки формы входа учителя
document.addEventListener('DOMContentLoaded', function() {
    const teacherLoginForm = document.getElementById('teacherLoginForm');
    if (teacherLoginForm) {
        teacherLoginForm.addEventListener('submit', handleTeacherLogin);
    }
});

async function handleTeacherLogin(e) {
    e.preventDefault();
    
    const form = e.target;
    const formData = new FormData(form);
    const password = formData.get('password');
    
    // Показываем индикатор загрузки
    const submitBtn = form.querySelector('button[type="submit"]');
    const originalText = submitBtn.innerHTML;
    submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Вход...';
    submitBtn.disabled = true;
    
    try {
        // Простая проверка пароля (в реальном приложении нужна более сложная авторизация)
        if (password === 'teacher2024') {
            showNotification('Успешный вход! Переходим в панель управления...', 'success');
            
            // Закрываем модальное окно
            setTimeout(() => {
                closeTeacherLoginModal();
                // Переходим в панель управления учителя
                window.location.href = '/teacher-dashboard';
            }, 1500);
        } else {
            showNotification('Неверный пароль. Попробуйте еще раз.', 'error');
        }
    } catch (error) {
        console.error('Error during teacher login:', error);
        showNotification('Произошла ошибка при входе. Попробуйте еще раз.', 'error');
    } finally {
        // Восстанавливаем кнопку
        submitBtn.innerHTML = originalText;
        submitBtn.disabled = false;
    }
}

// Мобильное меню
function toggleMobileMenu() {
    const navMenu = document.getElementById('navMenu');
    if (navMenu) {
        navMenu.classList.toggle('active');
    }
}

function closeMobileMenu() {
    const navMenu = document.getElementById('navMenu');
    if (navMenu) {
        navMenu.classList.remove('active');
    }
}

// Закрытие мобильного меню при клике вне его
document.addEventListener('click', function(e) {
    const navMenu = document.getElementById('navMenu');
    const mobileBtn = document.querySelector('.mobile-menu-btn');
    
    if (navMenu && !navMenu.contains(e.target) && !mobileBtn.contains(e.target)) {
        closeMobileMenu();
    }
});
