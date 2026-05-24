<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import type { Block } from './types'
import LeftSidebar from './components/LeftSidebar.vue'
import CenterTimeline from './components/CenterTimeline.vue'
import ProjectView from './components/ProjectView.vue'
import SmartInput from './components/SmartInput.vue'
import WebhookSettings from './components/WebhookSettings.vue'
import AIPanel from './components/AIPanel.vue'
import CommandPalette from './components/CommandPalette.vue'
import CalendarHeatmap from './components/CalendarHeatmap.vue'
import CardEditor from './components/CardEditor.vue'
import ToastContainer from './components/ToastContainer.vue'
import { useBlocks } from './composables/useBlocks'
import { useNodes } from './composables/useNodes'
import { useToast } from './composables/useToast'

const { blocks, visibleBlocks, hasMore, loading, filters, selectedBlockId, fetchBlocks, createBlock, updateBlock, deleteBlock, undoDelete, archiveBlock, loadMore, setFilter, hasActiveFilter, clearAllFilters } = useBlocks()
const { projects, people, fetchProjects, fetchPeople } = useNodes()
const { showToast } = useToast()

const currentView = ref<'project' | 'timeline'>('project')
const showWebhooks = ref(false)
const showAI = ref(false)
const editingBlockId = ref<string | null>(null)

// SmartInput prefill state (for follow-up button)
const inputPrefillProject = ref<string | undefined>(undefined)
const inputPrefillContent = ref<string | undefined>(undefined)

const editingBlock = computed<Block | null>(() =>
  editingBlockId.value ? blocks.value.find(b => b.id === editingBlockId.value) ?? null : null
)

function handleRequestEdit(id: string) {
  editingBlockId.value = id
}

async function handleEditorSave(id: string, content: string) {
  try {
    await updateBlock(id, { content })
    fetchProjects()
    fetchPeople()
  } catch (e: any) {
    console.error('Failed to update block:', e)
    alert('保存失败: ' + (e.message || '未知错误'))
  } finally {
    editingBlockId.value = null
  }
}

function handleEditorClose() {
  editingBlockId.value = null
}
const showHeaderMenu = ref(false)
const showHeatmap = ref(false)

function navigateToProject(project: string) {
  setFilter('project', project)
}

// ── Responsive ──
const screenSize = ref<'mobile' | 'tablet' | 'desktop'>('desktop')
const showLeftOverlay = ref(false)

function updateScreenSize() {
  const w = window.innerWidth
  if (w < 768) screenSize.value = 'mobile'
  else if (w <= 1024) screenSize.value = 'tablet'
  else screenSize.value = 'desktop'
}

// ── Resizable panels (desktop only) ──
const leftWidth = ref(224)
const dragging = ref(false)

function onResizeMove(e: MouseEvent) {
  if (dragging.value) {
    leftWidth.value = Math.max(160, Math.min(400, e.clientX))
  }
}
function onResizeUp() {
  dragging.value = false
  document.removeEventListener('mousemove', onResizeMove)
  document.removeEventListener('mouseup', onResizeUp)
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
}
function startResize() {
  dragging.value = true
  document.addEventListener('mousemove', onResizeMove)
  document.addEventListener('mouseup', onResizeUp)
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
}

function onDocumentClick() {
  if (showHeaderMenu.value) showHeaderMenu.value = false
}

onMounted(() => {
  updateScreenSize()
  window.addEventListener('resize', updateScreenSize)
  document.addEventListener('click', onDocumentClick)
  fetchBlocks(true)
  fetchProjects()
  fetchPeople()
})

onUnmounted(() => {
  window.removeEventListener('resize', updateScreenSize)
  document.removeEventListener('click', onDocumentClick)
  document.removeEventListener('mousemove', onResizeMove)
  document.removeEventListener('mouseup', onResizeUp)
})

// ── Actions ──

async function handleCreate(content: string, metadata?: Record<string, any>, parentBlockId?: string) {
  await createBlock(content, metadata)
  if (parentBlockId) {
    await updateBlock(parentBlockId, { status: 'completed' })
  }
  fetchProjects()
  fetchPeople()
  // Clear prefill state
  inputPrefillProject.value = undefined
  inputPrefillContent.value = undefined
}

async function handleToggleStatus(id: string, current: string) {
  const next = current === 'active' ? 'completed' : current === 'completed' ? 'blocked' : 'active'
  await updateBlock(id, { status: next })
}

function handleArchive(id: string) {
  archiveBlock(id)
  showToast('已归档 1 条日志')
}

function handleDelete(id: string) {
  deleteBlock(id)
  showToast('已删除 1 条日志', { label: '撤销', handler: () => undoDelete(id) })
}

// ── Phase 1.4: Drag-drop between projects ──
async function handleMoveBlock(id: string, newContent: string) {
  await updateBlock(id, { content: newContent })
  fetchProjects()
  fetchPeople()
}

// ── Phase 2.1: Follow-up from card ──
function handleRequestFollowup(block: Block) {
  const match = block.content.match(/(?:^|\s)[&]([^\s&][^\s]*)/)
  const project = match ? match[1] : undefined
  inputPrefillProject.value = project
  inputPrefillContent.value = `^${block.id} `
}

function handleSelectProject(name: string) {
  setFilter('project', name)
  showLeftOverlay.value = false
}
function handleSelectPerson(name: string) {
  setFilter('person', name)
  showLeftOverlay.value = false
}

function handleCommand(action: string) {
  switch (action) {
    case 'view-project': currentView.value = 'project'; break
    case 'view-timeline': currentView.value = 'timeline'; break
    case 'ai-report': showAI.value = true; break
    case 'webhooks': showWebhooks.value = true; break
    case 'clear-filters': clearAllFilters(); break
    case 'new-entry': break
  }
}
</script>

<template>
  <div class="h-full flex flex-col">
    <!-- Header — Liquid Glass Chrome -->
    <header class="h-14 px-5 flex items-center justify-between shrink-0 glass border-b border-white/10">
      <div class="flex items-center gap-3">
        <!-- Logo: two dots + connecting line (Phase 6) -->
        <div class="w-8 h-8 flex items-center justify-center">
          <svg viewBox="0 0 24 24" class="w-6 h-6 text-accent-600">
            <circle cx="7" cy="7" r="3.5" fill="currentColor" opacity="0.6" />
            <circle cx="17" cy="17" r="3.5" fill="currentColor" />
            <line x1="9.5" y1="9.5" x2="14.5" y2="14.5"
                  stroke="currentColor" stroke-width="2" stroke-linecap="round" />
          </svg>
        </div>
        <span class="font-semibold text-sm tracking-tight text-text-primary">LogGraph</span>

        <!-- Segmented control -->
        <div class="ml-2 inline-flex bg-surface-100 rounded-lg p-1">
          <button
            class="px-4 py-1.5 text-xs font-medium rounded-md transition-all duration-200"
            :class="currentView === 'project' ? 'bg-white shadow-sm text-text-primary' : 'text-text-secondary hover:text-text-primary'"
            @click="currentView = 'project'"
          >
            Projects
          </button>
          <button
            class="px-4 py-1.5 text-xs font-medium rounded-md transition-all duration-200"
            :class="currentView === 'timeline' ? 'bg-white shadow-sm text-text-primary' : 'text-text-secondary hover:text-text-primary'"
            @click="currentView = 'timeline'"
          >
            Timeline
          </button>
        </div>

        <!-- Active filter pill -->
        <button
          v-if="hasActiveFilter"
          class="flex items-center gap-1.5 px-3 py-1 text-[11px] text-accent-600 bg-accent-50 hover:bg-accent-100 rounded-full transition-colors font-medium border border-accent-200/50"
          @click="clearAllFilters()"
        >
          <span>Filtered</span>
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      <div class="flex items-center gap-1">
        <!-- Mobile panel toggles -->
        <button
          v-if="screenSize !== 'desktop'"
          class="text-xs text-text-secondary hover:text-accent-600 hover:bg-accent-50 transition-colors flex items-center gap-1 px-3 py-2 rounded-lg"
          @click="showLeftOverlay = true"
          title="Filters & Projects"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
          </svg>
        </button>

        <!-- Consolidated actions menu (Bug 2 fixed: no glass-strong, solid white) -->
        <div class="relative" @click.stop>
          <button
            class="text-xs text-text-secondary hover:text-accent-600 hover:bg-accent-50 transition-colors flex items-center gap-1 px-3 py-2 rounded-lg"
            @click="showHeaderMenu = !showHeaderMenu"
            title="More"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
            </svg>
          </button>
          <div
            v-if="showHeaderMenu"
            class="absolute right-0 top-full mt-2 bg-white border border-slate-200 shadow-elevated rounded-xl z-50 py-2 min-w-[160px]"
            @click.stop
          >
            <button
              class="w-full text-left px-4 py-2.5 text-sm text-text-primary hover:bg-accent-50 hover:text-accent-700 transition-colors flex items-center gap-3"
              @click="showAI = !showAI; showHeaderMenu = false"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09z" />
              </svg>
              AI Report
            </button>
            <button
              class="w-full text-left px-4 py-2.5 text-sm text-text-primary hover:bg-accent-50 hover:text-accent-700 transition-colors flex items-center gap-3"
              @click="showWebhooks = !showWebhooks; showHeaderMenu = false"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              </svg>
              Webhooks
            </button>
            <div class="border-t border-border-subtle my-2" />
            <button
              class="w-full text-left px-4 py-2.5 text-sm text-text-primary hover:bg-accent-50 hover:text-accent-700 transition-colors flex items-center gap-3"
              @click="showHeatmap = !showHeatmap; showHeaderMenu = false"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
              Activity
            </button>
          </div>
        </div>
      </div>
    </header>

    <!-- Heatmap popover -->
    <div
      v-if="showHeatmap"
      class="fixed top-14 right-4 z-40 bg-white border border-slate-200 rounded-2xl shadow-elevated p-4 w-80"
      @click.stop
    >
      <div class="flex items-center justify-between mb-3">
        <span class="text-xs font-semibold text-text-primary uppercase tracking-wide">Activity</span>
        <button class="text-text-muted hover:text-text-primary transition-colors p-1 rounded-lg hover:bg-surface-100" @click="showHeatmap = false">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      <CalendarHeatmap />
    </div>
    <div v-if="showHeatmap" class="fixed inset-0 z-30" @click="showHeatmap = false" />

    <!-- Main content area -->
    <div class="flex-1 flex overflow-hidden">
      <!-- Desktop layout -->
      <template v-if="screenSize === 'desktop'">
        <LeftSidebar
          :style="{ width: leftWidth + 'px' }"
          :projects="projects"
          :people="people"
          :active-project="filters.project"
          :screen-size="screenSize"
          @select-project="p => setFilter('project', p)"
          @select-person="p => setFilter('person', p)"
          @clear-filters="setFilter('project', undefined); setFilter('person', undefined)"
          @node-updated="fetchProjects(); fetchPeople()"
          @node-deleted="fetchProjects(); fetchPeople()"
        />
        <div
          class="w-1 cursor-col-resize bg-transparent hover:bg-accent-300/50 transition-colors shrink-0 rounded-full"
          @mousedown="startResize()"
        />

        <ProjectView
          v-if="currentView === 'project'"
          :screen-size="screenSize"
          :blocks="visibleBlocks"
          :loading="loading"
          :has-more="hasMore"
          :selected-id="selectedBlockId"
          @load-more="loadMore"
          @select="id => selectedBlockId = id"
          @toggle-status="handleToggleStatus"
          @navigate-to-project="navigateToProject"
          @archive="handleArchive"
          @delete="handleDelete"
          @request-edit="handleRequestEdit"
          @request-followup="handleRequestFollowup"
          @move-block="handleMoveBlock"
          @create-in-project="(content) => handleCreate(content)"
        />
        <CenterTimeline
          v-else
          :screen-size="screenSize"
          :blocks="visibleBlocks"
          :loading="loading"
          :has-more="hasMore"
          :selected-id="selectedBlockId"
          :projects="projects"
          :status-filter="filters.status"
          :project-filter="filters.project"
          :since-date="filters.since"
          :until-date="filters.until"
          @load-more="loadMore"
          @select="id => selectedBlockId = id"
          @toggle-status="handleToggleStatus"
          @filter-change="(key: string, value: string | undefined) => setFilter(key as any, value)"
          @archive="handleArchive"
          @delete="handleDelete"
          @request-edit="handleRequestEdit"
          @request-followup="handleRequestFollowup"
        />
      </template>

      <!-- Tablet layout -->
      <template v-else-if="screenSize === 'tablet'">
        <LeftSidebar
          style="width: 180px"
          :projects="projects"
          :people="people"
          :active-project="filters.project"
          :screen-size="screenSize"
          @select-project="p => setFilter('project', p)"
          @select-person="p => setFilter('person', p)"
          @clear-filters="setFilter('project', undefined); setFilter('person', undefined)"
          @node-updated="fetchProjects(); fetchPeople()"
          @node-deleted="fetchProjects(); fetchPeople()"
        />

        <ProjectView
          v-if="currentView === 'project'"
          :screen-size="screenSize"
          :blocks="visibleBlocks"
          :loading="loading"
          :has-more="hasMore"
          :selected-id="selectedBlockId"
          @load-more="loadMore"
          @select="id => selectedBlockId = id"
          @toggle-status="handleToggleStatus"
          @navigate-to-project="navigateToProject"
          @archive="handleArchive"
          @delete="handleDelete"
          @request-edit="handleRequestEdit"
          @request-followup="handleRequestFollowup"
          @move-block="handleMoveBlock"
          @create-in-project="(content) => handleCreate(content)"
        />
        <CenterTimeline
          v-else
          :screen-size="screenSize"
          :blocks="visibleBlocks"
          :loading="loading"
          :has-more="hasMore"
          :selected-id="selectedBlockId"
          :projects="projects"
          :status-filter="filters.status"
          :project-filter="filters.project"
          :since-date="filters.since"
          :until-date="filters.until"
          @load-more="loadMore"
          @select="id => selectedBlockId = id"
          @toggle-status="handleToggleStatus"
          @filter-change="(key: string, value: string | undefined) => setFilter(key as any, value)"
          @archive="handleArchive"
          @delete="handleDelete"
          @request-edit="handleRequestEdit"
          @request-followup="handleRequestFollowup"
        />
      </template>

      <!-- Mobile layout -->
      <template v-else>
        <ProjectView
          v-if="currentView === 'project'"
          :screen-size="screenSize"
          :blocks="visibleBlocks"
          :loading="loading"
          :has-more="hasMore"
          :selected-id="selectedBlockId"
          @load-more="loadMore"
          @select="id => selectedBlockId = id"
          @toggle-status="handleToggleStatus"
          @navigate-to-project="navigateToProject"
          @archive="handleArchive"
          @delete="handleDelete"
          @request-edit="handleRequestEdit"
          @request-followup="handleRequestFollowup"
          @move-block="handleMoveBlock"
          @create-in-project="(content) => handleCreate(content)"
        />
        <CenterTimeline
          v-else
          :screen-size="screenSize"
          :blocks="visibleBlocks"
          :loading="loading"
          :has-more="hasMore"
          :selected-id="selectedBlockId"
          :projects="projects"
          :status-filter="filters.status"
          :project-filter="filters.project"
          :since-date="filters.since"
          :until-date="filters.until"
          @load-more="loadMore"
          @select="id => selectedBlockId = id"
          @toggle-status="handleToggleStatus"
          @filter-change="(key: string, value: string | undefined) => setFilter(key as any, value)"
          @archive="handleArchive"
          @delete="handleDelete"
          @request-edit="handleRequestEdit"
          @request-followup="handleRequestFollowup"
        />
      </template>
    </div>

    <SmartInput
      :prefill-project="inputPrefillProject"
      :prefill-content="inputPrefillContent"
      @send="handleCreate"
      @clear-prefill="inputPrefillProject = undefined; inputPrefillContent = undefined"
    />

    <!-- Mobile/Tablet overlays -->
    <Teleport to="body">
      <!-- Left panel overlay -->
      <div
        v-if="showLeftOverlay"
        class="fixed inset-0 z-50"
      >
        <div class="absolute inset-0 bg-black/20 backdrop-blur-sm" @click="showLeftOverlay = false" />
        <aside class="relative w-80 max-w-[85vw] bg-white/95 backdrop-blur-md h-full shadow-elevated overflow-y-auto rounded-r-2xl">
          <div class="flex justify-end p-3">
            <button class="text-text-muted hover:text-text-primary text-lg leading-none p-2 rounded-xl hover:bg-surface-100 transition-colors" @click="showLeftOverlay = false">&times;</button>
          </div>
          <LeftSidebar
            :projects="projects"
            :people="people"
            :active-project="filters.project"
            :screen-size="screenSize"
            @select-project="handleSelectProject"
            @select-person="handleSelectPerson"
            @clear-filters="setFilter('project', undefined); setFilter('person', undefined); showLeftOverlay = false"
            @node-updated="fetchProjects(); fetchPeople()"
            @node-deleted="fetchProjects(); fetchPeople()"
          />
        </aside>
      </div>
    </Teleport>

    <ToastContainer />
    <CardEditor
      :block="editingBlock"
      @save="handleEditorSave"
      @close="handleEditorClose"
    />
    <CommandPalette @command="handleCommand" />
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
