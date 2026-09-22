<script setup lang="ts">
import { useChatStore } from '../stores/chatStore';
import type { Platform } from '../types';

const chatStore = useChatStore();

const emit = defineEmits<{
  (e: 'open-simulator'): void;
  (e: 'open-settings'): void;
  (e: 'toggle-theme'): void;
}>();

const channels = [
  { id: 'all', name: '全部訊息', icon: 'bi-chat-dots-fill', color: 'text-primary', badgeClass: 'bg-primary' },
  { id: 'line', name: 'LINE', icon: 'bi-line', color: 'color-line', badgeClass: 'bg-line' },
  { id: 'instagram', name: 'Instagram', icon: 'bi-instagram', color: 'color-ig', badgeClass: 'bg-ig' },
  { id: 'facebook', name: 'Messenger', icon: 'bi-messenger', color: 'color-fb', badgeClass: 'bg-fb' },
  { id: 'douyin', name: '抖音 / TikTok', icon: 'bi-tiktok', color: 'color-douyin', badgeClass: 'bg-douyin' },
  { id: 'telegram', name: 'Telegram', icon: 'bi-telegram', color: 'color-tg', badgeClass: 'bg-tg' },
];

function selectChannel(id: string) {
  chatStore.activePlatformFilter = id as Platform | 'all';
}
</script>

<template>
  <div class="channel-sidebar d-flex flex-column bg-body-tertiary border-end py-3 px-2">
    <!-- Brand / Logo -->
    <div class="d-flex align-items-center gap-2 px-2 mb-4">
      <div class="brand-icon rounded-3 bg-gradient d-flex align-items-center justify-content-center text-white shadow-sm">
        <i class="bi bi-globe-americas fs-4"></i>
      </div>
      <div class="d-none d-lg-block">
        <h6 class="mb-0 fw-bold tracking-tight">GlobalComm</h6>
        <span class="badge bg-success-subtle text-success border border-success-subtle py-0 px-1" style="font-size: 0.65rem;">
          <i class="bi bi-circle-fill me-1" style="font-size: 0.45rem;"></i>
          {{ chatStore.wsConnected ? '極速在線' : '重新連線中' }}
        </span>
      </div>
    </div>

    <!-- Navigation List -->
    <ul class="nav nav-pills flex-column gap-1 mb-auto">
      <li v-for="ch in channels" :key="ch.id" class="nav-item">
        <button 
          class="nav-link w-100 d-flex align-items-center justify-content-between text-start px-3 py-2 rounded-3 transition-all"
          :class="{ active: chatStore.activePlatformFilter === ch.id }"
          @click="selectChannel(ch.id)"
          :title="ch.name"
        >
          <div class="d-flex align-items-center gap-2">
            <i class="bi fs-5" :class="[ch.icon, ch.color]"></i>
            <span class="d-none d-lg-inline fw-medium fs-7">{{ ch.name }}</span>
          </div>

          <!-- Unread Badge -->
          <span 
            v-if="ch.id === 'all' && chatStore.totalUnreadCount > 0"
            class="badge rounded-pill bg-danger"
          >
            {{ chatStore.totalUnreadCount > 99 ? '99+' : chatStore.totalUnreadCount }}
          </span>
          <span 
            v-else-if="ch.id !== 'all' && (chatStore.platformUnreadMap[ch.id] || 0) > 0"
            class="badge rounded-pill bg-danger"
          >
            {{ chatStore.platformUnreadMap[ch.id] }}
          </span>
        </button>
      </li>
    </ul>

    <!-- Bottom Actions -->
    <div class="border-top pt-3 d-flex flex-column gap-2">
      <!-- Simulator Trigger Button -->
      <button 
        class="btn btn-outline-warning w-100 d-flex align-items-center justify-content-center justify-content-lg-start gap-2 py-2 rounded-3 shadow-sm"
        @click="emit('open-simulator')"
        title="開啟 Webhook 測試模擬器"
      >
        <i class="bi bi-cpu fs-5"></i>
        <span class="d-none d-lg-inline small fw-semibold">訊息模擬器</span>
      </button>

      <!-- Channel Settings Button -->
      <button 
        class="btn btn-outline-secondary w-100 d-flex align-items-center justify-content-center justify-content-lg-start gap-2 py-2 rounded-3"
        @click="emit('open-settings')"
        title="渠道 API 與 Webhook 設定"
      >
        <i class="bi bi-gear fs-5"></i>
        <span class="d-none d-lg-inline small fw-semibold">渠道設定</span>
      </button>

      <!-- Theme Switcher -->
      <button 
        class="btn btn-outline-light text-body-secondary w-100 d-flex align-items-center justify-content-center justify-content-lg-start gap-2 py-2 rounded-3"
        @click="emit('toggle-theme')"
        title="切換主題"
      >
        <i class="bi bi-moon-stars fs-5"></i>
        <span class="d-none d-lg-inline small">切換主題</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.channel-sidebar {
  width: 72px;
  min-width: 72px;
  transition: width 0.2s ease-in-out;
}

@media (min-width: 992px) {
  .channel-sidebar {
    width: 210px;
    min-width: 210px;
  }
}

.brand-icon {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #0d6efd, #00d2ff);
}

.fs-7 {
  font-size: 0.875rem;
}

.nav-link {
  color: var(--bs-body-color);
}

.nav-link:hover {
  background-color: var(--bs-secondary-bg);
}

.nav-link.active {
  background-color: var(--bs-primary);
  color: #fff;
}

.nav-link.active i {
  color: #fff !important;
}

/* Platform Brand Colors */
.color-line { color: #06C755; }
.color-ig { color: #E1306C; }
.color-fb { color: #0084FF; }
.color-douyin { color: #fe2c55; }
.color-tg { color: #229ED9; }

.bg-line { background-color: #06C755; }
.bg-ig { background: linear-gradient(45deg, #f09433, #e6683c, #dc2743, #cc2366, #bc1888); }
.bg-fb { background-color: #0084FF; }
.bg-douyin { background-color: #161823; }
.bg-tg { background-color: #229ED9; }
</style>
