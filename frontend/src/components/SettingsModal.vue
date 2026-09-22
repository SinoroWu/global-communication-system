<script setup lang="ts">
import { ref, watch } from 'vue';
import { useChatStore } from '../stores/chatStore';
import { apiService } from '../services/api';
import type { ChannelConfig, Platform } from '../types';

defineProps<{
  isOpen: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const chatStore = useChatStore();

const activeTab = ref<Platform>('line');
const channelsCopy = ref<ChannelConfig[]>([]);
const isSaving = ref(false);
const saveSuccess = ref(false);

watch(
  () => chatStore.channels,
  (ch) => {
    channelsCopy.value = JSON.parse(JSON.stringify(ch));
  },
  { immediate: true }
);

function getActiveConfig(): ChannelConfig | undefined {
  return channelsCopy.value.find((c) => c.platform === activeTab.value);
}

function copyWebhookUrl(path: string) {
  const origin = window.location.port === '5173' ? 'http://localhost:8080' : window.location.origin;
  const full = origin + path;
  navigator.clipboard.writeText(full);
  alert(`已複製 Webhook 完整端點：\n${full}`);
}

async function saveCurrentConfig() {
  const cfg = getActiveConfig();
  if (!cfg) return;

  isSaving.value = true;
  saveSuccess.value = false;
  try {
    await apiService.saveChannel(cfg);
    await chatStore.refreshChannels();
    saveSuccess.value = true;
    setTimeout(() => {
      saveSuccess.value = false;
    }, 2000);
  } catch (err) {
    alert('儲存失敗，請檢查後端連線！');
  } finally {
    isSaving.value = false;
  }
}
</script>

<template>
  <div v-if="isOpen" class="modal-backdrop fade show"></div>

  <div v-if="isOpen" class="modal fade show d-block" tabindex="-1">
    <div class="modal-dialog modal-lg modal-dialog-centered">
      <div class="modal-content border-0 shadow-lg rounded-4 overflow-hidden">
        <!-- Modal Header -->
        <div class="modal-header bg-body-tertiary border-bottom">
          <div class="d-flex align-items-center gap-2">
            <i class="bi bi-sliders fs-4 text-primary"></i>
            <div>
              <h6 class="modal-title fw-bold mb-0">五大通訊渠道 API & Webhook 配置</h6>
              <small class="text-secondary" style="font-size: 0.75rem;">管理各平台開發者金鑰與 Webhook 回呼網址</small>
            </div>
          </div>
          <button type="button" class="btn-close" @click="emit('close')"></button>
        </div>

        <!-- Modal Body with Tabs -->
        <div class="modal-body p-0 d-flex flex-column flex-md-row" style="min-height: 420px;">
          <!-- Left Platform Navigation -->
          <div class="bg-body-secondary p-3 border-end" style="width: 200px; min-width: 200px;">
            <div class="nav flex-column nav-pills gap-1">
              <button 
                class="nav-link text-start rounded-3 d-flex align-items-center gap-2 py-2"
                :class="{ active: activeTab === 'line' }"
                @click="activeTab = 'line'"
              >
                <i class="bi bi-line text-success fs-5"></i>
                <span class="small fw-semibold">LINE Official</span>
              </button>
              <button 
                class="nav-link text-start rounded-3 d-flex align-items-center gap-2 py-2"
                :class="{ active: activeTab === 'instagram' }"
                @click="activeTab = 'instagram'"
              >
                <i class="bi bi-instagram text-danger fs-5"></i>
                <span class="small fw-semibold">Instagram</span>
              </button>
              <button 
                class="nav-link text-start rounded-3 d-flex align-items-center gap-2 py-2"
                :class="{ active: activeTab === 'facebook' }"
                @click="activeTab = 'facebook'"
              >
                <i class="bi bi-messenger text-primary fs-5"></i>
                <span class="small fw-semibold">Facebook</span>
              </button>
              <button 
                class="nav-link text-start rounded-3 d-flex align-items-center gap-2 py-2"
                :class="{ active: activeTab === 'douyin' }"
                @click="activeTab = 'douyin'"
              >
                <i class="bi bi-tiktok text-dark fs-5"></i>
                <span class="small fw-semibold">抖音 / TikTok</span>
              </button>
              <button 
                class="nav-link text-start rounded-3 d-flex align-items-center gap-2 py-2"
                :class="{ active: activeTab === 'telegram' }"
                @click="activeTab = 'telegram'"
              >
                <i class="bi bi-telegram text-info fs-5"></i>
                <span class="small fw-semibold">Telegram Bot</span>
              </button>
            </div>

            <!-- Public Tunnel Hint -->
            <div class="card bg-body border-0 shadow-sm rounded-3 p-2 mt-4">
              <div class="small fw-semibold text-secondary mb-1">
                <i class="bi bi-shield-check text-success me-1"></i>公網 Webhook 提示
              </div>
              <p class="mb-0 text-secondary" style="font-size: 0.72rem;">
                若要接收真實社群平台訊息，請使用 ngrok 或 Cloudflare Tunnel 將本地埠號 8080 映射至公網 HTTPS 網址。
              </p>
            </div>
          </div>

          <!-- Right Content Form -->
          <div class="flex-grow-1 p-4 overflow-y-auto" v-if="getActiveConfig()">
            <div class="d-flex align-items-center justify-content-between mb-3">
              <h6 class="fw-bold mb-0">{{ getActiveConfig()!.name }} 設定</h6>
              <div class="form-check form-switch mb-0">
                <input 
                  class="form-check-input cursor-pointer" 
                  type="checkbox" 
                  role="switch" 
                  id="channelEnableSwitch"
                  v-model="getActiveConfig()!.enabled"
                />
                <label class="form-check-label small" for="channelEnableSwitch">啟用此渠道</label>
              </div>
            </div>

            <!-- Webhook URL Display -->
            <div class="mb-3">
              <label class="form-label small fw-semibold">Webhook 接收端點 (Webhook URL)</label>
              <div class="input-group input-group-sm">
                <input 
                  type="text" 
                  class="form-control font-monospace bg-body-tertiary" 
                  readonly 
                  :value="'/api/webhooks/' + getActiveConfig()!.platform" 
                />
                <button 
                  class="btn btn-outline-secondary" 
                  type="button" 
                  @click="copyWebhookUrl('/api/webhooks/' + getActiveConfig()!.platform)"
                >
                  <i class="bi bi-clipboard me-1"></i>複製端點
                </button>
              </div>
              <small class="text-secondary" style="font-size: 0.72rem;">請將此 URL 填入對應平台的開發者後台 Webhook 設定中。</small>
            </div>

            <!-- Webhook Secret / Verify Token -->
            <div class="mb-3">
              <label class="form-label small fw-semibold">
                {{ activeTab === 'line' ? 'Channel Secret' : activeTab === 'telegram' ? 'Secret Token' : '驗證權杖 (Verify Token / App Secret)' }}
              </label>
              <input 
                type="password" 
                class="form-control form-control-sm font-monospace" 
                placeholder="輸入簽章密鑰或 Token" 
                v-model="getActiveConfig()!.webhookSecret" 
              />
            </div>

            <!-- Outbound Access Token -->
            <div class="mb-3">
              <label class="form-label small fw-semibold">發送金鑰 (Channel Access Token / Bot Token)</label>
              <textarea 
                class="form-control form-control-sm font-monospace" 
                rows="3" 
                placeholder="輸入發送訊息所需的 Access Token" 
                v-model="getActiveConfig()!.accessToken"
              ></textarea>
            </div>

            <!-- Channel / App ID -->
            <div class="mb-3">
              <label class="form-label small fw-semibold">頻道帳號 / 機器人 ID</label>
              <input 
                type="text" 
                class="form-control form-control-sm font-monospace" 
                placeholder="@my_bot 或 Page ID" 
                v-model="getActiveConfig()!.channelId" 
              />
            </div>

            <div v-if="saveSuccess" class="alert alert-success py-2 px-3 small d-flex align-items-center gap-2 mb-3">
              <i class="bi bi-check-circle-fill"></i>設定已成功更新至 SQLite 資料庫！
            </div>

            <button 
              class="btn btn-primary btn-sm px-4 shadow-sm"
              @click="saveCurrentConfig"
              :disabled="isSaving"
            >
              <span v-if="isSaving" class="spinner-border spinner-border-sm me-1"></span>
              儲存此渠道設定
            </button>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="modal-footer bg-body-tertiary">
          <button type="button" class="btn btn-secondary btn-sm px-3" @click="emit('close')">關閉</button>
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
.cursor-pointer {
  cursor: pointer;
}
</style>
