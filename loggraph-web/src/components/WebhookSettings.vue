<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { WebhookToken } from '../types'
import { webhookApi, type TokenCreated } from '../api/webhook'

const emit = defineEmits<{ close: [] }>()

const tokens = ref<WebhookToken[]>([])
const newToken = ref<TokenCreated | null>(null)
const generating = ref(false)
const nameInput = ref('')
const error = ref('')
const baseURL = window.location.origin

onMounted(async () => {
  try { tokens.value = await webhookApi.listTokens() } catch { /* */ }
})

async function generate() {
  const name = nameInput.value.trim()
  if (!name) return
  generating.value = true
  error.value = ''
  try {
    newToken.value = await webhookApi.generateToken(name)
    nameInput.value = ''
    tokens.value = await webhookApi.listTokens()
  } catch (e: any) {
    error.value = e.message
  } finally {
    generating.value = false
  }
}

async function revoke(id: string) {
  try {
    await webhookApi.deleteToken(id)
    tokens.value = tokens.value.filter(t => t.id !== id)
  } catch { /* */ }
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex justify-end">
    <div class="absolute inset-0 bg-black/20" @click="emit('close')" />
    <div class="relative w-96 bg-white h-full shadow-xl overflow-y-auto">
      <div class="p-4 border-b border-gray-200 flex items-center justify-between">
        <h2 class="font-semibold text-sm">Webhook Settings</h2>
        <button class="text-gray-400 hover:text-gray-600 text-lg leading-none" @click="emit('close')">&times;</button>
      </div>

      <div class="p-4 space-y-4">
        <!-- New token -->
        <div>
          <label class="text-xs font-medium text-gray-500">Generate Token</label>
          <div class="flex gap-2 mt-1">
            <input
              v-model="nameInput"
              placeholder="e.g. CI Runner, Lark Bot"
              class="flex-1 text-sm border border-gray-200 rounded px-2 py-1.5 outline-none focus:border-blue-400"
              @keydown.enter="generate"
            />
            <button
              class="px-3 py-1.5 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 disabled:opacity-40 shrink-0"
              :disabled="generating || !nameInput.trim()"
              @click="generate"
            >
              {{ generating ? '...' : 'Create' }}
            </button>
          </div>
          <div v-if="error" class="text-xs text-red-500 mt-1">{{ error }}</div>
        </div>

        <!-- Freshly created token (shown once) -->
        <div v-if="newToken" class="bg-yellow-50 border border-yellow-200 rounded p-3">
          <div class="text-xs font-medium text-yellow-700 mb-1">Token created — copy it now:</div>
          <code class="text-xs break-all bg-yellow-100 px-2 py-1 rounded block mb-2">{{ newToken.token }}</code>
          <button class="text-xs text-blue-600 hover:text-blue-800" @click="copyToClipboard(newToken.token)">
            Copy
          </button>
          <button class="text-xs text-gray-400 ml-3" @click="newToken = null">Dismiss</button>
        </div>

        <!-- Usage examples -->
        <div class="bg-gray-50 rounded p-3">
          <div class="text-xs font-medium text-gray-500 mb-2">Usage (curl)</div>
          <pre class="text-[11px] text-gray-600 overflow-x-auto leading-relaxed"><code>curl -X POST {{ baseURL }}/api/v1/webhook/logs \
  -H "Authorization: Bearer &lt;token&gt;" \
  -H "Content-Type: application/json" \
  -d '{"content": "#项目A 测试通过","user_id": "ci"}'</code></pre>
        </div>

        <!-- Lark webhook hint -->
        <div class="bg-purple-50 rounded p-3">
          <div class="text-xs font-medium text-purple-700 mb-1">Lark Bot Webhook</div>
          <pre class="text-[11px] text-gray-600 overflow-x-auto leading-relaxed"><code>POST {{ baseURL }}/api/v1/webhook/lark
Authorization: Bearer &lt;token&gt;</code></pre>
          <div class="text-[11px] text-gray-500 mt-1">
            Configure this URL in Lark App → Event Subscriptions.
          </div>
        </div>

        <!-- Token list -->
        <div>
          <div class="text-xs font-medium text-gray-500 mb-2">Active Tokens</div>
          <div v-if="tokens.length === 0" class="text-xs text-gray-400">No tokens yet.</div>
          <ul class="space-y-1">
            <li
              v-for="t in tokens"
              :key="t.id"
              class="flex items-center justify-between text-sm px-2 py-1.5 rounded hover:bg-gray-50"
            >
              <div>
                <span class="font-medium">{{ t.name }}</span>
                <span class="text-xs text-gray-400 ml-2">{{ new Date(t.created_at).toLocaleDateString() }}</span>
              </div>
              <button
                class="text-xs text-red-500 hover:text-red-700"
                @click="revoke(t.id)"
              >
                Revoke
              </button>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>
