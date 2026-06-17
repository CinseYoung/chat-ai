<script setup lang="ts">
import Sidebar from './Sidebar.vue'
import MessageList from './MessageList.vue'
import ChatInput from './ChatInput.vue'
import WelcomePage from './WelcomePage.vue'
import { useChat } from '../composables/useChat'
import { ref, computed, type PropType } from 'vue'
import { Menu, Wifi, WifiOff, Bot, Plus } from 'lucide-vue-next'

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  isStreaming?: boolean
}

interface Session {
  id: string
  title: string
  messages: Message[]
  createdAt: Date
}

const props = defineProps({
  sessions: { type: Array as PropType<Session[]>, required: true },
  activeSession: { type: Object as PropType<Session | undefined>, required: true },
  isSidebarOpen: { type: Boolean, required: true },
  getOrCreateSession: { type: Function, required: true },
  updateTitle: { type: Function, required: true },
})

const emit = defineEmits<{
  'toggle-sidebar': []
  'new-session': []
  'select-session': [id: string]
  'delete-session': [id: string]
}>()

const {
  isLoading,
  ollamaStatus,
  ragDocsCount,
  checkHealth,
  ingest,
  sendMessage,
} = useChat()

const useRAG = ref(false)

const hasMessages = computed(() => (props.activeSession?.messages.length ?? 0) > 0)

function handleSend(message: string) {
  const session = props.getOrCreateSession()
  if (!session) return

  if (session.messages.length === 0) {
    props.updateTitle(session, message)
  }

  sendMessage(
    message,
    session,
    { useRAG: useRAG.value },
    () => checkHealth(),
    (err) => console.error('Chat error:', err),
  )
}

async function handleReingest() {
  try {
    const r = await ingest()
    console.log('[ingest]', r)
  } catch (e) {
    console.error('[ingest] failed', e)
  }
}

function handleQuickAction(text: string) {
  handleSend(text)
}
</script>

<template>
  <div class="flex h-screen w-full overflow-hidden bg-background">
    <Sidebar
      :sessions="sessions"
      :active-session-id="activeSession?.id"
      :is-open="isSidebarOpen"
      @new-session="emit('new-session')"
      @select-session="(id) => emit('select-session', id)"
      @delete-session="(id) => emit('delete-session', id)"
    />

    <div class="flex flex-1 flex-col min-w-0">
      <header class="fixed top-0 left-0 right-0 z-10 flex items-center justify-between h-14 px-4 bg-background/80 backdrop-blur-xl border-b border-border">
        <div class="flex items-center gap-3">
          <button
            class="p-2 rounded-lg hover:bg-muted transition-colors cursor-pointer"
            @click="emit('toggle-sidebar')"
          >
            <Menu class="w-5 h-5" />
          </button>
          <div class="flex items-center gap-2">
            <Bot class="w-5 h-5 text-primary" />
            <span class="font-semibold text-foreground">Qwen3</span>
            <span class="text-xs text-muted-foreground">+ RAG demo</span>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button
            class="flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-full transition-colors cursor-pointer"
            :class="ollamaStatus === 'connected'
              ? 'bg-primary/10 text-primary'
              : 'bg-red-500/10 text-red-400'"
            @click="checkHealth"
          >
            <component :is="ollamaStatus === 'connected' ? Wifi : WifiOff" class="w-3.5 h-3.5" />
            {{ ollamaStatus === 'connected' ? '已连接' : '未连接' }}
          </button>
          <button
            class="p-2 rounded-lg hover:bg-muted transition-colors cursor-pointer"
            @click="emit('new-session')"
            title="新对话"
          >
            <Plus class="w-4 h-4" />
          </button>
        </div>
      </header>

      <main class="flex-1 min-h-0 overflow-hidden pt-14">
        <WelcomePage
          v-if="!hasMessages"
          @quick-action="handleQuickAction"
        />
        <MessageList
          v-else
          :messages="activeSession?.messages ?? []"
        />
      </main>

      <ChatInput
        :is-loading="isLoading"
        :use-r-a-g="useRAG"
        :rag-docs-count="ragDocsCount"
        @update:use-r-a-g="useRAG = $event"
        @reingest="handleReingest"
        @send="handleSend"
      />
    </div>
  </div>
</template>
