<script setup lang="ts">
defineProps<{
  count: number
  hideCompleted?: boolean
  statusFilter?: string
  sinceDate?: string
  untilDate?: string
  screenSize?: 'mobile' | 'tablet' | 'desktop'
}>()

const emit = defineEmits<{
  'update:hideCompleted': [value: boolean]
  'filter-change': [key: string, value: string | undefined]
}>()

function applyDateFilter(since: string, until: string) {
  emit('filter-change', 'since', since ? new Date(since).toISOString() : undefined)
  emit('filter-change', 'until', until ? new Date(until + 'T23:59:59').toISOString() : undefined)
}
</script>

<template>
  <div class="px-3 sm:px-4 py-2 border-b border-gray-200 bg-white flex items-center gap-2 sm:gap-3 shrink-0 flex-wrap">
    <label class="flex items-center gap-1.5 text-xs cursor-pointer select-none text-gray-500">
      <input
        type="checkbox"
        :checked="hideCompleted"
        class="rounded-sm border-gray-300"
        @change="emit('update:hideCompleted', ($event.target as HTMLInputElement).checked)"
      />
      <span class="hidden sm:inline">Hide done</span>
    </label>
    <span class="text-[11px] text-gray-400 tabular-nums">{{ count }}</span>

    <div class="w-px h-4 bg-gray-200 hidden sm:block" />
    <select
      class="text-xs border border-gray-200 rounded-sm px-2 py-1 bg-white text-gray-500 outline-none focus:border-brand-300 transition-colors"
      :value="statusFilter || ''"
      @change="emit('filter-change', 'status', ($event.target as HTMLSelectElement).value || undefined)"
    >
      <option value="">All status</option>
      <option value="active">Active</option>
      <option value="completed">Completed</option>
      <option value="blocked">Blocked</option>
    </select>

    <div class="w-px h-4 bg-gray-200" />
    <input
      type="date"
      :value="sinceDate || ''"
      class="text-xs border border-gray-200 rounded-sm px-2 py-1 bg-white text-gray-500 w-32 outline-none focus:border-brand-300 transition-colors"
      title="From date"
      @change="applyDateFilter(($event.target as HTMLInputElement).value, untilDate || '')"
    />
    <span class="text-xs text-gray-300">-</span>
    <input
      type="date"
      :value="untilDate || ''"
      class="text-xs border border-gray-200 rounded-sm px-2 py-1 bg-white text-gray-500 w-32 outline-none focus:border-brand-300 transition-colors"
      title="To date"
      @change="applyDateFilter(sinceDate || '', ($event.target as HTMLInputElement).value)"
    />
  </div>
</template>
