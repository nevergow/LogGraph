<script setup lang="ts">
import { ref, computed } from 'vue'
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

function statusBadge(s: string): string {
  if (s === 'completed') return 'bg-emerald-100 text-emerald-700'
  if (s === 'blocked') return 'bg-red-100 text-red-700'
  return 'bg-blue-100 text-blue-700'
}

function statusLabel(s: string): string {
  if (s === 'completed') return 'Done'
  if (s === 'blocked') return 'Blocked'
  return 'Active'
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
  isDeadZone = touchStartX < 25
  swiping.value = true
}

function onTouchMove(e: TouchEvent) {
  if (!swiping.value || isDeadZone) return
  const dx = e.touches[0].clientX - touchStartX
  const dy = Math.abs(e.touches[0].clientY - touchStartY)
  // Require horizontal-dominant gesture
  if (Math.abs(dx) < Math.abs(dy) * 1.2) return
  swipeX.value = dx
}

function onTouchEnd() {
  if (!swiping.value) return
  swiping.value = false
  if (isDeadZone) { swipeX.value = 0; return }
  const threshold = 80
  if (swipeX.value > threshold) {
    // Right swipe → complete + archive
    emit('toggle-status', props.block.id, props.block.status)
  } else if (swipeX.value < -threshold) {
    // Left swipe → delete
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
  return Math.min(Math.abs(swipeX.value) / 80, 1)
})

const swipeRightOpacity = computed(() => {
  if (swipeX.value <= 0 || props.screenSize !== 'mobile') return 0
  return Math.min(Math.abs(swipeX.value) / 80, 1)
})
</script>

<template>
  <div
    ref="cardRef"
    class="group relative overflow-hidden rounded-xl"
    @touchstart.passive="onTouchStart"
    @touchmove="onTouchMove"
    @touchend="onTouchEnd"
  >
    <!-- Swipe action backgrounds (mobile only) -->
    <div v-if="screenSize === 'mobile'" class="absolute inset-0 flex">
      <!-- Left: Delete (red) -->
      <div
        class="flex-1 flex items-center justify-start pl-4 bg-red-500 rounded-l-xl transition-opacity"
        :style="{ opacity: swipeLeftOpacity }"
      >
        <span class="text-white text-sm font-medium">Delete</span>
      </div>
      <!-- Right: Complete (green) + Archive (amber) -->
      <div
        class="flex-1 flex flex-col rounded-r-xl overflow-hidden transition-opacity"
        :style="{ opacity: swipeRightOpacity }"
      >
        <div class="flex-1 flex items-center justify-end pr-4 bg-emerald-500">
          <span class="text-white text-sm font-medium">Done</span>
        </div>
        <div class="flex-1 flex items-center justify-end pr-4 bg-amber-500">
          <span class="text-white text-sm font-medium">Archive</span>
        </div>
      </div>
    </div>

    <!-- Card -->
    <div
      class="bg-white rounded-xl shadow-sm px-4 py-3 cursor-pointer transition-all duration-150 hover:shadow-md relative z-10"
      :class="{
        'ring-2 ring-blue-400 shadow-md': selected,
        'block-done': block.status === 'completed'
      }"
      :style="[swipeStyle, screenSize === 'mobile' ? {} : {}]"
      @click="emit('select', block.id)"
      @dblclick="emit('edit', block.id)"
    >
      <!-- Card header -->
      <div class="flex items-center justify-between mb-2">
        <div class="flex items-center gap-2">
          <span
            class="text-[11px] px-2 py-0.5 rounded-md font-medium cursor-pointer transition-all hover:ring-2 hover:ring-offset-1"
            :class="statusBadge(block.status)"
            :title="'Click to cycle: Active → Done → Blocked → Active'"
            @click.stop="emit('toggle-status', block.id, block.status)"
          >
            {{ statusLabel(block.status) }}
          </span>
          <span
            v-if="block.metadata?.priority === 'high'"
            class="text-[11px] px-1.5 py-0.5 rounded-md font-medium bg-red-50 text-red-600"
          >
            High
          </span>
          <span class="text-[11px] text-slate-400 font-medium">{{ block.user_id }}</span>
        </div>

        <!-- Desktop hover toolbar -->
        <div
          v-if="screenSize === 'desktop'"
          class="opacity-0 group-hover:opacity-100 transition-opacity duration-150 flex items-center gap-0.5"
          @click.stop
        >
          <button
            class="p-1 rounded hover:bg-emerald-50 text-slate-400 hover:text-emerald-600 transition-colors"
            title="Complete"
            @click="emit('toggle-status', block.id, block.status)"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
          </button>
          <button
            class="p-1 rounded hover:bg-amber-50 text-slate-400 hover:text-amber-600 transition-colors"
            title="Archive"
            @click="emit('archive', block.id)"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
            </svg>
          </button>
          <button
            class="p-1 rounded hover:bg-red-50 text-slate-400 hover:text-red-600 transition-colors"
            title="Delete"
            @click="emit('delete', block.id)"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        </div>
      </div>

      <!-- Markdown content -->
      <div
        class="text-sm sm:text-base leading-relaxed text-slate-700 prose prose-sm max-w-none"
        :class="{ 'max-h-[300px] overflow-hidden relative': !expanded && block.content.length > 200 }"
        v-html="renderContent(block.content)"
      />
      <button
        v-if="block.content.length > 200"
        class="text-xs text-blue-600 hover:text-blue-800 mt-1 font-medium"
        @click.stop="expanded = !expanded"
      >
        {{ expanded ? '收起' : '展开阅读' }}
      </button>
    </div>
  </div>
</template>
