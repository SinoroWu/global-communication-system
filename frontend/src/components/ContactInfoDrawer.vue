<script setup lang="ts">
import { ref, watch } from 'vue';
import { useChatStore } from '../stores/chatStore';

const props = defineProps<{
  isOpen: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const chatStore = useChatStore();

const notes = ref('');
const tags = ref<string[]>([]);
const newTagInput = ref('');
const isSaving = ref(false);

watch(
  () => chatStore.activeConversation,
  (conv) => {
    if (conv) {
      notes.value = conv.notes || '';
      tags.value = [...(conv.tags || [])];
    }
  },
  { immediate: true }
);

function addTag() {
  const val = newTagInput.value.trim().replace(/^#/, '');
  if (val && !tags.value.includes(val)) {
    tags.value.push(val);
    newTagInput.value = '';
    saveChanges();
  }
}

function removeTag(tag: string) {
  tags.value = tags.value.filter((t) => t !== tag);
  saveChanges();
}

async function saveChanges() {
  isSaving.value = true;
  try {
    await chatStore.updateCustomerNotes(notes.value, tags.value);
  } finally {
    isSaving.value = false;
  }
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text);
  alert('已複製外部 ID 到剪貼簿！');
}
</script>

<template>
  <div 
    v-if="isOpen && chatStore.activeConversation" 
    class="contact-info-drawer bg-body border-start d-flex flex-column h-100 shadow-lg p-3"
  >
    <!-- Drawer Header -->
    <div class="d-flex align-items-center justify-content-between pb-3 border-bottom mb-3">
      <h6 class="mb-0 fw-bold">
        <i class="bi bi-person-badge text-primary me-2"></i>客戶資訊卡
      </h6>
      <button class="btn btn-sm btn-light border-0" @click="emit('close')">
        <i class="bi bi-x-lg"></i>
      </button>
    </div>

    <div class="overflow-y-auto flex-grow-1">
      <!-- Profile Card -->
      <div class="text-center py-2 mb-3">
        <img 
          :src="chatStore.activeConversation.avatarUrl || 'https://api.dicebear.com/7.x/bottts/svg?seed=' + chatStore.activeConversation.externalUserId" 
          class="rounded-circle border mb-2 shadow-sm"
          width="64" 
          height="64" 
          alt="Avatar"
        />
        <h6 class="fw-bold mb-1">{{ chatStore.activeConversation.displayName }}</h6>
        <span class="badge bg-secondary-subtle text-secondary border">
          {{ chatStore.activeConversation.platform.toUpperCase() }} 客戶
        </span>
      </div>

      <!-- Detail Items -->
      <div class="card bg-body-tertiary border-0 rounded-3 p-3 mb-3">
        <div class="mb-2">
          <label class="small text-secondary mb-1">渠道外部識別碼 (External ID)</label>
          <div class="d-flex align-items-center justify-content-between">
            <span class="small font-monospace text-truncate" style="max-width: 170px;">
              {{ chatStore.activeConversation.externalUserId }}
            </span>
            <button class="btn btn-sm btn-link p-0 text-decoration-none" @click="copyToClipboard(chatStore.activeConversation.externalUserId)">
              <i class="bi bi-copy"></i>
            </button>
          </div>
        </div>

        <div>
          <label class="small text-secondary mb-1">初次建檔時間</label>
          <div class="small">
            {{ new Date(chatStore.activeConversation.createdAt).toLocaleString() }}
          </div>
        </div>
      </div>

      <!-- Tag Management -->
      <div class="mb-4">
        <label class="form-label small fw-semibold">客戶標籤 (Tags)</label>
        <div class="d-flex flex-wrap gap-1 mb-2">
          <span 
            v-for="tag in tags" 
            :key="tag" 
            class="badge bg-primary-subtle text-primary border border-primary-subtle d-flex align-items-center gap-1 py-1 px-2"
          >
            #{{ tag }}
            <button class="btn-close btn-close-xs" style="font-size: 0.5rem;" @click="removeTag(tag)"></button>
          </span>
          <span v-if="tags.length === 0" class="text-secondary small">尚無標籤</span>
        </div>

        <div class="input-group input-group-sm">
          <input 
            type="text" 
            class="form-control" 
            placeholder="新增標籤 (如: VIP, 待跟進)" 
            v-model="newTagInput"
            @keydown.enter.prevent="addTag"
          />
          <button class="btn btn-outline-primary" @click="addTag">新增</button>
        </div>
      </div>

      <!-- Customer Notes -->
      <div class="mb-3">
        <div class="d-flex align-items-center justify-content-between mb-1">
          <label class="form-label small fw-semibold mb-0">客服備忘筆記</label>
          <span v-if="isSaving" class="badge text-success small">儲存中...</span>
        </div>
        <textarea 
          class="form-control form-control-sm" 
          rows="4" 
          placeholder="記錄客戶喜好、特殊需求或跟進進度..."
          v-model="notes"
          @blur="saveChanges"
        ></textarea>
        <button class="btn btn-sm btn-primary w-100 mt-2" @click="saveChanges" :disabled="isSaving">
          儲存備忘筆記
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.contact-info-drawer {
  width: 280px;
  min-width: 280px;
  animation: slideIn 0.2s ease-out;
}

@keyframes slideIn {
  from {
    transform: translateX(100%);
  }
  to {
    transform: translateX(0);
  }
}
</style>
