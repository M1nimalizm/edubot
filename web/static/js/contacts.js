// Функции для контактов
function callPhone(phoneNumber) {
    if (window.Telegram && window.Telegram.WebApp) {
        // В Telegram WebApp
        window.Telegram.WebApp.openTelegramLink(`https://t.me/pugachev_teacher`);
    } else {
        // В обычном браузере
        window.open(`tel:${phoneNumber}`, '_self');
    }
}

function openTelegram(username) {
    if (window.Telegram && window.Telegram.WebApp) {
        // В Telegram WebApp
        window.Telegram.WebApp.openTelegramLink(`https://t.me/${username.replace('@', '')}`);
    } else {
        // В обычном браузере
        window.open(`https://t.me/${username.replace('@', '')}`, '_blank');
    }
}

function sendEmail(email) {
    if (window.Telegram && window.Telegram.WebApp) {
        // В Telegram WebApp
        window.Telegram.WebApp.openTelegramLink(`https://t.me/pugachev_teacher`);
    } else {
        // В обычном браузере
        window.open(`mailto:${email}`, '_self');
    }
}
