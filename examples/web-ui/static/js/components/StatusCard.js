function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function formatDuration(ms) {
    if (!ms || ms < 0) return '-';
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
}

function getStatusIcon(status) {
    switch (status) {
        case 'pending': return '⏳';
        case 'running': return '🔄';
        case 'success': return '✅';
        case 'error': return '❌';
        default: return '○';
    }
}

export function createStatusCard(status = {}) {
    const {
        traceId = '',
        startTime = null,
        agents = [],
        collapsed = false
    } = status;
    
    const container = document.createElement('div');
    container.className = collapsed ? 'status-card collapsed' : 'status-card';
    container.dataset.traceId = traceId;
    
    const header = document.createElement('div');
    header.className = 'status-card-header';
    
    const completedCount = agents.filter(a => a.status === 'success').length;
    const totalCount = agents.length;
    const hasError = agents.some(a => a.status === 'error');
    const isRunning = agents.some(a => a.status === 'running');
    
    let titleIcon = '📊';
    if (hasError) titleIcon = '⚠️';
    else if (isRunning) titleIcon = '⏳';
    else if (completedCount === totalCount && totalCount > 0) titleIcon = '✅';
    
    header.innerHTML = `
        <div class="status-card-title">
            <span>${titleIcon}</span>
            <span>处理进度 ${totalCount > 0 ? `(${completedCount}/${totalCount})` : ''}</span>
        </div>
        <span class="status-card-toggle">${collapsed ? '▶' : '▼'}</span>
    `;
    
    const content = document.createElement('div');
    content.className = 'status-card-content';
    content.style.display = collapsed ? 'none' : 'block';
    
    if (agents.length > 0) {
        agents.forEach(agent => {
            const item = document.createElement('div');
            item.className = 'status-item';
            item.dataset.agent = agent.name;
            item.innerHTML = `
                <span class="status-item-icon">${getStatusIcon(agent.status)}</span>
                <span>${escapeHtml(agent.name)}</span>
                ${agent.duration ? `<span style="margin-left: auto; color: var(--text-secondary); font-size: 12px;">${formatDuration(agent.duration)}</span>` : ''}
            `;
            content.appendChild(item);
        });
    } else {
        content.innerHTML = '<div class="status-item"><span style="color: var(--text-secondary);">暂无处理任务</span></div>';
    }
    
    const footer = document.createElement('div');
    footer.className = 'status-card-footer';
    
    const elapsed = startTime ? Date.now() - new Date(startTime).getTime() : 0;
    footer.innerHTML = `
        ${traceId ? `<span>Trace: ${escapeHtml(traceId)}</span>` : ''}
        ${startTime ? `<span style="margin-left: 16px;">耗时: ${formatDuration(elapsed)}</span>` : ''}
    `;
    
    header.addEventListener('click', () => {
        const isCollapsed = container.classList.toggle('collapsed');
        content.style.display = isCollapsed ? 'none' : 'block';
        header.querySelector('.status-card-toggle').textContent = isCollapsed ? '▶' : '▼';
    });
    
    container.appendChild(header);
    container.appendChild(content);
    if (traceId || startTime) {
        container.appendChild(footer);
    }
    
    return container;
}

export function updateStatusCard(container, status) {
    if (!container || !status) return;
    
    const {
        traceId = container.dataset.traceId,
        startTime = null,
        agents = []
    } = status;
    
    container.dataset.traceId = traceId;
    
    const header = container.querySelector('.status-card-header');
    const content = container.querySelector('.status-card-content');
    const footer = container.querySelector('.status-card-footer');
    
    const completedCount = agents.filter(a => a.status === 'success').length;
    const totalCount = agents.length;
    const hasError = agents.some(a => a.status === 'error');
    const isRunning = agents.some(a => a.status === 'running');
    
    let titleIcon = '📊';
    if (hasError) titleIcon = '⚠️';
    else if (isRunning) titleIcon = '⏳';
    else if (completedCount === totalCount && totalCount > 0) titleIcon = '✅';
    
    const titleEl = header.querySelector('.status-card-title');
    if (titleEl) {
        titleEl.innerHTML = `
            <span>${titleIcon}</span>
            <span>处理进度 ${totalCount > 0 ? `(${completedCount}/${totalCount})` : ''}</span>
        `;
    }
    
    if (content) {
        content.innerHTML = '';
        if (agents.length > 0) {
            agents.forEach(agent => {
                const item = document.createElement('div');
                item.className = 'status-item';
                item.dataset.agent = agent.name;
                item.innerHTML = `
                    <span class="status-item-icon">${getStatusIcon(agent.status)}</span>
                    <span>${escapeHtml(agent.name)}</span>
                    ${agent.duration ? `<span style="margin-left: auto; color: var(--text-secondary); font-size: 12px;">${formatDuration(agent.duration)}</span>` : ''}
                `;
                content.appendChild(item);
            });
        } else {
            content.innerHTML = '<div class="status-item"><span style="color: var(--text-secondary);">暂无处理任务</span></div>';
        }
    }
    
    if (footer) {
        const elapsed = startTime ? Date.now() - new Date(startTime).getTime() : 0;
        footer.innerHTML = `
            ${traceId ? `<span>Trace: ${escapeHtml(traceId)}</span>` : ''}
            ${startTime ? `<span style="margin-left: 16px;">耗时: ${formatDuration(elapsed)}</span>` : ''}
        `;
    } else if (traceId || startTime) {
        const newFooter = document.createElement('div');
        newFooter.className = 'status-card-footer';
        const elapsed = startTime ? Date.now() - new Date(startTime).getTime() : 0;
        newFooter.innerHTML = `
            ${traceId ? `<span>Trace: ${escapeHtml(traceId)}</span>` : ''}
            ${startTime ? `<span style="margin-left: 16px;">耗时: ${formatDuration(elapsed)}</span>` : ''}
        `;
        container.appendChild(newFooter);
    }
}

export function updateAgentStatus(container, agentName, status, duration = null) {
    if (!container) return;
    
    const content = container.querySelector('.status-card-content');
    if (!content) return;
    
    let item = content.querySelector(`[data-agent="${agentName}"]`);
    
    if (!item) {
        item = document.createElement('div');
        item.className = 'status-item';
        item.dataset.agent = agentName;
        content.appendChild(item);
    }
    
    item.innerHTML = `
        <span class="status-item-icon">${getStatusIcon(status)}</span>
        <span>${escapeHtml(agentName)}</span>
        ${duration ? `<span style="margin-left: auto; color: var(--text-secondary); font-size: 12px;">${formatDuration(duration)}</span>` : ''}
    `;
    
    const allItems = content.querySelectorAll('.status-item[data-agent]');
    const agents = Array.from(allItems).map(el => ({
        name: el.dataset.agent,
        status: el.querySelector('.status-item-icon')?.textContent === '✅' ? 'success' : 
                el.querySelector('.status-item-icon')?.textContent === '❌' ? 'error' :
                el.querySelector('.status-item-icon')?.textContent === '🔄' ? 'running' : 'pending'
    }));
    
    const completedCount = agents.filter(a => a.status === 'success').length;
    const totalCount = agents.length;
    const hasError = agents.some(a => a.status === 'error');
    const isRunning = agents.some(a => a.status === 'running');
    
    let titleIcon = '📊';
    if (hasError) titleIcon = '⚠️';
    else if (isRunning) titleIcon = '⏳';
    else if (completedCount === totalCount && totalCount > 0) titleIcon = '✅';
    
    const titleEl = container.querySelector('.status-card-title');
    if (titleEl) {
        titleEl.innerHTML = `
            <span>${titleIcon}</span>
            <span>处理进度 ${totalCount > 0 ? `(${completedCount}/${totalCount})` : ''}</span>
        `;
    }
}
