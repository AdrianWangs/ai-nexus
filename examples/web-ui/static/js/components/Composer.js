function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

export function createComposer(options = {}) {
    const {
        placeholder = '输入消息，使用 @ 提及 Agent...',
        hint = 'Enter 发送，Shift+Enter 换行',
        agents = [],
        onSend = () => {},
        onMention = () => {}
    } = options;
    
    const container = document.createElement('div');
    container.className = 'composer';
    
    const inputWrapper = document.createElement('div');
    inputWrapper.className = 'composer-input-wrapper';
    inputWrapper.style.position = 'relative';
    
    const dropdown = document.createElement('div');
    dropdown.className = 'mention-dropdown';
    
    const textarea = document.createElement('textarea');
    textarea.className = 'composer-input';
    textarea.placeholder = placeholder;
    textarea.rows = 1;
    
    const sendBtn = document.createElement('button');
    sendBtn.className = 'composer-send';
    sendBtn.innerHTML = '↑';
    sendBtn.disabled = true;
    
    const hintEl = document.createElement('div');
    hintEl.className = 'composer-hint';
    hintEl.textContent = hint;
    
    let currentAgents = [...agents];
    let selectedIndex = -1;
    let mentionStart = -1;
    
    function updateDropdown(filter = '') {
        const filtered = currentAgents.filter(agent => 
            agent.name.toLowerCase().includes(filter.toLowerCase()) ||
            (agent.description && agent.description.toLowerCase().includes(filter.toLowerCase()))
        );
        
        if (filtered.length === 0) {
            dropdown.classList.remove('show');
            return;
        }
        
        dropdown.innerHTML = filtered.map((agent, index) => `
            <div class="mention-item${index === selectedIndex ? ' selected' : ''}" data-name="${escapeHtml(agent.name)}" data-index="${index}">
                <span class="mention-item-name">@${escapeHtml(agent.name)}</span>
                ${agent.description ? `<span class="mention-item-desc">${escapeHtml(agent.description)}</span>` : ''}
            </div>
        `).join('');
        
        dropdown.querySelectorAll('.mention-item').forEach(item => {
            item.addEventListener('click', () => {
                selectMention(item.dataset.name);
            });
            item.addEventListener('mouseenter', () => {
                selectedIndex = parseInt(item.dataset.index);
                updateSelectedItem();
            });
        });
        
        dropdown.classList.add('show');
    }
    
    function updateSelectedItem() {
        dropdown.querySelectorAll('.mention-item').forEach((item, index) => {
            item.style.background = index === selectedIndex ? 'var(--bg-hover)' : '';
        });
    }
    
    function selectMention(name) {
        const value = textarea.value;
        const before = value.substring(0, mentionStart);
        const after = value.substring(textarea.selectionStart);
        textarea.value = before + '@' + name + ' ' + after;
        textarea.focus();
        const newPos = mentionStart + name.length + 2;
        textarea.setSelectionRange(newPos, newPos);
        dropdown.classList.remove('show');
        mentionStart = -1;
        selectedIndex = -1;
        onMention(name);
        updateSendButton();
    }
    
    function updateSendButton() {
        sendBtn.disabled = textarea.value.trim() === '';
    }
    
    function autoResize() {
        textarea.style.height = 'auto';
        const maxHeight = 200;
        textarea.style.height = Math.min(textarea.scrollHeight, maxHeight) + 'px';
    }
    
    function handleSend() {
        const message = textarea.value.trim();
        if (message) {
            onSend(message);
            textarea.value = '';
            autoResize();
            updateSendButton();
        }
    }
    
    textarea.addEventListener('input', (e) => {
        updateSendButton();
        autoResize();
        
        const value = textarea.value;
        const cursorPos = textarea.selectionStart;
        
        const textBeforeCursor = value.substring(0, cursorPos);
        const atIndex = textBeforeCursor.lastIndexOf('@');
        
        if (atIndex !== -1) {
            const charBefore = atIndex > 0 ? textBeforeCursor[atIndex - 1] : ' ';
            if (charBefore === ' ' || charBefore === '\n' || atIndex === 0) {
                const query = textBeforeCursor.substring(atIndex + 1);
                if (!query.includes(' ') && !query.includes('\n')) {
                    mentionStart = atIndex;
                    selectedIndex = 0;
                    updateDropdown(query);
                    return;
                }
            }
        }
        
        dropdown.classList.remove('show');
        mentionStart = -1;
    });
    
    textarea.addEventListener('keydown', (e) => {
        if (dropdown.classList.contains('show')) {
            const items = dropdown.querySelectorAll('.mention-item');
            
            if (e.key === 'ArrowDown') {
                e.preventDefault();
                selectedIndex = Math.min(selectedIndex + 1, items.length - 1);
                updateSelectedItem();
                return;
            }
            
            if (e.key === 'ArrowUp') {
                e.preventDefault();
                selectedIndex = Math.max(selectedIndex - 1, 0);
                updateSelectedItem();
                return;
            }
            
            if (e.key === 'Enter' || e.key === 'Tab') {
                if (selectedIndex >= 0 && items[selectedIndex]) {
                    e.preventDefault();
                    selectMention(items[selectedIndex].dataset.name);
                    return;
                }
            }
            
            if (e.key === 'Escape') {
                dropdown.classList.remove('show');
                mentionStart = -1;
                return;
            }
        }
        
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSend();
        }
    });
    
    sendBtn.addEventListener('click', handleSend);
    
    inputWrapper.appendChild(dropdown);
    inputWrapper.appendChild(textarea);
    inputWrapper.appendChild(sendBtn);
    container.appendChild(inputWrapper);
    container.appendChild(hintEl);
    
    container.setAgents = (newAgents) => {
        currentAgents = [...newAgents];
    };
    
    container.focus = () => {
        textarea.focus();
    };
    
    container.getValue = () => textarea.value;
    
    container.setValue = (value) => {
        textarea.value = value;
        autoResize();
        updateSendButton();
    };
    
    container.appendValue = (value) => {
        textarea.value += value;
        autoResize();
        updateSendButton();
        textarea.focus();
    };
    
    container.clear = () => {
        textarea.value = '';
        autoResize();
        updateSendButton();
    };
    
    container.setDisabled = (disabled) => {
        textarea.disabled = disabled;
        sendBtn.disabled = disabled || textarea.value.trim() === '';
    };
    
    return container;
}
