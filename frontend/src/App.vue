<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useChatStore } from './stores/chatStore';
import ChannelSidebar from './components/ChannelSidebar.vue';
import ConversationList from './components/ConversationList.vue';
import ChatWindow from './components/ChatWindow.vue';
import ContactInfoDrawer from './components/ContactInfoDrawer.vue';
import SimulatorModal from './components/SimulatorModal.vue';
import SettingsModal from './components/SettingsModal.vue';

const chatStore = useChatStore();

const isDrawerOpen = ref(false);
const isSimulatorOpen = ref(false);
const isSettingsOpen = ref(false);
const isDarkMode = ref(false);

function toggleTheme() {
  isDarkMode.value = !isDarkMode.value;
  document.documentElement.setAttribute('data-bs-theme', isDarkMode.value ? 'dark' : 'light');
}

onMounted(() => {
  chatStore.init();
  // Check system preferred theme
  if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
    toggleTheme();
  }
});
</script>

<template>
  <div class="app-layout d-flex vh-100 vw-100 overflow-hidden bg-body">
    <!-- Channel Sidebar (Leftmost) -->
    <ChannelSidebar 
      @open-simulator="isSimulatorOpen = true"
      @open-settings="isSettingsOpen = true"
      @toggle-theme="toggleTheme"
    />

    <!-- Conversation List (Middle) -->
    <ConversationList 
      @open-simulator="isSimulatorOpen = true"
    />

    <!-- Main Chat Workspace (Right) -->
    <ChatWindow 
      @toggle-drawer="isDrawerOpen = !isDrawerOpen"
    />

    <!-- Contact Info Offcanvas Drawer -->
    <ContactInfoDrawer 
      :is-open="isDrawerOpen"
      @close="isDrawerOpen = false"
    />

    <!-- Modals -->
    <SimulatorModal 
      :is-open="isSimulatorOpen"
      @close="isSimulatorOpen = false"
    />

    <SettingsModal 
      :is-open="isSettingsOpen"
      @close="isSettingsOpen = false"
    />
  </div>
</template>

<style>
/* Global app styles */
html, body {
  height: 100%;
  margin: 0;
  padding: 0;
  user-select: none;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
}

.app-layout {
  position: relative;
}

/* Custom scrollbars */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: rgba(150, 150, 150, 0.3);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(150, 150, 150, 0.6);
}
</style>
