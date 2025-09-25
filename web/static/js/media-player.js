/**
 * Универсальный MediaPlayer компонент для воспроизведения медиафайлов из Telegram
 */
class MediaPlayer {
    constructor(container, options = {}) {
        this.container = typeof container === 'string' ? document.querySelector(container) : container;
        this.options = {
            autoplay: false,
            controls: true,
            preload: 'metadata',
            poster: null,
            fallbackText: 'Медиафайл недоступен',
            onError: null,
            onLoad: null,
            onPlay: null,
            onPause: null,
            onEnded: null,
            ...options
        };
        
        this.mediaElement = null;
        this.isLoading = false;
        this.hasError = false;
        
        this.init();
    }

    init() {
        if (!this.container) {
            console.error('MediaPlayer: Container not found');
            return;
        }

        this.container.innerHTML = '';
        this.container.classList.add('media-player');
    }

    /**
     * Загружает и отображает медиафайл
     * @param {string} mediaId - ID медиафайла
     * @param {Object} mediaInfo - Информация о медиафайле
     */
    async loadMedia(mediaId, mediaInfo = {}) {
        if (!mediaId) {
            this.showError('ID медиафайла не указан');
            return;
        }

        this.isLoading = true;
        this.hasError = false;
        this.showLoading();

        try {
            // Определяем тип медиафайла
            const mediaType = mediaInfo.type || await this.detectMediaType(mediaId);
            
            // Создаем соответствующий элемент
            this.mediaElement = this.createMediaElement(mediaType, mediaId, mediaInfo);
            
            // Добавляем в контейнер
            this.container.appendChild(this.mediaElement);
            
            // Настраиваем обработчики событий
            this.setupEventListeners();
            
            // Загружаем медиафайл
            await this.loadMediaSource(mediaId);
            
            this.isLoading = false;
            this.hideLoading();
            
            if (this.options.onLoad) {
                this.options.onLoad(this.mediaElement);
            }
            
        } catch (error) {
            console.error('MediaPlayer: Error loading media:', error);
            this.showError(error.message);
            this.isLoading = false;
        }
    }

    /**
     * Определяет тип медиафайла
     */
    async detectMediaType(mediaId) {
        try {
            const response = await fetch(`/api/media/${mediaId}`);
            if (!response.ok) {
                throw new Error('Медиафайл не найден');
            }
            
            const data = await response.json();
            return data.media.type;
        } catch (error) {
            console.error('MediaPlayer: Error detecting media type:', error);
            return 'video'; // По умолчанию видео
        }
    }

    /**
     * Создает элемент медиа в зависимости от типа
     */
    createMediaElement(mediaType, mediaId, mediaInfo) {
        let element;
        
        switch (mediaType) {
            case 'video':
                element = document.createElement('video');
                element.controls = this.options.controls;
                element.autoplay = this.options.autoplay;
                element.preload = this.options.preload;
                if (this.options.poster) {
                    element.poster = this.options.poster;
                }
                break;
                
            case 'audio':
                element = document.createElement('audio');
                element.controls = this.options.controls;
                element.autoplay = this.options.autoplay;
                element.preload = this.options.preload;
                break;
                
            case 'image':
                element = document.createElement('img');
                element.style.maxWidth = '100%';
                element.style.height = 'auto';
                break;
                
            case 'document':
                element = document.createElement('div');
                element.className = 'document-viewer';
                element.innerHTML = `
                    <div class="document-icon">
                        <i class="fas fa-file-pdf"></i>
                    </div>
                    <div class="document-info">
                        <h4>${mediaInfo.caption || 'Документ'}</h4>
                        <p>Размер: ${this.formatFileSize(mediaInfo.size || 0)}</p>
                        <button class="btn btn-primary" onclick="window.open('/api/media/${mediaId}/stream', '_blank')">
                            <i class="fas fa-download"></i> Скачать
                        </button>
                    </div>
                `;
                return element;
                
            default:
                throw new Error('Неподдерживаемый тип медиафайла');
        }
        
        element.className = `media-element media-${mediaType}`;
        return element;
    }

    /**
     * Загружает источник медиафайла
     */
    async loadMediaSource(mediaId) {
        if (this.mediaElement.tagName === 'IMG') {
            this.mediaElement.src = `/api/media/${mediaId}/stream`;
        } else if (this.mediaElement.tagName === 'VIDEO' || this.mediaElement.tagName === 'AUDIO') {
            this.mediaElement.src = `/api/media/${mediaId}/stream`;
        }
    }

    /**
     * Настраивает обработчики событий
     */
    setupEventListeners() {
        if (!this.mediaElement) return;

        this.mediaElement.addEventListener('loadstart', () => {
            this.container.classList.add('loading');
        });

        this.mediaElement.addEventListener('canplay', () => {
            this.container.classList.remove('loading');
        });

        this.mediaElement.addEventListener('error', (e) => {
            console.error('MediaPlayer: Media error:', e);
            this.showError('Ошибка загрузки медиафайла');
        });

        this.mediaElement.addEventListener('play', () => {
            if (this.options.onPlay) {
                this.options.onPlay(this.mediaElement);
            }
        });

        this.mediaElement.addEventListener('pause', () => {
            if (this.options.onPause) {
                this.options.onPause(this.mediaElement);
            }
        });

        this.mediaElement.addEventListener('ended', () => {
            if (this.options.onEnded) {
                this.options.onEnded(this.mediaElement);
            }
        });
    }

    /**
     * Показывает состояние загрузки
     */
    showLoading() {
        this.container.innerHTML = `
            <div class="media-loading">
                <div class="spinner"></div>
                <p>Загрузка медиафайла...</p>
            </div>
        `;
    }

    /**
     * Скрывает состояние загрузки
     */
    hideLoading() {
        const loadingElement = this.container.querySelector('.media-loading');
        if (loadingElement) {
            loadingElement.remove();
        }
    }

    /**
     * Показывает ошибку
     */
    showError(message) {
        this.hasError = true;
        this.container.innerHTML = `
            <div class="media-error">
                <div class="error-icon">
                    <i class="fas fa-exclamation-triangle"></i>
                </div>
                <p>${message}</p>
                <button class="btn btn-secondary" onclick="this.closest('.media-player').mediaPlayer?.retry()">
                    <i class="fas fa-redo"></i> Попробовать снова
                </button>
            </div>
        `;
        
        if (this.options.onError) {
            this.options.onError(message);
        }
    }

    /**
     * Повторная попытка загрузки
     */
    retry() {
        if (this.lastMediaId) {
            this.loadMedia(this.lastMediaId, this.lastMediaInfo);
        }
    }

    /**
     * Форматирует размер файла
     */
    formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    /**
     * Получает миниатюру медиафайла
     */
    async getThumbnail(mediaId) {
        try {
            const response = await fetch(`/api/media/${mediaId}/thumbnail`);
            if (response.ok) {
                return response.url;
            }
        } catch (error) {
            console.error('MediaPlayer: Error getting thumbnail:', error);
        }
        return null;
    }

    /**
     * Уничтожает плеер
     */
    destroy() {
        if (this.mediaElement) {
            this.mediaElement.remove();
            this.mediaElement = null;
        }
        this.container.innerHTML = '';
    }
}

/**
 * MediaGallery компонент для отображения галереи медиафайлов
 */
class MediaGallery {
    constructor(container, options = {}) {
        this.container = typeof container === 'string' ? document.querySelector(container) : container;
        this.options = {
            columns: 3,
            showThumbnails: true,
            enableLightbox: true,
            onItemClick: null,
            ...options
        };
        
        this.mediaList = [];
        this.currentPlayer = null;
        
        this.init();
    }

    init() {
        if (!this.container) {
            console.error('MediaGallery: Container not found');
            return;
        }

        this.container.classList.add('media-gallery');
        this.container.style.display = 'grid';
        this.container.style.gridTemplateColumns = `repeat(${this.options.columns}, 1fr)`;
        this.container.style.gap = '1rem';
    }

    /**
     * Загружает список медиафайлов
     */
    async loadMedia(mediaList) {
        this.mediaList = mediaList;
        this.container.innerHTML = '';

        for (const media of mediaList) {
            const item = await this.createMediaItem(media);
            this.container.appendChild(item);
        }
    }

    /**
     * Создает элемент медиафайла для галереи
     */
    async createMediaItem(media) {
        const item = document.createElement('div');
        item.className = 'media-gallery-item';
        item.dataset.mediaId = media.id;
        item.dataset.mediaType = media.type;

        if (this.options.showThumbnails) {
            const thumbnail = await this.getThumbnail(media);
            if (thumbnail) {
                item.innerHTML = `
                    <div class="media-thumbnail">
                        <img src="${thumbnail}" alt="${media.caption || 'Медиафайл'}" loading="lazy">
                        <div class="media-overlay">
                            <i class="fas fa-play"></i>
                        </div>
                    </div>
                    <div class="media-info">
                        <h4>${media.caption || 'Без названия'}</h4>
                        <p>${this.formatFileSize(media.size || 0)}</p>
                    </div>
                `;
            } else {
                item.innerHTML = `
                    <div class="media-placeholder">
                        <i class="fas fa-file"></i>
                    </div>
                    <div class="media-info">
                        <h4>${media.caption || 'Без названия'}</h4>
                        <p>${this.formatFileSize(media.size || 0)}</p>
                    </div>
                `;
            }
        } else {
            item.innerHTML = `
                <div class="media-info">
                    <h4>${media.caption || 'Без названия'}</h4>
                    <p>${this.formatFileSize(media.size || 0)}</p>
                </div>
            `;
        }

        item.addEventListener('click', () => {
            if (this.options.onItemClick) {
                this.options.onItemClick(media);
            } else if (this.options.enableLightbox) {
                this.openLightbox(media);
            }
        });

        return item;
    }

    /**
     * Получает миниатюру медиафайла
     */
    async getThumbnail(media) {
        try {
            const response = await fetch(`/api/media/${media.id}/thumbnail`);
            if (response.ok) {
                return response.url;
            }
        } catch (error) {
            console.error('MediaGallery: Error getting thumbnail:', error);
        }
        return null;
    }

    /**
     * Открывает медиафайл в лайтбоксе
     */
    openLightbox(media) {
        const lightbox = document.createElement('div');
        lightbox.className = 'media-lightbox';
        lightbox.innerHTML = `
            <div class="lightbox-content">
                <button class="lightbox-close" onclick="this.closest('.media-lightbox').remove()">
                    <i class="fas fa-times"></i>
                </button>
                <div class="lightbox-player"></div>
            </div>
        `;

        document.body.appendChild(lightbox);

        const playerContainer = lightbox.querySelector('.lightbox-player');
        const player = new MediaPlayer(playerContainer, {
            controls: true,
            autoplay: true
        });

        player.loadMedia(media.id, media);
    }

    /**
     * Форматирует размер файла
     */
    formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
}

// Экспортируем классы для использования
window.MediaPlayer = MediaPlayer;
window.MediaGallery = MediaGallery;
