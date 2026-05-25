<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import type { Block } from '../types'

const props = defineProps<{
  block: Block | null
}>()

const emit = defineEmits<{
  save: [id: string, content: string]
  close: []
}>()

const editText = ref('')
const textareaRef = ref<HTMLTextAreaElement | null>(null)

watch(() => props.block, (b) => {
  if (b) {
    editText.value = b.content
    nextTick(() => {
      textareaRef.value?.focus()
      autoResize()
    })
  }
}, { immediate: true })

const projectTag = computed(() => {
  if (!props.block) return ''
  const m = props.block.content.match(/(?:^|\s)[#&]([^\s#&][^\s]*)/)
  return m ? m[1] : ''
})

const wordCount = computed(() => {
  return editText.value.trim().split(/\s+/).filter(Boolean).length
})

function autoResize() {
  const ta = textareaRef.value
  if (!ta) return
  ta.style.height = 'auto'
  ta.style.height = ta.scrollHeight + 'px'
}

function save() {
  if (!props.block) return
  const content = editText.value.trim()
  if (!content) return
  emit('save', props.block.id, content)
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    e.preventDefault()
    emit('close')
  }
  if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
    e.preventDefault()
    save()
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="card-editor">
      <div
        v-if="block"
        class="fixed inset-0 z-50 flex items-end sm:items-center justify-center"
        @keydown="onKeydown"
      >
        <!-- Backdrop -->
        <div
          class="absolute inset-0 bg-black/20 backdrop-blur-sm"
          @click="emit('close')"
        />

        <!-- Panel -->
        <div
          class="relative w-full sm:max-w-xl mx-auto glass-strong rounded-2xl shadow-glass border border-white/50 flex flex-col overflow-hidden"
          :style="{ maxHeight: '90vh' }"
        >
          <!-- Header -->
          <div class="flex items-center justify-between px-5 py-3 border-b border-border-subtle shrink-0">
            <div class="flex items-center gap-2.5 min-w-0">
              <div
                v-if="projectTag"
                class="inline-flex items-center gap-1 text-[10px] bg-brand-50 text-brand-600 px-2.5 py-1 rounded-full font-semibold"
              >
                <span class="w-1 h-1 rounded-full bg-brand-500" />
                &amp;{{ projectTag }}
              </div>
              <span class="text-[11px] text-text-muted">{{ wordCount }} words</span>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <button
                class="text-xs text-text-secondary hover:text-text-primary px-3 py-1.5 rounded-xl hover:bg-surface-100 transition-colors font-medium"
                @click="emit('close')"
              >
                Cancel
              </button>
              <button
                class="text-xs bg-gradient-to-r from-brand-500 to-violet-500 text-white px-4 py-1.5 rounded-xl hover:shadow-lg hover:shadow-brand-500/25 font-semibold transition-all disabled:opacity-40"
                :disabled="!editText.trim()"
                @click="save"
              >
                Save
              </button>
            </div>
          </div>

          <!-- Textarea -->
          <div class="flex-1 overflow-y-auto px-5 py-4">
            <textarea
              ref="textareaRef"
              v-model="editText"
              class="w-full resize-none outline-none text-sm leading-relaxed text-text-primary placeholder:text-text-muted bg-transparent"
              placeholder="Write your log entry..."
              rows="6"
              @input="autoResize"
              @keydown="onKeydown"
            />
          </div>

          <!-- Footer -->
          <div class="px-5 py-2.5 border-t border-border-subtle shrink-0 flex items-center justify-between">
            <span class="text-[10px] text-text-muted">Supports Markdown</span>
            <span class="text-[10px] text-text-muted font-mono">⌘↵ to save · esc to cancel</span>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.card-editor-enter-active {
  transition: transform 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}
.card-editor-leave-active {
  transition: transform 0.25s ease-in;
}
.card-editor-enter-from .relative,
.card-editor-leave-to .relative {
  transform: translateY(100%);
}
.card-editor-enter-from,
.card-editor-leave-to {
  opacity: 0;
}
.card-editor-enter-active .absolute {
  transition: opacity 0.3s ease;
}
.card-editor-leave-active .absolute {
  transition: opacity 0.2s ease;
}
</style>