<script setup lang="ts">
import { ref, watch, nextTick, type PropType } from 'vue'
import MessageItem from './MessageItem.vue'

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  isStreaming?: boolean
}

const props = defineProps({
  messages: { type: Array as PropType<Message[]>, required: true },
})

const scrollContainer = ref<HTMLElement | null>(null)

function scrollToBottom() {
  nextTick(() => {
    if (scrollContainer.value) {
      scrollContainer.value.scrollTop = scrollContainer.value.scrollHeight
    }
  })
}

watch(
  () => props.messages.length,
  () => scrollToBottom(),
  { immediate: true },
)

watch(
  () => {
    const last = props.messages[props.messages.length - 1]
    return last?.content?.length
  },
  () => scrollToBottom(),
)
</script>

<template>
  <div ref="scrollContainer" class="h-full overflow-y-auto">
    <div class="max-w-3xl mx-auto px-4 py-6 space-y-6">
      <MessageItem
        v-for="msg in messages"
        :key="msg.id"
        :message="msg"
      />
    </div>
  </div>
</template>
