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
    initializeCustomSelects();
});

// Инициализация Telegram WebApp
function initializeTelegramWebApp() {
    if (window.Telegram && window.Telegram.WebApp) {
        console.log('Telegram WebApp detected');
        
        // Добавляем класс для Telegram WebApp
        document.body.classList.add('telegram-webapp');
        
        // Настраиваем WebApp
        window.Telegram.WebApp.ready();
        window.Telegram.WebApp.expand();
        
        // Настраиваем тему
        document.body.style.backgroundColor = window.Telegram.WebApp.backgroundColor;
        document.body.style.color = window.Telegram.WebApp.textColor;
        
        // Скрываем кнопку закрытия, если мы в Telegram
        const closeButtons = document.querySelectorAll('.telegram-hidden');
        closeButtons.forEach(btn => btn.style.display = 'none');
        
        console.log('Telegram WebApp initialized successfully');

        // Авто-авторизация через Telegram WebApp
        // Жёсткая авто-авторизация в WebApp: если есть initDataUnsafe.user — авторизуемся без вопросов
        const tgUser = window.Telegram.WebApp.initDataUnsafe && window.Telegram.WebApp.initDataUnsafe.user;
        if (tgUser) {
            const telegramData = {
                id: tgUser.id,
                first_name: tgUser.first_name || '',
                last_name: tgUser.last_name || '',
                username: tgUser.username || '',
                photo_url: tgUser.photo_url || '',
                auth_date: Math.floor(Date.now() / 1000),
                hash: ''
            };
            authenticateWithTelegram(telegramData)
                .then((result) => {
                    // Никаких предложений входа в WebApp
                    // Если есть invite в URL — привязываем и ведём в кабинет ученика
                    const url = new URL(window.location.href);
                    const invite = url.searchParams.get('invite');
                    if (invite) {
                        handleInviteLinkPostAuth();
                        return;
                    }
                    // Иначе заходим сразу в соответствующий кабинет
                    const role = (result && result.user && result.user.role) || 'student';
                    if (role === 'teacher') {
                        window.location.href = '/teacher-dashboard';
                    } else if (role === 'student') {
                        window.location.href = '/student-dashboard';
                    }
                })
                .catch(() => {});
        }

        // В WebApp скрываем любые элементы логина
        try {
            const loginBtn = document.getElementById('studentLoginBtn');
            if (loginBtn) loginBtn.style.display = 'none';
            const studentLoginModal = document.getElementById('studentLoginModal');
            if (studentLoginModal) studentLoginModal.style.display = 'none';
        } catch {}
    } else {
        console.log('Running outside Telegram WebApp');
        // Вне Telegram: показываем кнопку входа (Telegram Login)
        const btn = document.getElementById('studentLoginBtn');
        if (btn) btn.style.display = 'inline-flex';
        prefillInviteFromURL();
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
        document.body.classList.add('modal-open');

        // Ничего не отключаем и не скрываем — по требованию
        
        // Фокус на первом поле формы
        const firstInput = modal.querySelector('input, select');
        if (firstInput) {
            setTimeout(() => firstInput.focus(), 100);
        }
        
        // Скрываем дублирующие кнопки на мобильных
        setTimeout(() => {
            hideDuplicateButtons();
        }, 100);

        // В Telegram WebApp скрываем MainButton, чтобы не было двух кнопок
        if (window.Telegram && window.Telegram.WebApp && window.Telegram.WebApp.MainButton) {
            try {
                window.Telegram.WebApp.MainButton.hide();
            } catch (e) {
                console.warn('Failed to hide Telegram MainButton:', e);
            }
        }
    }
}

// Закрытие модального окна
function closeTrialModal() {
    const modal = document.getElementById('trialModal');
    if (modal) {
        modal.classList.remove('active');
        document.body.style.overflow = '';
        document.body.classList.remove('modal-open');

        // Ничего не меняем — кнопки остаются как были
        
        // Очистка формы
        const form = document.getElementById('trialForm');
        if (form) {
            form.reset();
        }

        // Не возвращаем Telegram MainButton: он отключен в приложении
    }
}

// Функция для отправки формы через мобильную кнопку
function submitTrialForm() {
    const form = document.getElementById('trialForm');
    if (form) {
        // Создаем событие submit и отправляем форму
        const submitEvent = new Event('submit', { bubbles: true, cancelable: true });
        form.dispatchEvent(submitEvent);
    }
}

// Обработка отправки формы записи на пробное занятие
async function handleTrialSubmission(e) {
    e.preventDefault();
    
    const form = e.target;
    const formData = new FormData(form);
    const rawData = Object.fromEntries(formData.entries());
    
    // Преобразуем типы данных
    const contactType = rawData.contact_type || 'phone';
    const contactValue = contactType === 'phone' ? rawData.phone : rawData.telegram;
    
    const data = {
        name: rawData.name,
        grade: parseInt(rawData.grade),
        subject: rawData.subject,
        level: parseInt(rawData.level),
        contact_type: contactType,
        contact_value: contactValue,
        comment: rawData.comment || ''
    };
    
    // Валидация данных
    if (!validateTrialForm(data)) {
        return;
    }
    
    // Показываем индикатор загрузки
    const submitBtn = form.querySelector('.modal-footer button[type="submit"]');
    const mobileBtn = null;
    const mobileBtnText = null;
    
    let originalText, originalMobileText;
    
    if (submitBtn) {
        originalText = submitBtn.innerHTML;
        submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Отправка...';
        submitBtn.disabled = true;
    }
    
    if (mobileBtn && mobileBtnText) {
        originalMobileText = mobileBtnText.textContent;
        mobileBtnText.textContent = 'Отправка...';
        mobileBtn.disabled = true;
    }
    
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
        // Восстанавливаем кнопки
        if (submitBtn && originalText) {
            submitBtn.innerHTML = originalText;
            submitBtn.disabled = false;
        }
        
        if (mobileBtn && mobileBtnText && originalMobileText) {
            mobileBtnText.textContent = originalMobileText;
            mobileBtn.disabled = false;
        }
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

// Кастомные выпадающие списки
function initializeCustomSelects() {
    const customSelects = document.querySelectorAll('.custom-select');
    
    customSelects.forEach(customSelect => {
        const select = customSelect.querySelector('select');
        const selected = customSelect.querySelector('.select-selected');
        const items = customSelect.querySelector('.select-items');
        const options = items.querySelectorAll('div');
        
        // Устанавливаем начальное значение
        const firstOption = select.querySelector('option[value=""]');
        if (firstOption) {
            selected.textContent = firstOption.textContent;
        }
        
        // Клик по выбранному элементу
        selected.addEventListener('click', function(e) {
            e.stopPropagation();
            closeAllSelects();
            customSelect.classList.add('select-arrow-active');
            items.classList.remove('select-hide');
            
            // На мобильных устройствах добавляем класс для стилизации
            if (window.innerWidth <= 768) {
                items.classList.add('mobile-select');
            }
        });
        
        // Клик по опции
        options.forEach(option => {
            option.addEventListener('click', function() {
                const value = this.getAttribute('data-value');
                const text = this.textContent;
                
                // Обновляем select
                select.value = value;
                
                // Обновляем отображаемый текст
                selected.textContent = text;
                
                // Обновляем стили
                options.forEach(opt => opt.classList.remove('same-as-selected'));
                this.classList.add('same-as-selected');
                
                // Закрываем список
                customSelect.classList.remove('select-arrow-active');
                items.classList.add('select-hide');
                items.classList.remove('mobile-select');
                
                // Триггерим событие change
                select.dispatchEvent(new Event('change'));
            });
        });
    });
    
    // Закрытие при клике вне
    document.addEventListener('click', function(e) {
        if (!e.target.closest('.custom-select')) {
            closeAllSelects();
        }
    });
}

function closeAllSelects() {
    const customSelects = document.querySelectorAll('.custom-select');
    customSelects.forEach(customSelect => {
        customSelect.classList.remove('select-arrow-active');
        const items = customSelect.querySelector('.select-items');
        items.classList.add('select-hide');
        items.classList.remove('mobile-select');
    });
}

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
    
    if (!data.contact_value || data.contact_value.trim().length < 3) {
        const fieldName = data.contact_type === 'phone' ? 'номер телефона' : 'Telegram тег';
        errors.push(`Введите корректный ${fieldName}`);
    } else if (data.contact_type === 'phone' && !isValidPhone(data.contact_value)) {
        errors.push('Введите корректный номер телефона');
    } else if (data.contact_type === 'telegram' && !data.contact_value.startsWith('@')) {
        errors.push('Telegram тег должен начинаться с @');
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
            
            // Учитель в whitelist? предложим выбрать роль, иначе — студент по умолчанию
            if (result.allowed_teacher) {
                // Для веб-версии покажем компактный выбор роли один раз
                openRoleSelectModal(result.user);
            }
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
    // Переопределяем кнопку входа: одна кнопка «Войти»
    const trialButtons = document.querySelectorAll('[onclick="openTrialModal()"]');
    trialButtons.forEach(btn => {
        if (user.role === 'teacher') {
            btn.textContent = 'Панель преподавателя';
            btn.onclick = () => window.location.href = '/teacher-dashboard';
        } else {
            btn.textContent = 'Личный кабинет';
            btn.onclick = () => window.location.href = '/student-dashboard';
        }
    });
}

// UI модалка выбора роли (минимум кликов)
function openRoleSelectModal(user) {
    // Если в Telegram WebApp — не спрашиваем, оставляем роль последней, UI минимальный
    if (window.Telegram && window.Telegram.WebApp) return;
    // Создаём простую модалку с двумя кнопками
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
      <div class="modal-content">
        <div class="modal-header">
          <h3>Выбор роли</h3>
          <button class="modal-close" onclick="this.closest('.modal').remove()">&times;</button>
        </div>
        <div class="modal-body" style="display:flex; gap:12px; justify-content:center;">
          <button class="btn btn-primary" id="selectStudent">Ученик</button>
          <button class="btn btn-outline" id="selectTeacher">Учитель</button>
        </div>
      </div>`;
    document.body.appendChild(modal);

    const token = localStorage.getItem('authToken');
    const select = async (role) => {
        try {
            const resp = await fetch('/api/auth/select-role', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json', 'Authorization': 'Bearer ' + token },
                body: JSON.stringify({ role })
            });
            if (resp.ok) {
                const data = await resp.json();
                if (data.token) localStorage.setItem('authToken', data.token);
                modal.remove();
                if (role === 'teacher') {
                    window.location.href = '/teacher-dashboard';
                } else {
                    window.location.href = '/student-dashboard';
                }
            } else {
                const err = await resp.json().catch(() => ({}));
                showError(err.error || 'Не удалось выбрать роль');
            }
        } catch (e) {
            showError('Ошибка сети при выборе роли');
        }
    };
    modal.querySelector('#selectStudent').onclick = () => select('student');
    modal.querySelector('#selectTeacher').onclick = () => select('teacher');
}

// После Telegram-авторизации: если есть invite в URL — привязываем ученика и вводим в ЛК
function handleInviteLinkPostAuth() {
    const url = new URL(window.location.href);
    const invite = url.searchParams.get('invite');
    if (invite) {
        api.post('/api/public/register-student', { inviteCode: invite })
            .then((result) => {
                if (result && result.token) {
                    localStorage.setItem('authToken', result.token);
                }
                showSuccess('Добро пожаловать! Аккаунт ученика привязан.');
                setTimeout(() => window.location.href = '/student-dashboard', 1200);
            })
            .catch((e) => {
                console.error('Invite bind failed:', e);
                showError('Не удалось привязать аккаунт по ссылке. Введите код вручную.');
                prefillInviteFromURL();
            });
    }
}

function prefillInviteFromURL() {
    try {
        const url = new URL(window.location.href);
        const invite = url.searchParams.get('invite');
        if (invite) {
            const input = document.getElementById('studentInviteCode');
            if (input) input.value = invite;
        }
    } catch {}
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
    
    // Инициализация выбора контакта
    initializeContactChoice();
});

// Инициализация выбора контакта
function initializeContactChoice() {
    const phoneRadio = document.getElementById('contact_phone');
    const telegramRadio = document.getElementById('contact_telegram');
    const phoneInput = document.getElementById('phone');
    const telegramInput = document.getElementById('telegram');
    
    if (phoneRadio && telegramRadio && phoneInput && telegramInput) {
        // Обработка переключения между телефоном и Telegram
        phoneRadio.addEventListener('change', function() {
            if (this.checked) {
                phoneInput.style.display = 'block';
                telegramInput.style.display = 'none';
                phoneInput.required = true;
                telegramInput.required = false;
            }
        });
        
        telegramRadio.addEventListener('change', function() {
            if (this.checked) {
                phoneInput.style.display = 'none';
                telegramInput.style.display = 'block';
                phoneInput.required = false;
                telegramInput.required = true;
                
                // Автозаполнение Telegram тега из Telegram WebApp
                if (window.Telegram && window.Telegram.WebApp && window.Telegram.WebApp.initDataUnsafe) {
                    const user = window.Telegram.WebApp.initDataUnsafe.user;
                    if (user && user.username) {
                        telegramInput.value = '@' + user.username;
                    }
                }
            }
        });
    }
    
    // Инициализация мобильного UX
    initializeMobileUX();
}

// Инициализация мобильного UX
function initializeMobileUX() {
    // Проверяем, мобильное ли устройство
    const isMobile = window.innerWidth <= 768;
    
    if (isMobile) {
        // Добавляем класс для мобильных устройств
        document.body.classList.add('mobile-device');
        
        // Принудительно скрываем все кнопки кроме мобильной
        hideDuplicateButtons();
        
        // Улучшаем поведение кастомных селектов на мобильных
        const customSelects = document.querySelectorAll('.custom-select');
        customSelects.forEach(select => {
            const selectItems = select.querySelector('.select-items');
            if (selectItems) {
                // Добавляем класс для мобильных селектов
                selectItems.classList.add('mobile-select');
                
                // Обработка клика по элементам селекта
                const options = selectItems.querySelectorAll('div');
                options.forEach(option => {
                    option.addEventListener('click', function() {
                        // Закрываем селект после выбора
                        setTimeout(() => {
                            closeAllSelects();
                        }, 100);
                    });
                });
            }
        });
        
        // Убираем предотвращение скролла - позволяем нормальное поведение
    }
}

// Функция для скрытия дублирующих кнопок
function hideDuplicateButtons() {
    // Не скрываем никаких кнопок
    
    // Обработчик изменения размера окна
    window.addEventListener('resize', function() {
        const isMobile = window.innerWidth <= 768;
        if (isMobile) {
            document.body.classList.add('mobile-device');
        } else {
            document.body.classList.remove('mobile-device');
        }
    });
}

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
        const token = localStorage.getItem('authToken');
        if (!token) {
            showNotification('Сначала авторизуйтесь через Telegram', 'error');
            return;
        }

        const response = await fetch('/api/teacher/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json', 'Authorization': 'Bearer ' + token },
            body: JSON.stringify({ password })
        });

        if (response.ok) {
            const result = await response.json();
            if (result.token) {
                localStorage.setItem('authToken', result.token);
            }
            showNotification('Успешный вход! Переходим в панель управления...', 'success');
            setTimeout(() => {
                closeTeacherLoginModal();
                window.location.href = '/teacher-dashboard';
            }, 1200);
        } else {
            const err = await response.json().catch(() => ({}));
            showNotification(err.error || 'Неверный пароль или доступ запрещён', 'error');
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

// Кастомные выпадающие списки
function initializeCustomSelects() {
    const customSelects = document.querySelectorAll('.custom-select');
    
    customSelects.forEach(customSelect => {
        const select = customSelect.querySelector('select');
        const selected = customSelect.querySelector('.select-selected');
        const items = customSelect.querySelector('.select-items');
        const options = items.querySelectorAll('div');
        
        // Устанавливаем начальное значение
        const firstOption = select.querySelector('option[value=""]');
        if (firstOption) {
            selected.textContent = firstOption.textContent;
        }
        
        // Клик по выбранному элементу
        selected.addEventListener('click', function(e) {
            e.stopPropagation();
            closeAllSelects();
            customSelect.classList.add('select-arrow-active');
            items.classList.remove('select-hide');
            
            // На мобильных устройствах добавляем класс для стилизации
            if (window.innerWidth <= 768) {
                items.classList.add('mobile-select');
            }
        });
        
        // Клик по опции
        options.forEach(option => {
            option.addEventListener('click', function() {
                const value = this.getAttribute('data-value');
                const text = this.textContent;
                
                // Обновляем select
                select.value = value;
                
                // Обновляем отображаемый текст
                selected.textContent = text;
                
                // Обновляем стили
                options.forEach(opt => opt.classList.remove('same-as-selected'));
                this.classList.add('same-as-selected');
                
                // Закрываем список
                customSelect.classList.remove('select-arrow-active');
                items.classList.add('select-hide');
                items.classList.remove('mobile-select');
                
                // Триггерим событие change
                select.dispatchEvent(new Event('change'));
            });
        });
    });
    
    // Закрытие при клике вне
    document.addEventListener('click', function(e) {
        if (!e.target.closest('.custom-select')) {
            closeAllSelects();
        }
    });
}

function closeAllSelects() {
    const customSelects = document.querySelectorAll('.custom-select');
    customSelects.forEach(customSelect => {
        customSelect.classList.remove('select-arrow-active');
        const items = customSelect.querySelector('.select-items');
        items.classList.add('select-hide');
        items.classList.remove('mobile-select');
    });
}


// Функции для входа ученика
function openStudentLoginModal() {
    // В WebApp ничего не показываем — авторизация автоматическая
    if (window.Telegram && window.Telegram.WebApp) return;
    document.getElementById('studentLoginModal').style.display = 'block';
}

function closeStudentLoginModal() {
    document.getElementById('studentLoginModal').style.display = 'none';
}

// Telegram Login Widget для десктопа
function openStudentTelegramLogin() {
    // В WebApp не открываем Login Widget
    if (window.Telegram && window.Telegram.WebApp) return;
    openStudentLoginModal();
    const container = document.getElementById('telegramLoginContainer');
    if (!container) return;
    container.innerHTML = '';
    // Встраиваем Telegram Login Widget
    const script = document.createElement('script');
    // Замените на реальный bot_username
    script.src = 'https://telegram.org/js/telegram-widget.js?22';
    script.setAttribute('data-telegram-login', 'EduBot_by_Pugachev_bot');
    script.setAttribute('data-size', 'large');
    script.setAttribute('data-request-access', 'write');
    script.setAttribute('data-userpic', 'false');
    script.setAttribute('data-radius', '8');
    script.setAttribute('data-onauth', 'onTelegramAuth(user)');
    container.appendChild(script);
}

window.onTelegramAuth = async function(user) {
    // Конвертируем в формат нашего API
    const telegramData = {
        id: user.id,
        first_name: user.first_name || '',
        last_name: user.last_name || '',
        username: user.username || '',
        photo_url: user.photo_url || '',
        auth_date: Math.floor(Date.now() / 1000),
        hash: user.hash || ''
    };

    try {
        await authenticateWithTelegram(telegramData);
        handleInviteLinkPostAuth();
        closeStudentLoginModal();
        showSuccess('Вход выполнен');
        setTimeout(() => window.location.href = '/student-dashboard', 1000);
    } catch (e) {
        showError('Ошибка авторизации через Telegram');
    }
}


// Централизованная анимация успеха
function showSuccess(message) {
    // Удаляем предыдущие уведомления
    const existingToasts = document.querySelectorAll('.success-toast');
    existingToasts.forEach(toast => toast.remove());
    
    // Создаем новое уведомление
    const toast = document.createElement('div');
    toast.className = 'success-toast';
    toast.innerHTML = `<i class="fas fa-check-circle"></i> ${message}`;
    
    document.body.appendChild(toast);
    
    // Автоматически скрываем через 3 секунды
    setTimeout(() => {
        toast.style.animation = 'successSlideOut 0.3s ease-in';
        setTimeout(() => {
            if (toast.parentNode) {
                toast.parentNode.removeChild(toast);
            }
        }, 300);
    }, 3000);
}

function showError(message) {
    // Удаляем предыдущие уведомления
    const existingToasts = document.querySelectorAll('.success-toast');
    existingToasts.forEach(toast => toast.remove());
    
    // Создаем новое уведомление об ошибке
    const toast = document.createElement('div');
    toast.className = 'success-toast';
    toast.style.background = '#f44336';
    toast.innerHTML = `<i class="fas fa-exclamation-circle"></i> ${message}`;
    
    document.body.appendChild(toast);
    
    // Автоматически скрываем через 4 секунды
    setTimeout(() => {
        toast.style.animation = 'successSlideOut 0.3s ease-in';
        setTimeout(() => {
            if (toast.parentNode) {
                toast.parentNode.removeChild(toast);
            }
        }, 300);
    }, 4000);
}
