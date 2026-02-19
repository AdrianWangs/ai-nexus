class APIClient {
  async request(method, path, data) {
    const init = { method, headers: { 'Content-Type': 'application/json' } };
    if (data !== undefined) init.body = JSON.stringify(data);
    const resp = await fetch(path, init);
    if (!resp.ok) {
      const text = await resp.text().catch(() => resp.statusText);
      throw new Error(text || resp.statusText);
    }
    if (resp.status === 204) return null;
    return resp.json();
  }

  getAgents() { return this.request('GET', '/api/agents'); }
  getSessions() { return this.request('GET', '/api/sessions'); }
  createSession(title = 'New Chat') { return this.request('POST', '/api/sessions', { title }); }
  getMessages(sessionId) { return this.request('GET', `/api/sessions/${sessionId}/messages`); }

  sendMessage(sessionId, message, target) {
    return this.request('POST', '/api/chat', { session_id: sessionId, message, target });
  }

  createSSE(sessionId, onEvent) {
    const es = new EventSource(`/api/sessions/${sessionId}/stream`);
    es.addEventListener('message_start', (e) => onEvent({ type: 'message_start', ...JSON.parse(e.data) }));
    es.addEventListener('content_delta', (e) => onEvent({ type: 'content_delta', ...JSON.parse(e.data) }));
    es.addEventListener('message_end', (e) => onEvent({ type: 'message_end', ...JSON.parse(e.data) }));
    es.addEventListener('tool_call', (e) => onEvent({ type: 'tool_call', ...JSON.parse(e.data) }));
    es.addEventListener('tool_result', (e) => onEvent({ type: 'tool_result', ...JSON.parse(e.data) }));
    es.addEventListener('error', (e) => {
      const data = e?.data ? JSON.parse(e.data) : { error: 'SSE error' };
      onEvent({ type: 'error', ...data });
    });
    es.addEventListener('done', () => onEvent({ type: 'done' }));
    return es;
  }
}

const api = new APIClient();

