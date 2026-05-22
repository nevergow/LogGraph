import { ref, reactive, computed } from 'vue'
import type { Block } from '../types'
import { blocksApi, type BlockListParams } from '../api/blocks'

const pendingDeletes = new Map<string, { block: Block; timer: ReturnType<typeof setTimeout> }>()

export function useBlocks() {
  const blocks = ref<Block[]>([])
  const hasMore = ref(false)
  const nextCursor = ref<string | undefined>()
  const loading = ref(false)
  const error = ref<string | null>(null)
  const filters = reactive<BlockListParams>({ limit: 20 })
  const selectedBlockId = ref<string | null>(null)

  async function fetchBlocks(reset = false) {
    loading.value = true
    error.value = null
    try {
      const params = { ...filters }
      if (!reset && nextCursor.value) {
        params.cursor = nextCursor.value
      }
      const page = await blocksApi.list(params)
      if (reset) {
        blocks.value = page.data
      } else {
        blocks.value.push(...page.data)
      }
      hasMore.value = page.has_more
      nextCursor.value = page.next_cursor
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  async function createBlock(content: string, metadata?: Record<string, any>) {
    const data: { content: string; metadata?: Record<string, any> } = { content }
    if (metadata && Object.keys(metadata).length > 0) {
      data.metadata = metadata
    }
    const block = await blocksApi.create(data)
    blocks.value.unshift(block)
    return block
  }

  async function updateBlock(id: string, data: { content?: string; status?: string; metadata?: Record<string, any> }) {
    const updated = await blocksApi.update(id, data)
    const idx = blocks.value.findIndex(b => b.id === id)
    if (idx !== -1) blocks.value[idx] = updated
    return updated
  }

  function deleteBlock(id: string, undo = true): { block: Block } | undefined {
    const idx = blocks.value.findIndex(b => b.id === id)
    if (idx === -1) return undefined
    const [block] = blocks.value.splice(idx, 1)
    if (!undo) {
      blocksApi.delete(id).catch(() => {
        blocks.value.unshift(block)
      })
      return { block }
    }
    const timer = setTimeout(() => {
      blocksApi.delete(id).catch(() => {
        blocks.value.unshift(block)
      })
      pendingDeletes.delete(id)
    }, 3000)
    pendingDeletes.set(id, { block, timer })
    return { block }
  }

  function undoDelete(id: string) {
    const entry = pendingDeletes.get(id)
    if (!entry) return
    clearTimeout(entry.timer)
    blocks.value.unshift(entry.block)
    pendingDeletes.delete(id)
  }

  async function archiveBlock(id: string) {
    const block = blocks.value.find(b => b.id === id)
    if (!block) return
    const existingMeta = block.metadata || {}
    await updateBlock(id, { metadata: { ...existingMeta, isArchived: true } })
  }

  async function loadMore() {
    if (!hasMore.value || loading.value) return
    await fetchBlocks(false)
  }

  function setFilter(key: keyof BlockListParams, value: string | undefined) {
    if (value) {
      (filters as any)[key] = value
    } else {
      delete (filters as any)[key]
    }
    nextCursor.value = undefined
    fetchBlocks(true)
  }

  const hasActiveFilter = computed(() => {
    return !!(filters.project || filters.person || filters.status || filters.since || filters.until || filters.q)
  })

  function clearAllFilters() {
    const keys = Object.keys(filters) as (keyof BlockListParams)[]
    for (const k of keys) {
      if (k !== 'limit') delete (filters as any)[k]
    }
    nextCursor.value = undefined
    fetchBlocks(true)
  }

  const visibleBlocks = computed(() => {
    return blocks.value.filter(b => !(b.metadata?.isArchived === true))
  })

  return {
    blocks,
    visibleBlocks,
    hasMore,
    loading,
    error,
    filters,
    selectedBlockId,
    fetchBlocks,
    createBlock,
    updateBlock,
    deleteBlock,
    undoDelete,
    archiveBlock,
    loadMore,
    setFilter,
    hasActiveFilter,
    clearAllFilters,
  }
}
