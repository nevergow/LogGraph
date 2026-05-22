<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import type { Block } from './types'
import LeftSidebar from './components/LeftSidebar.vue'
import CenterTimeline from './components/CenterTimeline.vue'
import ProjectView from './components/ProjectView.vue'
import RightGraphPanel from './components/RightGraphPanel.vue'
import SmartInput from './components/SmartInput.vue'
import WebhookSettings from './components/WebhookSettings.vue'
import AIPanel from './components/AIPanel.vue'
import { useBlocks } from './composables/useBlocks'
import { useNodes } from './composables/useNodes'

const { blocks, hasMore, loading, filters, selectedBlockId, fetchBlocks, createBlock, updateBlock, loadMore, setFilter } = useBlocks()
const { projects, people, fetchProjects, fetchPeople } = useNodes()

const currentView = ref<'project' | 'timeline'>('project')
const showWebhooks = ref(false)
const showAI = ref(false)
const editingBlock = ref<Block | null>(null)

function navigateToProject(project: string) {
  setFilter('project', project)
  currentView.value = 'timeline'
}

// ── Responsive ──
const screenSize = ref<'mobile' | 'tablet' | 'desktop'>('desktop')
const showLeftOverlay = ref(false)
const showRightOverlay = ref(false)

function updateScreenSize() {
  const w = window.innerWidth
  if (w < 768) screenSize.value = 'mobile'
  else if (w <= 1024) screenSize.value = 'tablet'
  else screenSize.value = 'desktop'
}

// ── Resizable panels (desktop only) ──
const leftWidth = ref(224)
const rightWidth = ref(288)
const dragging = ref<'left' | 'right' | null>(null)

function onResizeMove(e: MouseEvent) {
  if (dragging.value === 'left') {
    leftWidth.value = Math.max(160, Math.min(400, e.clientX))
  } else if (dragging.value === 'right') {
    rightWidth.value = Math.max(200, Math.min(500, window.innerWidth - e.clientX))
  }
}
function onResizeUp() {
  dragging.value = null
  document.removeEventListener('mousemove', onResizeMove)
  document.removeEventListener('mouseup', onResizeUp)
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
}
function startResize(side: 'left' | 'right') {
  dragging.value = side
  document.addEventListener('mousemove', onResizeMove)
  document.addEventListener('mouseup', onResizeUp)
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
}

const selectedProjectNodeId = computed(() => {
  if (!filters.project) return null
  const node = projects.value.find(p => p.name === filters.project)
  return node?.id || null
})

onMounted(() => {
  updateScreenSize()
  window.addEventListener('resize', updateScreenSize)
  fetchBlocks(true)
  fetchProjects()
  fetchPeople()
})

onUnmounted(() => {
  window.removeEventListener('resize', updateScreenSize)
  document.removeEventListener('mousemove', onResizeMove)
  document.removeEventListener('mouseup', onResizeUp)
})

// ── Actions ──

async function handleCreate(content: string) {
  await createBlock(content)
  fetchProjects()
  fetchPeople()
}

function handleEdit(id: string) {
  const b = blocks.value.find(b => b.id === id)
  if (b) editingBlock.value = b
}

async function handleUpdate(id: string, content: string) {
  await updateBlock(id, { content })
  editingBlock.value = null
  fetchProjects()
  fetchPeople()
}

async function handleToggleStatus(id: string, current: string) {
  const next = current === 'active' ? 'completed' : current === 'completed' ? 'blocked' : 'active'
  await updateBlock(id, { status: next })
}

function handleCancelEdit() {
  editingBlock.value = null
}

function handleSelectProject(name: string) {
  setFilter('project', name)
  showLeftOverlay.value = false
}
function handleSelectPerson(name: string) {
  setFilter('person', name)
  showLeftOverlay.value = false
}
</script>

<template>
  <div class="h-full flex flex-col">
    <!-- Header -->
    <header class="h-12 border-b border-slate-200/60 flex items-center px-3 sm:px-5 shrink-0 bg-white justify-between">
      <div class="flex items-center gap-2 sm:gap-3">
        <div class="w-7 h-7 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
          <span class="text-white font-bold text-xs">LG</span>
        </div>
        <span class="font-semibold text-xs sm:text-sm tracking-tight text-slate-800">LogGraph</span>
        <span class="hidden sm:inline text-[10px] text-slate-400 bg-slate-100 px-1.5 py-0.5 rounded-full">v0.5</span>
      </div>
      <div class="flex items-center gap-0.5 sm:gap-1">
        <!-- Mobile panel toggles -->
        <button
          v-if="screenSize !== 'desktop'"
          class="text-xs text-slate-400 hover:text-blue-600 hover:bg-blue-50 transition-colors flex items-center gap-1 px-2 py-1.5 rounded-md"
          @click="showLeftOverlay = true"
          title="Filters & Projects"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
          </svg>
        </button>
        <button
          v-if="screenSize !== 'desktop'"
          class="text-xs text-slate-400 hover:text-purple-600 hover:bg-purple-50 transition-colors flex items-center gap-1 px-2 py-1.5 rounded-md"
          @click="showRightOverlay = true"
          title="Graph"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
          </svg>
        </button>

        <button
          class="text-xs text-slate-400 hover:text-purple-600 hover:bg-purple-50 transition-colors flex items-center gap-1 sm:gap-1.5 px-2 sm:px-3 py-1.5 rounded-md"
          @click="showAI = !showAI"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.455 2.456L21.75 6l-1.036.259a3.375 3.375 0 00-2.455 2.456z" />
          </svg>
          <span class="hidden sm:inline">AI</span>
        </button>
        <button
          class="text-xs text-slate-400 hover:text-blue-600 hover:bg-blue-50 transition-colors flex items-center gap-1 sm:gap-1.5 px-2 sm:px-3 py-1.5 rounded-md"
          @click="showWebhooks = !showWebhooks"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          <span class="hidden sm:inline">Webhooks</span>
        </button>
      </div>
    </header>

    <!-- Segmented control -->
    <div class="px-3 py-1.5 bg-white border-b border-slate-200/60 flex justify-center shrink-0">
      <div class="inline-flex bg-slate-100 rounded-lg p-0.5">
        <button
          class="px-4 py-1.5 text-xs font-medium rounded-md transition-all"
          :class="currentView === 'project' ? 'bg-white shadow-sm text-slate-800' : 'text-slate-500 hover:text-slate-700'"
          @click="currentView = 'project'"
        >
          Projects
        </button>
        <button
          class="px-4 py-1.5 text-xs font-medium rounded-md transition-all"
          :class="currentView === 'timeline' ? 'bg-white shadow-sm text-slate-800' : 'text-slate-500 hover:text-slate-700'"
          @click="currentView = 'timeline'"
        >
          Timeline
        </button>
      </div>
    </div>

    <!-- Main content area -->
    <div class="flex-1 flex overflow-hidden">
      <!-- Project View (default) -->
      <template v-if="currentView === 'project'">
        <ProjectView
          :screen-size="screenSize"
          :blocks="blocks"
          :loading="loading"
          :has-more="hasMore"
          :selected-id="selectedBlockId"
          @load-more="loadMore"
          @select="id => selectedBlockId = id"
          @edit="handleEdit"
          @toggle-status="handleToggleStatus"
          @navigate-to-project="navigateToProject"
        />
      </template>

      <!-- Timeline View -->
      <template v-else>
      <!-- Desktop layout: 3 columns with resize handles -->
      <template v-if="screenSize === 'desktop'">
        <LeftSidebar
          :style="{ width: leftWidth + 'px' }"
          :projects="projects"
          :people="people"
          :active-project="filters.project"
          @select-project="p => setFilter('project', p)"
          @select-person="p => setFilter('person', p)"
          @clear-filters="setFilter('project', undefined); setFilter('person', undefined)"
        />
        <div
          class="w-1 cursor-col-resize bg-transparent hover:bg-blue-300 transition-colors shrink-0"
          @mousedown="startResize('left')"
        />
        <CenterTimeline
          :screen-size="screenSize"
          :blocks="blocks"
          :loading="loading"
          :has-more="hasMore"
          :selected-id="selectedBlockId"
          @load-more="loadMore"
          @select="id => selectedBlockId = id"
          @edit="handleEdit"
          @toggle-status="handleToggleStatus"
          @filter-change="(key: string, value: string | undefined) => setFilter(key as any, value)"
        />
        <div
          class="w-1 cursor-col-resize bg-transparent hover:bg-blue-300 transition-colors shrink-0"
          @mousedown="startResize('right')"
        />
        <RightGraphPanel
          :style="{ width: rightWidth + 'px' }"
          :block-id="selectedBlockId"
          :project-node-id="selectedProjectNodeId"
          :project-name="filters.project"
        />
      </template>

      <!-- Tablet layout: left sidebar inline (narrow), center flex-1, right as overlay -->
      <template v-else-if="screenSize === 'tablet'">
        <LeftSidebar
          style="width: 180px"
          :projects="projects"
          :people="people"
          :active-project="filters.project"
          @select-project="p => setFilter('project', p)"
          @select-person="p => setFilter('person', p)"
          @clear-filters="setFilter('project', undefined); setFilter('person', undefined)"
        />
        <CenterTimeline
          :screen-size="screenSize"
          :blocks="blocks"
          :loading="loading"
          :has-more="hasMore"
          :selected-id="selectedBlockId"
          @load-more="loadMore"
          @select="id => selectedBlockId = id"
          @edit="handleEdit"
          @toggle-status="handleToggleStatus"
          @filter-change="(key: string, value: string | undefined) => setFilter(key as any, value)"
        />
      </template>

      <!-- Mobile layout: center only, both panels as overlays -->
      <template v-else>
        <CenterTimeline
          :screen-size="screenSize"
          :blocks="blocks"
          :loading="loading"
          :has-more="hasMore"
          :selected-id="selectedBlockId"
          @load-more="loadMore"
          @select="id => selectedBlockId = id"
          @edit="handleEdit"
          @toggle-status="handleToggleStatus"
          @filter-change="(key: string, value: string | undefined) => setFilter(key as any, value)"
        />
      </template>
      </template>
    </div>

    <SmartInput
      :editing-block="editingBlock"
      @send="handleCreate"
      @update="handleUpdate"
      @cancel-edit="handleCancelEdit"
    />

    <!-- Mobile/Tablet overlays (Teleported to body) -->
    <Teleport to="body">
      <!-- Left panel overlay -->
      <div
        v-if="showLeftOverlay"
        class="fixed inset-0 z-50"
      >
        <div class="absolute inset-0 bg-black/30" @click="showLeftOverlay = false" />
        <aside class="relative w-72 max-w-[85vw] bg-white h-full shadow-xl overflow-y-auto">
          <div class="flex justify-end p-2">
            <button class="text-slate-400 hover:text-slate-600 text-lg leading-none p-1" @click="showLeftOverlay = false">&times;</button>
          </div>
          <LeftSidebar
            :projects="projects"
            :people="people"
            :active-project="filters.project"
            @select-project="handleSelectProject"
            @select-person="handleSelectPerson"
            @clear-filters="setFilter('project', undefined); setFilter('person', undefined); showLeftOverlay = false"
          />
        </aside>
      </div>

      <!-- Right panel overlay -->
      <div
        v-if="showRightOverlay"
        class="fixed inset-0 z-50 flex justify-end"
      >
        <div class="absolute inset-0 bg-black/30" @click="showRightOverlay = false" />
        <aside class="relative w-80 max-w-[85vw] bg-white h-full shadow-xl overflow-y-auto">
          <div class="flex justify-end p-2">
            <button class="text-slate-400 hover:text-slate-600 text-lg leading-none p-1" @click="showRightOverlay = false">&times;</button>
          </div>
          <RightGraphPanel
            :block-id="selectedBlockId"
            :project-node-id="selectedProjectNodeId"
            :project-name="filters.project"
          />
        </aside>
      </div>
    </Teleport>

    <WebhookSettings v-if="showWebhooks" @close="showWebhooks = false" />
    <AIPanel
      v-if="showAI"
      :current-project="filters.project"
      :current-since="filters.since"
      :current-until="filters.until"
      @close="showAI = false"
    />
  </div>
</template>
