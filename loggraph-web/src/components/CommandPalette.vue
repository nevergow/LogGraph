<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'

const emit = defineEmits<{
  close: []
  command: [action: string]
}>()

const open = ref(false)
const search = ref('')
const selectedIndex = ref(0)
const inputRef = ref<HTMLInputElement | null>(null)

interface Command {
  id: string
  label: string
  shortcut?: string
  category: string
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

const filtered = ref<Command[]>(allCommands)

function filterCommands() {
  const q = search.value.toLowerCase()
  filtered.value = q
    ? allCommands.filter(c => c.label.toLowerCase().includes(q))
    : allCommands
  selectedIndex.value = 0
}

function execute(id: string) {
  emit('command', id)
  close()
}

function close() {
  open.value = false
  search.value = ''
  filtered.value = allCommands
  selectedIndex.value = 0
  emit('close')
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    selectedIndex.value = Math.min(selectedIndex.value + 1, filtered.value.length - 1)
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    selectedIndex.value = Math.max(selectedIndex.value - 1, 0)
  } else if (e.key === 'Enter') {
    e.preventDefault()
    if (filtered.value[selectedIndex.value]) {
      execute(filtered.value[selectedIndex.value].id)
    }
  } else if (e.key === 'Escape') {
    close()
  }
}

function onGlobalKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault()
    open.value = true
    nextTick(() => inputRef.value?.focus())
  }
}

onMounted(() => {
  document.addEventListener('keydown', onGlobalKeydown)
})
onUnmounted(() => {
  document.removeEventListener('keydown', onGlobalKeydown)
})
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
        class="relative w-full max-w-md bg-white rounded-md shadow-elevated border border-gray-200 overflow-hidden"
        @click.stop
      >
        <div class="flex items-center gap-2 px-4 py-3 border-b border-gray-100">
          <svg class="w-4 h-4 text-gray-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            ref="inputRef"
            v-model="search"
            class="flex-1 outline-none text-sm text-gray-700 placeholder:text-gray-400"
            placeholder="Type a command..."
            @input="filterCommands"
            @keydown="onKeydown"
          />
          <kbd class="text-[10px] text-gray-400 bg-gray-100 px-1.5 py-0.5 rounded-sm font-mono">esc</kbd>
        </div>
        <div class="max-h-64 overflow-y-auto py-1">
          <div
            v-for="(cmd, i) in filtered"
            :key="cmd.id"
            class="flex items-center justify-between px-4 py-2 text-sm cursor-pointer transition-colors"
            :class="i === selectedIndex ? 'bg-brand-50 text-brand-700' : 'text-gray-600 hover:bg-gray-50'"
            @click="execute(cmd.id)"
            @mouseenter="selectedIndex = i"
          >
            <span>{{ cmd.label }}</span>
            <span class="text-[10px] text-gray-400">{{ cmd.category }}</span>
          </div>
          <div v-if="filtered.length === 0" class="px-4 py-6 text-center text-sm text-gray-400">
            No commands found
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>
