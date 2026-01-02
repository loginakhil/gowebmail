// API Client
class APIClient {
    constructor(baseURL = '/api') {
        this.baseURL = baseURL;
    }

    async listEmails(params = {}) {
        const query = new URLSearchParams(params).toString();
        const response = await fetch(`${this.baseURL}/emails?${query}`);
        const data = await response.json();
        return data.success ? data.data : null;
    }

    async getEmail(id) {
        const response = await fetch(`${this.baseURL}/emails/${id}`);
        const data = await response.json();
        return data.success ? data.data : null;
    }

    async deleteEmail(id) {
        const response = await fetch(`${this.baseURL}/emails/${id}`, {
            method: 'DELETE'
        });
        return response.ok;
    }

    async deleteAllEmails() {
        const response = await fetch(`${this.baseURL}/emails`, {
            method: 'DELETE'
        });
        return response.ok;
    }

    async searchEmails(query) {
        const response = await fetch(`${this.baseURL}/emails/search?q=${encodeURIComponent(query)}`);
        const data = await response.json();
        return data.success ? data.data : null;
    }

    async getStats() {
        const response = await fetch(`${this.baseURL}/stats`);
        const data = await response.json();
        return data.success ? data.data : null;
    }
}

// WebSocket Client
class WebSocketClient {
    constructor(url = '/ws') {
        this.url = url;
        this.ws = null;
        this.listeners = {};
        this.reconnectDelay = 1000;
        this.maxReconnectDelay = 30000;
    }

    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsURL = `${protocol}//${window.location.host}${this.url}`;

        this.ws = new WebSocket(wsURL);

        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.reconnectDelay = 1000;
            this.emit('connected');
        };

        this.ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.emit(message.type, message.data);
            } catch (e) {
                console.error('Failed to parse WebSocket message:', e);
            }
        };

        this.ws.onclose = () => {
            console.log('WebSocket disconnected, reconnecting...');
            this.emit('disconnected');
            setTimeout(() => this.connect(), this.reconnectDelay);
            this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay);
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }

    on(event, callback) {
        if (!this.listeners[event]) {
            this.listeners[event] = [];
        }
        this.listeners[event].push(callback);
    }

    emit(event, data) {
        if (this.listeners[event]) {
            this.listeners[event].forEach(callback => callback(data));
        }
    }
}

// Main Application
class App {
    constructor() {
        this.api = new APIClient();
        this.ws = new WebSocketClient();
        this.emails = [];
        this.selectedEmail = null;
        this.currentView = 'html';

        this.init();
    }

    init() {
        this.setupEventListeners();
        this.setupWebSocket();
        this.loadEmails();
        this.updateStats();
    }

    setupEventListeners() {
        // Refresh button
        document.getElementById('refresh-btn').addEventListener('click', () => {
            this.loadEmails();
        });

        // Delete all button
        document.getElementById('delete-all-btn').addEventListener('click', () => {
            this.deleteAllEmails();
        });

        // Delete selected button
        document.getElementById('delete-selected-btn').addEventListener('click', () => {
            if (this.selectedEmail) {
                this.deleteEmail(this.selectedEmail.id);
            }
        });

        // Search button
        document.getElementById('search-btn').addEventListener('click', () => {
            this.search();
        });

        // Clear filters button
        document.getElementById('clear-filters-btn').addEventListener('click', () => {
            document.getElementById('search-input').value = '';
            this.loadEmails();
        });

        // Search on Enter key
        document.getElementById('search-input').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.search();
            }
        });
    }

    setupWebSocket() {
        this.ws.on('connected', () => {
            document.getElementById('connection-status').className = 'stat status-connected';
        });

        this.ws.on('disconnected', () => {
            document.getElementById('connection-status').className = 'stat status-disconnected';
        });

        this.ws.on('email.new', (data) => {
            this.handleNewEmail(data);
        });

        this.ws.on('email.deleted', (data) => {
            this.handleEmailDeleted(data);
        });

        this.ws.on('emails.cleared', () => {
            this.handleEmailsCleared();
        });

        this.ws.connect();
    }

    async loadEmails() {
        this.showLoading(true);
        const result = await this.api.listEmails({ limit: 100 });
        this.showLoading(false);

        if (result) {
            this.emails = result.emails || [];
            this.renderEmailList();
            this.updateStats();
        }
    }

    async search() {
        const query = document.getElementById('search-input').value.trim();
        if (!query) {
            this.loadEmails();
            return;
        }

        this.showLoading(true);
        const result = await this.api.searchEmails(query);
        this.showLoading(false);

        if (result) {
            this.emails = result.emails || [];
            this.renderEmailList();
        }
    }

    renderEmailList() {
        const listEl = document.getElementById('email-list');
        const emptyEl = document.getElementById('empty-state');

        if (this.emails.length === 0) {
            listEl.innerHTML = '';
            emptyEl.style.display = 'flex';
            return;
        }

        emptyEl.style.display = 'none';
        listEl.innerHTML = this.emails.map(email => this.createEmailItem(email)).join('');

        // Add click listeners
        listEl.querySelectorAll('.email-item').forEach((item, index) => {
            item.addEventListener('click', () => {
                this.selectEmail(this.emails[index]);
            });
        });
    }

    createEmailItem(email) {
        const date = new Date(email.receivedAt);
        const timeStr = this.formatTime(date);
        const from = email.from || 'Unknown';
        const subject = email.subject || '(No subject)';

        return `
            <div class="email-item ${email.id === this.selectedEmail?.id ? 'selected' : ''}" data-id="${email.id}">
                <div class="email-from">${this.escapeHtml(from)}</div>
                <div class="email-subject">${this.escapeHtml(subject)}</div>
                <div class="email-meta">
                    <span>${timeStr}</span>
                    <span>${this.formatSize(email.size)}</span>
                </div>
            </div>
        `;
    }

    async selectEmail(email) {
        this.selectedEmail = email;
        this.renderEmailList();

        // Enable delete selected button
        document.getElementById('delete-selected-btn').disabled = false;

        // Load full email details
        const fullEmail = await this.api.getEmail(email.id);
        if (fullEmail) {
            this.renderEmailPreview(fullEmail);
        }
    }

    renderEmailPreview(email) {
        const previewEl = document.getElementById('email-preview');
        
        const hasHTML = email.bodyHTML && email.bodyHTML.trim() !== '';
        const hasPlain = email.bodyPlain && email.bodyPlain.trim() !== '';

        previewEl.innerHTML = `
            <div class="email-header">
                <div class="email-subject-line">${this.escapeHtml(email.subject || '(No subject)')}</div>
                <div class="email-details">
                    <div class="email-detail">
                        <div class="email-detail-label">From:</div>
                        <div class="email-detail-value">${this.escapeHtml(email.from)}</div>
                    </div>
                    <div class="email-detail">
                        <div class="email-detail-label">To:</div>
                        <div class="email-detail-value">${this.escapeHtml(email.to.join(', '))}</div>
                    </div>
                    ${email.cc && email.cc.length > 0 ? `
                    <div class="email-detail">
                        <div class="email-detail-label">CC:</div>
                        <div class="email-detail-value">${this.escapeHtml(email.cc.join(', '))}</div>
                    </div>
                    ` : ''}
                    <div class="email-detail">
                        <div class="email-detail-label">Date:</div>
                        <div class="email-detail-value">${new Date(email.receivedAt).toLocaleString()}</div>
                    </div>
                </div>
            </div>
            <div class="email-body">
                <div class="email-tabs">
                    ${hasHTML ? '<button class="email-tab active" data-view="html">HTML</button>' : ''}
                    ${hasPlain ? `<button class="email-tab ${!hasHTML ? 'active' : ''}" data-view="plain">Plain Text</button>` : ''}
                    <button class="email-tab" data-view="raw">Raw</button>
                </div>
                <div class="email-content" id="email-content">
                    ${this.renderEmailContent(email, hasHTML ? 'html' : 'plain')}
                </div>
                ${email.attachments && email.attachments.length > 0 ? `
                <div class="email-attachments">
                    <h3>Attachments (${email.attachments.length})</h3>
                    ${email.attachments.map(att => `
                        <div class="attachment-item">
                            📎 <a href="/api/emails/${email.id}/attachments/${att.id}" download="${att.filename}">
                                ${this.escapeHtml(att.filename)} (${this.formatSize(att.size)})
                            </a>
                        </div>
                    `).join('')}
                </div>
                ` : ''}
            </div>
        `;

        // Add tab click listeners
        previewEl.querySelectorAll('.email-tab').forEach(tab => {
            tab.addEventListener('click', () => {
                const view = tab.dataset.view;
                this.currentView = view;
                
                // Update active tab
                previewEl.querySelectorAll('.email-tab').forEach(t => t.classList.remove('active'));
                tab.classList.add('active');
                
                // Update content
                document.getElementById('email-content').innerHTML = this.renderEmailContent(email, view);
            });
        });
    }

    renderEmailContent(email, view) {
        switch (view) {
            case 'html':
                return `<iframe srcdoc="${this.escapeHtml(email.bodyHTML)}" style="width: 100%; min-height: 400px; border: none;"></iframe>`;
            case 'plain':
                return `<pre>${this.escapeHtml(email.bodyPlain)}</pre>`;
            case 'raw':
                let raw = '';
                for (const [key, values] of Object.entries(email.headers || {})) {
                    values.forEach(value => {
                        raw += `${key}: ${value}\n`;
                    });
                }
                raw += '\n' + (email.bodyPlain || email.bodyHTML || '');
                return `<pre>${this.escapeHtml(raw)}</pre>`;
            default:
                return '';
        }
    }

    async deleteEmail(id) {
        if (!confirm('Are you sure you want to delete this email?')) {
            return;
        }

        const success = await this.api.deleteEmail(id);
        if (success) {
            this.emails = this.emails.filter(e => e.id !== id);
            this.selectedEmail = null;
            this.renderEmailList();
            document.getElementById('email-preview').innerHTML = '<div class="preview-empty"><div class="empty-icon">📧</div><p>Select an email to view</p></div>';
            document.getElementById('delete-selected-btn').disabled = true;
            this.updateStats();
        }
    }

    async deleteAllEmails() {
        if (!confirm('Are you sure you want to delete ALL emails? This cannot be undone.')) {
            return;
        }

        const success = await this.api.deleteAllEmails();
        if (success) {
            this.emails = [];
            this.selectedEmail = null;
            this.renderEmailList();
            document.getElementById('email-preview').innerHTML = '<div class="preview-empty"><div class="empty-icon">📧</div><p>Select an email to view</p></div>';
            document.getElementById('delete-selected-btn').disabled = true;
            this.updateStats();
        }
    }

    handleNewEmail(data) {
        // Add to beginning of list
        this.emails.unshift(data);
        this.renderEmailList();
        this.updateStats();
    }

    handleEmailDeleted(data) {
        this.emails = this.emails.filter(e => e.id !== data.id);
        if (this.selectedEmail && this.selectedEmail.id === data.id) {
            this.selectedEmail = null;
            document.getElementById('email-preview').innerHTML = '<div class="preview-empty"><div class="empty-icon">📧</div><p>Select an email to view</p></div>';
            document.getElementById('delete-selected-btn').disabled = true;
        }
        this.renderEmailList();
        this.updateStats();
    }

    handleEmailsCleared() {
        this.emails = [];
        this.selectedEmail = null;
        this.renderEmailList();
        document.getElementById('email-preview').innerHTML = '<div class="preview-empty"><div class="empty-icon">📧</div><p>Select an email to view</p></div>';
        document.getElementById('delete-selected-btn').disabled = true;
        this.updateStats();
    }

    async updateStats() {
        const stats = await this.api.getStats();
        if (stats) {
            document.getElementById('email-count').textContent = `${stats.totalEmails} email${stats.totalEmails !== 1 ? 's' : ''}`;
        }
    }

    showLoading(show) {
        document.getElementById('loading').style.display = show ? 'flex' : 'none';
    }

    formatTime(date) {
        const now = new Date();
        const diff = now - date;
        const seconds = Math.floor(diff / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
        const days = Math.floor(hours / 24);

        if (days > 0) return `${days}d ago`;
        if (hours > 0) return `${hours}h ago`;
        if (minutes > 0) return `${minutes}m ago`;
        return 'Just now';
    }

    formatSize(bytes) {
        if (bytes < 1024) return bytes + ' B';
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
        return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Initialize app when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => new App());
} else {
    new App();
}
