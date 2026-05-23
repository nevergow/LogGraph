<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { blocksApi } from '../api/blocks'

const heatmapData = ref<Map<string, number>>(new Map())
const loading = ref(false)
const maxCount = ref(0)

const weeks = computed(() => {
  const now = new Date()
  const result: { week: number; days: { date: string; count: number; level: number }[] }[] = []

  // Start 12 weeks ago from today (Sunday)
  const start = new Date(now)
  start.setDate(start.getDate() - start.getDay() - 11 * 7) // Go back to Sunday 12 weeks ago

  const current = new Date(start)
  let week: { date: string; count: number; level: number }[] = []

  while (current <= now) {
    const key = formatDateKey(current)
    const count = heatmapData.value.get(key) || 0
    const level = maxCount.value > 0 ? Math.min(4, Math.ceil((count / maxCount.value) * 4)) : 0

    week.push({ date: key, count, level })

    if (current.getDay() === 6) {
      result.push({ week: current.getTime(), days: week })
      week = []
    }
    current.setDate(current.getDate() + 1)
  }
  if (week.length > 0) result.push({ week: Date.now(), days: week })

  return result
})

const monthLabels = computed(() => {
  const labels: { label: string; col: number }[] = []
  let lastMonth = -1
  const flat = weeks.value.flatMap(w => w.days)
  flat.forEach((d, i) => {
    const m = new Date(d.date + 'T00:00:00').getMonth()
    if (m !== lastMonth) {
      labels.push({ label: monthName(m), col: Math.floor(i / 7) })
      lastMonth = m
    }
  })
  return labels
})

function formatDateKey(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function monthName(m: number): string {
  return ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'][m]
}

function cellColor(level: number): string {
  if (level === 0) return 'bg-gray-100'
  if (level === 1) return 'bg-emerald-200'
  if (level === 2) return 'bg-emerald-400'
  if (level === 3) return 'bg-emerald-600'
  return 'bg-emerald-800'
}

function cellTitle(date: string, count: number): string {
  return `${date}: ${count} entries`
}

async function fetchData() {
  loading.value = true
  try {
    const now = new Date()
    const since = new Date(now)
    since.setMonth(since.getMonth() - 3)
    const sinceStr = since.toISOString()

    const page = await blocksApi.list({ since: sinceStr, limit: 500 })
    const map = new Map<string, number>()
    let mx = 0
    for (const b of (page.data || [])) {
      const key = b.created_at.slice(0, 10)
      const v = (map.get(key) || 0) + 1
      map.set(key, v)
      if (v > mx) mx = v
    }
    heatmapData.value = map
    maxCount.value = mx
  } catch { /* silent */ }
  loading.value = false
}

onMounted(fetchData)
</script>

<template>
  <div class="px-2 pb-3">
    <div class="text-xs font-medium text-gray-400 px-1 mb-1.5">Activity</div>

    <div v-if="loading" class="flex justify-center py-2">
      <div class="w-3 h-3 border-2 border-gray-300 border-t-gray-500 rounded-full animate-spin" />
    </div>

    <div v-else class="flex flex-col gap-0.5">
      <!-- Month labels -->
      <div class="flex text-[9px] text-gray-400 ml-6">
        <template v-for="(ml, i) in monthLabels" :key="i">
          <span :style="{ marginLeft: i === 0 ? ml.col * 13 + 'px' : (ml.col - (monthLabels[i - 1]?.col || 0)) * 13 + 'px' }">{{ ml.label }}</span>
        </template>
      </div>

      <!-- Grid: day-of-week labels + cells -->
      <div class="flex gap-0.5">
        <!-- Day labels -->
        <div class="flex flex-col gap-0.5 mr-0.5">
          <span class="text-[8px] text-gray-300 leading-none h-2.5 w-5 text-right">M</span>
          <span class="text-[8px] text-gray-300 leading-none h-2.5 w-5 text-right">W</span>
          <span class="text-[8px] text-gray-300 leading-none h-2.5 w-5 text-right">F</span>
        </div>

        <!-- Week columns -->
        <div v-for="(wk, wi) in weeks" :key="wi" class="flex flex-col gap-0.5">
          <div
            v-for="(day, di) in wk.days"
            :key="di"
            class="w-2.5 h-2.5 rounded-[2px] hover:ring-1 hover:ring-brand-300 transition-shadow"
            :class="cellColor(day.level)"
            :title="cellTitle(day.date, day.count)"
          />
        </div>
      </div>

      <!-- Legend -->
      <div class="flex items-center gap-1 mt-1.5 text-[9px] text-gray-400">
        <span>Less</span>
        <div class="w-2.5 h-2.5 rounded-[2px] bg-gray-100" />
        <div class="w-2.5 h-2.5 rounded-[2px] bg-emerald-200" />
        <div class="w-2.5 h-2.5 rounded-[2px] bg-emerald-400" />
        <div class="w-2.5 h-2.5 rounded-[2px] bg-emerald-600" />
        <div class="w-2.5 h-2.5 rounded-[2px] bg-emerald-800" />
        <span>More</span>
      </div>
    </div>
  </div>
</template>
