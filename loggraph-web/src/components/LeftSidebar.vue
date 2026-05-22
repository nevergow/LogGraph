<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Node } from '../types'

const props = defineProps<{
  projects: Node[]
  people: Node[]
  activeProject?: string
}>()

const emit = defineEmits<{
  'select-project': [name: string]
  'select-person': [name: string]
  'clear-filters': []
}>()

const searchQuery = ref('')

const filteredProjects = computed(() => {
  if (!searchQuery.value) return props.projects
  const q = searchQuery.value.toLowerCase()
  return props.projects.filter(p => p.name.toLowerCase().includes(q))
})

const filteredPeople = computed(() => {
  if (!searchQuery.value) return props.people
  const q = searchQuery.value.toLowerCase()
  return props.people.filter(p => p.name.toLowerCase().includes(q))
})

const showSearch = computed(() => props.projects.length + props.people.length > 8)
</script>

<template>
  <aside class="border-r border-slate-200 bg-white flex flex-col shrink-0 overflow-hidden">
    <div class="p-3 border-b border-slate-100 space-y-2">
      <div class="flex items-center justify-between">
        <span class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Filters</span>
        <button
          v-if="activeProject"
          class="text-[10px] text-blue-600 hover:text-blue-800 bg-blue-50 hover:bg-blue-100 px-2 py-0.5 rounded-full transition-colors"
          @click="emit('clear-filters')"
        >
          Clear
        </button>
      </div>
      <input
        v-if="showSearch"
        v-model="searchQuery"
        type="text"
        placeholder="Search..."
        class="w-full text-xs border border-slate-200 rounded-md px-2 py-1.5 outline-none focus:border-blue-400 focus:ring-1 focus:ring-blue-100 transition-all"
      />
    </div>

    <div class="flex-1 overflow-y-auto">
      <!-- Projects / Standards -->
      <div class="px-3 pt-3 pb-1.5 text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
        Projects / Standards
        <span class="ml-1 text-slate-300 font-normal">{{ filteredProjects.length }}</span>
      </div>
      <ul class="px-2">
        <li
          v-for="p in filteredProjects"
          :key="p.id"
          class="group px-2.5 py-1.5 rounded-md text-sm cursor-pointer transition-all truncate flex items-center gap-2"
          :class="activeProject === p.name
            ? 'bg-blue-50 text-blue-700 font-medium'
            : 'text-slate-600 hover:bg-slate-50'"
          @click="emit('select-project', p.name)"
        >
          <span class="w-1.5 h-1.5 rounded-full shrink-0" :class="p.type === 'standard' ? 'bg-orange-400' : 'bg-blue-400'" />
          <span class="truncate">#{{ p.name }}</span>
          <span v-if="p.type === 'standard'" class="text-[10px] text-orange-400 shrink-0 ml-auto opacity-60">std</span>
        </li>
        <li v-if="filteredProjects.length === 0" class="px-2 py-2 text-xs text-slate-400 italic">
          {{ searchQuery ? 'No matches' : 'No projects yet' }}
        </li>
      </ul>

      <!-- People -->
      <div class="px-3 pt-4 pb-1.5 text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
        People
        <span class="ml-1 text-slate-300 font-normal">{{ filteredPeople.length }}</span>
      </div>
      <ul class="px-2 pb-4">
        <li
          v-for="p in filteredPeople"
          :key="p.id"
          class="group px-2.5 py-1.5 rounded-md text-sm cursor-pointer transition-all truncate flex items-center gap-2 text-slate-600 hover:bg-slate-50"
          @click="emit('select-person', p.name)"
        >
          <span class="w-1.5 h-1.5 rounded-full bg-green-400 shrink-0" />
          <span class="truncate">@{{ p.name }}</span>
        </li>
        <li v-if="filteredPeople.length === 0" class="px-2 py-2 text-xs text-slate-400 italic">
          {{ searchQuery ? 'No matches' : 'No people yet' }}
        </li>
      </ul>
    </div>
  </aside>
</template>
