const state = {
    agents: [],
    groups: [],
    currentGroup: null,
    isLoading: false,
    statusItems: []
};

const listeners = new Set();

function notify() {
    listeners.forEach(listener => {
        try {
            listener(state);
        } catch (e) {
            console.error('Store listener error:', e);
        }
    });
}

export function subscribe(listener) {
    listeners.add(listener);
    return () => listeners.delete(listener);
}

export function getState() {
    return state;
}

export function setAgents(agents) {
    state.agents = agents || [];
    notify();
}

export function setGroups(groups) {
    state.groups = groups || [];
    notify();
}

export function setCurrentGroup(group) {
    state.currentGroup = group;
    notify();
}

export function setLoading(loading) {
    state.isLoading = loading;
    notify();
}

export function addMessage(message) {
    if (!state.currentGroup) return;

    if (!state.currentGroup.messages) {
        state.currentGroup.messages = [];
    }

    if (!message.id) {
        message.id = 'msg_' + Date.now() + '_' + Math.random().toString(36).substr(2, 6);
    }

    const exists = state.currentGroup.messages.some(m => m.id && m.id === message.id);
    if (!exists) {
        state.currentGroup.messages.push(message);
    }

    if (message.from && message.from !== 'user') {
        const agentName = message.from.startsWith('@') ? message.from.slice(1) : message.from;
        if (!state.currentGroup.participants) {
            state.currentGroup.participants = [];
        }
        if (!state.currentGroup.participants.includes(agentName)) {
            state.currentGroup.participants.push(agentName);
        }
    }

    notify();
}

export function updateStatus(statusItem) {
    const existingIndex = state.statusItems.findIndex(
        item => item.type === statusItem.type && item.agent === statusItem.agent
    );

    if (existingIndex >= 0) {
        state.statusItems[existingIndex] = { ...state.statusItems[existingIndex], ...statusItem };
    } else {
        state.statusItems.push(statusItem);
    }

    notify();
}

export function clearStatus() {
    state.statusItems = [];
    notify();
}

export function removeGroup(groupId) {
    state.groups = state.groups.filter(g => g.id !== groupId);
    if (state.currentGroup && state.currentGroup.id === groupId) {
        state.currentGroup = null;
    }
    notify();
}

export function addGroup(group) {
    const exists = state.groups.some(g => g.id === group.id);
    if (!exists) {
        state.groups.unshift(group);
    }
    notify();
}

export function updateGroup(group) {
    const index = state.groups.findIndex(g => g.id === group.id);
    if (index >= 0) {
        state.groups[index] = group;
    } else {
        state.groups.unshift(group);
    }
    if (state.currentGroup && state.currentGroup.id === group.id) {
        state.currentGroup = group;
    }
    notify();
}

export const store = {
    subscribe,
    getState,
    setAgents,
    setGroups,
    setCurrentGroup,
    setLoading,
    addMessage,
    updateStatus,
    clearStatus,
    removeGroup,
    addGroup,
    updateGroup
};
