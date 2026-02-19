export function formatTime(timestamp) {
    if (!timestamp) return '';

    const date = typeof timestamp === 'string' ? new Date(timestamp) : timestamp;
    if (isNaN(date.getTime())) return '';

    const now = new Date();
    const diff = now - date;
    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (seconds < 60) return '刚刚';
    if (minutes < 60) return `${minutes}分钟前`;
    if (hours < 24) return `${hours}小时前`;
    if (days < 7) return `${days}天前`;

    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hour = String(date.getHours()).padStart(2, '0');
    const minute = String(date.getMinutes()).padStart(2, '0');

    if (date.getFullYear() === now.getFullYear()) {
        return `${month}-${day} ${hour}:${minute}`;
    }
    return `${date.getFullYear()}-${month}-${day} ${hour}:${minute}`;
}

export function formatAgent(name) {
    if (!name) return '';
    return name.startsWith('@') ? name.slice(1) : name;
}

const agentThemes = {
    'orchestrator': { bg: 'linear-gradient(135deg, #3b82f6, #a855f7)', color: '#fff' },
    'user-service': { bg: '#3b82f6', color: '#fff' },
    'order-service': { bg: '#22c55e', color: '#fff' },
    'product-service': { bg: '#a855f7', color: '#fff' },
    'payment-service': { bg: '#f97316', color: '#fff' },
    'inventory-service': { bg: '#06b6d4', color: '#fff' },
    'notification-service': { bg: '#ec4899', color: '#fff' },
    'analytics-service': { bg: '#8b5cf6', color: '#fff' }
};

const defaultThemes = [
    { bg: '#3b82f6', color: '#fff' },
    { bg: '#22c55e', color: '#fff' },
    { bg: '#a855f7', color: '#fff' },
    { bg: '#f97316', color: '#fff' },
    { bg: '#06b6d4', color: '#fff' },
    { bg: '#ec4899', color: '#fff' },
    { bg: '#8b5cf6', color: '#fff' },
    { bg: '#14b8a6', color: '#fff' }
];

export function getAgentTheme(name) {
    const normalized = formatAgent(name).toLowerCase();

    if (agentThemes[normalized]) {
        return agentThemes[normalized];
    }

    let hash = 0;
    for (let i = 0; i < normalized.length; i++) {
        hash = ((hash << 5) - hash) + normalized.charCodeAt(i);
        hash = hash & hash;
    }
    const index = Math.abs(hash) % defaultThemes.length;
    return defaultThemes[index];
}

const agentIcons = {
    'orchestrator': '🎯',
    'user-service': '👤',
    'order-service': '📦',
    'product-service': '🏷️',
    'payment-service': '💳',
    'inventory-service': '📊',
    'notification-service': '🔔',
    'analytics-service': '📈',
    'user': '👤'
};

const defaultIcons = ['🤖', '⚡', '🔧', '💡', '🎨', '🔮', '🌟', '🚀'];

export function getAgentIcon(name) {
    if (!name) return '🤖';

    const normalized = formatAgent(name).toLowerCase();

    if (agentIcons[normalized]) {
        return agentIcons[normalized];
    }

    let hash = 0;
    for (let i = 0; i < normalized.length; i++) {
        hash = ((hash << 5) - hash) + normalized.charCodeAt(i);
        hash = hash & hash;
    }
    const index = Math.abs(hash) % defaultIcons.length;
    return defaultIcons[index];
}

export function escapeHtml(str) {
    if (!str) return '';
    const escapeMap = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#39;'
    };
    return str.replace(/[&<>"']/g, char => escapeMap[char]);
}

export function renderMarkdown(content) {
    if (!content) return '';

    let html = escapeHtml(content);

    html = html.replace(/```(\w*)\n([\s\S]*?)```/g, (match, lang, code) => {
        return `<pre><code class="language-${lang}">${code.trim()}</code></pre>`;
    });

    html = html.replace(/`([^`]+)`/g, '<code>$1</code>');

    html = html.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');

    html = html.replace(/\*([^*]+)\*/g, '<em>$1</em>');

    html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener">$1</a>');

    html = html.replace(/^### (.+)$/gm, '<h3>$1</h3>');
    html = html.replace(/^## (.+)$/gm, '<h2>$1</h2>');
    html = html.replace(/^# (.+)$/gm, '<h1>$1</h1>');

    html = html.replace(/^[-*] (.+)$/gm, '<li>$1</li>');
    html = html.replace(/(<li>.*<\/li>\n?)+/g, '<ul>$&</ul>');

    html = html.replace(/^\d+\. (.+)$/gm, '<li>$1</li>');

    html = html.replace(/\n\n/g, '</p><p>');
    html = `<p>${html}</p>`;
    html = html.replace(/<p><\/p>/g, '');
    html = html.replace(/<p>(<h[1-3]>)/g, '$1');
    html = html.replace(/(<\/h[1-3]>)<\/p>/g, '$1');
    html = html.replace(/<p>(<pre>)/g, '$1');
    html = html.replace(/(<\/pre>)<\/p>/g, '$1');
    html = html.replace(/<p>(<ul>)/g, '$1');
    html = html.replace(/(<\/ul>)<\/p>/g, '$1');

    return html;
}
