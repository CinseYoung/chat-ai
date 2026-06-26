<script setup lang="ts">
import { MessageSquare, Plus, Trash2, Search } from 'lucide-vue-next'
import { ref, type PropType } from 'vue'

interface Session {
  id: string
  title: string
  messages: any[]
  createdAt: Date
}

const props = defineProps({
  sessions: { type: Array as PropType<Session[]>, required: true },
  activeSessionId: { type: String as PropType<string | null>, default: null },
  isOpen: { type: Boolean, required: true },
})

const emit = defineEmits<{
  'new-session': []
  'select-session': [id: string]
  'delete-session': [id: string]
}>()

const searchQuery = ref('')

const filteredSessions = computed(() => {
  if (!searchQuery.value) return props.sessions
  return props.sessions.filter(s =>
    s.title.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

function formatDate(date: Date | string) {
  const d = new Date(date)
  const now = new Date()
  const isToday = d.toDateString() === now.toDateString()
  if (isToday) return '今天'
  const yesterday = new Date(now)
  yesterday.setDate(yesterday.getDate() - 1)
  if (d.toDateString() === yesterday.toDateString()) return '昨天'
  return `${d.getMonth() + 1}/${d.getDate()}`
}

import { computed } from 'vue'
</script>

<template>
  <aside
    class="flex-shrink-0 flex flex-col h-full bg-card border-r border-border transition-all duration-300"
    :class="isOpen ? 'w-64' : 'w-0 overflow-hidden'"
  >
    <!-- New chat button -->
    <div class="p-3">
      <button
        class="flex items-center justify-center gap-2 w-full h-10 rounded-lg border border-border bg-muted/50 hover:bg-muted transition-colors cursor-pointer text-sm font-medium"
        @click="emit('new-session')"
      >
        <Plus class="w-4 h-4" />
        新对话
      </button>
    </div>

    <!-- Search -->
    <div class="px-3 pb-2">
      <div class="relative">
        <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="搜索对话..."
          class="w-full h-8 pl-8 pr-3 text-sm bg-muted/50 border border-border rounded-lg outline-none focus:ring-1 focus:ring-primary/50 transition-colors"
        />
      </div>
    </div>

    <!-- Session list -->
    <div class="flex-1 overflow-y-auto px-2 pb-3">
      <div
        v-for="session in filteredSessions"
        :key="session.id"
        class="group flex items-center gap-2 px-3 py-2.5 rounded-lg cursor-pointer transition-colors mb-0.5"
        :class="session.id === activeSessionId
          ? 'bg-muted text-foreground'
          : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground'"
        @click="emit('select-session', session.id)"
      >
        <MessageSquare class="w-4 h-4 flex-shrink-0 opacity-60" />
        <div class="flex-1 min-w-0">
          <div class="text-sm truncate">{{ session.title }}</div>
          <div class="text-xs text-muted-foreground/60 mt-0.5">{{ formatDate(session.createdAt) }}</div>
        </div>
        <button
          class="opacity-0 group-hover:opacity-100 p-1 rounded hover:bg-red-500/20 hover:text-red-400 transition-all cursor-pointer"
          @click.stop="emit('delete-session', session.id)"
        >
          <Trash2 class="w-3.5 h-3.5" />
        </button>
      </div>

      <div
        v-if="filteredSessions.length === 0"
        class="text-center text-sm text-muted-foreground/50 py-8"
      >
        {{ searchQuery ? '未找到匹配的对话' : '暂无对话记录' }}
      </div>
    </div>

    <!-- Footer -->
    <div class="p-3 border-t border-border">
      <div class="text-xs text-muted-foreground/50 text-center">
        Qwen3 + chromem-go
      </div>
    </div>
  </aside>
</template>
