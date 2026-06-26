<script setup lang="ts">
import { Database, MessageCircleQuestion, Hash, AlertCircle } from 'lucide-vue-next'

const emit = defineEmits<{
  'quick-action': [text: string]
}>()

// 4 个示例问题:左两个命中知识库,右两个故意问 LLM 不知道的内容
// 先关 RAG 试一遍,再开 RAG 对比答案
const quickActions = [
  {
    icon: Database,
    title: '我们组的代号是什么?周会几点?',
    description: '知识库内有答案,开 RAG 应给出 Phoenix-7 + 周三下午 4 点',
    color: 'from-emerald-500/20 to-teal-500/20',
    borderHover: 'hover:border-emerald-500/40',
    prompt: '我们组的代号是什么?周会几点开?',
  },
  {
    icon: Hash,
    title: '单 PR 行数上限是多少?',
    description: '通用知识里没有,RAG 应答 800 行',
    color: 'from-blue-500/20 to-cyan-500/20',
    borderHover: 'hover:border-blue-500/40',
    prompt: '我们组单个 PR 的修改行数上限是多少?',
  },
  {
    icon: MessageCircleQuestion,
    title: '什么是 RAG?Hybrid Search?',
    description: '知识库内的分享提纲会被命中',
    color: 'from-purple-500/20 to-violet-500/20',
    borderHover: 'hover:border-purple-500/40',
    prompt: '简要解释 RAG 的标准流水线,以及什么是 Hybrid Search?',
  },
  {
    icon: AlertCircle,
    title: '一个知识库不知道的问题',
    description: '应答 "资料中未提及",对照 LLM 自由发挥',
    color: 'from-orange-500/20 to-red-500/20',
    borderHover: 'hover:border-orange-500/40',
    prompt: 'CEO 的办公室在哪一层?',
  },
]
</script>

<template>
  <div class="flex-1 flex items-center justify-center p-4">
    <div class="max-w-2xl w-full text-center animate-slide-up">
      <div class="mb-10">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-primary/20 to-primary/5 mb-4 animate-pulse-glow">
          <Database class="w-8 h-8 text-primary" />
        </div>
        <h1 class="text-3xl font-semibold text-foreground mb-2">
          Chat AI
        </h1>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <button
          v-for="action in quickActions"
          :key="action.title"
          class="gradient-border group flex items-start gap-3 p-4 rounded-xl bg-card border border-border text-left transition-all duration-300 cursor-pointer hover:bg-muted/50"
          :class="action.borderHover"
          @click="emit('quick-action', action.prompt)"
        >
          <div
            class="flex-shrink-0 w-10 h-10 rounded-lg bg-gradient-to-br flex items-center justify-center"
            :class="action.color"
          >
            <component :is="action.icon" class="w-5 h-5 text-foreground/80" />
          </div>
          <div>
            <div class="text-sm font-medium text-foreground group-hover:text-primary transition-colors">
              {{ action.title }}
            </div>
            <div class="text-xs text-muted-foreground mt-1">
              {{ action.description }}
            </div>
          </div>
        </button>
      </div>

    </div>
  </div>
</template>
