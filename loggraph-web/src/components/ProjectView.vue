<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Block } from '../types'
import { renderMarkdown } from '../composables/useMarkdown'

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

const projectGroups = computed(() => {
  const map = new Map<string, Block[]>()
  for (const b of props.blocks) {
    // Match #tag but not markdown headings (##, ###, etc.)
    const match = b.content.match(/(?:^|\s)#([^\s#][^\s]*)/)
    const project = match ? match[1] : 'Unfiled'
    if (!map.has(project)) map.set(project, [])
    map.get(project)!.push(b)
  }
  return [...map.entries()].sort((a, b) => {
    if (a[0] === 'Unfiled') return 1
    if (b[0] === 'Unfiled') return -1
    return a[0].localeCompare(b[0])
  })
})

function statusBadge(s: string): string {
  if (s === 'completed') return 'bg-emerald-100 text-emerald-700'
  if (s === 'blocked') return 'bg-red-100 text-red-700'
  return 'bg-blue-100 text-blue-700'
}

function statusLabel(s: string): string {
  if (s === 'completed') return 'Done'
  if (s === 'blocked') return 'Blocked'
  return 'Active'
}

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

function renderContent(text: string): string {
  return renderMarkdown(text)
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
          v-for="[project, projectBlocks] in projectGroups"
          :key="project"
          class="bg-white rounded-xl shadow-sm overflow-hidden"
        >
          <!-- Project header -->
          <div
            class="flex items-center justify-between px-4 py-3 hover:bg-slate-50 transition-colors cursor-pointer"
            @click="emit('navigate-to-project', project)"
          >
            <div class="flex items-center gap-2.5">
              <span
                class="w-2 h-2 rounded-full shrink-0"
                :class="nodeColor(projectBlocks[0]?.status || 'active')"
              />
              <span class="font-semibold text-sm text-slate-800">{{ project }}</span>
              <span class="text-[11px] text-slate-400 bg-slate-100 px-1.5 py-0.5 rounded-full">
                {{ projectBlocks.length }}
              </span>
            </div>
            <button
              class="flex items-center gap-2"
              @click.stop="toggleProject(project)"
            >
              <svg
                class="w-4 h-4 text-slate-400 transition-transform duration-200"
                :class="{ 'rotate-180': !collapsedProjects.has(project) }"
                fill="none" stroke="currentColor" viewBox="0 0 24 24"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </button>
          </div>

          <!-- Project blocks -->
          <div
            v-if="!collapsedProjects.has(project)"
            class="border-t border-slate-100 divide-y divide-slate-50"
          >
            <div
              v-for="block in projectBlocks"
              :key="block.id"
              class="px-4 py-2.5 cursor-pointer transition-colors hover:bg-slate-50"
              :class="{ 'bg-blue-50/50': selectedId === block.id }"
              @click="emit('select', block.id)"
              @dblclick="emit('edit', block.id)"
            >
              <div class="flex items-center justify-between mb-1">
                <div class="flex items-center gap-2">
                  <span
                    class="text-[10px] px-1.5 py-0.5 rounded-md font-medium cursor-pointer transition-all hover:ring-2 hover:ring-offset-1"
                    :class="statusBadge(block.status)"
                    @click.stop="emit('toggle-status', block.id, block.status)"
                  >
                    {{ statusLabel(block.status) }}
                  </span>
                  <span class="text-[11px] text-slate-400 font-medium">{{ block.user_id }}</span>
                </div>
                <span class="text-[11px] text-slate-400 shrink-0">{{ formatTime(block.created_at) }}</span>
              </div>
              <div
                class="text-sm leading-relaxed text-slate-600 prose prose-sm max-w-none"
                v-html="renderContent(block.content)"
              />
            </div>
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
