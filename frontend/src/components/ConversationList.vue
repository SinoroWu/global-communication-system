<script setup lang="ts">
import { useChatStore } from '../stores/chatStore';
import type { Platform } from '../types';

const chatStore = useChatStore();

const emit = defineEmits<{
  (e: 'open-simulator'): void;
}>();

function formatTime(timestamp: number): string {
  if (!timestamp) return '';
  const date = new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMin = Math.floor(diffMs / (1000 * 60));

  if (diffMin < 1) return '剛剛';
  if (diffMin < 60) return `${diffMin} 分鐘前`;

  const isToday = date.toDateString() === now.toDateString();
  if (isToday) {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }

  const yesterday = new Date(now);
  yesterday.setDate(yesterday.getDate() - 1);
  if (date.toDateString() === yesterday.toDateString()) {
    return '昨天 ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }

  return `${date.getMonth() + 1}/${date.getDate()}`;
}

function getPlatformIcon(p: Platform) {
  switch (p) {
    case 'line': return 'bi-line text-success';
    case 'instagram': return 'bi-instagram text-danger';
    case 'facebook': return 'bi-messenger text-primary';
    case 'douyin': return 'bi-tiktok text-dark';
    case 'telegram': return 'bi-telegram text-info';
    default: return 'bi-chat-dots';
  }
}

function getPlatformBadgeBg(p: Platform) {
  switch (p) {
    case 'line': return 'bg-success';
    case 'instagram': return 'bg-danger';
    case 'facebook': return 'bg-primary';
    case 'douyin': return 'bg-dark';
    case 'telegram': return 'bg-info';
    default: return 'bg-secondary';
  }
}
</script>

<template>
  <div class="conversation-list d-flex flex-column h-100 border-end bg-body">
    <!-- Header with Search -->
    <div class="p-3 border-bottom">
      <div class="d-flex align-items-center justify-content-between mb-2">
        <h6 class="mb-0 fw-bold">
          {{ chatStore.activePlatformFilter === 'all' ? '全部對話' : chatStore.activePlatformFilter.toUpperCase() + ' 訊息' }}
          <span class="text-secondary small fw-normal ms-1">({{ chatStore.filteredConversations.length }})</span>
        </h6>
        <button 
          class="btn btn-sm btn-light border rounded-pill px-2 py-0"
          @click="chatStore.refreshConversations()"
          title="重新整理"
        >
          <i class="bi bi-arrow-clockwise"></i>
        </button>
      </div>

      <!-- Search Box -->
      <div class="input-group input-group-sm">
        <span class="input-group-text bg-body-tertiary border-end-0">
          <i class="bi bi-search text-secondary"></i>
        </span>
        <input 
          type="text" 
          class="form-control bg-body-tertiary border-start-0" 
          placeholder="搜尋聯絡人或訊息..."
          v-model="chatStore.searchQuery"
        />
        <button 
          v-if="chatStore.searchQuery" 
          class="btn btn-outline-secondary border-start-0"
          @click="chatStore.searchQuery = ''"
        >
          <i class="bi bi-x"></i>
        </button>
      </div>
    </div>

    <!-- Scrollable Conversation List -->
    <div class="flex-grow-1 overflow-y-auto">
      <div v-if="chatStore.filteredConversations.length === 0" class="text-center py-5 px-3">
        <div class="mb-3 text-secondary">
          <i class="bi bi-chat-square-text fs-1 opacity-50"></i>
        </div>
        <p class="text-secondary mb-2 small">尚無符合條件的對話</p>
        <button 
          class="btn btn-sm btn-outline-primary rounded-pill px-3"
          @click="emit('open-simulator')"
        >
          <i class="bi bi-play-circle me-1"></i>
          觸發模擬測試訊息
        </button>
      </div>

      <div 
        v-for="conv in chatStore.filteredConversations" 
        :key="conv.id"
        class="conversation-item d-flex gap-3 p-3 border-bottom cursor-pointer transition-colors"
        :class="{ 'active-conversation': chatStore.activeConversationId === conv.id }"
        @click="chatStore.selectConversation(conv.id)"
      >
        <!-- Avatar with Platform Badge -->
        <div class="position-relative flex-shrink-0">
          <img 
            :src="conv.avatarUrl || 'https://api.dicebear.com/7.x/bottts/svg?seed=' + conv.externalUserId" 
            class="rounded-circle border bg-body-tertiary"
            width="46" 
            height="46" 
            alt="avatar"
          />
          <span 
            class="position-absolute bottom-0 end-0 badge p-1 rounded-circle border border-2 border-white shadow-sm"
            :class="getPlatformBadgeBg(conv.platform)"
            style="font-size: 0.65rem;"
          >
            <i class="bi" :class="getPlatformIcon(conv.platform).replace('text-success', 'text-white').replace('text-danger', 'text-white').replace('text-primary', 'text-white').replace('text-dark', 'text-white').replace('text-info', 'text-white')"></i>
          </span>
        </div>

        <!-- Meta Info -->
        <div class="flex-grow-1 min-w-0">
          <div class="d-flex align-items-center justify-content-between mb-1">
            <span class="fw-semibold text-truncate d-inline-block" style="max-width: 140px;">
              {{ conv.displayName }}
            </span>
            <span class="small text-secondary" style="font-size: 0.75rem;">
              {{ formatTime(conv.lastMessageAt) }}
            </span>
          </div>

          <div class="d-flex align-items-center justify-content-between">
            <p 
              class="small text-truncate mb-0" 
              :class="conv.unreadCount > 0 ? 'fw-bold text-body' : 'text-secondary'"
              style="max-width: 170px;"
            >
              {{ conv.lastMessage || '尚無訊息' }}
            </p>

            <span 
              v-if="conv.unreadCount > 0" 
              class="badge rounded-pill bg-danger shadow-sm ms-2"
              style="font-size: 0.7rem;"
            >
              {{ conv.unreadCount }}
            </span>
          </div>

          <!-- Tags -->
          <div v-if="conv.tags && conv.tags.length > 0" class="d-flex gap-1 mt-1 flex-wrap">
            <span 
              v-for="tag in conv.tags" 
              :key="tag" 
              class="badge bg-secondary-subtle text-secondary border border-secondary-subtle px-1 py-0"
              style="font-size: 0.65rem;"
            >
              #{{ tag }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.conversation-list {
  width: 320px;
  min-width: 320px;
}

.cursor-pointer {
  cursor: pointer;
}

.conversation-item:hover {
  background-color: var(--bs-tertiary-bg);
}

.active-conversation {
  background-color: var(--bs-primary-bg-subtle) !important;
  border-left: 3px solid var(--bs-primary);
}

.transition-colors {
  transition: background-color 0.15s ease-in-out;
}
</style>
