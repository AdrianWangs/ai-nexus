(function () {
  const el = {
    sessionList: document.getElementById('session-list'),
    messages: document.getElementById('messages'),
    agentList: document.getElementById('agent-list'),
    agentPanel: document.getElementById('agent-panel'),
    agentCount: document.getElementById('agent-count'),
    chatTitle: document.getElementById('chat-title'),
    input: document.getElementById('message-input'),
    sendBtn: document.getElementById('send-btn'),
    newBtn: document.getElementById('new-session-btn'),
    toggleAgentsBtn: document.getElementById('toggle-agents-btn'),
    closeAgentsBtn: document.getElementById('close-panel-btn'),
    mentionPopup: document.getElementById('mention-popup'),
  };

  let eventSource = null;
  let mentionQuery = '';
  let mentionIndex = 0;

  function initials(s) {
    if (!s) return '?';
    const x = String(s).replace(/^@/, '');
    return x.slice(0, 2).toUpperCase();
  }

  function escapeHtml(str) {
    const div = document.createElement('div');
    div.textContent = str || '';
    return div.innerHTML;
  }

  function render(state) {
    // sessions
    el.sessionList.innerHTML = (state.sessions || []).map(sess => {
      const active = sess.id === state.currentSessionId ? 'active' : '';
      const title = escapeHtml(sess.title || sess.id.slice(0, 8));
      const participants = (sess.participants || []).join(' ');
      return `
        <div class="session-item ${active}" data-id="${sess.id}">
          <div class="session-title">${title}</div>
          <div class="session-meta">${escapeHtml(participants)}</div>
        </div>`;
    }).join('');
    el.sessionList.querySelectorAll('.session-item').forEach(x => x.addEventListener('click', () => selectSession(x.dataset.id)));

    // agents
    el.agentCount.textContent = String((state.agents || []).length);
    el.agentList.innerHTML = (state.agents || []).map(a => {
      return `
        <div class="agent-card">
          <div class="agent-name">@${escapeHtml(a.name)}</div>
          <div class="agent-desc">${escapeHtml(a.description || '')}</div>
        </div>`;
    }).join('');

    // messages
    const msgs = state.messages || [];
    let html = msgs.map(m => renderMessage(m)).join('');
    if (state.isStreaming && state.streamingMessage) {
      html += renderMessage({
        id: state.streamingMessage.id,
        from: state.streamingMessage.from,
        to: 'user',
        content: state.streamingMessage.content,
        type: 'reply',
        timestamp: new Date().toISOString(),
      }, true);
    }
    el.messages.innerHTML = html;
    el.messages.parentElement?.scrollTo({ top: el.messages.parentElement.scrollHeight });

    // input enable
    const enabled = Boolean(state.currentSessionId);
    el.input.disabled = !enabled;
    el.sendBtn.disabled = !enabled;
  }

  function renderMessage(m, streaming) {
    const isUser = m.from === 'user';
    const klass = m.type === 'internal' ? 'internal' : (isUser ? 'user' : 'agent');
    const avatar = isUser ? '你' : initials(m.from);
    const from = isUser ? '你' : escapeHtml(m.from);
    const time = m.timestamp ? new Date(m.timestamp).toLocaleTimeString() : '';
    const body = escapeHtml(m.content || '').replace(/\n/g, '<br/>');
    return `
      <div class="message ${klass}" data-id="${m.id}">
        <div class="avatar">${avatar}</div>
        <div class="bubble">
          <div class="meta"><span class="from">${from}</span><span>${time}${streaming ? ' · …' : ''}</span></div>
          <div>${body}</div>
        </div>
      </div>`;
  }

  async function loadInitial() {
    const [agents, sessions] = await Promise.all([api.getAgents().catch(() => []), api.getSessions().catch(() => [])]);
    store.setAgents(agents || []);
    store.setSessions(sessions || []);
    if ((sessions || []).length === 0) {
      const sess = await api.createSession('New Chat');
      store.setSessions([sess]);
      await selectSession(sess.id);
    } else {
      await selectSession(sessions[0].id);
    }
  }

  async function selectSession(sessionId) {
    store.setCurrentSession(sessionId);
    el.chatTitle.textContent = sessionId ? `会话 ${sessionId.slice(0, 8)}` : '选择或创建对话';

    if (eventSource) {
      eventSource.close();
      eventSource = null;
    }

    const msgs = await api.getMessages(sessionId).catch(() => []);
    store.setMessages(msgs || []);

    eventSource = api.createSSE(sessionId, handleEvent);
  }

  async function createSession() {
    const sess = await api.createSession('New Chat');
    const st = store.getState();
    store.setSessions([sess, ...(st.sessions || [])]);
    await selectSession(sess.id);
  }

  function handleEvent(e) {
    switch (e.type) {
      case 'message_start':
        store.startStreaming(e.from || 'Agent');
        break;
      case 'content_delta':
        store.appendDelta(e.content || '');
        break;
      case 'message_end':
      case 'done':
        store.finishStreaming();
        break;
      case 'tool_call':
      case 'tool_result':
        // optional: append internal log
        store.addMessage({
          id: Date.now().toString(),
          from: e.from || 'tools',
          to: e.to || 'user',
          content: e.type === 'tool_call' ? `Tool call: ${e.tool_call?.name || ''}` : `Tool result: ${e.tool_result?.content || ''}`,
          type: 'internal',
          timestamp: new Date().toISOString(),
        });
        break;
      case 'error':
        store.setError(e.error || 'Unknown error');
        store.finishStreaming();
        break;
    }
  }

  async function send() {
    const st = store.getState();
    const contentRaw = (el.input.value || '').trim();
    if (!st.currentSessionId || !contentRaw) return;

    const match = contentRaw.match(/^@(\S+)\s+(.*)$/);
    const target = match ? match[1] : '';
    const content = match ? match[2] : contentRaw;

    store.addMessage({ id: Date.now().toString(), from: 'user', to: target ? '@' + target : 'orchestrator', content, type: 'user', timestamp: new Date().toISOString() });
    el.input.value = '';
    await api.sendMessage(st.currentSessionId, content, target);
  }

  function showAgentPanel(show) {
    el.agentPanel.classList.toggle('show', show);
  }

  function updateMentionPopup() {
    const st = store.getState();
    const agents = st.agents || [];
    const filtered = mentionQuery ? agents.filter(a => a.name.toLowerCase().includes(mentionQuery)) : agents;
    if (!filtered.length) {
      el.mentionPopup.classList.remove('show');
      return;
    }
    el.mentionPopup.innerHTML = filtered.map((a, idx) => `
      <div class="mention-item ${idx === mentionIndex ? 'selected' : ''}" data-name="${a.name}">@${a.name}</div>
    `).join('');
    el.mentionPopup.classList.add('show');
    el.mentionPopup.querySelectorAll('.mention-item').forEach((node) => {
      node.addEventListener('click', () => insertMention(node.dataset.name));
    });
  }

  function insertMention(name) {
    const v = el.input.value;
    const cursor = el.input.selectionStart;
    const before = v.slice(0, cursor).replace(/@(\w*)$/, '@' + name + ' ');
    const after = v.slice(cursor);
    el.input.value = before + after;
    const pos = before.length;
    el.input.setSelectionRange(pos, pos);
    el.mentionPopup.classList.remove('show');
    mentionQuery = '';
    mentionIndex = 0;
    el.input.focus();
  }

  // wiring
  store.subscribe(render);

  el.newBtn.addEventListener('click', () => createSession().catch(console.error));
  el.sendBtn.addEventListener('click', () => send().catch(console.error));
  el.toggleAgentsBtn.addEventListener('click', () => showAgentPanel(true));
  el.closeAgentsBtn.addEventListener('click', () => showAgentPanel(false));

  el.input.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      send().catch(console.error);
    }
  });

  el.input.addEventListener('input', (e) => {
    const value = e.target.value;
    const pos = e.target.selectionStart;
    const before = value.slice(0, pos);
    const m = before.match(/@(\w*)$/);
    if (m) {
      mentionQuery = (m[1] || '').toLowerCase();
      mentionIndex = 0;
      updateMentionPopup();
    } else {
      el.mentionPopup.classList.remove('show');
    }
  });

  document.addEventListener('click', (e) => {
    if (e.target !== el.input && !el.mentionPopup.contains(e.target)) {
      el.mentionPopup.classList.remove('show');
    }
  });

  loadInitial().catch(err => {
    store.setError(err.message || String(err));
  });
})();

