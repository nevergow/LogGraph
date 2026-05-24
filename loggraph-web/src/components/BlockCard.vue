<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import type { Block } from '../types'
import { renderMarkdown, extractTitle } from '../composables/useMarkdown'

const props = defineProps<{
  block: Block
  selected: boolean
  screenSize?: 'mobile' | 'tablet' | 'desktop'
  draggable?: boolean
}>()

const emit = defineEmits<{
  select: [id: string]
  'toggle-status': [id: string, current: string]
  archive: [id: string]
  delete: [id: string]
  'request-edit': [id: string]
  'request-followup': [block: Block]
  'request-graph': [id: string]
}>()

const viewMode = ref<'compact' | 'expanded'>('compact')
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
  if (s === 'completed') return 'border-l-emerald-500'
  if (s === 'blocked') return 'border-l-red-500'
  return 'border-l-blue-500'
}

function statusBg(s: string): string {
  if (s === 'blocked') return 'bg-red-50/30'
  if (s === 'completed') return ''
  return ''
}

function renderContent(text: string): string {
  return renderMarkdown(text)
}

// Related blocks get amber signal
const hasRelations = computed(() =>
  props.block.referenced_by && props.block.referenced_by.length > 0
)

// ── Swipe (mobile only) ──
const swipeX = ref(0)
const swiping = ref(false)
let touchStartX = 0
let touchStartY = 0
let isDeadZone = false

function onTouchStart(e: TouchEvent) {
  if (props.screenSize !== 'mobile') return
  touchStartX = e.touches[0].clientX
  touchStartY = e.touches[0].clientY
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

// ── Drag (Phase 1.4) ──
function onDragStart(e: DragEvent) {
  if (!props.draggable) return
  const match = props.block.content.match(/(?:^|\s)[&]([^\s&][^\s]*)/)
  const project = match ? match[1] : 'Unfiled'
  e.dataTransfer!.effectAllowed = 'move'
  e.dataTransfer!.setData('application/x-block-id', props.block.id)
  e.dataTransfer!.setData('application/x-from-project', project)
}
</script>

<template>
  <div
    class="group relative overflow-hidden rounded-xl"
    :class="{ 'cursor-grab active:cursor-grabbing': draggable }"
    :draggable="draggable"
    @dragstart="onDragStart"
    @touchstart.passive="onTouchStart"
    @touchmove="onTouchMove"
    @touchend="onTouchEnd"
  >
    <!-- Swipe action backgrounds (mobile only) -->
    <div v-if="screenSize === 'mobile'" class="absolute inset-0 flex">
      <div
        class="flex-1 flex items-center justify-start pl-5 bg-danger rounded-l-2xl transition-opacity"
        :style="{ opacity: swipeLeftOpacity }"
      >
        <span class="text-white text-sm font-semibold">Delete</span>
      </div>
      <div
        class="flex-1 flex items-center justify-end pr-5 bg-success rounded-r-2xl transition-opacity"
        :style="{ opacity: swipeRightOpacity }"
      >
        <span class="text-white text-sm font-semibold">Done</span>
      </div>
    </div>

    <!-- Card — card-surface replaces glass -->
    <div
      class="card-surface rounded-xl cursor-pointer transition-all duration-200 hover:shadow-card-hover relative z-10 border-l-[3px]"
      :class="[
        hasRelations ? 'block-related' : borderColor(block.status),
        statusBg(block.status),
        {
          'bg-blue-50/70 border-l-blue-500': selected,
          'block-done': block.status === 'completed',
          'block-done-transition': true,
          'p-4': viewMode !== 'compact',
          'p-3': viewMode === 'compact',
        }
      ]"
      :style="swipeStyle"
      @click="emit('select', block.id)"
      @dblclick="viewMode === 'compact' ? viewMode = 'expanded' : (viewMode === 'expanded' ? viewMode = 'compact' : null)"
    >
      <!-- Desktop: three-dot dropdown -->
      <div v-if="viewMode !== 'compact'" class="flex justify-end mb-3">
        <div
          v-if="screenSize === 'desktop'"
          class="opacity-0 group-hover:opacity-100 transition-opacity duration-150 relative"
          @click.stop
        >
          <button
            class="p-2 rounded-lg text-text-muted hover:text-text-primary hover:bg-surface-100 transition-colors"
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
            class="absolute right-0 top-full mt-2 bg-white border border-slate-200 rounded-xl shadow-elevated z-50 py-2 min-w-[140px]"
            @click.stop
          >
            <button
              class="w-full text-left px-4 py-2.5 text-sm text-text-primary hover:bg-accent-50 hover:text-accent-700 transition-colors flex items-center gap-3"
              @click="emit('request-edit', block.id); closeMoreMenu()"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
              Edit
            </button>
            <button
              class="w-full text-left px-4 py-2.5 text-sm text-text-primary hover:bg-success-light hover:text-success transition-colors flex items-center gap-3"
              @click="emit('toggle-status', block.id, block.status); closeMoreMenu()"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
              {{ block.status === 'completed' ? 'Reopen' : 'Complete' }}
            </button>
            <div class="border-t border-border-subtle my-2" />
            <button
              class="w-full text-left px-4 py-2.5 text-sm text-danger hover:bg-danger-light transition-colors flex items-center gap-3"
              @click="emit('delete', block.id); closeMoreMenu()"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
              Delete
            </button>
          </div>
        </div>
      </div>

      <!-- ── Compact mode ── -->
      <div v-if="viewMode === 'compact'" class="flex items-center gap-3">
        <span class="text-sm text-text-primary truncate flex-1 font-medium" :class="{ 'text-text-muted line-through': block.status === 'completed' }">{{ extractTitle(block.content) }}</span>
        <svg class="w-4 h-4 text-text-muted shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </div>

      <!-- ── Expanded mode ── -->
      <template v-if="viewMode === 'expanded'">
        <div class="card-body expanded">
          <div class="card-body-inner">
            <div
              class="text-sm leading-relaxed text-text-primary prose prose-sm max-w-none"
              v-html="renderContent(block.content)"
            />
          </div>
        </div>
        <div class="flex items-center justify-between mt-3 pt-2 border-t border-border-subtle">
          <div class="flex items-center gap-3">
            <button
              class="text-xs text-accent-600 hover:text-accent-800 font-semibold transition-colors"
              @click.stop="emit('request-edit', block.id)"
            >
              Edit
            </button>
            <button
              class="text-xs text-text-muted hover:text-text-primary font-medium transition-colors"
              @click.stop="viewMode = 'compact'"
            >
              Collapse
            </button>
            <button
              class="text-xs text-text-muted hover:text-accent-600 font-medium transition-colors"
              @click.stop="emit('request-graph', block.id)"
            >
              Connections
            </button>
          </div>
          <!-- Follow-up button (Phase 2.1) -->
          <button
            class="flex items-center gap-1 px-3 py-1.5 text-xs font-semibold text-accent-600 bg-accent-50 hover:bg-accent-100 rounded-lg transition-colors"
            :class="screenSize === 'mobile' ? 'min-w-[44px] min-h-[44px] justify-center' : ''"
            @click.stop="emit('request-followup', block)"
            title="Add follow-up"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span v-if="screenSize !== 'mobile'">Follow-up</span>
          </button>
        </div>
      </template>
    </div>
  </div>
</template>
