import { ref, reactive } from 'vue'
import type { Block } from '../types'
import { blocksApi, type BlockListParams } from '../api/blocks'

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

  async function createBlock(content: string) {
    const block = await blocksApi.create({ content })
    blocks.value.unshift(block)
    return block
  }

  async function updateBlock(id: string, data: { content?: string; status?: string }) {
    const updated = await blocksApi.update(id, data)
    const idx = blocks.value.findIndex(b => b.id === id)
    if (idx !== -1) blocks.value[idx] = updated
    return updated
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

  return {
    blocks,
    hasMore,
    loading,
    error,
    filters,
    selectedBlockId,
    fetchBlocks,
    createBlock,
    updateBlock,
    loadMore,
    setFilter,
  }
}
