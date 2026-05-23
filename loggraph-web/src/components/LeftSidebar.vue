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

const collapsed = ref(false)
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

const showSearch = computed(() => true)
</script>

<template>
  <aside class="border-r border-gray-200 bg-white flex flex-col shrink-0 overflow-hidden" :class="collapsed ? 'items-center' : ''">
    <!-- Collapse toggle -->
    <div class="p-2 border-b border-gray-100 flex" :class="collapsed ? 'justify-center' : 'justify-between items-center'">
      <button
        v-if="!collapsed && activeProject"
        class="text-[10px] text-brand-600 hover:text-brand-800 bg-brand-50 hover:bg-brand-100 px-2 py-0.5 rounded-full transition-colors"
        @click="emit('clear-filters')"
      >
        Clear
      </button>
      <span v-if="!collapsed && !activeProject" />
      <button
        class="text-gray-400 hover:text-gray-600 hover:bg-gray-100 p-1 rounded-sm transition-colors shrink-0"
        :title="collapsed ? 'Expand sidebar' : 'Collapse sidebar'"
        @click="collapsed = !collapsed"
      >
        <svg v-if="collapsed" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 5l7 7-7 7M5 5l7 7-7 7" />
        </svg>
        <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
        </svg>
      </button>
    </div>

    <template v-if="collapsed">
      <!-- Collapsed: icon-only strip -->
      <div class="flex-1 flex flex-col items-center gap-4 pt-3 overflow-y-auto">
        <button
          class="p-2 rounded-sm hover:bg-brand-50 text-gray-400 hover:text-brand-600 transition-colors"
          title="Projects"
          @click="collapsed = false"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
        </button>
        <button
          class="p-2 rounded-sm hover:bg-brand-50 text-gray-400 hover:text-brand-600 transition-colors"
          title="People"
          @click="collapsed = false"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
          </svg>
        </button>
        <div class="flex-1" />
      </div>
    </template>

    <template v-else>
      <div class="px-3 py-2 border-b border-gray-100">
        <input
          v-if="showSearch"
          v-model="searchQuery"
          type="text"
          placeholder="Search..."
          class="w-full text-xs border border-gray-200 rounded-md px-2 py-1.5 outline-none focus:border-brand-300 focus:ring-1 focus:ring-brand-50 transition-all"
        />
      </div>

      <div class="flex-1 overflow-y-auto">
        <!-- Projects / Standards -->
        <div class="px-3 pt-3 pb-1.5 text-xs font-medium text-gray-400">
          Projects
        </div>
        <ul class="px-2">
          <li
            v-for="p in filteredProjects"
            :key="p.id"
            class="group px-2.5 py-1.5 rounded-md text-sm cursor-pointer transition-all truncate flex items-center gap-2"
            :class="activeProject === p.name
              ? 'bg-brand-50 text-brand-700 font-medium'
              : 'text-gray-600 hover:bg-gray-50'"
            @click="emit('select-project', p.name)"
          >
            <span class="w-1 h-1 rounded-full shrink-0" :class="p.type === 'standard' ? 'bg-brand-300' : 'bg-brand-400'" />
            <span class="truncate">{{ p.name }}</span>
            <span v-if="p.type === 'standard'" class="text-[10px] text-brand-400 shrink-0 ml-auto opacity-60">std</span>
          </li>
          <li v-if="filteredProjects.length === 0" class="px-2 py-2 text-xs text-gray-400 italic">
            {{ searchQuery ? 'No matches' : 'No projects yet' }}
          </li>
        </ul>

        <!-- People -->
        <div class="px-3 pt-4 pb-1.5 text-xs font-medium text-gray-400">
          People
        </div>
        <ul class="px-2 pb-4">
          <li
            v-for="p in filteredPeople"
            :key="p.id"
            class="group px-2.5 py-1.5 rounded-md text-sm cursor-pointer transition-all truncate flex items-center gap-2 text-gray-600 hover:bg-gray-50"
            @click="emit('select-person', p.name)"
          >
            <span class="w-1 h-1 rounded-full bg-emerald-400 shrink-0" />
            <span class="truncate">{{ p.name }}</span>
          </li>
          <li v-if="filteredPeople.length === 0" class="px-2 py-2 text-xs text-gray-400 italic">
            {{ searchQuery ? 'No matches' : 'No people yet' }}
          </li>
        </ul>

      </div>
    </template>
  </aside>
</template>
