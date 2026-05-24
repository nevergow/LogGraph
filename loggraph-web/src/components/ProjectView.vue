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
  'request-followup': [block: Block]
  'move-block': [id: string, newContent: string]
  'create-in-project': [content: string]
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
  isUnfiled: boolean
}

const projectGroups = computed<ProjectGroup[]>(() => {
  const map = new Map<string, Block[]>()
  for (const b of filteredBlocks.value) {
    const match = b.content.match(/(?:^|\s)[&]([^\s&][^\s]*)/)
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
      return { name, blocks, counts, allDone: counts.active === 0 && blocks.length > 0, isUnfiled: name === 'Unfiled' }
    })
    .sort((a, b) => {
      if (a.name === 'Unfiled') return 1
      if (b.name === 'Unfiled') return -1
      return a.name.localeCompare(b.name)
    })
})

// ── Drag-and-drop (Phase 1.4) ──
const dragOverProject = ref<string | null>(null)

function onDragOver(e: DragEvent, project: string) {
  e.preventDefault()
  if (e.dataTransfer) e.dataTransfer.dropEffect = 'move'
  dragOverProject.value = project
}

function onDragLeave() {
  dragOverProject.value = null
}

async function onDrop(e: DragEvent, toProject: string) {
  e.preventDefault()
  dragOverProject.value = null
  const blockId = e.dataTransfer?.getData('application/x-block-id')
  const fromProject = e.dataTransfer?.getData('application/x-from-project')
  if (!blockId || fromProject === toProject) return

  const block = props.blocks.find(b => b.id === blockId)
  if (!block) return

  let newContent = block.content
  if (fromProject === 'Unfiled') {
    newContent = `&${toProject} ` + block.content
  } else {
    // Replace &oldName or #oldName with &newName
    newContent = block.content
      .replace(new RegExp(`&${escapeRegex(fromProject)}(?=\\s|$)`, 'g'), `&${toProject}`)
      .replace(new RegExp(`#${escapeRegex(fromProject)}(?=\\s|$)`, 'g'), `&${toProject}`)
  }

  emit('move-block', blockId, newContent)
}

function escapeRegex(s: string): string {
  return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

// ── Per-project add button (Phase 1.5) ──
const addingToProject = ref<string | null>(null)
const inlineAddText = ref('')

function startAddToProject(project: string) {
  addingToProject.value = project
  inlineAddText.value = ''
}

function submitInlineAdd() {
  const content = inlineAddText.value.trim()
  if (!content || !addingToProject.value) return
  emit('create-in-project', `&${addingToProject.value} ${content}`)
  addingToProject.value = null
  inlineAddText.value = ''
}

function cancelInlineAdd() {
  addingToProject.value = null
  inlineAddText.value = ''
}

function onInlineKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    submitInlineAdd()
  } else if (e.key === 'Escape') {
    cancelInlineAdd()
  }
}
</script>

<template>
  <main class="flex-1 flex flex-col overflow-hidden bg-surface-canvas">
    <FilterBar
      :count="filteredBlocks.length"
      :hide-completed="hideCompleted"
      :status-filter="localStatusFilter"
      :screen-size="screenSize"
      @update:hide-completed="hideCompleted = $event"
      @filter-change="(key, value) => { if (key === 'status') localStatusFilter = value || '' }"
    />
    <div class="flex-1 overflow-y-auto px-4 py-4">
      <SkeletonCard v-if="loading && projectGroups.length === 0" :count="3" />

      <template v-else>
      <div v-if="hasMore" class="text-center pb-4">
        <button
          class="text-xs text-accent-600 hover:text-accent-800 disabled:opacity-40 font-semibold px-4 py-2 rounded-lg hover:bg-accent-50 transition-colors"
          :disabled="loading"
          @click="emit('load-more')"
        >
          {{ loading ? 'Loading...' : 'Load older entries' }}
        </button>
      </div>

      <TransitionGroup name="card-list" tag="div" class="space-y-6">
        <div
          v-for="group in projectGroups"
          :key="group.name"
          class="transition-all"
          :class="{
            'opacity-60': group.allDone,
            'ring-2 ring-accent-300/30 rounded-xl': dragOverProject === group.name,
          }"
          @dragover="onDragOver($event, group.name)"
          @dragleave="onDragLeave"
          @drop="onDrop($event, group.name)"
        >
          <!-- Section header (no card wrapper) -->
          <div
            class="flex items-center justify-between px-1 py-2 cursor-pointer"
            :class="{ 'opacity-50': group.allDone }"
            @click="emit('navigate-to-project', group.name)"
          >
            <div class="flex items-center gap-3">
              <!-- Icon: inbox for unfiled, folder for projects -->
              <div
                class="w-7 h-7 rounded-lg flex items-center justify-center"
                :class="group.isUnfiled
                  ? 'bg-stone-100 border border-dashed border-stone-300'
                  : 'bg-accent-50'"
              >
                <svg v-if="group.isUnfiled" class="w-3.5 h-3.5 text-stone-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
                </svg>
                <svg v-else class="w-3.5 h-3.5 text-accent-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                </svg>
              </div>
              <div>
                <span
                  class="font-semibold text-sm"
                  :class="group.allDone ? 'text-text-muted' : 'text-text-primary'"
                >{{ group.name }}</span>
                <!-- Unfiled subtitle -->
                <p v-if="group.isUnfiled" class="text-[10px] text-text-muted mt-0.5">Entries without a project</p>
                <!-- Status badges -->
                <div v-if="!group.isUnfiled" class="flex items-center gap-1.5 mt-0.5">
                  <span v-if="group.counts.active > 0" class="inline-flex items-center gap-1 text-[10px] bg-accent-50 text-accent-600 px-2 py-0.5 rounded-full font-semibold">
                    <span class="w-1.5 h-1.5 rounded-full bg-accent-500" />
                    {{ group.counts.active }} Active
                  </span>
                  <span v-if="group.counts.completed > 0" class="inline-flex items-center gap-1 text-[10px] bg-success-light text-success px-2 py-0.5 rounded-full font-semibold">
                    <span class="w-1.5 h-1.5 rounded-full bg-success" />
                    {{ group.counts.completed }} Done
                  </span>
                  <span v-if="group.counts.blocked > 0" class="inline-flex items-center gap-1 text-[10px] bg-danger-light text-danger px-2 py-0.5 rounded-full font-semibold">
                    <span class="w-1.5 h-1.5 rounded-full bg-danger" />
                    {{ group.counts.blocked }} Blocked
                  </span>
                </div>
              </div>
            </div>
            <button
              class="flex items-center gap-2 p-2 rounded-lg hover:bg-surface-100 transition-colors"
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

          <!-- CTA for unfiled pool -->
          <p v-if="group.isUnfiled && !collapsedProjects.has(group.name) && group.blocks.length > 0"
             class="px-1 pb-2 text-[10px] text-text-muted">
            Use <code class="text-accent-500 bg-accent-50 px-1 py-0.5 rounded font-medium">&amp;projectName</code> to assign
          </p>

          <!-- Project blocks -->
          <div
            v-if="!collapsedProjects.has(group.name)"
            class="space-y-3"
          >
            <BlockCard
              v-for="block in group.blocks"
              :key="block.id"
              :block="block"
              :selected="selectedId === block.id"
              :screen-size="screenSize"
              :draggable="true"
              @select="id => emit('select', id)"
              @toggle-status="(id, current) => emit('toggle-status', id, current)"
              @archive="id => emit('archive', id)"
              @delete="id => emit('delete', id)"
              @request-edit="id => emit('request-edit', id)"
              @request-followup="block => emit('request-followup', block)"
            />

            <!-- Per-project add button (Phase 1.5) -->
            <div v-if="addingToProject === group.name" class="px-1">
              <textarea
                v-model="inlineAddText"
                :placeholder="`Add to &${group.name}...`"
                rows="2"
                class="w-full resize-none outline-none text-sm p-3 bg-white border border-accent-200 rounded-xl focus:ring-2 focus:ring-accent-200/50 transition-all placeholder:text-text-muted"
                @keydown="onInlineKeydown"
              />
              <div class="flex items-center gap-2 mt-2">
                <button
                  class="px-3 py-1.5 text-xs font-semibold bg-accent-600 text-white rounded-lg hover:bg-accent-700 transition-colors"
                  :disabled="!inlineAddText.trim()"
                  @click="submitInlineAdd"
                >
                  Add
                </button>
                <button
                  class="px-3 py-1.5 text-xs text-text-muted hover:text-text-primary rounded-lg hover:bg-surface-100 transition-colors"
                  @click="cancelInlineAdd"
                >
                  Cancel
                </button>
              </div>
            </div>
            <button
              v-else
              class="w-full flex items-center justify-center gap-1.5 py-2 text-xs text-text-muted hover:text-accent-600 hover:bg-accent-50/50 rounded-lg transition-colors"
              :class="screenSize === 'mobile' ? 'min-h-[44px]' : 'opacity-0 group-hover:opacity-100'"
              @click="startAddToProject(group.name)"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              Add to {{ group.name }}
            </button>
          </div>
        </div>
      </TransitionGroup>

      <!-- Empty state -->
      <div v-if="projectGroups.length === 0 && !loading" class="text-center py-20">
        <div class="w-16 h-16 rounded-2xl bg-accent-50 flex items-center justify-center mx-auto mb-4">
          <svg class="w-8 h-8 text-accent-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
        </div>
        <div class="text-sm font-semibold text-text-primary mb-1">Your work memory starts here.</div>
        <div class="text-xs text-text-muted leading-relaxed max-w-xs mx-auto">
          Type <code class="text-accent-500 bg-accent-50 px-1 py-0.5 rounded font-medium">&amp;projectName</code> to organize,
          <code class="text-accent-500 bg-accent-50 px-1 py-0.5 rounded font-medium">@person</code> to mention,
          or just write what you're working on.
        </div>
        <div class="mt-4 inline-block w-2 h-4 bg-accent-400 rounded-sm cursor-blink" />
      </div>
      </template>
    </div>
  </main>
</template>
