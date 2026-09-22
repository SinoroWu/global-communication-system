<script setup lang="ts">
import { ref } from 'vue';
import { useChatStore } from '../stores/chatStore';
import type { Platform, ContentType } from '../types';

defineProps<{
  isOpen: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const chatStore = useChatStore();

const selectedPlatform = ref<Platform>('line');
const senderName = ref('');
const senderId = ref('');
const messageContent = ref('您好！想請問這款商品的最新優惠活動？');
const contentType = ref<ContentType>('text');
const mediaUrl = ref('');
const isSubmitting = ref(false);
const successNotice = ref(false);

const presetScenarios: Record<Platform, { name: string; id: string; text: string }> = {
  line: {
    name: 'LINE 顧客 - 王小美',
    id: 'U' + Math.random().toString(36).substring(2, 10),
    text: '您好！我在 LINE 官方帳號看到特惠推播，請問這個方案還有名額嗎？',
  },
  instagram: {
    name: 'IG 粉絲 - alex_visuals',
    id: 'ig_' + Math.random().toString(36).substring(2, 9),
    text: 'Hello! 請問你們今天下午門市有營業嗎？想過去體驗新品！✨',
  },
  facebook: {
    name: 'FB 買家 - 陳先生',
    id: 'fb_' + Math.random().toString(36).substring(2, 9),
    text: '您好，想詢問訂單 #TW-90218 目前的出貨進度，謝謝！',
  },
  douyin: {
    name: '抖音用戶 - 科技小達人',
    id: 'dy_' + Math.random().toString(36).substring(2, 9),
    text: '主播你好！剛在短影片看到的折疊鍵盤真的很輕巧，請問有白色款嗎？',
  },
  telegram: {
    name: 'TG 用戶 - @dev_samuel',
    id: 'tg_' + Math.floor(100000 + Math.random() * 900000),
    text: '嗨！我想了解 Global Communication System 的 API 整合規格文件。',
  },
};

function selectPreset(p: Platform) {
  selectedPlatform.value = p;
  const preset = presetScenarios[p];
  senderName.value = preset.name;
  senderId.value = preset.id;
  messageContent.value = preset.text;
}

// Initialize with LINE preset
selectPreset('line');

async function handleSimulate() {
  if (!messageContent.value.trim() && !mediaUrl.value.trim()) return;

  isSubmitting.value = true;
  successNotice.value = false;

  try {
    await chatStore.triggerSimulator({
      platform: selectedPlatform.value,
      senderId: senderId.value || undefined,
      senderName: senderName.value || undefined,
      content: messageContent.value,
      contentType: contentType.value,
      mediaUrl: mediaUrl.value || undefined,
    });

    successNotice.value = true;
    setTimeout(() => {
      successNotice.value = false;
      emit('close');
    }, 900);
  } catch (err) {
    alert('模擬訊息發送失敗，請確認後端是否運行！');
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <div v-if="isOpen" class="modal-backdrop fade show"></div>

  <div v-if="isOpen" class="modal fade show d-block" tabindex="-1">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content border-0 shadow-lg rounded-4 overflow-hidden">
        <!-- Modal Header -->
        <div class="modal-header bg-warning-subtle border-bottom border-warning-subtle">
          <div class="d-flex align-items-center gap-2">
            <i class="bi bi-cpu-fill fs-4 text-warning"></i>
            <div>
              <h6 class="modal-title fw-bold mb-0">全渠道 Webhook 訊息模擬器</h6>
              <small class="text-secondary" style="font-size: 0.75rem;">本地快速模擬五大通訊軟體進線訊息</small>
            </div>
          </div>
          <button type="button" class="btn-close" @click="emit('close')"></button>
        </div>

        <!-- Modal Body -->
        <div class="modal-body p-4">
          <!-- Platform Tabs -->
          <label class="form-label small fw-semibold mb-2">1. 選擇模擬的進線渠道：</label>
          <div class="btn-group w-100 mb-3" role="group">
            <button 
              type="button" 
              class="btn btn-sm"
              :class="selectedPlatform === 'line' ? 'btn-success' : 'btn-outline-success'"
              @click="selectPreset('line')"
            >
              <i class="bi bi-line me-1"></i>LINE
            </button>
            <button 
              type="button" 
              class="btn btn-sm"
              :class="selectedPlatform === 'instagram' ? 'btn-danger' : 'btn-outline-danger'"
              @click="selectPreset('instagram')"
            >
              <i class="bi bi-instagram me-1"></i>IG
            </button>
            <button 
              type="button" 
              class="btn btn-sm"
              :class="selectedPlatform === 'facebook' ? 'btn-primary' : 'btn-outline-primary'"
              @click="selectPreset('facebook')"
            >
              <i class="bi bi-messenger me-1"></i>FB
            </button>
            <button 
              type="button" 
              class="btn btn-sm"
              :class="selectedPlatform === 'douyin' ? 'btn-dark' : 'btn-outline-dark'"
              @click="selectPreset('douyin')"
            >
              <i class="bi bi-tiktok me-1"></i>抖音
            </button>
            <button 
              type="button" 
              class="btn btn-sm"
              :class="selectedPlatform === 'telegram' ? 'btn-info' : 'btn-outline-info'"
              @click="selectPreset('telegram')"
            >
              <i class="bi bi-telegram me-1"></i>TG
            </button>
          </div>

          <!-- Sender Details -->
          <div class="row g-2 mb-3">
            <div class="col-6">
              <label class="form-label small text-secondary">訪客暱稱</label>
              <input type="text" class="form-control form-control-sm" v-model="senderName" />
            </div>
            <div class="col-6">
              <label class="form-label small text-secondary">外部 User ID</label>
              <input type="text" class="form-control form-control-sm font-monospace" v-model="senderId" />
            </div>
          </div>

          <!-- Message Content -->
          <div class="mb-3">
            <label class="form-label small fw-semibold">2. 訊息內容：</label>
            <textarea 
              class="form-control" 
              rows="3" 
              v-model="messageContent"
              placeholder="輸入客戶傳送的文字訊息..."
            ></textarea>
          </div>

          <!-- Optional Media -->
          <div class="mb-2">
            <label class="form-label small text-secondary">附帶圖片 URL (可選)</label>
            <input 
              type="url" 
              class="form-control form-control-sm" 
              placeholder="https://images.unsplash.com/..." 
              v-model="mediaUrl" 
            />
          </div>

          <div v-if="successNotice" class="alert alert-success py-2 px-3 small d-flex align-items-center gap-2 mb-0 mt-3">
            <i class="bi bi-check-circle-fill"></i>
            模擬訊息已成功透過 WebSocket 推播至前端視窗！
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="modal-footer bg-body-tertiary">
          <button type="button" class="btn btn-secondary btn-sm px-3" @click="emit('close')">取消</button>
          <button 
            type="button" 
            class="btn btn-warning btn-sm px-4 fw-bold shadow-sm"
            @click="handleSimulate"
            :disabled="isSubmitting"
          >
            <span v-if="isSubmitting" class="spinner-border spinner-border-sm me-1"></span>
            <i v-else class="bi bi-lightning-fill me-1"></i>
            立即模擬發送
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-backdrop {
  z-index: 1040;
}
.modal {
  z-index: 1050;
}
</style>
