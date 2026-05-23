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
  <div class="px-4 py-2.5 border-b border-border-subtle bg-white/60 backdrop-blur-sm flex items-center gap-3 shrink-0 flex-wrap">
    <label class="flex items-center gap-2 text-xs cursor-pointer select-none text-text-secondary">
      <input
        type="checkbox"
        :checked="hideCompleted"
        class="w-4 h-4 rounded-lg border-border-light text-brand-500 focus:ring-brand-200"
        @change="emit('update:hideCompleted', ($event.target as HTMLInputElement).checked)"
      />
      <span class="hidden sm:inline">Hide done</span>
    </label>
    <span class="text-[11px] text-text-muted tabular-nums font-medium">{{ count }}</span>

    <div class="w-px h-4 bg-border-light hidden sm:block" />
    <select
      class="text-xs border-0 bg-surface-100 rounded-xl px-3 py-1.5 text-text-secondary outline-none focus:ring-2 focus:ring-brand-200/50 transition-colors"
      :value="statusFilter || ''"
      @change="emit('filter-change', 'status', ($event.target as HTMLSelectElement).value || undefined)"
    >
      <option value="">All status</option>
      <option value="active">Active</option>
      <option value="completed">Completed</option>
      <option value="blocked">Blocked</option>
    </select>

    <div class="w-px h-4 bg-border-light" />
    <input
      type="date"
      :value="sinceDate || ''"
      class="text-xs border-0 bg-surface-100 rounded-xl px-3 py-1.5 text-text-secondary w-32 outline-none focus:ring-2 focus:ring-brand-200/50 transition-colors"
      title="From date"
      @change="applyDateFilter(($event.target as HTMLInputElement).value, untilDate || '')"
    />
    <span class="text-xs text-text-muted">-</span>
    <input
      type="date"
      :value="untilDate || ''"
      class="text-xs border-0 bg-surface-100 rounded-xl px-3 py-1.5 text-text-secondary w-32 outline-none focus:ring-2 focus:ring-brand-200/50 transition-colors"
      title="To date"
      @change="applyDateFilter(sinceDate || '', ($event.target as HTMLInputElement).value)"
    />
  </div>
</template>
