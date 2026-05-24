<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { extractTitle } from '../composables/useMarkdown'

const emit = defineEmits<{
  close: []
  command: [action: string]
  'select-block': [id: string]
}>()

const open = ref(false)
const search = ref('')
const selectedIndex = ref(0)
const inputRef = ref<HTMLInputElement | null>(null)
const searching = ref(false)

interface Command {
  id: string
  label: string
  shortcut?: string
  category: string
}

interface SearchResult {
  id: string
  title: string
  created_at: string
}

const allCommands: Command[] = [
  { id: 'view-project', label: 'Go to Project View', category: 'Navigation' },
  { id: 'view-timeline', label: 'Go to Timeline View', category: 'Navigation' },
  { id: 'graph', label: 'Open Knowledge Graph', category: 'Panels' },
  { id: 'ai-report', label: 'AI Report', category: 'Panels' },
  { id: 'webhooks', label: 'Webhook Settings', category: 'Panels' },
  { id: 'clear-filters', label: 'Clear All Filters', category: 'Actions' },
  { id: 'new-entry', label: 'New Log Entry', category: 'Actions' },
]

const filteredCommands = ref<Command[]>(allCommands)
const searchResults = ref<SearchResult[]>([])

// Combined list for unified keyboard navigation
interface ListItem {
  type: 'command' | 'result' | 'section'
  data?: Command | SearchResult
  label?: string
}

const listItems = ref<ListItem[]>([])

function rebuildList() {
  const items: ListItem[] = []
  if (filteredCommands.value.length > 0) {
    items.push({ type: 'section', label: 'Commands' })
    for (const c of filteredCommands.value) {
      items.push({ type: 'command', data: c })
    }
  }
  if (searchResults.value.length > 0) {
    items.push({ type: 'section', label: 'Search Results' })
    for (const r of searchResults.value) {
      items.push({ type: 'result', data: r })
    }
  }
  listItems.value = items
  // Clamp selectedIndex
  const selectable = items.filter(i => i.type !== 'section')
  if (selectedIndex.value >= selectable.length) {
    selectedIndex.value = Math.max(0, selectable.length - 1)
  }
}

function filterCommands() {
  const q = search.value.toLowerCase()
  filteredCommands.value = q
    ? allCommands.filter(c => c.label.toLowerCase().includes(q))
    : allCommands
}

let searchTimer: ReturnType<typeof setTimeout> | null = null

async function doSearch() {
  const q = search.value.trim()
  if (!q || q.length < 2) {
    searchResults.value = []
    rebuildList()
    return
  }
  searching.value = true
  try {
    const res = await fetch(`/api/v1/blocks?q=${encodeURIComponent(q)}&limit=5`)
    if (res.ok) {
      const page = await res.json()
      searchResults.value = (page.data || []).map((b: any) => ({
        id: b.id,
        title: extractTitle(b.content || ''),
        created_at: b.created_at,
      }))
    } else {
      searchResults.value = []
    }
  } catch {
    searchResults.value = []
  } finally {
    searching.value = false
    rebuildList()
  }
}

watch(search, () => {
  filterCommands()
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(doSearch, 200)
})

function onInputChange() {
  selectedIndex.value = 0
}

function getSelectableItems(): ListItem[] {
  return listItems.value.filter(i => i.type !== 'section')
}

function executeCurrent() {
  const items = getSelectableItems()
  const item = items[selectedIndex.value]
  if (!item) return
  if (item.type === 'command') {
    emit('command', (item.data as Command).id)
  } else if (item.type === 'result') {
    emit('select-block', (item.data as SearchResult).id)
  }
  close()
}

function close() {
  open.value = false
  search.value = ''
  filteredCommands.value = allCommands
  searchResults.value = []
  rebuildList()
  selectedIndex.value = 0
  emit('close')
}

function onKeydown(e: KeyboardEvent) {
  const items = getSelectableItems()
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    selectedIndex.value = Math.min(selectedIndex.value + 1, items.length - 1)
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    selectedIndex.value = Math.max(selectedIndex.value - 1, 0)
  } else if (e.key === 'Enter') {
    e.preventDefault()
    executeCurrent()
  } else if (e.key === 'Escape') {
    close()
  }
}

function onGlobalKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault()
    open.value = true
    filteredCommands.value = allCommands
    searchResults.value = []
    rebuildList()
    nextTick(() => inputRef.value?.focus())
  }
}

onMounted(() => {
  document.addEventListener('keydown', onGlobalKeydown)
  rebuildList()
})
onUnmounted(() => {
  document.removeEventListener('keydown', onGlobalKeydown)
})

function formatTime(ts: string): string {
  const d = new Date(ts)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fixed inset-0 z-50 flex items-start justify-center pt-[20vh]"
      @click="close"
    >
      <div class="absolute inset-0 bg-black/30" />
      <div
        class="relative w-full max-w-lg bg-white rounded-xl shadow-elevated border border-slate-200 overflow-hidden"
        @click.stop
      >
        <div class="flex items-center gap-2 px-4 py-3 border-b border-slate-100">
          <svg class="w-4 h-4 text-text-muted shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            ref="inputRef"
            v-model="search"
            class="flex-1 outline-none text-sm text-text-primary placeholder:text-text-muted"
            placeholder="Search commands or log entries..."
            @input="onInputChange"
            @keydown="onKeydown"
          />
          <span v-if="searching" class="w-4 h-4 border-2 border-accent-300 border-t-transparent rounded-full animate-spin" />
          <kbd class="text-[10px] text-text-muted bg-surface-100 px-1.5 py-0.5 rounded font-mono">esc</kbd>
        </div>
        <div class="max-h-80 overflow-y-auto py-1">
          <template v-for="(item, i) in listItems" :key="i + (item.type === 'section' ? item.label! : (item.data as any).id)">
            <div
              v-if="item.type === 'section'"
              class="px-4 py-2 text-[10px] text-text-muted uppercase tracking-wider font-semibold"
            >
              {{ item.label }}
            </div>
            <div
              v-else
              class="flex items-center justify-between px-4 py-2.5 text-sm cursor-pointer transition-colors"
              :class="selectedIndex === getSelectableItems().indexOf(item) ? 'bg-accent-50 text-accent-700' : 'text-text-primary hover:bg-surface-100'"
              @click="executeCurrent()"
              @mouseenter="selectedIndex = getSelectableItems().indexOf(item)"
            >
              <template v-if="item.type === 'command'">
                <span>{{ (item.data as Command).label }}</span>
                <span class="text-[10px] text-text-muted">{{ (item.data as Command).category }}</span>
              </template>
              <template v-else-if="item.type === 'result'">
                <span class="truncate flex-1">{{ (item.data as SearchResult).title }}</span>
                <span class="text-[10px] text-text-muted font-mono ml-3 shrink-0">{{ formatTime((item.data as SearchResult).created_at) }}</span>
              </template>
            </div>
          </template>
          <div v-if="listItems.length === 0 && !searching && search.length >= 2" class="px-4 py-6 text-center text-sm text-text-muted">
            No results found
          </div>
          <div v-if="!search" class="px-4 py-6 text-center text-xs text-text-muted">
            Type to search commands, or enter 2+ characters to search log entries
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>
