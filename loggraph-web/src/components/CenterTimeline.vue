<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'
import type { Block } from '../types'
import BlockCard from './BlockCard.vue'

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
  edit: [id: string]
  'toggle-status': [id: string, current: string]
  'filter-change': [key: string, value: string | undefined]
  archive: [id: string]
  delete: [id: string]
}>()

const hideCompleted = ref(false)
const dateFrom = ref('')
const dateTo = ref('')
const showFilters = ref(false)

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
  return 'bg-blue-400'
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
      nearest = list[i] || null
    }
  })
  if (nearest) {
    tooltipTime.value = formatTime(nearest.created_at)
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

function applyDateFilter() {
  emit('filter-change', 'since', dateFrom.value ? new Date(dateFrom.value).toISOString() : undefined)
  emit('filter-change', 'until', dateTo.value ? new Date(dateTo.value + 'T23:59:59').toISOString() : undefined)
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
  <main class="flex-1 flex flex-col overflow-hidden bg-slate-50/50">
    <!-- Filter bar -->
    <div class="px-3 sm:px-4 py-2 border-b border-slate-200 bg-white flex items-center gap-2 sm:gap-3 shrink-0 flex-wrap">
      <label class="flex items-center gap-1.5 text-xs cursor-pointer select-none text-slate-500">
        <input v-model="hideCompleted" type="checkbox" class="rounded border-slate-300" />
        <span class="hidden sm:inline">Hide done</span>
      </label>
      <span class="text-[11px] text-slate-400 tabular-nums">{{ visibleBlocks.length }}</span>

      <!-- Mobile filter toggle -->
      <button
        v-if="props.screenSize === 'mobile'"
        class="text-xs text-blue-600 ml-auto"
        @click="showFilters = !showFilters"
      >
        {{ showFilters ? 'Hide' : 'Filters' }}
      </button>

      <template v-if="props.screenSize !== 'mobile' || showFilters">
        <div class="w-px h-4 bg-slate-200 hidden sm:block" />
        <select
          class="text-xs border border-slate-200 rounded-md px-2 py-1 bg-white text-slate-500 outline-none focus:border-blue-400 transition-colors"
          :value="props.statusFilter || ''"
          @change="emit('filter-change', 'status', ($event.target as HTMLSelectElement).value || undefined)"
        >
          <option value="">All status</option>
          <option value="active">Active</option>
          <option value="completed">Completed</option>
          <option value="blocked">Blocked</option>
        </select>

        <div class="w-px h-4 bg-slate-200" />
        <select
          class="text-xs border border-slate-200 rounded-md px-2 py-1 bg-white text-slate-500 outline-none focus:border-blue-400 transition-colors max-w-[140px]"
          :value="props.projectFilter || ''"
          @change="emit('filter-change', 'project', ($event.target as HTMLSelectElement).value || undefined)"
        >
          <option value="">All projects</option>
          <option v-for="p in props.projects" :key="p.name" :value="p.name">{{ p.name }}</option>
        </select>

        <div class="w-px h-4 bg-slate-200" />
        <input
          type="date" :value="dateFrom"
          @change="dateFrom = ($event.target as HTMLInputElement).value; applyDateFilter()"
          class="text-xs border border-slate-200 rounded-md px-2 py-1 bg-white text-slate-500 w-32 outline-none focus:border-blue-400 transition-colors"
          title="From date"
        />
        <span class="text-xs text-slate-300">-</span>
        <input
          type="date" :value="dateTo"
          @change="dateTo = ($event.target as HTMLInputElement).value; applyDateFilter()"
          class="text-xs border border-slate-200 rounded-md px-2 py-1 bg-white text-slate-500 w-32 outline-none focus:border-blue-400 transition-colors"
          title="To date"
        />
      </template>
    </div>

    <!-- Timeline -->
    <div id="timeline-scroll" class="flex-1 overflow-y-auto px-4 py-3">
      <div v-if="hasMore" class="text-center pb-3">
        <button
          class="text-xs text-blue-600 hover:text-blue-800 disabled:opacity-40 font-medium"
          :disabled="loading"
          @click="emit('load-more')"
        >
          {{ loading ? 'Loading...' : 'Load older entries' }}
        </button>
      </div>

      <div class="space-y-0">
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
            <div class="absolute top-0 bottom-0 left-1/2 w-0.5 bg-slate-200 -translate-x-1/2" />
            <div
              class="relative z-10 w-2.5 h-2.5 rounded-full mt-3 shrink-0 ring-2 ring-white"
              :class="nodeColor(block.status)"
            />
            <span class="relative z-10 text-[9px] text-slate-400 mt-0.5 tabular-nums leading-none whitespace-nowrap">{{ formatTime(block.created_at) }}</span>
          </div>

          <!-- Card -->
          <div class="flex-1 min-w-0 pb-2">
            <BlockCard
              :block="block"
              :selected="selectedId === block.id"
              :screen-size="screenSize"
              @select="id => emit('select', id)"
              @edit="id => emit('edit', id)"
              @toggle-status="(id, current) => emit('toggle-status', id, current)"
              @archive="id => emit('archive', id)"
              @delete="id => emit('delete', id)"
            />
          </div>
        </div>
      </div>

      <!-- Touch drag tooltip -->
      <Teleport to="body">
        <div
          v-if="tooltipVisible"
          class="fixed z-50 px-3 py-1.5 bg-white/80 backdrop-blur-md rounded-lg shadow-lg text-xs font-mono text-slate-700 border border-white/50 pointer-events-none transition-opacity"
          :style="tooltipStyle()"
        >
          {{ tooltipTime }}
        </div>
      </Teleport>

      <div v-if="visibleBlocks.length === 0 && !loading" class="text-center py-16 text-slate-400 text-sm">
        <div class="text-2xl mb-2">-</div>
        No entries. Type below to get started.
      </div>
    </div>
  </main>
</template>
