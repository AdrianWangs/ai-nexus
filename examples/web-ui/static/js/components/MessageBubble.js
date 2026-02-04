function getAgentClass(name) {
    const normalized = name.replace('@', '').toLowerCase().replace(/[_-]/g, '-');
    if (normalized.includes('orchestrator')) return 'orchestrator';
    if (normalized.includes('user-service') || normalized.includes('userservice')) return 'user-service';
    if (normalized.includes('order-service') || normalized.includes('orderservice')) return 'order-service';
    if (normalized.includes('product-service') || normalized.includes('productservice')) return 'product-service';
    return '';
}

function getAgentIcon(name) {
    const normalized = name.replace('@', '').toLowerCase();
    if (normalized === 'user') return '👤';
    if (normalized.includes('orchestrator')) return '🎯';
    if (normalized.includes('user-service') || normalized.includes('userservice')) return '👥';
    if (normalized.includes('order-service') || normalized.includes('orderservice')) return '📦';
    if (normalized.includes('product-service') || normalized.includes('productservice')) return '🛍️';
    return '🤖';
}

function formatTime(timestamp) {
    if (!timestamp) return '';
    const date = new Date(timestamp);
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function renderToolCalls(toolCalls) {
    if (!toolCalls || toolCalls.length === 0) return '';
    
    const items = toolCalls.map(tool => {
        const name = tool.name || tool.function?.name || 'unknown';
        return `
            <div class="tool-item">
                <span class="tool-item-icon">🔧</span>
                <span class="tool-item-name">${escapeHtml(name)}</span>
            </div>
        `;
    }).join('');
    
    return `<div class="message-tools">${items}</div>`;
}

function renderToolResults(toolResults) {
    if (!toolResults || toolResults.length === 0) return '';
    
    const items = toolResults.map(result => {
        const status = result.success !== false ? '✅' : '❌';
        const name = result.name || 'tool';
        return `
            <div class="tool-item">
                <span class="tool-item-icon">${status}</span>
                <span class="tool-item-name">${escapeHtml(name)}</span>
            </div>
        `;
    }).join('');
    
    return `<div class="message-tools">${items}</div>`;
}

export function createMessageBubble(message) {
    const { from, to, content, type, timestamp, toolCalls, toolResults } = message;
    
    const isUser = type === 'user' || from === 'user';
    const bubbleClass = isUser ? 'message-bubble user' : 'message-bubble';
    const fromName = from || 'unknown';
    const toName = to || '';
    const agentClass = getAgentClass(fromName);
    const avatarClass = isUser ? 'message-avatar user' : `message-avatar ${agentClass}`;
    const icon = getAgentIcon(fromName);
    
    const showHeader = !isUser && toName;
    const headerHtml = showHeader ? `
        <div class="message-header">
            <span class="message-from">${escapeHtml(fromName)}</span>
            <span class="message-arrow">→</span>
            <span class="message-to">${escapeHtml(toName)}</span>
            <span class="message-time">${formatTime(timestamp)}</span>
        </div>
    ` : '';
    
    const toolCallsHtml = renderToolCalls(toolCalls);
    const toolResultsHtml = renderToolResults(toolResults);
    
    const html = `
        <div class="${bubbleClass}" data-type="${type || 'message'}">
            <div class="${avatarClass}">${icon}</div>
            <div class="message-content">
                ${headerHtml}
                <div class="message-body">${escapeHtml(content || '')}</div>
                ${toolCallsHtml}
                ${toolResultsHtml}
            </div>
        </div>
    `;
    
    const template = document.createElement('template');
    template.innerHTML = html.trim();
    return template.content.firstChild;
}

export function createInternalMessagesGroup(messages, title = '内部通信') {
    const container = document.createElement('div');
    container.className = 'internal-messages';
    
    const header = document.createElement('div');
    header.className = 'internal-messages-header';
    header.innerHTML = `
        <div class="internal-messages-title">
            <span>🔄</span>
            <span>${escapeHtml(title)} (${messages.length})</span>
        </div>
        <span class="status-card-toggle">▼</span>
    `;
    
    const content = document.createElement('div');
    content.className = 'internal-messages-content';
    content.style.display = 'none';
    
    messages.forEach(msg => {
        const bubble = createMessageBubble({ ...msg, type: 'internal' });
        content.appendChild(bubble);
    });
    
    header.addEventListener('click', () => {
        const isHidden = content.style.display === 'none';
        content.style.display = isHidden ? 'block' : 'none';
        header.querySelector('.status-card-toggle').textContent = isHidden ? '▲' : '▼';
    });
    
    container.appendChild(header);
    container.appendChild(content);
    
    return container;
}
