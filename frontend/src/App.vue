<script setup lang="ts">
import { ref, computed } from 'vue'
import ChatLayout from './components/ChatLayout.vue'
import { generateId } from './lib/utils'

const isSidebarOpen = ref(true)

interface RagHit {
  source: string
  content: string
  similarity: number
}

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  isStreaming?: boolean
  ragHits?: RagHit[]   // RAG 模式下检索命中的片段
  ragMode?: boolean    // 该消息是否使用了 RAG
}

interface Session {
  id: string
  title: string
  messages: Message[]
  createdAt: Date
}

const sessions = ref<Session[]>([])
const activeSessionId = ref<string | null>(null)

const activeSession = computed(() =>
  sessions.value.find(s => s.id === activeSessionId.value)
)

function createNewSession() {
  const session: Session = {
    id: generateId(),
    title: '新对话',
    messages: [],
    createdAt: new Date(),
  }
  sessions.value.unshift(session)
  activeSessionId.value = session.id
  return session
}

function selectSession(id: string) {
  activeSessionId.value = id
}

function deleteSession(id: string) {
  const idx = sessions.value.findIndex(s => s.id === id)
  if (idx !== -1) {
    sessions.value.splice(idx, 1)
    if (activeSessionId.value === id) {
      activeSessionId.value = sessions.value[0]?.id || null
    }
  }
}

function getOrCreateSession(): Session {
  if (!activeSession.value) {
    return createNewSession()
  }
  return activeSession.value
}

function updateSessionTitle(session: Session, firstMessage: string) {
  session.title = firstMessage.slice(0, 30) + (firstMessage.length > 30 ? '...' : '')
}
</script>

<template>
  <ChatLayout
    :sessions="sessions"
    :active-session="activeSession"
    :is-sidebar-open="isSidebarOpen"
    :get-or-create-session="getOrCreateSession"
    :update-title="updateSessionTitle"
    @toggle-sidebar="isSidebarOpen = !isSidebarOpen"
    @new-session="createNewSession"
    @select-session="selectSession"
    @delete-session="deleteSession"
  />
</template>
