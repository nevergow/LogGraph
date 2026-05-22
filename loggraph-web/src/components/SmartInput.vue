<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { nodesApi } from '../api/nodes'
import type { Node, Block } from '../types'

const props = defineProps<{ editingBlock: Block | null }>()
const emit = defineEmits<{
  send: [content: string, metadata?: Record<string, any>]
  update: [id: string, content: string, metadata?: Record<string, any>]
  'cancel-edit': []
}>()

const text = ref('')
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const expandedTextareaRef = ref<HTMLTextAreaElement | null>(null)

// ── Progressive input state ──
const isExpanded = ref(false)

function expand() {
  isExpanded.value = true
  nextTick(() => {
    expandedTextareaRef.value?.focus()
    showSuggest.value = false
  })
}

function collapse() {
  isExpanded.value = false
  nextTick(() => {
    textareaRef.value?.focus()
    showSuggest.value = false
  })
}

function onExpandedKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
    e.preventDefault()
    collapse()
    submit()
  } else if (e.key === 'Escape') {
    e.preventDefault()
    collapse()
  }
}

// ── Autocomplete state ──
const showSuggest = ref(false)
const suggestType = ref<'project' | 'person' | 'reference'>('project')
const suggestItems = ref<Node[]>([])
const suggestIndex = ref(0)
const triggerPos = ref(0)
const triggerChar = ref('')

// ── Toolbar state ──
const showToolbar = ref(false)
const cursorStart = ref(0)
const cursorEnd = ref(0)

function trackCursor() {
  const ta = textareaRef.value
  if (!ta) return
  cursorStart.value = ta.selectionStart
  cursorEnd.value = ta.selectionEnd
}

// ── Trigger detection on every keystroke ──

watch(text, async (val) => {
  if (isExpanded.value) return
  const ta = textareaRef.value
  if (!ta) return
  const pos = ta.selectionStart
  const before = val.slice(0, pos)

  const hashIdx = before.lastIndexOf('#')
  const atIdx = before.lastIndexOf('@')
  const caretIdx = before.lastIndexOf('^')

  const lastIdx = Math.max(hashIdx, atIdx, caretIdx)
  if (lastIdx === -1) {
    showSuggest.value = false
    return
  }

  const after = before.slice(lastIdx)
  if (after.includes(' ') || after.includes('\n')) {
    showSuggest.value = false
    return
  }

  if (lastIdx === hashIdx) {
    suggestType.value = 'project'
    triggerChar.value = '#'
  } else if (lastIdx === atIdx) {
    suggestType.value = 'person'
    triggerChar.value = '@'
  } else {
    suggestType.value = 'reference'
    triggerChar.value = '^'
  }

  triggerPos.value = lastIdx
  const query = after.slice(1)
  suggestIndex.value = 0

  if (suggestType.value === 'reference') {
    try {
      const res = await fetch(`/api/v1/blocks?q=${encodeURIComponent(query)}&limit=5`)
      if (res.ok) {
        const page = await res.json()
        suggestItems.value = (page.data || []).map((b: any) => ({
          id: b.id,
          name: b.content.slice(0, 40),
          type: 'custom' as const,
          created_at: b.created_at,
        }))
        showSuggest.value = suggestItems.value.length > 0
      }
    } catch { showSuggest.value = false }
  } else {
    try {
      const type = suggestType.value === 'project' ? undefined : suggestType.value
      suggestItems.value = await nodesApi.suggest(query, type)
      showSuggest.value = suggestItems.value.length > 0
    } catch { showSuggest.value = false }
  }
})

// ── Keyboard navigation (compact mode) ──

function onKeydown(e: KeyboardEvent) {
  if (isExpanded.value) return

  if (!showSuggest.value) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      submit()
    }
    return
  }

  if (e.key === 'ArrowDown') {
    e.preventDefault()
    suggestIndex.value = Math.min(suggestIndex.value + 1, suggestItems.value.length - 1)
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    suggestIndex.value = Math.max(suggestIndex.value - 1, 0)
  } else if (e.key === 'Enter' || e.key === 'Tab') {
    e.preventDefault()
    applySuggestion(suggestItems.value[suggestIndex.value])
  } else if (e.key === 'Escape') {
    showSuggest.value = false
  }
}

// ── Apply suggestion ──

function applySuggestion(item: Node) {
  const before = text.value.slice(0, triggerPos.value)
  const ta = textareaRef.value!
  const pos = ta.selectionStart
  const after = text.value.slice(pos)

  if (suggestType.value === 'reference') {
    text.value = before + `^${item.id} ` + after
  } else {
    text.value = before + triggerChar.value + item.name + ' ' + after
  }
  showSuggest.value = false
  nextTick(() => {
    ta.focus()
    const newPos = before.length + (suggestType.value === 'reference' ? item.id.length + 2 : item.name.length + 2)
    ta.setSelectionRange(newPos, newPos)
  })
}

// ── Formatting helpers ──

function wrapSelection(prefix: string, suffix: string) {
  const ta = isExpanded.value ? expandedTextareaRef.value : textareaRef.value
  if (!ta) return
  const start = cursorStart.value
  const end = cursorEnd.value
  const selected = text.value.slice(start, end)

  if (selected) {
    text.value = text.value.slice(0, start) + prefix + selected + suffix + text.value.slice(end)
    nextTick(() => {
      ta.focus()
      const newPos = start + prefix.length + selected.length + suffix.length
      ta.setSelectionRange(newPos, newPos)
    })
  } else {
    text.value = text.value.slice(0, start) + prefix + suffix + text.value.slice(end)
    nextTick(() => {
      ta.focus()
      ta.setSelectionRange(start + prefix.length, start + prefix.length)
    })
  }
}

function insertAtCursor(value: string) {
  const ta = isExpanded.value ? expandedTextareaRef.value : textareaRef.value
  if (!ta) return
  const start = cursorStart.value
  text.value = text.value.slice(0, start) + value + text.value.slice(cursorEnd.value)
  nextTick(() => {
    ta.focus()
    const newPos = start + value.length
    ta.setSelectionRange(newPos, newPos)
  })
}

function insertAtLineStart(prefix: string) {
  const start = cursorStart.value
  const lineStart = text.value.lastIndexOf('\n', start - 1) + 1
  text.value = text.value.slice(0, lineStart) + prefix + text.value.slice(lineStart)
  nextTick(() => {
    const ta = isExpanded.value ? expandedTextareaRef.value! : textareaRef.value!
    ta.focus()
    const newPos = lineStart + prefix.length
    ta.setSelectionRange(newPos, newPos)
  })
}

// ── Track cursor in expanded mode ──

function trackExpandedCursor() {
  const ta = expandedTextareaRef.value
  if (!ta) return
  cursorStart.value = ta.selectionStart
  cursorEnd.value = ta.selectionEnd
}

// ── Priority parsing ──

function parsePriority(content: string): { cleanContent: string; priority: string } {
  let priority = 'normal'
  let clean = content.trim()

  const startMatch = clean.match(/^(!!|!high)\s+/)
  if (startMatch) {
    priority = 'high'
    clean = clean.slice(startMatch[0].length)
  }

  if (priority === 'normal') {
    const endMatch = clean.match(/\s+(!!|!high)$/)
    if (endMatch) {
      priority = 'high'
      clean = clean.slice(0, endMatch.index)
    }
  }

  return { cleanContent: clean.trim(), priority }
}

// ── Submit ──

function submit() {
  const trimmed = text.value.trim()
  if (!trimmed) return

  const { cleanContent, priority } = parsePriority(trimmed)
  if (!cleanContent) return

  const metadata: Record<string, any> = {}
  if (priority === 'high') metadata.priority = 'high'

  if (props.editingBlock) {
    emit('update', props.editingBlock.id, cleanContent, metadata)
  } else {
    emit('send', cleanContent, metadata)
  }
  text.value = ''
  showSuggest.value = false
  isExpanded.value = false
}

function cancelEdit() {
  text.value = ''
  isExpanded.value = false
  emit('cancel-edit')
}

// Watch for edit mode
watch(() => props.editingBlock, (b) => {
  if (b) {
    text.value = b.content
    isExpanded.value = false
    nextTick(() => {
      const ta = textareaRef.value
      if (ta) {
        ta.focus()
        ta.setSelectionRange(ta.value.length, ta.value.length)
      }
    })
  }
})

// ── File paste ──

function onPaste(e: ClipboardEvent) {
  const file = e.clipboardData?.files?.[0]
  if (!file) return
  e.preventDefault()
  const placeholder = `[${file.name}](uploading...)`
  const ta = isExpanded.value ? expandedTextareaRef.value! : textareaRef.value!
  const pos = ta.selectionStart
  text.value = text.value.slice(0, pos) + placeholder + text.value.slice(ta.selectionEnd)
}

// Auto-resize (compact mode only)
function autoResize() {
  if (isExpanded.value) return
  const ta = textareaRef.value
  if (!ta) return
  ta.style.height = 'auto'
  ta.style.height = Math.min(ta.scrollHeight, 160) + 'px'
}

watch(text, autoResize)

onMounted(() => {
  document.addEventListener('keydown', onKeydown)
})
onUnmounted(() => {
  document.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div class="bg-white px-3 py-2 shrink-0 relative safe-area-bottom">

    <!-- Suggest popover (compact mode only) -->
    <div
      v-if="showSuggest && !isExpanded"
      class="absolute bottom-full left-4 mb-1 w-64 max-h-40 overflow-y-auto bg-white border border-slate-200 rounded-lg shadow-xl z-50"
    >
      <div
        v-for="(item, i) in suggestItems"
        :key="item.id"
        class="px-3 py-2 text-sm cursor-pointer hover:bg-blue-50 flex items-center gap-2 transition-colors"
        :class="{ 'bg-blue-50': i === suggestIndex }"
        @click="applySuggestion(item)"
        @mouseenter="suggestIndex = i"
      >
        <span class="text-xs text-slate-400 w-5 shrink-0 font-mono">{{ triggerChar }}</span>
        <span class="truncate text-slate-700">{{ item.name }}</span>
        <span v-if="item.type !== 'custom'" class="text-[10px] text-slate-300 ml-auto">{{ item.type }}</span>
      </div>
    </div>

    <!-- Editing indicator -->
    <div v-if="editingBlock" class="flex items-center gap-2 mb-1.5">
      <span class="text-xs text-blue-600 font-medium">Editing</span>
      <button class="text-xs text-slate-400 hover:text-slate-600 underline" @click="cancelEdit">Cancel</button>
    </div>

    <!-- ── Compact mode ── -->
    <div v-if="!isExpanded" class="flex items-center gap-2">
      <!-- Expand toggle -->
      <button
        class="shrink-0 text-slate-400 hover:text-slate-600 hover:bg-slate-100 p-1.5 rounded-lg transition-colors"
        @click="expand"
        title="Expand editor"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
        </svg>
      </button>

      <!-- Capsule input -->
      <textarea
        ref="textareaRef"
        v-model="text"
        rows="1"
        :placeholder="editingBlock ? 'Edit and press Enter to save...' : '#project @person ^reference — Enter to send'"
        class="flex-1 resize-none outline-none text-sm sm:text-base py-2 px-4 bg-slate-100 rounded-full border-0 focus:bg-white focus:ring-2 focus:ring-blue-100 focus:border-blue-400 transition-all placeholder:text-slate-300"
        @click="trackCursor"
        @keyup="trackCursor"
        @select="trackCursor"
        @paste="onPaste"
      />

      <button
        class="shrink-0 px-4 py-2 bg-slate-900 text-white text-sm rounded-full hover:bg-slate-800 transition-colors disabled:opacity-30 font-medium"
        :disabled="!text.trim()"
        @click="submit"
      >
        {{ editingBlock ? 'Update' : 'Send' }}
      </button>
    </div>

    <!-- ── Expanded mode (Teleported) ── -->
    <Teleport to="body">
      <div
        v-if="isExpanded"
        class="fixed inset-0 z-40 bg-black/50"
        @click="collapse"
      />
      <div
        v-if="isExpanded"
        class="fixed bottom-0 left-0 right-0 z-50 bg-white rounded-t-2xl shadow-2xl flex flex-col"
        :style="{ maxHeight: '80vh', paddingBottom: 'env(safe-area-inset-bottom)' }"
      >
        <!-- Toolbar header -->
        <div class="flex items-center gap-0.5 px-4 py-2 border-b border-slate-100 shrink-0">
          <button class="toolbar-btn font-bold" @click="wrapSelection('**', '**')" title="Bold">B</button>
          <button class="toolbar-btn italic" @click="wrapSelection('*', '*')" title="Italic">I</button>
          <button class="toolbar-btn" @click="wrapSelection('~~', '~~')" title="Strikethrough"><span class="line-through">S</span></button>

          <span class="w-px h-5 bg-slate-200 mx-1" />

          <button class="toolbar-btn font-semibold text-[15px]" @click="insertAtLineStart('## ')" title="Heading">H</button>
          <button class="toolbar-btn" @click="insertAtLineStart('- ')" title="Bullet list">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
          <button class="toolbar-btn" @click="wrapSelection('`', '`')" title="Inline code">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
            </svg>
          </button>
          <button class="toolbar-btn" @click="insertAtLineStart('> ')" title="Blockquote">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </button>

          <span class="w-px h-5 bg-slate-200 mx-1" />

          <button class="toolbar-btn text-blue-600 font-semibold text-xs" @click="insertAtCursor('#')" title="Insert project tag">#</button>
          <button class="toolbar-btn text-emerald-600 font-semibold text-xs" @click="insertAtCursor('@')" title="Insert person mention">@</button>
          <button class="toolbar-btn text-purple-600 font-semibold text-xs" @click="insertAtCursor('^')" title="Insert block reference">^</button>

          <button class="toolbar-btn ml-auto text-slate-400" @click="collapse" title="Collapse">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </button>
        </div>

        <!-- Large textarea -->
        <textarea
          ref="expandedTextareaRef"
          v-model="text"
          class="flex-1 resize-none outline-none px-4 py-3 text-base leading-relaxed"
          :placeholder="editingBlock ? 'Editing... Cmd/Ctrl+Enter to save' : 'Write in markdown... Cmd/Ctrl+Enter to send'"
          @click="trackExpandedCursor"
          @keyup="trackExpandedCursor"
          @select="trackExpandedCursor"
          @keydown="onExpandedKeydown"
          @paste="onPaste"
        />

        <!-- Footer -->
        <div class="px-4 py-2 border-t border-slate-100 flex items-center justify-between shrink-0">
          <span class="text-[11px] text-slate-400">&#8984;/Ctrl+Enter to send &middot; Esc to close</span>
          <button
            class="px-5 py-2 bg-slate-900 text-white text-sm rounded-xl hover:bg-slate-800 transition-colors disabled:opacity-30 font-medium"
            :disabled="!text.trim()"
            @click="collapse(); submit()"
          >
            {{ editingBlock ? 'Update' : 'Send' }}
          </button>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.toolbar-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  font-size: 13px;
  color: #475569;
  transition: all 0.15s;
  cursor: pointer;
  border: none;
  background: transparent;
}
.toolbar-btn:hover {
  background: #f1f5f9;
  color: #1e293b;
}
.toolbar-btn:active {
  background: #e2e8f0;
}
</style>
