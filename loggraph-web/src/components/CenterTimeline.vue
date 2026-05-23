<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'
import type { Block } from '../types'
import BlockCard from './BlockCard.vue'
import FilterBar from './FilterBar.vue'
import SkeletonCard from './SkeletonCard.vue'

const props = defineProps<{
  blocks: Block[]
  loading: boolean
  hasMore: boolean
  selectedId: string | null
  screenSize?: 'mobile' | 'tablet' | 'desktop'
  projects?: { name: string; id: string }[]
  statusFilter?: string
  projectFilter?: string
  sinceDate?: string
  untilDate?: string
}>()

const emit = defineEmits<{
  'load-more': []
  select: [id: string]
  'toggle-status': [id: string, current: string]
  'filter-change': [key: string, value: string | undefined]
  archive: [id: string]
  delete: [id: string]
  'request-edit': [id: string]
}>()

const hideCompleted = ref(false)

// ── Touch drag tooltip ──
const tooltipVisible = ref(false)
const tooltipTime = ref('')
const tooltipY = ref(0)
let tooltipTimer: ReturnType<typeof setTimeout> | null = null

onUnmounted(() => {
  if (tooltipTimer) clearTimeout(tooltipTimer)
})

function nodeColor(s: string): string {
  if (s === 'completed') return 'bg-emerald-400'
  if (s === 'blocked') return 'bg-red-400'
  return 'bg-brand-400'
}

function onTimelineTouchMove(e: TouchEvent, list: Block[]) {
  const touch = e.touches[0]
  tooltipY.value = touch.clientY
  const container = document.getElementById('timeline-scroll')
  if (!container) return
  const cards = container.querySelectorAll('.timeline-card')
  let nearest: Block | null = null
  let minDist = Infinity
  cards.forEach((card, i) => {
    const rect = card.getBoundingClientRect()
    const midY = rect.top + rect.height / 2
    const dist = Math.abs(touch.clientY - midY)
    if (dist < minDist) {
      minDist = dist
      nearest = list[i]
    }
  })
  if (nearest) {
    tooltipTime.value = formatTime((nearest as Block).created_at)
    tooltipVisible.value = true
  }
}

function onTimelineTouchEnd() {
  if (tooltipTimer) clearTimeout(tooltipTimer)
  tooltipTimer = setTimeout(() => {
    tooltipVisible.value = false
  }, 1000)
}

function tooltipStyle(): Record<string, string> {
  const maxY = typeof window !== 'undefined' ? window.innerHeight - 60 : 600
  return { top: Math.min(tooltipY.value, maxY) + 'px', right: '12px' }
}

const visibleBlocks = computed(() => {
  if (!hideCompleted.value) return props.blocks
  return props.blocks.filter(b => b.status !== 'completed')
})

function formatTime(ts: string): string {
  const d = new Date(ts)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}
</script>

<template>
  <main class="flex-1 flex flex-col overflow-hidden bg-gray-50/50">
    <!-- Filter bar -->
    <div class="px-3 sm:px-4 py-2 border-b border-gray-200 bg-white flex items-center gap-2 sm:gap-3 shrink-0 flex-wrap">
      <FilterBar
        :count="visibleBlocks.length"
        :hide-completed="hideCompleted"
        :status-filter="props.statusFilter"
        :since-date="props.sinceDate ? props.sinceDate.slice(0, 10) : ''"
        :until-date="props.untilDate ? props.untilDate.slice(0, 10) : ''"
        :screen-size="screenSize"
        @update:hide-completed="hideCompleted = $event"
        @filter-change="(key, value) => emit('filter-change', key, value)"
      />

      <div class="w-px h-4 bg-gray-200" />
      <select
        class="text-xs border border-gray-200 rounded-sm px-2 py-1 bg-white text-gray-500 outline-none focus:border-brand-300 transition-colors max-w-[140px]"
        :value="props.projectFilter || ''"
        @change="emit('filter-change', 'project', ($event.target as HTMLSelectElement).value || undefined)"
      >
        <option value="">All projects</option>
        <option v-for="p in props.projects" :key="p.name" :value="p.name">{{ p.name }}</option>
      </select>
    </div>

    <!-- Timeline -->
    <div id="timeline-scroll" class="flex-1 overflow-y-auto px-4 py-3">
      <div v-if="hasMore" class="text-center pb-3">
        <button
          class="text-xs text-brand-600 hover:text-brand-800 disabled:opacity-40 font-medium"
          :disabled="loading"
          @click="emit('load-more')"
        >
          {{ loading ? 'Loading...' : 'Load older entries' }}
        </button>
      </div>

      <!-- Skeleton loading (initial load only) -->
      <SkeletonCard v-if="loading && visibleBlocks.length === 0" :count="4" />

      <TransitionGroup v-else name="card-list" tag="div" class="space-y-0">
        <div
          v-for="block in visibleBlocks"
          :key="block.id"
          class="timeline-card flex gap-3"
        >
          <!-- Vertical timeline axis -->
          <div
            class="w-14 shrink-0 flex flex-col items-center relative cursor-pointer"
            @touchstart.passive="() => {}"
            @touchmove.prevent="onTimelineTouchMove($event, visibleBlocks)"
            @touchend="onTimelineTouchEnd"
          >
            <div class="absolute top-0 bottom-0 left-1/2 w-0.5 bg-gray-200 -translate-x-1/2" />
            <div
              class="relative z-10 w-2.5 h-2.5 rounded-full mt-3 shrink-0 ring-2 ring-white"
              :class="nodeColor(block.status)"
            />
            <span class="relative z-10 text-[9px] text-gray-400 mt-0.5 tabular-nums leading-none whitespace-nowrap">{{ formatTime(block.created_at) }}</span>
          </div>

          <!-- Card -->
          <div class="flex-1 min-w-0 pb-2">
            <BlockCard
              :block="block"
              :selected="selectedId === block.id"
              :screen-size="screenSize"
              @select="id => emit('select', id)"
              @toggle-status="(id, current) => emit('toggle-status', id, current)"
              @archive="id => emit('archive', id)"
              @delete="id => emit('delete', id)"
              @request-edit="id => emit('request-edit', id)"
            />
          </div>
        </div>
      </TransitionGroup>

      <!-- Touch drag tooltip -->
      <Teleport to="body">
        <div
          v-if="tooltipVisible"
          class="fixed z-50 px-3 py-1.5 bg-gray-900 rounded-sm shadow-elevated text-xs font-mono text-white pointer-events-none transition-opacity"
          :style="tooltipStyle()"
        >
          {{ tooltipTime }}
        </div>
      </Teleport>

      <div v-if="visibleBlocks.length === 0 && !loading" class="text-center py-20 text-gray-400">
        <svg class="w-10 h-10 text-gray-300 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <div class="text-sm font-medium text-gray-500 mb-1">No entries yet</div>
        <div class="text-xs text-gray-400">Type a message below to start logging.</div>
      </div>
    </div>
  </main>
</template>
