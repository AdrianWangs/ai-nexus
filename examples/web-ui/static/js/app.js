import { fetchAgents, fetchGroups, fetchGroup, deleteGroup, sendMessageStream } from './api.js';
import { store } from './store.js';
import { formatTime, formatAgent, getAgentIcon, getAgentTheme } from './utils.js';
import { createMessageBubble, createInternalMessagesGroup } from './components/MessageBubble.js';
import { createStatusCard, updateStatusCard, updateAgentStatus } from './components/StatusCard.js';
import { createComposer } from './components/Composer.js';
import { createAgentCard, createAgentList, createQuickMentionBar } from './components/AgentCard.js';

class App {
    constructor() {
        this.elements = {};
        this.composer = null;
        this.statusCard = null;
        this.isProcessing = false;
    }

    async init() {
        this.cacheElements();
        this.setupEventListeners();
        this.setupStoreSubscription();
        await this.loadInitialData();
        this.render();
    }

    cacheElements() {
        this.elements = {
            sessionList: document.getElementById('session-list'),
            messageList: document.getElementById('message-list'),
            composerContainer: document.getElementById('composer-container'),
            agentList: document.getElementById('agent-list'),
            quickMention: document.getElementById('quick-mention'),
            statusCardContainer: document.getElementById('status-card-container'),
            chatHeader: document.getElementById('chat-header'),
            newChatBtn: document.getElementById('new-chat-btn'),
            agentCount: document.getElementById('agent-count'),
            toast: document.getElementById('toast')
        };
    }

    setupEventListeners() {
        this.elements.newChatBtn?.addEventListener('click', () => this.createNewChat());

        this.composer = createComposer({
            placeholder: '输入消息，@agent-name 指定服务',
            agents: store.getState().agents,
            onSend: (message) => this.sendMessage(message),
            onMention: (agent) => this.handleMention(agent)
        });
        this.elements.composerContainer?.appendChild(this.composer);
    }

    setupStoreSubscription() {
        store.subscribe((state) => {
            this.renderSessionList(state.groups);
            this.renderAgentPanel(state.agents);
            this.composer?.setAgents(state.agents);
            
            if (state.currentGroup) {
                this.renderMessages(state.currentGroup);
            }
            
            if (state.statusItems.length > 0 || this.isProcessing) {
                this.renderStatusCard(state.statusItems);
            }
        });
    }

    async loadInitialData() {
        try {
            const [agents, groups] = await Promise.all([
                fetchAgents(),
                fetchGroups()
            ]);
            store.setAgents(agents);
            store.setGroups(groups);
            
            if (groups.length > 0) {
                const latestGroup = groups.sort((a, b) => 
                    new Date(b.updated_at) - new Date(a.updated_at)
                )[0];
                await this.selectGroup(latestGroup.id);
            }
        } catch (error) {
            this.showToast('加载数据失败: ' + error.message, true);
        }
    }

    render() {
        const state = store.getState();
        this.renderSessionList(state.groups);
        this.renderAgentPanel(state.agents);
        
        if (state.currentGroup) {
            this.renderMessages(state.currentGroup);
            this.renderChatHeader(state.currentGroup);
        } else {
            this.renderEmptyState();
        }
    }

    renderSessionList(groups) {
        if (!this.elements.sessionList) return;

        if (groups.length === 0) {
            this.elements.sessionList.innerHTML = `
                <div class="empty-state" style="padding: 20px;">
                    <p style="font-size: 13px; color: var(--text-secondary);">暂无会话</p>
                </div>
            `;
            return;
        }

        const currentId = store.getState().currentGroup?.id;
        
        this.elements.sessionList.innerHTML = groups
            .sort((a, b) => new Date(b.updated_at) - new Date(a.updated_at))
            .map(group => {
                const title = group.messages?.[0]?.content?.slice(0, 30) || '新会话';
                const participants = group.participants || [];
                const messageCount = group.messages?.length || 0;
                const time = formatTime(group.updated_at);
                
                return `
                    <div class="session-item ${group.id === currentId ? 'active' : ''}" 
                         data-id="${group.id}">
                        <div class="session-item-title">${this.escapeHtml(title)}</div>
                        <div class="session-item-meta">
                            <div class="session-item-participants">
                                ${participants.slice(0, 3).map(p => 
                                    `<span class="participant-badge">@${p}</span>`
                                ).join('')}
                            </div>
                            <span>${time} · ${messageCount}条</span>
                        </div>
                    </div>
                `;
            }).join('');

        this.elements.sessionList.querySelectorAll('.session-item').forEach(item => {
            item.addEventListener('click', () => {
                const id = item.dataset.id;
                this.selectGroup(id);
            });
        });
    }

    renderAgentPanel(agents) {
        if (this.elements.agentCount) {
            this.elements.agentCount.textContent = `当前在线：${agents.length}`;
        }

        if (this.elements.quickMention) {
            this.elements.quickMention.innerHTML = '';
            const quickBar = createQuickMentionBar(agents, (agent) => this.handleMention(agent));
            this.elements.quickMention.appendChild(quickBar);
        }

        if (this.elements.agentList) {
            this.elements.agentList.innerHTML = '';
            const list = createAgentList(agents, (agent) => this.handleMention(agent));
            this.elements.agentList.appendChild(list);
        }
    }

    renderMessages(group) {
        if (!this.elements.messageList) return;

        this.elements.messageList.innerHTML = '';

        if (!group.messages || group.messages.length === 0) {
            this.elements.messageList.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">💬</div>
                    <div class="empty-state-title">开始对话</div>
                    <div class="empty-state-desc">发送消息开始与 AI 助手对话，使用 @agent-name 指定特定服务</div>
                </div>
            `;
            return;
        }

        const userMessages = [];
        const internalMessages = [];

        group.messages.forEach(msg => {
            if (msg.type === 'internal') {
                internalMessages.push(msg);
            } else {
                if (internalMessages.length > 0) {
                    const internalGroup = createInternalMessagesGroup(
                        [...internalMessages],
                        `内部协商 (${internalMessages.length}条)`
                    );
                    this.elements.messageList.appendChild(internalGroup);
                    internalMessages.length = 0;
                }
                userMessages.push(msg);
                const bubble = createMessageBubble(msg);
                this.elements.messageList.appendChild(bubble);
            }
        });

        if (internalMessages.length > 0) {
            const internalGroup = createInternalMessagesGroup(
                internalMessages,
                `内部协商 (${internalMessages.length}条)`
            );
            this.elements.messageList.appendChild(internalGroup);
        }

        this.elements.messageList.scrollTop = this.elements.messageList.scrollHeight;
    }

    renderChatHeader(group) {
        if (!this.elements.chatHeader) return;

        const title = group.messages?.[0]?.content?.slice(0, 50) || '新会话';
        const participants = group.participants || [];
        const messageCount = group.messages?.length || 0;

        this.elements.chatHeader.innerHTML = `
            <div class="chat-header-title">${this.escapeHtml(title)}</div>
            <div class="chat-header-meta">
                ${participants.map(p => `@${p}`).join(' ')} · ${messageCount}条消息
            </div>
        `;
    }

    renderEmptyState() {
        if (!this.elements.messageList) return;

        this.elements.messageList.innerHTML = `
            <div class="empty-state">
                <div class="empty-state-icon">🤖</div>
                <div class="empty-state-title">AI Nexus</div>
                <div class="empty-state-desc">点击"新建群聊"开始与 AI 助手对话</div>
            </div>
        `;

        if (this.elements.chatHeader) {
            this.elements.chatHeader.innerHTML = `
                <div class="chat-header-title">AI Nexus</div>
                <div class="chat-header-meta">群聊调度台</div>
            `;
        }
    }

    renderStatusCard(statusItems) {
        if (!this.elements.statusCardContainer) return;

        const agents = statusItems.map(item => ({
            name: item.agent || item.name || 'unknown',
            status: item.status || 'pending',
            message: item.message,
            duration: item.duration
        }));

        const status = {
            isProcessing: this.isProcessing,
            agents: agents,
            traceId: this.currentTraceId,
            startTime: this.processingStartTime
        };

        if (!this.statusCard) {
            this.statusCard = createStatusCard(status);
            this.elements.statusCardContainer.innerHTML = '';
            this.elements.statusCardContainer.appendChild(this.statusCard);
        } else {
            updateStatusCard(this.statusCard, status);
        }

        if (!this.isProcessing && statusItems.length === 0) {
            this.elements.statusCardContainer.innerHTML = '';
            this.statusCard = null;
        }
    }

    async selectGroup(groupId) {
        try {
            const group = await fetchGroup(groupId);
            if (group) {
                store.setCurrentGroup(group);
                this.renderChatHeader(group);
            }
        } catch (error) {
            this.showToast('加载会话失败', true);
        }
    }

    async createNewChat() {
        const newGroup = {
            id: 'group_' + Math.random().toString(36).substr(2, 8),
            messages: [],
            participants: [],
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
        };
        store.addGroup(newGroup);
        store.setCurrentGroup(newGroup);
        this.renderChatHeader(newGroup);
        this.composer?.focus();
    }

    async sendMessage(message) {
        if (!message.trim() || this.isProcessing) return;

        let currentGroup = store.getState().currentGroup;
        if (!currentGroup) {
            await this.createNewChat();
            currentGroup = store.getState().currentGroup;
        }

        this.isProcessing = true;
        this.processingStartTime = Date.now();
        this.currentTraceId = null;
        store.clearStatus();
        this.composer?.setDisabled(true);

        const mention = this.extractMention(message);
        const cleanMessage = mention ? message.replace(`@${mention}`, '').trim() : message;

        try {
            await sendMessageStream(
                currentGroup.id,
                message,
                mention,
                (event) => this.handleStreamEvent(event)
            );
        } catch (error) {
            this.showToast('发送失败: ' + error.message, true);
        } finally {
            this.isProcessing = false;
            this.composer?.setDisabled(false);
            this.composer?.clear();
            this.composer?.focus();
            
            setTimeout(() => {
                store.clearStatus();
                this.renderStatusCard([]);
            }, 3000);
        }
    }

    handleStreamEvent(event) {
        this.currentTraceId = event.trace_id;

        switch (event.type) {
            case 'group_message':
                if (event.data) {
                    store.addMessage(event.data);
                }
                break;

            case 'thinking':
                store.updateStatus({
                    agent: 'system',
                    status: 'running',
                    message: event.content
                });
                break;

            case 'agent_call':
                store.updateStatus({
                    agent: event.agent,
                    status: 'running',
                    message: event.content
                });
                break;

            case 'agent_result':
                store.updateStatus({
                    agent: event.agent,
                    status: event.status || 'success',
                    message: event.content
                });
                break;

            case 'content':
                break;

            case 'error':
                store.updateStatus({
                    agent: 'system',
                    status: 'error',
                    message: event.error
                });
                this.showToast(event.error, true);
                break;

            case 'done':
                break;
        }
    }

    handleMention(agent) {
        if (this.composer) {
            const agentName = typeof agent === 'string' ? agent : agent.name;
            const currentValue = this.composer.getValue();
            if (currentValue && !currentValue.startsWith('@')) {
                this.composer.setValue(`@${agentName} ${currentValue}`);
            } else {
                this.composer.appendValue(`@${agentName} `);
            }
            this.composer.focus();
        }
    }

    extractMention(message) {
        const match = message.match(/^@(\S+)/);
        return match ? match[1] : null;
    }

    showToast(message, isError = false) {
        if (!this.elements.toast) return;

        this.elements.toast.textContent = message;
        this.elements.toast.className = 'toast show' + (isError ? ' error' : '');

        setTimeout(() => {
            this.elements.toast.className = 'toast';
        }, 3000);
    }

    escapeHtml(str) {
        if (!str) return '';
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const app = new App();
    app.init();
});
