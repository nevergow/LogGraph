<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { aiApi, type ReportResult } from '../api/ai'
import { renderMarkdown } from '../composables/useMarkdown'
import { useNodes } from '../composables/useNodes'

const props = defineProps<{
  currentProject?: string
  currentSince?: string
  currentUntil?: string
}>()

const emit = defineEmits<{ close: [] }>()

const { projects, fetchProjects } = useNodes()

const tab = ref<'report' | 'settings'>('report')

// Report state
const selectedProjects = ref<string[]>([])
const since = ref('')
const until = ref('')
const loading = ref(false)
const result = ref<ReportResult | null>(null)
const error = ref('')

// Settings state
const settingsBaseUrl = ref('')
const settingsApiKey = ref('')
const settingsModel = ref('')
const settingsLoading = ref(false)
const settingsMsg = ref('')

function toggleProject(name: string) {
  const idx = selectedProjects.value.indexOf(name)
  if (idx >= 0) {
    selectedProjects.value.splice(idx, 1)
  } else {
    selectedProjects.value.push(name)
  }
}

onMounted(() => {
  if (props.currentProject) selectedProjects.value = [props.currentProject]
  if (props.currentSince) since.value = props.currentSince.split('T')[0]
  if (props.currentUntil) until.value = props.currentUntil.split('T')[0]
  fetchProjects()
  fetchSettings()
})

async function fetchSettings() {
  try {
    const s = await aiApi.getSettings()
    settingsBaseUrl.value = s.base_url
    settingsApiKey.value = s.api_key
    settingsModel.value = s.model
  } catch { /* ignore */ }
}

async function generate() {
  if (selectedProjects.value.length === 0) return
  loading.value = true
  error.value = ''
  result.value = null
  try {
    result.value = await aiApi.generateReport(
      selectedProjects.value[0],
      since.value || undefined,
      until.value || undefined,
    )
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function saveSettings() {
  settingsLoading.value = true
  settingsMsg.value = ''
  try {
    await aiApi.updateSettings({
      base_url: settingsBaseUrl.value,
      api_key: settingsApiKey.value,
      model: settingsModel.value,
    })
    settingsMsg.value = 'Saved'
    setTimeout(() => settingsMsg.value = '', 2000)
  } catch (e: any) {
    settingsMsg.value = 'Error: ' + e.message
  } finally {
    settingsLoading.value = false
  }
}

</script>

<template>
  <div class="fixed inset-0 z-50 flex justify-end">
    <div class="absolute inset-0 bg-black/20" @click="emit('close')" />
    <div class="relative w-96 bg-white h-full shadow-xl overflow-y-auto">
      <div class="p-4 border-b border-gray-200 flex items-center justify-between">
        <div class="flex gap-3">
          <button
            class="text-sm font-medium pb-0.5 border-b-2 transition-colors"
            :class="tab === 'report' ? 'text-brand-600 border-brand-600' : 'text-gray-400 border-transparent hover:text-gray-600'"
            @click="tab = 'report'"
          >Report</button>
          <button
            class="text-sm font-medium pb-0.5 border-b-2 transition-colors"
            :class="tab === 'settings' ? 'text-brand-600 border-brand-600' : 'text-gray-400 border-transparent hover:text-gray-600'"
            @click="tab = 'settings'"
          >Settings</button>
        </div>
        <button class="text-gray-400 hover:text-gray-600 text-lg leading-none" @click="emit('close')">&times;</button>
      </div>

      <!-- Report tab -->
      <div v-if="tab === 'report'" class="p-4 space-y-3">
        <div>
          <label class="text-xs font-medium text-gray-500">Projects</label>
          <div class="mt-1 flex flex-wrap gap-2">
            <button
              v-for="p in projects"
              :key="p.name"
              class="text-xs px-3 py-1.5 rounded-xl font-medium transition-colors border"
              :class="selectedProjects.includes(p.name)
                ? 'bg-accent-600 text-white border-accent-600'
                : 'bg-surface-100 text-text-secondary border-border-light hover:bg-accent-50 hover:text-accent-600'"
              @click="toggleProject(p.name)"
            >
              &amp;{{ p.name }}
            </button>
          </div>
          <p v-if="projects.length === 0" class="text-xs text-text-muted mt-1">No projects found</p>
        </div>
        <div class="flex gap-2">
          <div class="flex-1">
            <label class="text-xs text-text-muted">Since</label>
            <input v-model="since" type="date" class="w-full text-sm border border-slate-200 rounded px-2 py-1.5 mt-0.5 outline-none focus:border-accent-400" />
          </div>
          <div class="flex-1">
            <label class="text-xs text-text-muted">Until</label>
            <input v-model="until" type="date" class="w-full text-sm border border-slate-200 rounded px-2 py-1.5 mt-0.5 outline-none focus:border-accent-400" />
          </div>
        </div>

        <button
          class="w-full py-2 bg-accent-600 text-white text-sm rounded-lg hover:bg-accent-700 disabled:opacity-40 transition-colors font-semibold"
          :disabled="loading || selectedProjects.length === 0"
          @click="generate"
        >
          {{ loading ? 'Generating...' : 'Generate Report' }}
        </button>

        <div v-if="error" class="text-xs text-red-500 bg-red-50 rounded p-2">{{ error }}</div>

        <div v-if="result" class="bg-gray-50 rounded-sm p-4">
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs font-medium text-gray-500">{{ result.project }}</span>
            <span class="text-[10px] text-gray-400">{{ result.block_count }} blocks · {{ result.period }}</span>
          </div>
          <div class="text-sm leading-relaxed prose prose-sm max-w-none" v-html="renderMarkdown(result.markdown)" />
        </div>
      </div>

      <!-- Settings tab -->
      <div v-if="tab === 'settings'" class="p-4 space-y-3">
        <div>
          <label class="text-xs font-medium text-gray-500">API Base URL</label>
          <input
            v-model="settingsBaseUrl"
            placeholder="https://api.deepseek.com/v1"
            class="w-full text-sm border border-gray-200 rounded px-2 py-1.5 mt-1 outline-none focus:border-brand-400"
          />
        </div>
        <div>
          <label class="text-xs font-medium text-gray-500">API Key</label>
          <input
            v-model="settingsApiKey"
            type="password"
            placeholder="sk-..."
            class="w-full text-sm border border-gray-200 rounded px-2 py-1.5 mt-1 outline-none focus:border-brand-400"
          />
        </div>
        <div>
          <label class="text-xs font-medium text-gray-500">Model</label>
          <input
            v-model="settingsModel"
            placeholder="deepseek-chat"
            class="w-full text-sm border border-gray-200 rounded px-2 py-1.5 mt-1 outline-none focus:border-brand-400"
          />
        </div>

        <button
          class="w-full py-2 bg-brand-500 text-white text-sm rounded-sm hover:bg-brand-600 disabled:opacity-40 transition-colors"
          :disabled="settingsLoading"
          @click="saveSettings"
        >
          {{ settingsLoading ? 'Saving...' : 'Save Settings' }}
        </button>

        <div v-if="settingsMsg" class="text-xs rounded p-2" :class="settingsMsg.startsWith('Error') ? 'text-red-500 bg-red-50' : 'text-green-600 bg-green-50'">
          {{ settingsMsg }}
        </div>

        <div class="text-[10px] text-gray-400 text-center pt-2">
          Settings are stored in memory. Restarting the AI service resets to .env values.
        </div>
      </div>

      <div class="px-4 pb-4 text-[10px] text-gray-400 text-center">
        Powered by LLM · Data from Go API
      </div>
    </div>
  </div>
</template>
