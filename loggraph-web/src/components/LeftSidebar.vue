<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onUnmounted } from 'vue'
import type { Node } from '../types'
import { nodesApi } from '../api/nodes'

const props = defineProps<{
  projects: Node[]
  people: Node[]
  activeProject?: string
  screenSize?: 'mobile' | 'tablet' | 'desktop'
}>()

const emit = defineEmits<{
  'select-project': [name: string]
  'select-person': [name: string]
  'clear-filters': []
  'node-updated': []
  'node-deleted': []
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

// ── Inline edit state ──
const editingNodeId = ref<string | null>(null)
const editingNodeName = ref('')
const editInputRef = ref<HTMLInputElement | null>(null)
const deletingNodeId = ref<string | null>(null)
const activeNodeId = ref<string | null>(null)
const savingNodeId = ref<string | null>(null)

// Close active state when clicking outside
function onDocumentClick(e: MouseEvent) {
  // Don't clear if click is inside the sidebar
  const target = e.target as HTMLElement
  if (target && target.closest('aside')) {
    return
  }
  activeNodeId.value = null
}
onMounted(() => document.addEventListener('click', onDocumentClick))
onUnmounted(() => document.removeEventListener('click', onDocumentClick))

function startEdit(node: Node) {
  editingNodeId.value = node.id
  editingNodeName.value = node.name
  nextTick(() => editInputRef.value?.focus())
}

async function saveEdit(id: string) {
  if (savingNodeId.value === id) return
  const name = editingNodeName.value.trim()
  if (!name) return
  savingNodeId.value = id
  try {
    await nodesApi.update(id, name, true)
    editingNodeId.value = null
    emit('node-updated')
  } catch (e: any) {
    console.error('Failed to update node:', e)
    alert('保存失败: ' + (e.message || '未知错误'))
  } finally {
    savingNodeId.value = null
  }
}

function cancelEdit() {
  editingNodeId.value = null
}

function onEditKeydown(e: KeyboardEvent, id: string) {
  if (e.key === 'Enter') {
    e.preventDefault()
    saveEdit(id)
  } else if (e.key === 'Escape') {
    e.preventDefault()
    cancelEdit()
  }
}

async function confirmDelete(node: Node) {
  deletingNodeId.value = node.id
}

async function doDelete(id: string) {
  try {
    await nodesApi.delete(id)
    deletingNodeId.value = null
    activeNodeId.value = null
    emit('node-deleted')
  } catch (e: any) {
    console.error('Failed to delete node:', e)
    alert('删除失败: ' + (e.message || '未知错误'))
  }
}

const isDesktop = computed(() => props.screenSize === 'desktop')

function handleItemClick(nodeId: string, name: string, type: 'project' | 'person') {
  if (isDesktop.value) {
    if (type === 'project') {
      emit('select-project', name)
    } else {
      emit('select-person', name)
    }
  } else {
    // On mobile/tablet, toggle selection to show action buttons
    activeNodeId.value = activeNodeId.value === nodeId ? null : nodeId
  }
}
</script>

<template>
  <aside class="border-r border-white/10 flex flex-col shrink-0 overflow-hidden glass" :class="collapsed ? 'items-center' : ''">
    <!-- Collapse toggle -->
    <div class="p-3 border-b border-border-subtle flex" :class="collapsed ? 'justify-center' : 'justify-between items-center'">
      <button
        v-if="!collapsed && activeProject"
        class="text-[10px] text-accent-600 hover:text-accent-800 bg-accent-50 hover:bg-accent-100 px-3 py-1 rounded-full transition-colors font-medium border border-accent-200/50"
        @click="emit('clear-filters')"
      >
        Clear
      </button>
      <span v-if="!collapsed && !activeProject" />
      <button
        class="text-text-muted hover:text-accent-600 hover:bg-accent-50 p-2 rounded-xl transition-colors shrink-0"
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
      <div class="flex-1 flex flex-col items-center gap-3 pt-4 overflow-y-auto">
        <button
          class="p-3 rounded-xl hover:bg-accent-50 text-text-muted hover:text-accent-600 transition-colors"
          title="Projects"
          @click="collapsed = false"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
        </button>
        <button
          class="p-3 rounded-xl hover:bg-accent-50 text-text-muted hover:text-accent-600 transition-colors"
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
      <div class="px-4 py-3 border-b border-border-subtle">
        <input
          v-if="showSearch"
          v-model="searchQuery"
          type="text"
          placeholder="Search..."
          class="w-full text-xs border-0 bg-surface-100 rounded-full px-4 py-2.5 outline-none focus:ring-2 focus:ring-brand-200/50 transition-all placeholder:text-text-muted"
        />
      </div>

      <div class="flex-1 overflow-y-auto">
        <!-- Projects / Standards -->
        <div class="px-4 pt-4 pb-2 text-[10px] font-semibold text-text-muted uppercase tracking-wider">
          Projects
        </div>
        <ul class="px-3 pb-4">
          <li
            v-for="p in filteredProjects"
            :key="p.id"
            class="group relative px-3 py-2.5 rounded-xl text-sm cursor-pointer transition-all truncate flex items-center gap-3 card-lift select-none"
            :class="[
              activeProject === p.name
                ? 'bg-accent-50 text-accent-700 font-medium'
                : 'text-text-secondary',
              activeNodeId === p.id ? 'bg-surface-100 ring-2 ring-accent-300' : 'hover:bg-surface-100'
            ]"
            @click.stop="handleItemClick(p.id, p.name, 'project')"
          >
            <span class="w-2 h-2 rounded-full shrink-0" :class="p.type === 'standard' ? 'bg-accent-300' : 'bg-accent-500'" />
            <input
              v-if="editingNodeId === p.id"
              ref="editInputRef"
              v-model="editingNodeName"
              class="flex-1 min-w-0 text-xs border border-accent-300 rounded-lg px-2 py-1 outline-none bg-white"
              @click.stop
              @keydown="onEditKeydown($event, p.id)"
              @blur="saveEdit(p.id)"
            />
            <span v-else class="truncate">{{ p.name }}</span>
            <span v-if="p.type === 'standard'" class="text-[9px] text-accent-400 shrink-0 ml-auto opacity-60 font-medium">STD</span>
            <!-- Mobile tap actions / Desktop hover actions -->
            <div
              v-if="editingNodeId !== p.id"
              class="flex items-center gap-1 shrink-0 ml-auto"
              :class="[
                isDesktop ? 'hidden group-hover:flex' : (activeNodeId === p.id ? 'flex' : 'hidden')
              ]"
              @click.stop
            >
              <button
                class="p-1.5 rounded-lg text-text-muted hover:text-accent-600 hover:bg-accent-50 transition-colors"
                title="Rename"
                @click.stop="startEdit(p)"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
              </button>
              <button
                class="p-1.5 rounded-lg text-text-muted hover:text-danger hover:bg-danger-light transition-colors"
                title="Delete"
                @click.stop="confirmDelete(p)"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          </li>
          <li v-if="filteredProjects.length === 0" class="px-3 py-3 text-xs text-text-muted italic">
            {{ searchQuery ? 'No matches' : 'No projects yet' }}
          </li>
        </ul>

        <!-- People -->
        <div class="px-4 pt-5 pb-2 text-[10px] font-semibold text-text-muted uppercase tracking-wider">
          People
        </div>
        <ul class="px-3 pb-4">
          <li
            v-for="p in filteredPeople"
            :key="p.id"
            class="group relative px-3 py-2.5 rounded-xl text-sm cursor-pointer transition-all truncate flex items-center gap-3 card-lift text-text-secondary select-none"
            :class="activeNodeId === p.id ? 'bg-surface-100 ring-2 ring-accent-300' : 'hover:bg-surface-100'"
            @click.stop="handleItemClick(p.id, p.name, 'person')"
          >
            <span class="w-2 h-2 rounded-full bg-success shrink-0" />
            <input
              v-if="editingNodeId === p.id"
              ref="editInputRef"
              v-model="editingNodeName"
              class="flex-1 min-w-0 text-xs border border-accent-300 rounded-lg px-2 py-1 outline-none bg-white"
              @click.stop
              @keydown="onEditKeydown($event, p.id)"
              @blur="saveEdit(p.id)"
            />
            <span v-else class="truncate">{{ p.name }}</span>
            <!-- Mobile tap actions / Desktop hover actions -->
            <div
              v-if="editingNodeId !== p.id"
              class="flex items-center gap-1 shrink-0 ml-auto"
              :class="[
                isDesktop ? 'hidden group-hover:flex' : (activeNodeId === p.id ? 'flex' : 'hidden')
              ]"
              @click.stop
            >
              <button
                class="p-1.5 rounded-lg text-text-muted hover:text-accent-600 hover:bg-accent-50 transition-colors"
                title="Rename"
                @click.stop="startEdit(p)"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
              </button>
              <button
                class="p-1.5 rounded-lg text-text-muted hover:text-danger hover:bg-danger-light transition-colors"
                title="Delete"
                @click.stop="confirmDelete(p)"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          </li>
          <li v-if="filteredPeople.length === 0" class="px-3 py-3 text-xs text-text-muted italic">
            {{ searchQuery ? 'No matches' : 'No people yet' }}
          </li>
        </ul>

      </div>
    </template>
  </aside>

  <!-- Delete confirmation modal -->
  <Teleport to="body">
    <div
      v-if="deletingNodeId"
      class="fixed inset-0 z-50 flex items-center justify-center"
      @click="deletingNodeId = null"
    >
      <div class="absolute inset-0 bg-black/20 backdrop-blur-sm" />
      <div class="relative glass-strong rounded-2xl shadow-glass border border-white/50 p-6 max-w-sm w-full mx-4" @click.stop>
        <p class="text-sm text-text-primary mb-4">Delete this node? Blocks will keep their text but lose the association.</p>
        <div class="flex justify-end gap-2">
          <button
            class="px-4 py-2 text-sm text-text-secondary hover:bg-surface-100 rounded-xl transition-colors"
            @click="deletingNodeId = null"
          >
            Cancel
          </button>
          <button
            class="px-4 py-2 text-sm bg-danger text-white rounded-xl hover:bg-danger/90 transition-colors shadow-md"
            @click="doDelete(deletingNodeId!)"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
