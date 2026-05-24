<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
  currentView: 'project' | 'timeline'
  showFilter: boolean
}>()

const emit = defineEmits<{
  'toggle-view': []
  'toggle-filter': []
  'open-search': []
}>()

const expanded = ref(false)

function toggleExpanded() {
  expanded.value = !expanded.value
}
</script>

<template>
  <div class="fixed bottom-20 left-3 right-3 z-30 safe-area-bottom">
    <!-- Collapsed: floating pill -->
    <div
      v-if="!expanded"
      class="flex items-center justify-center"
    >
      <button
        class="flex items-center gap-2 px-5 py-2.5 bg-white/95 backdrop-blur-md border border-slate-200 rounded-full shadow-elevated text-xs font-medium text-text-secondary active:scale-95 transition-transform"
        @click="toggleExpanded"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
        </svg>
        <span>Quick Actions</span>
      </button>
    </div>

    <!-- Expanded: action bar -->
    <div
      v-else
      class="flex items-center gap-2 p-2 bg-white/95 backdrop-blur-md border border-slate-200 rounded-2xl shadow-elevated"
    >
      <button
        class="flex-1 flex items-center justify-center gap-1.5 py-2.5 text-xs font-medium rounded-xl transition-colors"
        :class="currentView === 'project' ? 'bg-accent-50 text-accent-700' : 'text-text-secondary'"
        @click="emit('toggle-view')"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
        Projects
      </button>
      <button
        class="flex-1 flex items-center justify-center gap-1.5 py-2.5 text-xs font-medium rounded-xl transition-colors"
        :class="currentView === 'timeline' ? 'bg-accent-50 text-accent-700' : 'text-text-secondary'"
        @click="emit('toggle-view')"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        Timeline
      </button>
      <div class="w-px h-6 bg-slate-200" />
      <button
        class="flex items-center justify-center w-10 h-10 rounded-xl text-text-secondary transition-colors"
        :class="showFilter ? 'bg-accent-50 text-accent-600' : ''"
        @click="emit('toggle-filter')"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
        </svg>
      </button>
      <button
        class="flex items-center justify-center w-10 h-10 rounded-xl text-text-secondary hover:bg-surface-100 transition-colors"
        title="Search"
        @click="emit('open-search')"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      </button>
      <button
        class="flex items-center justify-center w-10 h-10 rounded-xl text-text-muted hover:bg-surface-100 transition-colors"
        @click="toggleExpanded"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <!-- Backdrop when expanded -->
    <div
      v-if="expanded"
      class="fixed inset-0 -z-10"
      @click="expanded = false"
    />
  </div>
</template>
