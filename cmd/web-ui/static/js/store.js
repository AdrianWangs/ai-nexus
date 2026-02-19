class Store {
  constructor() {
    this.state = {
      sessions: [],
      currentSessionId: null,
      messages: [],
      agents: [],
      isStreaming: false,
      streamingMessage: null,
      error: null,
    };
    this.listeners = new Set();
  }

  getState() { return { ...this.state }; }
  subscribe(fn) { this.listeners.add(fn); return () => this.listeners.delete(fn); }
  setState(upd) { this.state = { ...this.state, ...upd }; this.listeners.forEach(fn => fn(this.getState())); }

  setAgents(agents) { this.setState({ agents }); }
  setSessions(sessions) { this.setState({ sessions }); }
  setCurrentSession(sessionId) { this.setState({ currentSessionId: sessionId, messages: [] }); }
  setMessages(messages) { this.setState({ messages }); }
  addMessage(msg) { this.setState({ messages: [...this.state.messages, msg] }); }

  startStreaming(from) {
    this.setState({
      isStreaming: true,
      streamingMessage: { id: Date.now().toString(), from, content: '' },
    });
  }

  appendDelta(delta) {
    if (!this.state.streamingMessage) return;
    this.setState({
      streamingMessage: { ...this.state.streamingMessage, content: this.state.streamingMessage.content + (delta || '') },
    });
  }

  finishStreaming() {
    if (!this.state.streamingMessage) {
      this.setState({ isStreaming: false, streamingMessage: null });
      return;
    }
    const m = this.state.streamingMessage;
    this.setState({
      messages: [...this.state.messages, { id: m.id, from: m.from, to: 'user', content: m.content, type: 'reply', timestamp: new Date().toISOString() }],
      isStreaming: false,
      streamingMessage: null,
    });
  }

  setError(err) { this.setState({ error: err }); }
}

const store = new Store();

