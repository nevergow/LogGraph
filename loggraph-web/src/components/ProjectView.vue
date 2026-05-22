<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Block } from '../types'
import BlockCard from './BlockCard.vue'

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
  for (const b of props.blocks) {
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
  <main class="flex-1 flex flex-col overflow-hidden bg-slate-50/50">
    <div class="flex-1 overflow-y-auto px-4 py-3">
      <!-- Load more -->
      <div v-if="hasMore" class="text-center pb-3">
        <button
          class="text-xs text-blue-600 hover:text-blue-800 disabled:opacity-40 font-medium"
          :disabled="loading"
          @click="emit('load-more')"
        >
          {{ loading ? 'Loading...' : 'Load older entries' }}
        </button>
      </div>

      <div class="space-y-4">
        <div
          v-for="group in projectGroups"
          :key="group.name"
          class="bg-white rounded-xl shadow-sm overflow-hidden"
          :class="{ 'opacity-70': group.allDone }"
        >
          <!-- Project header -->
          <div
            class="flex items-center justify-between px-4 py-3 hover:bg-slate-50 transition-colors cursor-pointer"
            :class="{ 'bg-slate-50/50': group.allDone }"
            @click="emit('navigate-to-project', group.name)"
          >
            <div class="flex items-center gap-2.5">
              <span
                class="w-2 h-2 rounded-full shrink-0"
                :class="group.allDone ? 'bg-gray-300' : nodeColor(group.blocks[0]?.status || 'active')"
              />
              <span
                class="font-semibold text-sm"
                :class="group.allDone ? 'text-gray-400' : 'text-slate-800'"
              >{{ group.name }}</span>
              <!-- Status count bubbles -->
              <div class="flex items-center gap-1">
                <span
                  v-if="group.counts.active > 0"
                  class="text-[11px] bg-blue-50 text-blue-600 px-1.5 py-0.5 rounded-full font-medium"
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
                class="w-4 h-4 text-slate-400 transition-transform duration-200"
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
            class="border-t border-slate-100 px-3 py-2 space-y-2"
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
      </div>

      <div v-if="projectGroups.length === 0 && !loading" class="text-center py-16 text-slate-400 text-sm">
        <div class="text-2xl mb-2">#</div>
        No project entries. Type a message with #project to get started.
      </div>
    </div>
  </main>
</template>
