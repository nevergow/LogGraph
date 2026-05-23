<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import type { Block } from '../types'
import { renderMarkdown } from '../composables/useMarkdown'

const props = defineProps<{
  block: Block
  selected: boolean
  screenSize?: 'mobile' | 'tablet' | 'desktop'
}>()

const emit = defineEmits<{
  select: [id: string]
  edit: [id: string]
  'toggle-status': [id: string, current: string]
  archive: [id: string]
  delete: [id: string]
}>()

const expanded = ref(false)
const showMoreMenu = ref(false)

function toggleMoreMenu() {
  showMoreMenu.value = !showMoreMenu.value
}
function closeMoreMenu() {
  showMoreMenu.value = false
}
function onDocumentClick() {
  if (showMoreMenu.value) showMoreMenu.value = false
}

onMounted(() => {
  document.addEventListener('click', onDocumentClick)
})
onUnmounted(() => {
  document.removeEventListener('click', onDocumentClick)
})

function borderColor(s: string): string {
  if (s === 'completed') return 'border-l-emerald-400'
  if (s === 'blocked') return 'border-l-red-400'
  return 'border-l-brand-400'
}

function renderContent(text: string): string {
  return renderMarkdown(text)
}

// ── Swipe (mobile only) ──
const cardRef = ref<HTMLElement | null>(null)
const swipeX = ref(0)
const swiping = ref(false)
let touchStartX = 0
let touchStartY = 0
let touchStartTime = 0
let isDeadZone = false

function onTouchStart(e: TouchEvent) {
  if (props.screenSize !== 'mobile') return
  touchStartX = e.touches[0].clientX
  touchStartY = e.touches[0].clientY
  touchStartTime = Date.now()
  isDeadZone = touchStartX < 15
  swiping.value = true
}

function onTouchMove(e: TouchEvent) {
  if (!swiping.value || isDeadZone) return
  const dx = e.touches[0].clientX - touchStartX
  const dy = Math.abs(e.touches[0].clientY - touchStartY)
  if (Math.abs(dx) < Math.abs(dy) * 1.2) return
  swipeX.value = dx
}

function onTouchEnd() {
  if (!swiping.value) return
  swiping.value = false
  if (isDeadZone) { swipeX.value = 0; return }
  const deleteThreshold = 45
  const doneThreshold = 60
  if (swipeX.value > doneThreshold) {
    emit('toggle-status', props.block.id, props.block.status)
  } else if (swipeX.value < -deleteThreshold) {
    emit('delete', props.block.id)
  }
  swipeX.value = 0
}

const swipeStyle = computed(() => {
  if (!swipeX.value) return {}
  return {
    transform: `translateX(${swipeX.value}px)`,
    transition: swiping.value ? 'none' : 'transform 0.2s ease',
  }
})

const swipeLeftOpacity = computed(() => {
  if (swipeX.value >= 0 || props.screenSize !== 'mobile') return 0
  const d = Math.abs(swipeX.value)
  if (d < 5) return 0
  return Math.min(d / 50, 1)
})

const swipeRightOpacity = computed(() => {
  if (swipeX.value <= 0 || props.screenSize !== 'mobile') return 0
  return Math.min(Math.abs(swipeX.value) / 60, 1)
})
</script>

<template>
  <div
    ref="cardRef"
    class="group relative overflow-hidden rounded-md"
    @touchstart.passive="onTouchStart"
    @touchmove="onTouchMove"
    @touchend="onTouchEnd"
  >
    <!-- Swipe action backgrounds (mobile only) -->
    <div v-if="screenSize === 'mobile'" class="absolute inset-0 flex">
      <div
        class="flex-1 flex items-center justify-start pl-4 bg-red-500 rounded-l-md transition-opacity"
        :style="{ opacity: swipeLeftOpacity }"
      >
        <span class="text-white text-sm font-medium">Delete</span>
      </div>
      <div
        class="flex-1 flex items-center justify-end pr-4 bg-emerald-500 rounded-r-md transition-opacity"
        :style="{ opacity: swipeRightOpacity }"
      >
        <span class="text-white text-sm font-medium">Done</span>
      </div>
    </div>

    <!-- Card -->
    <div
      class="bg-white rounded-md px-4 py-3 cursor-pointer transition-all duration-150 hover:shadow-elevated relative z-10 border-l-[3px]"
      :class="[
        borderColor(block.status),
        {
          'bg-brand-50/50 border-l-brand-500': selected,
          'block-done': block.status === 'completed',
        }
      ]"
      :style="[swipeStyle, screenSize === 'mobile' ? {} : {}]"
      @click="emit('select', block.id)"
      @dblclick="emit('edit', block.id)"
    >
      <!-- Desktop: three-dot dropdown (hover only) -->
      <div class="flex justify-end mb-2">
        <div
          v-if="screenSize === 'desktop'"
          class="opacity-0 group-hover:opacity-100 transition-opacity duration-150 relative"
          @click.stop
        >
          <button
            class="p-1 rounded-sm text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
            title="More actions"
            @click="toggleMoreMenu"
          >
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
              <circle cx="12" cy="5" r="1.5" />
              <circle cx="12" cy="12" r="1.5" />
              <circle cx="12" cy="19" r="1.5" />
            </svg>
          </button>
          <div
            v-if="showMoreMenu"
            class="absolute right-0 top-full mt-1 bg-white border border-gray-200 rounded-md shadow-elevated z-50 py-1 min-w-[120px]"
            @click.stop
          >
            <button
              class="w-full text-left px-3 py-1.5 text-sm text-gray-700 hover:bg-brand-50 hover:text-brand-700 transition-colors flex items-center gap-2"
              @click="emit('edit', block.id); closeMoreMenu()"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
              Edit
            </button>
            <button
              class="w-full text-left px-3 py-1.5 text-sm text-gray-700 hover:bg-emerald-50 hover:text-emerald-700 transition-colors flex items-center gap-2"
              @click="emit('toggle-status', block.id, block.status); closeMoreMenu()"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
              {{ block.status === 'completed' ? 'Reopen' : 'Complete' }}
            </button>
            <div class="border-t border-gray-100 my-1" />
            <button
              class="w-full text-left px-3 py-1.5 text-sm text-red-600 hover:bg-red-50 transition-colors flex items-center gap-2"
              @click="emit('delete', block.id); closeMoreMenu()"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
              Delete
            </button>
          </div>
        </div>
      </div>

      <!-- Markdown content -->
      <div
        class="text-sm leading-relaxed text-gray-700 prose prose-sm max-w-none"
        :class="{ 'max-h-[300px] overflow-hidden relative': !expanded && block.content.length > 200 }"
        v-html="renderContent(block.content)"
      />
      <button
        v-if="block.content.length > 200"
        class="text-xs text-brand-600 hover:text-brand-800 mt-1 font-medium"
        @click.stop="expanded = !expanded"
      >
        {{ expanded ? '收起' : '展开阅读' }}
      </button>
    </div>
  </div>
</template>
