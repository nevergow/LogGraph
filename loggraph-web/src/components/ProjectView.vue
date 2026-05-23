<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Block } from '../types'
import BlockCard from './BlockCard.vue'
import FilterBar from './FilterBar.vue'
import SkeletonCard from './SkeletonCard.vue'

const hideCompleted = ref(false)
const localStatusFilter = ref('')

const filteredBlocks = computed(() => {
  let result = props.blocks
  if (hideCompleted.value) result = result.filter(b => b.status !== 'completed')
  if (localStatusFilter.value) result = result.filter(b => b.status === localStatusFilter.value)
  return result
})

const props = defineProps<{
  blocks: Block[]
  loading: boolean
  hasMore: boolean
  selectedId: string | null
  screenSize?: 'mobile' | 'tablet' | 'desktop'
}>()

const emit = defineEmits<{
  'load-more': []
  select: [id: string]
  edit: [id: string]
  'toggle-status': [id: string, current: string]
  'navigate-to-project': [project: string]
  archive: [id: string]
  delete: [id: string]
}>()

const collapsedProjects = ref<Set<string>>(new Set())

function toggleProject(name: string) {
  if (collapsedProjects.value.has(name)) {
    collapsedProjects.value.delete(name)
  } else {
    collapsedProjects.value.add(name)
  }
  collapsedProjects.value = new Set(collapsedProjects.value)
}

interface ProjectGroup {
  name: string
  blocks: Block[]
  counts: { active: number; completed: number; blocked: number }
  allDone: boolean
}

const projectGroups = computed<ProjectGroup[]>(() => {
  const map = new Map<string, Block[]>()
  for (const b of filteredBlocks.value) {
    const match = b.content.match(/(?:^|\s)#([^\s#][^\s]*)/)
    const project = match ? match[1] : 'Unfiled'
    if (!map.has(project)) map.set(project, [])
    map.get(project)!.push(b)
  }
  return [...map.entries()]
    .map(([name, blocks]) => {
      const counts = { active: 0, completed: 0, blocked: 0 }
      for (const b of blocks) {
        if (b.status === 'active') counts.active++
        else if (b.status === 'completed') counts.completed++
        else if (b.status === 'blocked') counts.blocked++
      }
      return { name, blocks, counts, allDone: counts.active === 0 && blocks.length > 0 }
    })
    .sort((a, b) => {
      if (a.name === 'Unfiled') return 1
      if (b.name === 'Unfiled') return -1
      return a.name.localeCompare(b.name)
    })
})

function nodeColor(s: string): string {
  if (s === 'completed') return 'bg-emerald-400'
  if (s === 'blocked') return 'bg-red-400'
  return 'bg-blue-400'
}

function formatTime(ts: string): string {
  const d = new Date(ts)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

</script>

<template>
  <main class="flex-1 flex flex-col overflow-hidden bg-gray-50/50">
    <FilterBar
      :count="filteredBlocks.length"
      :hide-completed="hideCompleted"
      :status-filter="localStatusFilter"
      :screen-size="screenSize"
      @update:hide-completed="hideCompleted = $event"
      @filter-change="(key, value) => { if (key === 'status') localStatusFilter = value || '' }"
    />
    <div class="flex-1 overflow-y-auto px-4 py-3">
      <!-- Skeleton loading (initial load only) -->
      <SkeletonCard v-if="loading && projectGroups.length === 0" :count="3" />

      <template v-else>
      <!-- Load more -->
      <div v-if="hasMore" class="text-center pb-3">
        <button
          class="text-xs text-brand-600 hover:text-brand-800 disabled:opacity-40 font-medium"
          :disabled="loading"
          @click="emit('load-more')"
        >
          {{ loading ? 'Loading...' : 'Load older entries' }}
        </button>
      </div>

      <TransitionGroup name="card-list" tag="div" class="space-y-4">
        <div
          v-for="group in projectGroups"
          :key="group.name"
          class="bg-white rounded-md hover:shadow-elevated transition-shadow overflow-hidden"
          :class="{ 'opacity-50': group.allDone }"
        >
          <!-- Project header -->
          <div
            class="flex items-center justify-between px-4 py-3 hover:bg-gray-50 transition-colors cursor-pointer"
            :class="{ 'bg-gray-50/50': group.allDone }"
            @click="emit('navigate-to-project', group.name)"
          >
            <div class="flex items-center gap-2.5">
              <span
                class="font-medium text-sm"
                :class="group.allDone ? 'text-gray-400' : 'text-gray-800'"
              >{{ group.name }}</span>
              <!-- Status count bubbles -->
              <div class="flex items-center gap-1">
                <span
                  v-if="group.counts.active > 0"
                  class="text-[11px] bg-brand-50 text-brand-600 px-1.5 py-0.5 rounded-full font-medium"
                >{{ group.counts.active }} Active</span>
                <span
                  v-if="group.counts.completed > 0"
                  class="text-[11px] bg-emerald-50 text-emerald-600 px-1.5 py-0.5 rounded-full font-medium"
                >{{ group.counts.completed }} Done</span>
                <span
                  v-if="group.counts.blocked > 0"
                  class="text-[11px] bg-red-50 text-red-600 px-1.5 py-0.5 rounded-full font-medium"
                >{{ group.counts.blocked }} Blocked</span>
              </div>
            </div>
            <button
              class="flex items-center gap-2"
              @click.stop="toggleProject(group.name)"
            >
              <svg
                class="w-4 h-4 text-gray-400 transition-transform duration-200"
                :class="{ 'rotate-180': !collapsedProjects.has(group.name) }"
                fill="none" stroke="currentColor" viewBox="0 0 24 24"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </button>
          </div>

          <!-- Project blocks -->
          <div
            v-if="!collapsedProjects.has(group.name)"
            class="border-t border-gray-100 px-3 py-2 space-y-2"
          >
            <BlockCard
              v-for="block in group.blocks"
              :key="block.id"
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
      </TransitionGroup>

      <div v-if="projectGroups.length === 0 && !loading" class="text-center py-20 text-gray-400">
        <svg class="w-10 h-10 text-gray-300 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
        <div class="text-sm font-medium text-gray-500 mb-1">No project entries</div>
        <div class="text-xs text-gray-400">Type <code class="text-brand-500 bg-brand-50 px-1 rounded-sm">#project</code> in your message to organize entries.</div>
      </div>
      </template>
    </div>
  </main>
</template>
