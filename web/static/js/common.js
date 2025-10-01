// Унифицированные функции для всех страниц EduBot

// Мобильное меню
function toggleMobileMenu() {
    const mobileMenu = document.getElementById('mobileMenu');
    if (mobileMenu) {
        mobileMenu.classList.toggle('active');
    }
}

// Кнопка "Назад" - унифицированная логика
function goBack() {
    if (window.history.length > 1) {
        window.history.back();
    } else {
        // Определяем базовую страницу в зависимости от текущего URL
        const currentPath = window.location.pathname;
        if (currentPath.includes('/teacher')) {
            window.location.href = '/teacher-dashboard';
        } else if (currentPath.includes('/student')) {
            window.location.href = '/student-dashboard';
        } else {
            window.location.href = '/app';
        }
    }
}

// Закрытие мобильного меню при клике вне его
document.addEventListener('click', function(event) {
    const mobileMenu = document.getElementById('mobileMenu');
    const toggle = document.querySelector('.mobile-menu-toggle');
    
    if (mobileMenu && mobileMenu.classList.contains('active') && 
        !mobileMenu.contains(event.target) && 
        !toggle.contains(event.target)) {
        mobileMenu.classList.remove('active');
    }
});

// Унифицированная функция выхода
function logout() {
    if (confirm('Вы уверены, что хотите выйти?')) {
        localStorage.removeItem('authToken');
        window.location.href = '/app';
    }
}

// Унифицированные функции уведомлений
function showSuccess(message) {
    // Создаем уведомление если нет контейнера
    let container = document.getElementById('successContainer');
    if (!container) {
        container = document.createElement('div');
        container.id = 'successContainer';
        container.style.cssText = 'position: fixed; top: 20px; right: 20px; z-index: 1000;';
        document.body.appendChild(container);
    }
    
    container.innerHTML = `
        <div style="background: #27AE60; color: white; padding: 1rem; border-radius: 6px; margin-bottom: 0.5rem; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
            <i class="fas fa-check-circle"></i> ${message}
        </div>
    `;
    
    setTimeout(() => {
        if (container) container.innerHTML = '';
    }, 5000);
}

function showError(message) {
    // Создаем уведомление если нет контейнера
    let container = document.getElementById('errorContainer');
    if (!container) {
        container = document.createElement('div');
        container.id = 'errorContainer';
        container.style.cssText = 'position: fixed; top: 20px; right: 20px; z-index: 1000;';
        document.body.appendChild(container);
    }
    
    // Заменяем слово "ошибка" на более дружелюбные формулировки
    let friendlyMessage = message
        .replace(/ошибка/gi, 'проблема')
        .replace(/не удалось/gi, 'не получилось')
        .replace(/неизвестная ошибка/gi, 'неизвестная проблема');
    
    container.innerHTML = `
        <div style="background: #E74C3C; color: white; padding: 1rem; border-radius: 6px; margin-bottom: 0.5rem; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
            <i class="fas fa-exclamation-triangle"></i> ${friendlyMessage}
        </div>
    `;
    
    setTimeout(() => {
        if (container) container.innerHTML = '';
    }, 5000);
}

// Унифицированная функция загрузки
function showLoading(containerId, message = 'Загрузка...') {
    const container = document.getElementById(containerId);
    if (container) {
        container.innerHTML = `
            <div style="text-align: center; padding: 2rem; color: #666;">
                <i class="fas fa-spinner fa-spin" style="font-size: 2rem; margin-bottom: 1rem;"></i>
                <p>${message}</p>
            </div>
        `;
    }
}

// Унифицированная функция пустого состояния
function showEmptyState(containerId, icon, title, message) {
    const container = document.getElementById(containerId);
    if (container) {
        container.innerHTML = `
            <div style="text-align: center; padding: 3rem; color: #666;">
                <i class="${icon}" style="font-size: 3rem; margin-bottom: 1rem; color: #BDC3C7;"></i>
                <h3 style="margin-bottom: 0.5rem;">${title}</h3>
                <p>${message}</p>
            </div>
        `;
    }
}
