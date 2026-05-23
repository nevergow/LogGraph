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
  'toggle-status': [id: string, current: string]
  'navigate-to-project': [project: string]
  archive: [id: string]
  delete: [id: string]
  'request-edit': [id: string]
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

</script>

<template>
  <main class="flex-1 flex flex-col overflow-hidden bg-gradient-to-br from-surface-50/50 to-violet-50/30">
    <FilterBar
      :count="filteredBlocks.length"
      :hide-completed="hideCompleted"
      :status-filter="localStatusFilter"
      :screen-size="screenSize"
      @update:hide-completed="hideCompleted = $event"
      @filter-change="(key, value) => { if (key === 'status') localStatusFilter = value || '' }"
    />
    <div class="flex-1 overflow-y-auto px-4 py-4">
      <!-- Skeleton loading (initial load only) -->
      <SkeletonCard v-if="loading && projectGroups.length === 0" :count="3" />

      <template v-else>
      <!-- Load more -->
      <div v-if="hasMore" class="text-center pb-4">
        <button
          class="text-xs text-brand-600 hover:text-brand-800 disabled:opacity-40 font-semibold px-4 py-2 rounded-xl hover:bg-brand-50 transition-colors"
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
          class="bg-white/80 backdrop-blur-sm rounded-2xl overflow-hidden border border-border-subtle hover:shadow-card transition-shadow"
          :class="{ 'opacity-60': group.allDone }"
        >
          <!-- Project header -->
          <div
            class="flex items-center justify-between px-5 py-4 hover:bg-surface-50/50 transition-colors cursor-pointer"
            :class="{ 'bg-surface-50/30': group.allDone }"
            @click="emit('navigate-to-project', group.name)"
          >
            <div class="flex items-center gap-3">
              <div
                class="w-8 h-8 rounded-xl bg-gradient-to-br from-brand-100 to-violet-100 flex items-center justify-center"
              >
                <svg class="w-4 h-4 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                </svg>
              </div>
              <div>
                <span
                  class="font-semibold text-sm"
                  :class="group.allDone ? 'text-text-muted' : 'text-text-primary'"
                >{{ group.name }}</span>
                <!-- Status count badges -->
                <div class="flex items-center gap-1.5 mt-1">
                  <span
                    v-if="group.counts.active > 0"
                    class="inline-flex items-center gap-1 text-[10px] bg-brand-50 text-brand-600 px-2 py-0.5 rounded-full font-semibold"
                  >
                    <span class="w-1.5 h-1.5 rounded-full bg-brand-500" />
                    {{ group.counts.active }} Active
                  </span>
                  <span
                    v-if="group.counts.completed > 0"
                    class="inline-flex items-center gap-1 text-[10px] bg-success-light text-success px-2 py-0.5 rounded-full font-semibold"
                  >
                    <span class="w-1.5 h-1.5 rounded-full bg-success" />
                    {{ group.counts.completed }} Done
                  </span>
                  <span
                    v-if="group.counts.blocked > 0"
                    class="inline-flex items-center gap-1 text-[10px] bg-danger-light text-danger px-2 py-0.5 rounded-full font-semibold"
                  >
                    <span class="w-1.5 h-1.5 rounded-full bg-danger" />
                    {{ group.counts.blocked }} Blocked
                  </span>
                </div>
              </div>
            </div>
            <button
              class="flex items-center gap-2 p-2 rounded-xl hover:bg-surface-100 transition-colors"
              @click.stop="toggleProject(group.name)"
            >
              <svg
                class="w-4 h-4 text-text-muted transition-transform duration-200"
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
            class="border-t border-border-subtle px-4 py-3 space-y-3"
          >
            <BlockCard
              v-for="block in group.blocks"
              :key="block.id"
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

      <div v-if="projectGroups.length === 0 && !loading" class="text-center py-20">
        <div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-brand-100 to-violet-100 flex items-center justify-center mx-auto mb-4">
          <svg class="w-8 h-8 text-brand-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
        </div>
        <div class="text-sm font-semibold text-text-primary mb-1">No project entries</div>
        <div class="text-xs text-text-muted">Type <code class="text-brand-500 bg-brand-50 px-1.5 py-0.5 rounded-lg font-medium">#project</code> in your message to organize entries.</div>
      </div>
      </template>
    </div>
  </main>
</template>
