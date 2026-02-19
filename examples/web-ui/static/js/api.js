const API_BASE = '';

export async function fetchAgents() {
    const response = await fetch(`${API_BASE}/agents`);
    if (!response.ok) {
        throw new Error(`Failed to fetch agents: ${response.statusText}`);
    }
    return response.json();
}

export async function fetchGroups() {
    const response = await fetch(`${API_BASE}/groups`);
    if (!response.ok) {
        throw new Error(`Failed to fetch groups: ${response.statusText}`);
    }
    return response.json();
}

export async function fetchGroup(id) {
    const response = await fetch(`${API_BASE}/groups/${id}`);
    if (!response.ok) {
        if (response.status === 404) {
            return null;
        }
        throw new Error(`Failed to fetch group: ${response.statusText}`);
    }
    return response.json();
}

export async function deleteGroup(id) {
    const response = await fetch(`${API_BASE}/groups/${id}`, {
        method: 'DELETE'
    });
    if (!response.ok) {
        throw new Error(`Failed to delete group: ${response.statusText}`);
    }
    return response.json();
}

export async function sendMessage(sessionId, message, mention) {
    const body = {
        session_id: sessionId || '',
        message: message
    };
    if (mention) {
        body.mention = mention;
    }

    const response = await fetch(`${API_BASE}/chat`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(body)
    });

    if (!response.ok) {
        throw new Error(`Failed to send message: ${response.statusText}`);
    }
    return response.json();
}

export async function sendMessageStream(sessionId, message, mention, onEvent) {
    const body = {
        session_id: sessionId || '',
        message: message
    };
    if (mention) {
        body.mention = mention;
    }

    const response = await fetch(`${API_BASE}/chat/stream`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Accept': 'text/event-stream'
        },
        body: JSON.stringify(body)
    });

    if (!response.ok) {
        throw new Error(`Failed to send message: ${response.statusText}`);
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';

    while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
            const trimmed = line.trim();
            if (!trimmed || !trimmed.startsWith('data: ')) continue;

            const data = trimmed.slice(6);
            if (data === '[DONE]') {
                return;
            }

            try {
                const event = JSON.parse(data);
                if (onEvent) {
                    onEvent(event);
                }
            } catch (e) {
                console.warn('Failed to parse SSE event:', data);
            }
        }
    }
}
