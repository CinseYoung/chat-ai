<script setup lang="ts">
import { ref } from 'vue'
import { Send, Loader2, Database, RefreshCw } from 'lucide-vue-next'

const props = defineProps({
  isLoading: { type: Boolean, default: false },
  useRAG: { type: Boolean, default: false },
  ragDocsCount: { type: Number, default: 0 },
})

const emit = defineEmits<{
  send: [message: string]
  'update:useRAG': [value: boolean]
  'reingest': []
}>()

const inputText = ref('')
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const isReingesting = ref(false)

function handleSend() {
  const text = inputText.value.trim()
  if (!text || props.isLoading) return
  emit('send', text)
  inputText.value = ''
  resizeTextarea()
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSend()
  }
}

function resizeTextarea() {
  const el = textareaRef.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = Math.min(el.scrollHeight, 200) + 'px'
}

async function handleReingest() {
  isReingesting.value = true
  try {
    emit('reingest')
  } finally {
    setTimeout(() => { isReingesting.value = false }, 800)
  }
}
</script>

<template>
  <div class="border-t border-border bg-background/80 backdrop-blur-xl">
    <div class="max-w-3xl mx-auto px-4 py-3">
      <div class="relative flex items-end gap-2 bg-card border border-border rounded-2xl px-4 py-3 focus-within:ring-1 focus-within:ring-primary/50 transition-shadow">
        <textarea
          ref="textareaRef"
          v-model="inputText"
          :disabled="isLoading"
          placeholder="输入消息,Shift+Enter 换行..."
          rows="1"
          class="flex-1 resize-none bg-transparent outline-none text-[15px] text-foreground placeholder:text-muted-foreground/60 max-h-[200px]"
          @keydown="handleKeydown"
          @input="resizeTextarea"
        />

        <button
          class="flex-shrink-0 w-9 h-9 rounded-xl flex items-center justify-center transition-all cursor-pointer"
          :class="inputText.trim() && !isLoading
            ? 'bg-primary text-primary-foreground hover:bg-primary/90 shadow-lg shadow-primary/25'
            : 'bg-muted text-muted-foreground'"
          :disabled="isLoading"
          @click="handleSend"
        >
          <Loader2 v-if="isLoading" class="w-4 h-4 animate-spin" />
          <Send v-else class="w-4 h-4" />
        </button>
      </div>

      <!-- Bottom toolbar -->
      <div class="flex items-center justify-between mt-2 px-1">
        <div class="flex items-center gap-3">
          <!-- RAG 开关 -->
          <button
            class="flex items-center gap-1.5 text-xs px-2.5 py-1 rounded-full transition-colors cursor-pointer border"
            :class="useRAG
              ? 'bg-primary/15 text-primary border-primary/30'
              : 'bg-muted/40 text-muted-foreground border-transparent hover:border-border'"
            @click="emit('update:useRAG', !useRAG)"
          >
            <Database class="w-3 h-3" />
            RAG: {{ useRAG ? '开' : '关' }}
            <span class="text-[10px] opacity-60">({{ ragDocsCount }} chunks)</span>
          </button>

          <!-- 重新灌库 -->
          <button
            class="flex items-center gap-1 text-[11px] px-2 py-0.5 rounded-full text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
            :disabled="isReingesting"
            @click="handleReingest"
            title="读取 backend-go/data/docs 重新灌入向量库"
          >
            <RefreshCw class="w-3 h-3" :class="{ 'animate-spin': isReingesting }" />
            重灌知识库
          </button>
        </div>
        <span class="text-xs text-muted-foreground/40">
          本地模型,数据不上传
        </span>
      </div>
    </div>
  </div>
</template>
