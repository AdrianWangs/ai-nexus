function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function getAgentIconClass(name) {
    const normalized = name.toLowerCase().replace(/[_-]/g, '-');
    if (normalized.includes('orchestrator')) return 'orchestrator';
    if (normalized.includes('user-service') || normalized.includes('userservice')) return 'user-service';
    if (normalized.includes('order-service') || normalized.includes('orderservice')) return 'order-service';
    if (normalized.includes('product-service') || normalized.includes('productservice')) return 'product-service';
    return '';
}

function getAgentIcon(name) {
    const normalized = name.toLowerCase();
    if (normalized.includes('orchestrator')) return '🎯';
    if (normalized.includes('user-service') || normalized.includes('userservice')) return '👥';
    if (normalized.includes('order-service') || normalized.includes('orderservice')) return '📦';
    if (normalized.includes('product-service') || normalized.includes('productservice')) return '🛍️';
    return '🤖';
}

export function createAgentCard(agent, onMention = () => {}) {
    const {
        name = 'Unknown',
        description = '',
        capabilities = [],
        status = 'online'
    } = agent;
    
    const container = document.createElement('div');
    container.className = 'agent-card';
    container.dataset.agent = name;
    
    const iconClass = getAgentIconClass(name);
    const icon = getAgentIcon(name);
    
    const header = document.createElement('div');
    header.className = 'agent-card-header';
    header.innerHTML = `
        <div class="agent-card-icon ${iconClass}">${icon}</div>
        <div class="agent-card-info">
            <h3>${escapeHtml(name)}</h3>
            <p>${escapeHtml(description || '暂无描述')}</p>
        </div>
    `;
    
    container.appendChild(header);
    
    if (capabilities && capabilities.length > 0) {
        const caps = document.createElement('div');
        caps.className = 'agent-card-caps';
        capabilities.forEach(cap => {
            const capEl = document.createElement('span');
            capEl.className = 'agent-card-cap';
            capEl.textContent = cap;
            caps.appendChild(capEl);
        });
        container.appendChild(caps);
    }
    
    const action = document.createElement('div');
    action.className = 'agent-card-action';
    
    const mentionBtn = document.createElement('button');
    mentionBtn.className = 'agent-card-mention-btn';
    mentionBtn.innerHTML = `<span>@</span> 提及`;
    mentionBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        onMention(name);
    });
    
    action.appendChild(mentionBtn);
    container.appendChild(action);
    
    return container;
}

export function createAgentList(agents, onMention = () => {}) {
    const container = document.createElement('div');
    container.className = 'agent-list';
    
    if (!agents || agents.length === 0) {
        container.innerHTML = `
            <div class="empty-state" style="padding: 40px 20px;">
                <div class="empty-state-icon">🤖</div>
                <div class="empty-state-title">暂无可用 Agent</div>
                <div class="empty-state-desc">请检查服务是否正常运行</div>
            </div>
        `;
        return container;
    }
    
    agents.forEach(agent => {
        const card = createAgentCard(agent, onMention);
        container.appendChild(card);
    });
    
    return container;
}

export function createQuickMentionBar(agents, onMention = () => {}) {
    const container = document.createElement('div');
    container.className = 'quick-mention';
    
    const title = document.createElement('div');
    title.className = 'quick-mention-title';
    title.textContent = '快速提及';
    
    const list = document.createElement('div');
    list.className = 'quick-mention-list';
    
    agents.forEach(agent => {
        const btn = document.createElement('button');
        btn.className = 'quick-mention-btn';
        btn.textContent = `@${agent.name}`;
        btn.addEventListener('click', () => {
            onMention(agent.name);
        });
        list.appendChild(btn);
    });
    
    container.appendChild(title);
    container.appendChild(list);
    
    return container;
}
