import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { 
  Conversation, 
  Message, 
  Platform, 
  ChannelConfig, 
  SystemStats,
  SimulatorPayload,
  ContentType 
} from '../types';
import { apiService } from '../services/api';
import { wsService } from '../services/websocket';

export const useChatStore = defineStore('chat', () => {
  const conversations = ref<Conversation[]>([]);
  const activeConversationId = ref<string | null>(null);
  const activePlatformFilter = ref<Platform | 'all'>('all');
  const messages = ref<Record<string, Message[]>>({});
  const channels = ref<ChannelConfig[]>([]);
  const stats = ref<SystemStats | null>(null);
  const wsConnected = ref<boolean>(false);
  const searchQuery = ref<string>('');
  const loading = ref<boolean>(false);
  const sendingMessage = ref<boolean>(false);

  // Audio notification sound using Web Audio API (zero external asset dependency)
  const playNotificationSound = () => {
    try {
      const ctx = new (window.AudioContext || (window as any).webkitAudioContext)();
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      osc.type = 'sine';
      osc.frequency.setValueAtTime(587.33, ctx.currentTime); // D5
      osc.frequency.setValueAtTime(880, ctx.currentTime + 0.08); // A5
      gain.gain.setValueAtTime(0.15, ctx.currentTime);
      gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.3);
      osc.connect(gain);
      gain.connect(ctx.destination);
      osc.start();
      osc.stop(ctx.currentTime + 0.3);
    } catch (e) {
      // Audio might be blocked by browser autoplay policy until user gesture
    }
  };

  const activeConversation = computed(() => {
    if (!activeConversationId.value) return null;
    return conversations.value.find((c) => c.id === activeConversationId.value) || null;
  });

  const activeMessages = computed(() => {
    if (!activeConversationId.value) return [];
    return messages.value[activeConversationId.value] || [];
  });

  const filteredConversations = computed(() => {
    return conversations.value.filter((conv) => {
      const matchPlatform = activePlatformFilter.value === 'all' || conv.platform === activePlatformFilter.value;
      const matchSearch = !searchQuery.value || 
        conv.displayName.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
        conv.lastMessage.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
        conv.externalUserId.toLowerCase().includes(searchQuery.value.toLowerCase());
      return matchPlatform && matchSearch;
    });
  });

  const totalUnreadCount = computed(() => {
    return conversations.value.reduce((sum, c) => sum + (c.unreadCount || 0), 0);
  });

  const platformUnreadMap = computed(() => {
    const map: Record<string, number> = {
      line: 0,
      instagram: 0,
      facebook: 0,
      douyin: 0,
      telegram: 0,
    };
    for (const c of conversations.value) {
      if (c.platform in map) {
        map[c.platform] += c.unreadCount || 0;
      }
    }
    return map;
  });

  // Actions
  async function init() {
    loading.value = true;
    try {
      // Connect WebSocket
      wsService.connect();
      wsService.onStatusChange((status) => {
        wsConnected.value = status;
      });

      // Register real-time handlers
      wsService.on('new_message', (msg: Message) => {
        handleIncomingMessage(msg);
      });

      wsService.on('conversation_updated', (conv: Conversation) => {
        handleConversationUpdated(conv);
      });

      // Load initial data
      await Promise.all([
        refreshConversations(),
        refreshChannels(),
        refreshStats(),
      ]);

      // Automatically select first conversation if available and none selected
      if (conversations.value.length > 0 && !activeConversationId.value) {
        await selectConversation(conversations.value[0].id);
      }
    } catch (err) {
      console.error('[Store] Init error:', err);
    } finally {
      loading.value = false;
    }
  }

  async function refreshConversations() {
    try {
      const list = await apiService.getConversations();
      conversations.value = list;
    } catch (err) {
      console.error('[Store] Fetch conversations error:', err);
    }
  }

  async function refreshChannels() {
    try {
      channels.value = await apiService.getChannels();
    } catch (err) {
      console.error('[Store] Fetch channels error:', err);
    }
  }

  async function refreshStats() {
    try {
      stats.value = await apiService.getStats();
    } catch (err) {
      console.error('[Store] Fetch stats error:', err);
    }
  }

  async function selectConversation(id: string) {
    activeConversationId.value = id;
    const conv = conversations.value.find((c) => c.id === id);

    // Fetch messages for this conversation
    try {
      const msgs = await apiService.getMessages(id);
      messages.value[id] = msgs;

      // If there are unread messages, mark as read
      if (conv && conv.unreadCount > 0) {
        conv.unreadCount = 0;
        await apiService.markAsRead(id);
      }
    } catch (err) {
      console.error('[Store] Load messages error:', err);
    }
  }

  async function sendMessage(content: string, contentType: ContentType = 'text', mediaUrl?: string) {
    if (!activeConversationId.value || !content.trim()) return;

    sendingMessage.value = true;
    try {
      const sent = await apiService.sendMessage(activeConversationId.value, content.trim(), contentType, mediaUrl);
      if (!messages.value[activeConversationId.value]) {
        messages.value[activeConversationId.value] = [];
      }
      messages.value[activeConversationId.value].push(sent);

      // Update local last message preview
      const conv = conversations.value.find((c) => c.id === activeConversationId.value);
      if (conv) {
        conv.lastMessage = content.trim();
        conv.lastMessageAt = sent.timestamp;
      }
    } catch (err) {
      console.error('[Store] Send message error:', err);
      throw err;
    } finally {
      sendingMessage.value = false;
    }
  }

  function handleIncomingMessage(msg: Message) {
    // Append to messages array if loaded
    if (!messages.value[msg.conversationId]) {
      messages.value[msg.conversationId] = [];
    }

    const exists = messages.value[msg.conversationId].some((m) => m.id === msg.id);
    if (!exists) {
      messages.value[msg.conversationId].push(msg);
    }

    // If this message belongs to active conversation, mark as read immediately
    if (activeConversationId.value === msg.conversationId) {
      apiService.markAsRead(msg.conversationId);
    } else {
      // Play chime notification
      playNotificationSound();
    }
  }

  function handleConversationUpdated(updated: Conversation) {
    const idx = conversations.value.findIndex((c) => c.id === updated.id);
    if (idx !== -1) {
      // Keep unread count as 0 if user is currently looking at this conversation
      if (activeConversationId.value === updated.id) {
        updated.unreadCount = 0;
      }
      conversations.value.splice(idx, 1);
      conversations.value.unshift(updated);
    } else {
      conversations.value.unshift(updated);
      if (!activeConversationId.value) {
        selectConversation(updated.id);
      }
    }
  }

  async function updateCustomerNotes(notes: string, tags: string[]) {
    if (!activeConversationId.value) return;
    try {
      const updated = await apiService.updateMetadata(activeConversationId.value, notes, tags);
      const conv = conversations.value.find((c) => c.id === activeConversationId.value);
      if (conv) {
        conv.notes = updated.notes;
        conv.tags = updated.tags;
      }
    } catch (err) {
      console.error('[Store] Update notes error:', err);
    }
  }

  async function triggerSimulator(payload: SimulatorPayload) {
    try {
      const res = await apiService.triggerSimulator(payload);
      return res.message;
    } catch (err) {
      console.error('[Store] Simulator error:', err);
      throw err;
    }
  }

  return {
    conversations,
    activeConversationId,
    activePlatformFilter,
    messages,
    channels,
    stats,
    wsConnected,
    searchQuery,
    loading,
    sendingMessage,
    activeConversation,
    activeMessages,
    filteredConversations,
    totalUnreadCount,
    platformUnreadMap,
    init,
    refreshConversations,
    refreshChannels,
    refreshStats,
    selectConversation,
    sendMessage,
    updateCustomerNotes,
    triggerSimulator,
  };
});
