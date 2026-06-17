<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { nodesApi } from '../api/nodes'
import { useNodes } from '../composables/useNodes'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import { Markdown } from 'tiptap-markdown'
import type { Node } from '../types'

const props = defineProps<{
  screenSize?: 'mobile' | 'tablet' | 'desktop'
  prefillProject?: string
  prefillContent?: string
}>()

const emit = defineEmits<{
  send: [content: string, metadata?: Record<string, any>, parentBlockId?: string]
  'clear-prefill': []
}>()

const text = ref('')
const textareaRef = ref<HTMLTextAreaElement | null>(null)

const { projects, fetchProjects } = useNodes()

const selectedProject = ref<string>('')

// ── Progressive input state ──
const isExpanded = ref(false)
const isFullscreen = ref(false)
const isFocused = ref(false)
const charCount = ref(0)

// ── TipTap editor ──
const editor = useEditor({
  extensions: [
    StarterKit,
    Placeholder.configure({ placeholder: 'Write in markdown... Cmd/Ctrl+Enter to send' }),
    Markdown.configure({ html: false, breaks: true, linkify: true }),
  ],
  editorProps: {
    attributes: { class: 'flex-1 outline-none px-4 py-3 text-base leading-relaxed overflow-y-auto prose prose-sm max-w-none' },
    handleKeyDown: (_: any, e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
        e.preventDefault()
        collapse()
        submit()
        return true
      }
      if (e.key === 'Escape') {
        if (showSlashMenu.value) {
          showSlashMenu.value = false
          return true
        }
        e.preventDefault()
        collapse()
        return true
      }
      if (e.key === '/') {
        const { $from } = editor.value!.state.selection
        const isLineStart = $from.parentOffset === 0
        if (isLineStart) {
          showSlashMenu.value = true
          slashMenuIndex.value = 0
          // Don't prevent default: allow / to be typed, then filter it on select
        }
      }
      if (showSlashMenu.value) {
        onSlashKeydown(e)
        if (['ArrowDown', 'ArrowUp', 'Enter', 'Escape'].includes(e.key)) {
          return true
        }
      }
      return false
    },
  },
  onUpdate: () => {
    if (editor.value) {
      // @ts-ignore
      text.value = editor.value.storage.markdown?.getMarkdown?.() || ''
    }
  },
} as any)

function expand() {
  isExpanded.value = true
  showSuggest.value = false
  nextTick(() => {
    if (editor.value) {
      editor.value.commands.setContent(text.value || '')
      editor.value.commands.focus('end')
    }
  })
}

function onCompactFocus() {
  isFocused.value = true
}

function onCompactBlur() {
  isFocused.value = false
}

function onCompactClick() {
  trackCursor()
  if (text.value.trim() || props.prefillContent) {
    expand()
  }
}

function collapse() {
  if (editor.value) {
    // @ts-ignore
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
function tipTapBulletList() { editor.value?.chain().focus().toggleBulletList().run() }
function tipTapCode() { editor.value?.chain().focus().toggleCode().run() }
function tipTapBlockquote() { editor.value?.chain().focus().toggleBlockquote().run() }
function tipTapInsertTag(ch: string) {
  editor.value?.chain().focus().insertContent(ch).run()
}

// ── Auto-expand on content length ──
watch(text, (val) => {
  charCount.value = val.length
  // Auto-expand when typing long content (> 80 chars) in compact mode
  if (!isExpanded.value && val.length > 80 && !showSuggest.value) {
    expand()
  }
})

// ── Slash menu ──
const showSlashMenu = ref(false)
const slashMenuIndex = ref(0)
const slashItems = [
  { id: 'text', label: 'Text', icon: 'T', action: () => insertSlash('') },
  { id: 'bullet', label: 'Bullet List', icon: '☰', action: () => insertSlash('- ') },
  { id: 'code', label: 'Code Block', icon: '</>', action: () => insertSlash('```\n\n```', true) },
  { id: 'quote', label: 'Blockquote', icon: '"', action: () => insertSlash('> ') },
  { id: 'divider', label: 'Divider', icon: '—', action: () => insertSlash('---\n') },
]

function insertSlash(text: string, moveUp?: boolean) {
  if (!editor.value) return
  showSlashMenu.value = false
  const { $from } = editor.value.state.selection
  const lineStart = $from.start()
  editor.value.chain().focus().deleteRange({ from: lineStart, to: $from.pos }).run()
  if (moveUp) {
    const pos = editor.value.state.selection.from
    editor.value.chain().focus().insertContent(text).run()
    editor.value.commands.setTextSelection(pos + 3)
  } else {
    editor.value.chain().focus().insertContent(text).run()
  }
}

function onSlashKeydown(e: KeyboardEvent) {
  if (!showSlashMenu.value) return
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    slashMenuIndex.value = Math.min(slashMenuIndex.value + 1, slashItems.length - 1)
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    slashMenuIndex.value = Math.max(slashMenuIndex.value - 1, 0)
  } else if (e.key === 'Enter') {
    e.preventDefault()
    slashItems[slashMenuIndex.value].action()
  } else if (e.key === 'Escape') {
    e.preventDefault()
    showSlashMenu.value = false
  }
}

// ── Autocomplete state ──
const showSuggest = ref(false)
const suggestType = ref<'project' | 'person' | 'reference'>('project')
const suggestItems = ref<any[]>([])
const suggestIndex = ref(0)
const triggerPos = ref(0)
const triggerChar = ref('')

// ── Toolbar state ──
const cursorStart = ref(0)
const cursorEnd = ref(0)

function trackCursor() {
  const ta = textareaRef.value
  if (!ta) return
  cursorStart.value = ta.selectionStart
  cursorEnd.value = ta.selectionEnd
}

// ── Trigger detection: & (project), @ (person), ^ (reference) ──

watch(text, async (val) => {
  if (isExpanded.value) return
  const ta = textareaRef.value
  if (!ta) return
  const pos = ta.selectionStart
  const before = val.slice(0, pos)

  const ampIdx = before.lastIndexOf('&')
  const atIdx = before.lastIndexOf('@')
  const caretIdx = before.lastIndexOf('^')

  const lastIdx = Math.max(ampIdx, atIdx, caretIdx)
  if (lastIdx === -1) {
    showSuggest.value = false
    return
  }

  const after = before.slice(lastIdx)
  if (after.includes(' ') || after.includes('\n')) {
    showSuggest.value = false
    return
  }

  if (lastIdx === ampIdx) {
    suggestType.value = 'project'
    triggerChar.value = '&'
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
      const params = new URLSearchParams({ q: query, limit: '5' })
      if (selectedProject.value) {
        // Filter references to current project
        params.set('project', selectedProject.value)
      }
      const res = await fetch(`/api/v1/blocks?${params}`)
      if (res.ok) {
        const page = await res.json()
        suggestItems.value = (page.data || []).map((b: any) => ({
          id: b.id,
          name: (b.content || '').slice(0, 40),
          type: 'custom' as const,
          created_at: b.created_at,
        }))
        showSuggest.value = suggestItems.value.length > 0
      }
    } catch { showSuggest.value = false }
  } else {
    try {
      const type = suggestType.value === 'person' ? 'person' : undefined
      suggestItems.value = await nodesApi.suggest(query, type)
      // If no exact match, offer to create a new project/person
      if (query && !suggestItems.value.some((it: any) => it.name === query)) {
        suggestItems.value.push({ id: '__create__', name: query, type: suggestType.value, _create: true } as any)
      }
      showSuggest.value = suggestItems.value.length > 0
    } catch { showSuggest.value = false }
  }
})

// ── Keyboard navigation ──

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

async function applySuggestion(item: any) {
  // Handle "Create new project/person" action
  if (item._create) {
    try {
      const nodeType = suggestType.value === 'person' ? 'person' : 'project'
      await nodesApi.create(item.name, nodeType)
      await fetchProjects()
    } catch {
      // If creation fails, just insert the text as-is without a capsule
    }
  }

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

// ── Prefill handling (Phase 2.1 follow-up) ──
watch(() => props.prefillProject, (val) => {
  if (val) selectedProject.value = val
})
watch(() => props.prefillContent, (val) => {
  if (val) {
    text.value = val
    nextTick(() => {
      textareaRef.value?.focus()
      const len = val.length
      textareaRef.value?.setSelectionRange(len, len)
    })
  }
})

// ── Submit ──

function submit() {
  if (isExpanded.value && editor.value) {
    // @ts-ignore
    text.value = editor.value.storage.markdown?.getMarkdown?.() || ''
  }

  let trimmed = text.value.trim()
  if (!trimmed) return

  if (selectedProject.value) {
    trimmed = `&${selectedProject.value} ` + trimmed
  }

  const metadata: Record<string, any> = {}

  // Extract parent block ID from prefill content (^uuid at start)
  let parentBlockId: string | undefined
  if (props.prefillContent) {
    const refMatch = props.prefillContent.match(/^\^([0-9a-fA-F-]{36})/)
    if (refMatch) parentBlockId = refMatch[1]
  }

  emit('send', trimmed, metadata, parentBlockId)
  text.value = ''
  selectedProject.value = ''
  showSuggest.value = false
  isExpanded.value = false
  isFullscreen.value = false
  if (editor.value) editor.value.commands.clearContent()
  emit('clear-prefill')
}

// ── File paste ──

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

// Auto-resize
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
  fetchProjects()
  // Apply prefill props on mount
  if (props.prefillProject) selectedProject.value = props.prefillProject
  if (props.prefillContent) {
    text.value = props.prefillContent
    nextTick(() => {
      textareaRef.value?.focus()
      const len = props.prefillContent!.length
      textareaRef.value?.setSelectionRange(len, len)
    })
  }
})
onUnmounted(() => {
  document.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div class="bg-white/92 backdrop-blur-sm px-4 py-3 shrink-0 relative safe-area-bottom border-t border-white/10 shadow-[inset_0_1px_0_rgba(255,255,255,0.1)]" :class="{ 'pb-6': screenSize === 'mobile' }">

    <!-- Suggest popover (compact mode only) -->
    <div
      v-if="showSuggest && !isExpanded"
      class="absolute bottom-full left-4 mb-2 w-72 max-h-48 overflow-y-auto bg-white border border-slate-200 rounded-xl shadow-elevated z-50"
    >
      <div
        v-for="(item, i) in suggestItems"
        :key="item.id"
        class="px-4 py-3 cursor-pointer hover:bg-accent-50 flex items-center gap-3 transition-colors"
        :class="{ 'bg-accent-50': i === suggestIndex }"
        @click="applySuggestion(item)"
        @mouseenter="suggestIndex = i"
      >
        <template v-if="item._create">
          <span class="text-xs text-green-600 w-5 shrink-0 font-mono font-bold">+</span>
          <span class="truncate text-text-primary text-sm">Create "{{ item.name }}"</span>
        </template>
        <template v-else>
          <span class="text-xs text-accent-500 w-5 shrink-0 font-mono font-semibold">{{ triggerChar }}</span>
          <span class="truncate text-text-primary text-sm">{{ item.name }}</span>
          <span v-if="item.type !== 'custom'" class="text-[10px] text-text-muted ml-auto uppercase tracking-wide">{{ item.type }}</span>
        </template>
      </div>
      <div v-if="suggestItems.length === 0" class="px-4 py-3 text-xs text-text-muted italic">
        No matches
      </div>
    </div>

    <!-- ── Compact mode ── -->
    <div v-if="!isExpanded" class="flex items-center gap-3">
      <button
        class="shrink-0 text-text-muted hover:text-accent-600 hover:bg-accent-50 p-2 rounded-lg transition-all flex items-center justify-center"
        :class="{ 'min-w-[44px] min-h-[44px]': screenSize === 'mobile', 'bg-accent-50 text-accent-600': isFocused }"
        @click="expand"
        title="Expand editor"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
        </svg>
      </button>

      <textarea
        ref="textareaRef"
        v-model="text"
        rows="1"
        placeholder="&project @person ^reference — Enter to send"
        class="flex-1 resize-none outline-none text-sm py-3 px-5 rounded-2xl border-0 transition-all placeholder:text-text-muted"
        :class="{
          'min-h-[44px]': screenSize === 'mobile',
          'bg-white ring-2 ring-accent-200/50 shadow-sm': isFocused,
          'bg-gradient-to-r from-surface-100 to-surface-50': !isFocused
        }"
        @focus="onCompactFocus"
        @blur="onCompactBlur"
        @click="onCompactClick"
        @keyup="trackCursor"
        @select="trackCursor"
        @paste="onPaste"
      />

      <!-- Project dropdown -->
      <select v-model="selectedProject" class="shrink-0 text-[11px] border-0 bg-surface-100 rounded-lg px-3 py-2 text-text-secondary outline-none focus:ring-2 focus:ring-accent-200/50 transition-colors max-w-[100px]" :class="{ 'min-h-[44px]': screenSize === 'mobile' }">
        <option value="">No project</option>
        <option v-for="p in projects" :key="p.name" :value="p.name">{{ p.name }}</option>
      </select>

      <button
        class="shrink-0 px-5 py-2.5 bg-accent-600 text-white text-sm rounded-lg hover:bg-accent-700 hover:shadow-md transition-all disabled:opacity-30 font-semibold btn-press"
        :class="{ 'min-h-[44px]': screenSize === 'mobile' }"
        :disabled="!text.trim()"
        @click="submit"
      >
        Send
      </button>
    </div>

    <!-- ── Expanded mode (non-modal: no backdrop overlay) ── -->
    <Teleport to="body">
      <Transition name="editor-expand">
        <div
          v-if="isExpanded"
          class="fixed z-50 bg-white border border-slate-200 shadow-elevated flex flex-col overflow-hidden"
          :class="isFullscreen
            ? 'inset-4 rounded-2xl'
            : 'bottom-4 left-4 right-4 sm:left-1/2 sm:-translate-x-1/2 sm:max-w-2xl sm:w-full sm:top-24 rounded-2xl'"
          :style="isFullscreen ? {} : { maxHeight: '80vh', paddingBottom: 'env(safe-area-inset-bottom)' }"
        >
        <!-- Toolbar header -->
        <div class="flex items-center gap-1 px-4 py-3 border-b border-border-subtle shrink-0">
          <button class="toolbar-btn font-bold" @click="tipTapBold" title="Bold">B</button>
          <button class="toolbar-btn italic" @click="tipTapItalic" title="Italic">I</button>

          <span class="w-px h-5 bg-border-light mx-2" />

          <button class="toolbar-btn" :class="{ 'text-accent-600 bg-accent-50': editor?.isActive('bulletList') }" @click="tipTapBulletList" title="Bullet list">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
          <button class="toolbar-btn" :class="{ 'text-accent-600 bg-accent-50': editor?.isActive('code') }" @click="tipTapCode" title="Inline code">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
            </svg>
          </button>
          <button class="toolbar-btn" :class="{ 'text-accent-600 bg-accent-50': editor?.isActive('blockquote') }" @click="tipTapBlockquote" title="Blockquote">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </button>

          <span class="w-px h-5 bg-border-light mx-2" />

          <button class="toolbar-btn text-accent-600 font-semibold text-xs" @click="tipTapInsertTag('&')" title="Insert project tag">&amp;</button>
          <button class="toolbar-btn text-accent-600 font-semibold text-xs" @click="tipTapInsertTag('@')" title="Insert person mention">@</button>
          <button class="toolbar-btn text-accent-600 font-semibold text-xs" @click="tipTapInsertTag('^')" title="Insert block reference">^</button>

          <div class="flex-1" />

          <button class="toolbar-btn" @click="toggleFullscreen" :title="isFullscreen ? 'Exit fullscreen' : 'Fullscreen'">
            <svg v-if="isFullscreen" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
            <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
            </svg>
          </button>

          <button class="toolbar-btn text-text-muted" @click="collapse" title="Collapse">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </button>
        </div>

        <div class="flex-1 relative">
          <EditorContent :editor="editor" class="flex-1 overflow-hidden h-full" />
          <!-- Slash menu -->
          <div
            v-if="showSlashMenu"
            class="absolute bottom-4 left-4 z-10 flex items-center gap-1 p-1.5 bg-white rounded-xl shadow-elevated border border-slate-200"
          >
            <button
              v-for="(item, i) in slashItems"
              :key="item.id"
              class="flex items-center gap-2 px-3 py-2 text-xs rounded-lg transition-colors whitespace-nowrap"
              :class="i === slashMenuIndex ? 'bg-accent-50 text-accent-700' : 'text-text-secondary hover:bg-surface-100'"
              @click="item.action()"
              @mouseenter="slashMenuIndex = i"
            >
              <span class="w-5 h-5 flex items-center justify-center font-mono text-[11px] font-semibold rounded bg-surface-100">{{ item.icon }}</span>
              {{ item.label }}
            </button>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-4 py-3 border-t border-border-subtle flex items-center justify-between shrink-0 gap-3">
          <div class="flex items-center gap-2">
            <select v-model="selectedProject" class="text-[11px] border-0 bg-surface-100 rounded-lg px-3 py-2 text-text-secondary outline-none focus:ring-2 focus:ring-accent-200/50 transition-colors max-w-[120px]" :class="{ 'min-h-[44px]': screenSize === 'mobile' }">
              <option value="">No project</option>
              <option v-for="p in projects" :key="p.name" :value="p.name">{{ p.name }}</option>
            </select>
            <span class="text-[10px] text-text-muted font-medium">
              {{ charCount }} chars
            </span>
          </div>
          <button
            class="px-6 py-2.5 bg-accent-600 text-white text-sm rounded-lg hover:bg-accent-700 hover:shadow-md transition-all disabled:opacity-30 font-semibold btn-press"
            :class="{ 'min-h-[44px]': screenSize === 'mobile' }"
            :disabled="!text.trim()"
            @click="collapse(); submit()"
          >
            Send
          </button>
        </div>
      </div>
    </Transition>
  </Teleport>
</div>
</template>

<style scoped>
.editor-expand-enter-active,
.editor-expand-leave-active {
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
.editor-expand-enter-from,
.editor-expand-leave-to {
  opacity: 0;
  transform: scale(0.96) translateY(12px);
}

.toolbar-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
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
