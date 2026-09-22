<script setup lang="ts">
import { ref, watch, nextTick, onMounted } from 'vue';
import { useChatStore } from '../stores/chatStore';
import MessageComposer from './MessageComposer.vue';
import type { Platform } from '../types';

const chatStore = useChatStore();
const messagesContainerRef = ref<HTMLDivElement | null>(null);

const emit = defineEmits<{
  (e: 'toggle-drawer'): void;
}>();

function scrollToBottom(smooth = true) {
  nextTick(() => {
    if (messagesContainerRef.value) {
      messagesContainerRef.value.scrollTo({
        top: messagesContainerRef.value.scrollHeight,
        behavior: smooth ? 'smooth' : 'auto',
      });
    }
  });
}

// Watch messages length to auto scroll
watch(
  () => chatStore.activeMessages.length,
  () => {
    scrollToBottom(true);
  }
);

watch(
  () => chatStore.activeConversationId,
  () => {
    scrollToBottom(false);
  }
);

onMounted(() => {
  scrollToBottom(false);
});

function formatMessageTime(timestamp: number): string {
  if (!timestamp) return '';
  const d = new Date(timestamp);
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

function getPlatformBadge(platform: Platform) {
  switch (platform) {
    case 'line': return { name: 'LINE', class: 'bg-success' };
    case 'instagram': return { name: 'Instagram', class: 'bg-danger' };
    case 'facebook': return { name: 'Messenger', class: 'bg-primary' };
    case 'douyin': return { name: '抖音', class: 'bg-dark' };
    case 'telegram': return { name: 'Telegram', class: 'bg-info' };
    default: return { name: '社群', class: 'bg-secondary' };
  }
}
</script>

<template>
  <div class="chat-window d-flex flex-column h-100 bg-body-tertiary flex-grow-1">
    <!-- Conversation Active Header -->
    <div v-if="chatStore.activeConversation" class="chat-header px-4 py-3 border-bottom bg-body d-flex align-items-center justify-content-between shadow-sm">
      <div class="d-flex align-items-center gap-3">
        <div class="position-relative">
          <img 
            :src="chatStore.activeConversation.avatarUrl || 'https://api.dicebear.com/7.x/bottts/svg?seed=' + chatStore.activeConversation.externalUserId" 
            class="rounded-circle border"
            width="44" 
            height="44" 
            alt="avatar"
          />
        </div>
        <div>
          <div class="d-flex align-items-center gap-2">
            <h6 class="mb-0 fw-bold">{{ chatStore.activeConversation.displayName }}</h6>
            <span class="badge rounded-pill" :class="getPlatformBadge(chatStore.activeConversation.platform).class">
              {{ getPlatformBadge(chatStore.activeConversation.platform).name }}
            </span>
          </div>
          <span class="text-secondary small font-monospace">
            ID: {{ chatStore.activeConversation.externalUserId }}
          </span>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="d-flex align-items-center gap-2">
        <button 
          class="btn btn-outline-secondary btn-sm rounded-pill px-3"
          @click="emit('toggle-drawer')"
          title="查看或編輯客戶資料"
        >
          <i class="bi bi-person-lines-fill me-1"></i>
          客戶資料
        </button>
      </div>
    </div>

    <!-- Empty State (No conversation selected) -->
    <div v-if="!chatStore.activeConversation" class="flex-grow-1 d-flex flex-column align-items-center justify-content-center p-4 text-center">
      <div class="mb-3 text-secondary opacity-25">
        <i class="bi bi-chat-square-dots" style="font-size: 5rem;"></i>
      </div>
      <h5 class="fw-bold mb-2">歡迎使用 全渠道通訊整合系統</h5>
      <p class="text-secondary small max-w-sm mb-4">
        請在左側選擇一個對話以開始回覆，或透過上方模擬器產生測試訊息。
      </p>
    </div>

    <!-- Messages Timeline -->
    <div 
      v-else 
      ref="messagesContainerRef" 
      class="chat-messages flex-grow-1 p-4 overflow-y-auto d-flex flex-column gap-3"
    >
      <div 
        v-for="(msg, idx) in chatStore.activeMessages" 
        :key="msg.id || idx"
        class="message-row d-flex gap-2"
        :class="msg.senderRole === 'agent' ? 'justify-content-end' : 'justify-content-start'"
      >
        <!-- User Avatar -->
        <img 
          v-if="msg.senderRole !== 'agent'" 
          :src="msg.avatar || 'https://api.dicebear.com/7.x/bottts/svg?seed=' + msg.senderId" 
          class="rounded-circle border mt-1 align-self-end flex-shrink-0"
          width="32" 
          height="32" 
          alt="avatar"
        />

        <!-- Bubble Container -->
        <div 
          class="message-bubble-wrapper d-flex flex-column"
          :class="msg.senderRole === 'agent' ? 'align-items-end' : 'align-items-start'"
        >
          <!-- Sender Name -->
          <span class="small text-secondary mb-1 px-1" style="font-size: 0.72rem;">
            {{ msg.senderName }}
          </span>

          <!-- Message Bubble -->
          <div 
            class="message-bubble p-3 rounded-4 shadow-sm"
            :class="msg.senderRole === 'agent' ? 'bg-primary text-white agent-bubble' : 'bg-body border user-bubble'"
          >
            <!-- Media Attachment -->
            <div v-if="msg.mediaUrl || msg.contentType === 'image'" class="mb-2">
              <img 
                :src="msg.mediaUrl || 'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=400'" 
                class="rounded-3 img-fluid border" 
                style="max-height: 220px; object-fit: cover;"
                alt="Attachment"
              />
            </div>

            <!-- Text Content -->
            <p class="mb-0 text-break" style="white-space: pre-wrap; font-size: 0.925rem;">
              {{ msg.content }}
            </p>
          </div>

          <!-- Timestamp & Status -->
          <div class="d-flex align-items-center gap-1 mt-1 px-1 text-secondary small" style="font-size: 0.7rem;">
            <span>{{ formatMessageTime(msg.timestamp) }}</span>
            <i v-if="msg.senderRole === 'agent'" class="bi bi-check2-all text-primary"></i>
          </div>
        </div>

        <!-- Agent Avatar -->
        <img 
          v-if="msg.senderRole === 'agent'" 
          :src="msg.avatar || 'https://api.dicebear.com/7.x/bottts/svg?seed=agent'" 
          class="rounded-circle border mt-1 align-self-end flex-shrink-0"
          width="32" 
          height="32" 
          alt="avatar"
        />
      </div>
    </div>

    <!-- Message Composer -->
    <MessageComposer v-if="chatStore.activeConversation" />
  </div>
</template>

<style scoped>
.chat-window {
  min-width: 0;
}

.max-w-sm {
  max-width: 380px;
}

.message-bubble-wrapper {
  max-width: 75%;
}

@media (min-width: 992px) {
  .message-bubble-wrapper {
    max-width: 60%;
  }
}

.agent-bubble {
  border-bottom-right-radius: 4px !important;
}

.user-bubble {
  border-bottom-left-radius: 4px !important;
}
</style>
