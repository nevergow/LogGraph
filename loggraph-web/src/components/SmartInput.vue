<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, nextTick, shallowRef } from 'vue'
import { nodesApi } from '../api/nodes'
import { useNodes } from '../composables/useNodes'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import { Markdown } from 'tiptap-markdown'
import type { Node, Block } from '../types'

const props = defineProps<{ editingBlock: Block | null }>()
const emit = defineEmits<{
  send: [content: string, metadata?: Record<string, any>]
  update: [id: string, content: string, metadata?: Record<string, any>]
  'cancel-edit': []
}>()

const text = ref('')
const textareaRef = ref<HTMLTextAreaElement | null>(null)

const { projects } = useNodes()

// ── Priority quadrant ──
// q1: 紧急重要, q2: 紧急不重要, q3: 不紧急重要, q4: 不紧急不重要
const selectedPriority = ref<string>('q3')
const selectedProject = ref<string>('')

const quadrantLabels: Record<string, string> = {
  q1: '紧急重要',
  q2: '紧急不重要',
  q3: '不紧急重要',
  q4: '不紧急不重要',
}

function quadrantLabelShort(v: string): string {
  const m: Record<string, string> = { q1: 'Q1 紧急重要', q2: 'Q2 紧急不重要', q3: 'Q3 不紧急重要', q4: 'Q4 不紧急不重要' }
  return m[v] || 'Q3'
}

// ── Progressive input state ──
const isExpanded = ref(false)
const isFullscreen = ref(false)

// ── TipTap editor (expanded mode only) ──
const editor = useEditor({
  extensions: [
    StarterKit,
    Placeholder.configure({ placeholder: 'Write in markdown... Cmd/Ctrl+Enter to send' }),
    Markdown.configure({ html: false, breaks: true, linkify: true }),
  ],
  editorProps: {
    attributes: { class: 'flex-1 outline-none px-4 py-3 text-base leading-relaxed overflow-y-auto prose prose-sm max-w-none' },
    handleKeyDown: (_, e) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
        e.preventDefault()
        collapse()
        submit()
        return true
      }
      if (e.key === 'Escape') {
        e.preventDefault()
        collapse()
        return true
      }
      return false
    },
  },
  onUpdate: () => {
    if (editor.value) {
      text.value = editor.value.storage.markdown?.getMarkdown?.() || ''
    }
  },
  immediatelyRender: false,
})

function expand() {
  isExpanded.value = true
  showSuggest.value = false
  nextTick(() => {
    if (editor.value) {
      editor.value.commands.setContent(text.value || '')
      editor.value.commands.focus()
    }
  })
}

function collapse() {
  if (editor.value) {
    text.value = editor.value.storage.markdown?.getMarkdown?.() || ''
    editor.value.commands.clearContent()
  }
  isExpanded.value = false
  isFullscreen.value = false
  nextTick(() => {
    textareaRef.value?.focus()
    showSuggest.value = false
  })
}

function toggleFullscreen() {
  isFullscreen.value = !isFullscreen.value
  nextTick(() => editor.value?.commands.focus())
}

// ── TipTap toolbar actions ──
function tipTapBold() { editor.value?.chain().focus().toggleBold().run() }
function tipTapItalic() { editor.value?.chain().focus().toggleItalic().run() }
function tipTapStrike() { editor.value?.chain().focus().toggleStrike().run() }
function tipTapHeading() { editor.value?.chain().focus().toggleHeading({ level: 2 }).run() }
function tipTapBulletList() { editor.value?.chain().focus().toggleBulletList().run() }
function tipTapCode() { editor.value?.chain().focus().toggleCode().run() }
function tipTapBlockquote() { editor.value?.chain().focus().toggleBlockquote().run() }
function tipTapInsertTag(ch: string) {
  editor.value?.chain().focus().insertContent(ch).run()
}

// ── Autocomplete state ──
const showSuggest = ref(false)
const suggestType = ref<'project' | 'person' | 'reference'>('project')
const suggestItems = ref<Node[]>([])
const suggestIndex = ref(0)
const triggerPos = ref(0)
const triggerChar = ref('')

// ── Toolbar state (compact mode only) ──
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

// ── Submit ──

function submit() {
  // In expanded mode, get latest markdown from TipTap
  if (isExpanded.value && editor.value) {
    text.value = editor.value.storage.markdown?.getMarkdown?.() || ''
  }

  let trimmed = text.value.trim()
  if (!trimmed) return

  if (selectedProject.value) {
    trimmed = `#${selectedProject.value} ` + trimmed
  }

  const metadata: Record<string, any> = {}
  metadata.quadrant = selectedPriority.value
  if (selectedPriority.value === 'q1' || selectedPriority.value === 'q2') {
    metadata.priority = 'high'
  }

  if (props.editingBlock) {
    emit('update', props.editingBlock.id, trimmed, metadata)
  } else {
    emit('send', trimmed, metadata)
  }
  text.value = ''
  selectedPriority.value = 'q3'
  selectedProject.value = ''
  showSuggest.value = false
  isExpanded.value = false
  isFullscreen.value = false
  if (editor.value) editor.value.commands.clearContent()
}

function cancelEdit() {
  text.value = ''
  selectedPriority.value = 'q3'
  selectedProject.value = ''
  isExpanded.value = false
  isFullscreen.value = false
  if (editor.value) editor.value.commands.clearContent()
  emit('cancel-edit')
}

// Watch for edit mode
watch(() => props.editingBlock, (b) => {
  if (b) {
    text.value = b.content
    const q = b.metadata?.quadrant as string
    selectedPriority.value = (q === 'q1' || q === 'q2' || q === 'q3' || q === 'q4') ? q : 'q3'
    const projectMatch = b.content.match(/(?:^|\s)#([^\s#][^\s]*)/)
    selectedProject.value = projectMatch ? projectMatch[1] : ''
    isExpanded.value = false
    isFullscreen.value = false
    nextTick(() => {
      const ta = textareaRef.value
      if (ta) {
        ta.focus()
        ta.setSelectionRange(ta.value.length, ta.value.length)
      }
    })
  }
})

// ── File paste (compact mode only) ──

function onPaste(e: ClipboardEvent) {
  if (isExpanded.value) return
  const file = e.clipboardData?.files?.[0]
  if (!file) return
  e.preventDefault()
  const placeholder = `[${file.name}](uploading...)`
  const ta = textareaRef.value!
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
      class="absolute bottom-full left-4 mb-1 w-64 max-h-40 overflow-y-auto bg-white border border-gray-200 rounded-md shadow-elevated z-50"
    >
      <div
        v-for="(item, i) in suggestItems"
        :key="item.id"
        class="px-3 py-2 text-sm cursor-pointer hover:bg-brand-50 flex items-center gap-2 transition-colors"
        :class="{ 'bg-brand-50': i === suggestIndex }"
        @click="applySuggestion(item)"
        @mouseenter="suggestIndex = i"
      >
        <span class="text-xs text-gray-400 w-5 shrink-0 font-mono">{{ triggerChar }}</span>
        <span class="truncate text-gray-700">{{ item.name }}</span>
        <span v-if="item.type !== 'custom'" class="text-[10px] text-gray-300 ml-auto">{{ item.type }}</span>
      </div>
    </div>

    <!-- Editing indicator -->
    <div v-if="editingBlock" class="flex items-center gap-2 mb-1.5">
      <span class="text-xs text-brand-600 font-medium">Editing</span>
      <button class="text-xs text-gray-400 hover:text-gray-600 underline" @click="cancelEdit">Cancel</button>
    </div>

    <!-- ── Compact mode ── -->
    <div v-if="!isExpanded" class="flex items-center gap-2">
      <button
        class="shrink-0 text-gray-400 hover:text-gray-600 hover:bg-gray-100 p-1.5 rounded-sm transition-colors"
        @click="expand"
        title="Expand editor"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
        </svg>
      </button>

      <textarea
        ref="textareaRef"
        v-model="text"
        rows="1"
        :placeholder="editingBlock ? 'Edit and press Enter to save...' : '#project @person ^reference — Enter to send'"
        class="flex-1 resize-none outline-none text-sm sm:text-base py-2 px-4 bg-gray-50 rounded-sm border border-gray-200 focus:bg-white focus:border-brand-300 focus:ring-2 focus:ring-brand-50 transition-all placeholder:text-gray-300"
        @click="trackCursor"
        @keyup="trackCursor"
        @select="trackCursor"
        @paste="onPaste"
      />

      <!-- Priority dropdown -->
      <select v-model="selectedPriority" class="shrink-0 text-[10px] border border-gray-200 rounded-md px-1 py-1 bg-white text-gray-500 outline-none focus:border-brand-300 transition-colors">
        <option value="q1">Q1 紧急重要</option>
        <option value="q2">Q2 紧急不重要</option>
        <option value="q3">Q3 不紧急重要</option>
        <option value="q4">Q4 不紧急不重要</option>
      </select>

      <!-- Project dropdown -->
      <select v-model="selectedProject" class="shrink-0 text-[11px] border border-gray-200 rounded-md px-1.5 py-1 bg-white text-gray-500 outline-none focus:border-brand-300 transition-colors max-w-[110px]">
        <option value="">No project</option>
        <option v-for="p in projects" :key="p.name" :value="p.name">{{ p.name }}</option>
      </select>

      <button
        class="shrink-0 px-4 py-2 bg-brand-500 text-white text-sm rounded-sm hover:bg-brand-600 transition-colors disabled:opacity-30 font-medium"
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
        class="fixed inset-0 z-40 bg-black/40"
        :class="{ 'bg-black/60': isFullscreen }"
        @click="collapse"
      />
      <div
        v-if="isExpanded"
        class="fixed z-50 bg-white shadow-2xl flex flex-col overflow-hidden"
        :class="isFullscreen
          ? 'inset-4 rounded-xl'
          : 'bottom-4 left-4 right-4 sm:left-1/2 sm:-translate-x-1/2 sm:max-w-2xl sm:w-full rounded-xl'"
        :style="isFullscreen ? {} : { maxHeight: '80vh', paddingBottom: 'env(safe-area-inset-bottom)' }"
      >
        <!-- Toolbar header -->
        <div class="flex items-center gap-0.5 px-4 py-2 border-b border-gray-100 shrink-0">
          <button class="toolbar-btn font-bold" @click="tipTapBold" title="Bold">B</button>
          <button class="toolbar-btn italic" @click="tipTapItalic" title="Italic">I</button>
          <button class="toolbar-btn" @click="tipTapStrike" title="Strikethrough"><span class="line-through">S</span></button>

          <span class="w-px h-5 bg-gray-200 mx-1" />

          <button class="toolbar-btn font-semibold text-[15px]" :class="{ 'text-brand-600 bg-brand-50': editor?.isActive('heading', { level: 2 }) }" @click="tipTapHeading" title="Heading">H</button>
          <button class="toolbar-btn" :class="{ 'text-brand-600 bg-brand-50': editor?.isActive('bulletList') }" @click="tipTapBulletList" title="Bullet list">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
          <button class="toolbar-btn" :class="{ 'text-brand-600 bg-brand-50': editor?.isActive('code') }" @click="tipTapCode" title="Inline code">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
            </svg>
          </button>
          <button class="toolbar-btn" :class="{ 'text-brand-600 bg-brand-50': editor?.isActive('blockquote') }" @click="tipTapBlockquote" title="Blockquote">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </button>

          <span class="w-px h-5 bg-gray-200 mx-1" />

          <button class="toolbar-btn text-brand-600 font-semibold text-xs" @click="tipTapInsertTag('#')" title="Insert project tag">#</button>
          <button class="toolbar-btn text-brand-600 font-semibold text-xs" @click="tipTapInsertTag('@')" title="Insert person mention">@</button>
          <button class="toolbar-btn text-brand-600 font-semibold text-xs" @click="tipTapInsertTag('^')" title="Insert block reference">^</button>

          <div class="flex-1" />

          <!-- Fullscreen toggle -->
          <button class="toolbar-btn" @click="toggleFullscreen" :title="isFullscreen ? 'Exit fullscreen' : 'Fullscreen'">
            <svg v-if="isFullscreen" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
            <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
            </svg>
          </button>

          <!-- Collapse -->
          <button class="toolbar-btn text-gray-400" @click="collapse" title="Collapse">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </button>
        </div>

        <!-- TipTap WYSIWYG Editor -->
        <EditorContent :editor="editor" class="flex flex-1 overflow-hidden" />

        <!-- Footer -->
        <div class="px-4 py-2 border-t border-gray-100 flex items-center justify-end shrink-0 gap-2">
          <div class="flex items-center gap-1.5">
            <select v-model="selectedPriority" class="text-[10px] border border-gray-200 rounded-md px-1 py-1 bg-white text-gray-500 outline-none focus:border-brand-300 transition-colors">
              <option value="q1">Q1 紧急重要</option>
              <option value="q2">Q2 紧急不重要</option>
              <option value="q3">Q3 不紧急重要</option>
              <option value="q4">Q4 不紧急不重要</option>
            </select>
            <select v-model="selectedProject" class="text-[11px] border border-gray-200 rounded-md px-1.5 py-1 bg-white text-gray-500 outline-none focus:border-brand-300 transition-colors max-w-[110px]">
              <option value="">No project</option>
              <option v-for="p in projects" :key="p.name" :value="p.name">{{ p.name }}</option>
            </select>
            <button
              class="px-5 py-2 bg-brand-500 text-white text-sm rounded-sm hover:bg-brand-600 transition-colors disabled:opacity-30 font-medium"
              :disabled="!text.trim()"
              @click="collapse(); submit()"
            >
              {{ editingBlock ? 'Update' : 'Send' }}
            </button>
          </div>
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
.toolbar-btn:focus {
  outline: none;
}
</style>
