<script setup lang="ts">
import { computed, ref, type PropType } from 'vue'
import { Bot, User, Database, ChevronDown, ChevronRight } from 'lucide-vue-next'
import Markdown from 'markdown-it'
import hljs from 'highlight.js'

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
  ragHits?: RagHit[]
  ragMode?: boolean
}

const props = defineProps({
  message: { type: Object as PropType<Message>, required: true },
})

const md = new Markdown({
  html: false,
  linkify: true,
  breaks: true,
  highlight(str: string, lang: string) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return hljs.highlight(str, { language: lang }).value
      } catch {}
    }
    return ''
  },
})

const renderedContent = computed(() =>
  props.message.role === 'user' ? props.message.content : md.render(props.message.content)
)

const isUser = computed(() => props.message.role === 'user')
const hasRagHits = computed(() => (props.message.ragHits?.length ?? 0) > 0)

const showHits = ref(false)
</script>

<template>
  <div class="flex gap-3 animate-slide-up" :class="isUser ? 'flex-row-reverse' : 'flex-row'">
    <div
      class="flex-shrink-0 w-8 h-8 rounded-lg flex items-center justify-center"
      :class="isUser ? 'bg-primary/20 text-primary' : 'bg-muted text-foreground'"
    >
      <component :is="isUser ? User : Bot" class="w-4 h-4" />
    </div>

    <div class="max-w-[80%] flex flex-col gap-2">
      <!-- RAG 命中片段(可折叠),用于演示中可视化"检索"过程 -->
      <div v-if="!isUser && hasRagHits" class="bg-primary/5 border border-primary/20 rounded-xl text-xs">
        <button
          class="w-full flex items-center gap-1.5 px-3 py-2 text-primary cursor-pointer hover:bg-primary/10 rounded-xl transition-colors"
          @click="showHits = !showHits"
        >
          <Database class="w-3.5 h-3.5" />
          <span class="font-medium">RAG 命中 {{ message.ragHits!.length }} 条</span>
          <component :is="showHits ? ChevronDown : ChevronRight" class="w-3.5 h-3.5 ml-auto" />
        </button>
        <div v-if="showHits" class="px-3 pb-3 space-y-2 border-t border-primary/10 pt-2">
          <div
            v-for="(hit, idx) in message.ragHits"
            :key="idx"
            class="bg-background/60 border border-border rounded-lg p-2"
          >
            <div class="flex items-center justify-between text-[10px] text-muted-foreground mb-1">
              <span class="font-mono">[{{ idx + 1 }}] {{ hit.source }}</span>
              <span class="text-primary">sim {{ hit.similarity.toFixed(3) }}</span>
            </div>
            <div class="text-foreground/90 whitespace-pre-wrap leading-snug">{{ hit.content }}</div>
          </div>
        </div>
      </div>

      <!-- RAG 模式但知识库无命中,提示用户 -->
      <div
        v-else-if="!isUser && message.ragMode && !message.isStreaming"
        class="text-[11px] text-muted-foreground/70 inline-flex items-center gap-1"
      >
        <Database class="w-3 h-3" />
        RAG 模式 · 知识库无命中
      </div>

      <div
        class="rounded-2xl px-4 py-3 text-[15px] leading-relaxed"
        :class="isUser ? 'bg-primary/15 text-foreground rounded-tr-md' : 'bg-card border border-border rounded-tl-md'"
      >
        <div v-if="isUser" class="whitespace-pre-wrap">{{ message.content }}</div>
        <div
          v-else
          class="markdown-body"
          :class="{ 'typing-cursor': message.isStreaming }"
          v-html="renderedContent"
        />
      </div>
    </div>
  </div>
</template>
