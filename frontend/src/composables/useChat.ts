import { ref, reactive } from 'vue'

// useChat 只关心 session.messages.push() 的能力,定义一个最小契约。
interface ChatSession {
  id: string
  messages: any[]
}

export function useChat() {
  const isLoading = ref(false)
  const ollamaStatus = ref<'connected' | 'disconnected'>('disconnected')
  const ragDocsCount = ref(0)

  async function checkHealth() {
    try {
      const res = await fetch('/api/health')
      const data = await res.json()
      ollamaStatus.value = data.ollama === 'connected' ? 'connected' : 'disconnected'
      ragDocsCount.value = data.docs_count ?? 0
    } catch {
      ollamaStatus.value = 'disconnected'
    }
  }

  async function ingest() {
    const res = await fetch('/api/ingest', { method: 'POST' })
    const data = await res.json()
    await checkHealth()
    return data
  }

  function sendMessage(
    message: string,
    session: ChatSession,
    options: { useRAG: boolean },
    onDone: () => void,
    onError: (err: string) => void,
  ) {
    isLoading.value = true

    session.messages.push({
      id: crypto.randomUUID(),
      role: 'user',
      content: message,
    })

    const assistantMsg = reactive({
      id: crypto.randomUUID(),
      role: 'assistant' as const,
      content: '',
      isStreaming: true,
      ragMode: options.useRAG,
      ragHits: [] as any[],
    })
    session.messages.push(assistantMsg)

    const params = new URLSearchParams({
      message,
      use_rag: String(options.useRAG),
    })
    const evtSource = new EventSource(`/api/chat/stream?${params}`)

    evtSource.addEventListener('context', (e) => {
      const data = JSON.parse((e as MessageEvent).data)
      assistantMsg.ragHits = data.hits || []
    })

    evtSource.addEventListener('content', (e) => {
      const data = JSON.parse((e as MessageEvent).data)
      assistantMsg.content += data.content
    })

    evtSource.addEventListener('done', () => {
      assistantMsg.isStreaming = false
      evtSource.close()
      isLoading.value = false
      onDone()
    })

    evtSource.addEventListener('error', (e) => {
      assistantMsg.isStreaming = false
      evtSource.close()
      isLoading.value = false
      const me = e as MessageEvent
      const data = me.data ? JSON.parse(me.data) : {}
      onError(data.content || '连接中断,请检查后端服务是否运行')
    })
  }

  checkHealth()

  return {
    isLoading,
    ollamaStatus,
    ragDocsCount,
    checkHealth,
    ingest,
    sendMessage,
  }
}
