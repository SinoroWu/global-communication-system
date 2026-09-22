<script setup lang="ts">
import { ref, nextTick } from 'vue';
import { useChatStore } from '../stores/chatStore';
import type { ContentType } from '../types';

const chatStore = useChatStore();

const inputText = ref('');
const mediaUrl = ref('');
const showMediaInput = ref(false);
const textareaRef = ref<HTMLTextAreaElement | null>(null);

const cannedResponses = [
  '您好！很高興為您服務，請問有什麼我可以協助您的？',
  '感謝您的耐心等候，我們已為您查詢相關資訊。',
  '這部分已為您建立處理工單，專員將於稍後與您聯繫確認。',
  '若有任何其他問題，歡迎隨時透過此對話詢問，祝您有美好的一天！',
  '目前系統連線與服務一切正常，感謝您的支持！',
];

function applyCannedResponse(text: string) {
  inputText.value = text;
  nextTick(() => {
    textareaRef.value?.focus();
  });
}

async function handleSend() {
  const content = inputText.value.trim();
  if (!content && !mediaUrl.value.trim()) return;

  const cType: ContentType = mediaUrl.value ? 'image' : 'text';
  const mUrl = mediaUrl.value.trim() || undefined;

  try {
    await chatStore.sendMessage(content || '[圖片]', cType, mUrl);
    inputText.value = '';
    mediaUrl.value = '';
    showMediaInput.value = false;
  } catch (err) {
    alert('訊息發送失敗，請檢查後端連線！');
  }
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    handleSend();
  }
}
</script>

<template>
  <div class="message-composer p-3 border-top bg-body">
    <!-- Quick Replies & Media Bar -->
    <div class="d-flex align-items-center justify-content-between mb-2">
      <!-- Canned Responses Dropdown -->
      <div class="dropdown">
        <button 
          class="btn btn-sm btn-outline-secondary dropdown-toggle rounded-pill px-3" 
          type="button" 
          data-bs-toggle="dropdown"
        >
          <i class="bi bi-lightning-charge-fill text-warning me-1"></i>
          快捷回覆範本
        </button>
        <ul class="dropdown-menu shadow-sm">
          <li v-for="(resp, i) in cannedResponses" :key="i">
            <button class="dropdown-item small py-2" @click="applyCannedResponse(resp)">
              {{ resp }}
            </button>
          </li>
        </ul>
      </div>

      <!-- Attachment Button -->
      <button 
        class="btn btn-sm rounded-pill px-3"
        :class="showMediaInput ? 'btn-primary' : 'btn-outline-secondary'"
        @click="showMediaInput = !showMediaInput"
        title="插入圖片連結"
      >
        <i class="bi bi-image me-1"></i>
        圖片連結
      </button>
    </div>

    <!-- Media URL Input Field -->
    <div v-if="showMediaInput" class="mb-2">
      <div class="input-group input-group-sm">
        <span class="input-group-text bg-body-tertiary">
          <i class="bi bi-link-45deg"></i>
        </span>
        <input 
          type="url" 
          class="form-control" 
          placeholder="請輸入圖片 URL (https://...)" 
          v-model="mediaUrl"
        />
        <button class="btn btn-outline-secondary" @click="mediaUrl = ''">清除</button>
      </div>
    </div>

    <!-- Main Text Input Area -->
    <div class="d-flex gap-2 align-items-end">
      <textarea
        ref="textareaRef"
        class="form-control"
        rows="2"
        placeholder="輸入回覆訊息... (按 Enter 發送，Shift+Enter 換行)"
        v-model="inputText"
        @keydown="handleKeydown"
        :disabled="chatStore.sendingMessage || !chatStore.activeConversation"
      ></textarea>

      <button 
        class="btn btn-primary h-100 px-4 d-flex align-items-center justify-content-center shadow-sm"
        @click="handleSend"
        :disabled="chatStore.sendingMessage || (!inputText.trim() && !mediaUrl.trim()) || !chatStore.activeConversation"
      >
        <span v-if="chatStore.sendingMessage" class="spinner-border spinner-border-sm me-1"></span>
        <i v-else class="bi bi-send-fill fs-5"></i>
      </button>
    </div>
  </div>
</template>

<style scoped>
.message-composer textarea {
  resize: none;
}
</style>
