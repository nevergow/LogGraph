<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { VueFlow, MarkerType, type Node as VFNode, type Edge as VFEdge } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import type { GraphData } from '../types'
import { blocksApi } from '../api/blocks'

const props = defineProps<{
  blockId: string | null
  projectNodeId: string | null
  projectName: string | null
}>()

const graphData = ref<GraphData | null>(null)
const loading = ref(false)
const mode = ref<'block' | 'project' | 'empty'>('empty')

async function fetchBlockGraph(id: string) {
  loading.value = true
  try {
    const raw = await blocksApi.graph(id)
    graphData.value = { ...raw, nodes: raw.nodes || [], edges: raw.edges || [] }
    mode.value = 'block'
  } catch {
    graphData.value = null
    mode.value = 'empty'
  } finally {
    loading.value = false
  }
}

async function fetchNodeGraph(id: string) {
  loading.value = true
  try {
    const raw = await blocksApi.nodeGraph(id)
    graphData.value = { ...raw, nodes: raw.nodes || [], edges: raw.edges || [] }
    mode.value = 'project'
  } catch {
    graphData.value = null
    mode.value = 'empty'
  } finally {
    loading.value = false
  }
}

watch(() => props.blockId, async (id) => {
  if (id) { await fetchBlockGraph(id); return }
  if (props.projectNodeId) { await fetchNodeGraph(props.projectNodeId); return }
  graphData.value = null
  mode.value = 'empty'
}, { immediate: true })

watch(() => props.projectNodeId, async (id) => {
  if (props.blockId) return // block view takes precedence
  if (id) { await fetchNodeGraph(id); return }
  graphData.value = null
  mode.value = 'empty'
})

const nodeTypeColors: Record<string, string> = {
  project: '#3b82f6',
  person: '#22c55e',
  standard: '#f97316',
  custom: '#6b7280',
}

function blockLabel(content: string): string {
  const cleaned = content.replace(/~~(.+?)~~/g, '$1').replace(/\[BLOCK\]/g, '').replace(/[#@^]\S+/g, '').trim()
  return cleaned.slice(0, 40) + (cleaned.length > 40 ? '...' : '') || '(empty)'
}

const vfNodes = computed<VFNode[]>(() => {
  if (!graphData.value) return []
  const g = graphData.value
  const nodes: VFNode[] = []

  if (mode.value === 'block') {
    // Central block
    nodes.push({
      id: g.block.id,
      type: 'default',
      position: { x: 250, y: 150 },
      data: { label: blockLabel(g.block.content), status: g.block.status },
      style: {
        background: '#eff6ff', border: '2px solid #3b82f6',
        borderRadius: '8px', padding: '10px 14px', fontSize: '12px', maxWidth: '220px',
      },
    })
    // Node entities around the center
    g.nodes.forEach((n, i) => {
      const angle = (2 * Math.PI * i) / g.nodes.length
      nodes.push({
        id: n.id, type: 'default',
        position: { x: 250 + 160 * Math.cos(angle), y: 150 + 160 * Math.sin(angle) },
        data: { label: (n.type === 'person' ? '@' : '#') + n.name, nodeType: n.type },
        style: {
          background: '#fff', border: `2px solid ${nodeTypeColors[n.type] || '#6b7280'}`,
          borderRadius: '20px', padding: '6px 12px', fontSize: '11px', fontWeight: 500,
        },
      })
    })
  } else {
    // Project-centric: center is the selected node, surrounding are blocks and other nodes
    const centerNode = g.nodes.find(n => n.id === props.projectNodeId)
    const otherNodes = g.nodes.filter(n => n.id !== props.projectNodeId)

    // Center: project/standard node
    nodes.push({
      id: props.projectNodeId || 'center',
      type: 'default',
      position: { x: 250, y: 150 },
      data: { label: (centerNode?.type === 'person' ? '@' : '#') + (centerNode?.name || props.projectName || '') },
      style: {
        background: '#fff', border: '3px solid #3b82f6',
        borderRadius: '24px', padding: '10px 18px', fontSize: '13px', fontWeight: 600,
      },
    })

    // Other related nodes in a ring
    otherNodes.forEach((n, i) => {
      const angle = (2 * Math.PI * i) / Math.max(otherNodes.length, 1)
      nodes.push({
        id: n.id, type: 'default',
        position: { x: 250 + 160 * Math.cos(angle), y: 150 + 160 * Math.sin(angle) },
        data: { label: (n.type === 'person' ? '@' : '#') + n.name, nodeType: n.type },
        style: {
          background: '#fff', border: `2px solid ${nodeTypeColors[n.type] || '#6b7280'}`,
          borderRadius: '20px', padding: '6px 12px', fontSize: '11px', fontWeight: 500,
        },
      })
    })
  }

  return nodes
})

const vfEdges = computed<VFEdge[]>(() => {
  if (!graphData.value) return []
  return graphData.value.edges.map((e, i) => ({
    id: `e-${i}`,
    source: e.source,
    target: e.target,
    label: e.label,
    type: 'default',
    markerEnd: { type: MarkerType.ArrowClosed, width: 16, height: 16 },
    style: { stroke: '#94a3b8', strokeWidth: 1.5 },
    labelStyle: { fill: '#94a3b8', fontSize: '9px' },
    labelBgStyle: { fill: '#f8fafc' },
  }))
})
</script>

<template>
  <aside class="border-l border-gray-200 bg-white shrink-0 overflow-hidden flex flex-col">
    <div v-if="!blockId && !projectNodeId" class="p-4 text-xs text-gray-400 text-center pt-12 flex-1">
      Click a block or select a project to see its graph
    </div>

    <div v-else-if="loading" class="p-4 text-xs text-gray-400 text-center pt-12 flex-1">
      Loading graph...
    </div>

    <div v-else-if="graphData" class="flex-1 flex flex-col">
      <!-- Header -->
      <div class="px-3 py-2 border-b border-gray-100 text-xs text-gray-500 flex gap-3 shrink-0">
        <span v-if="mode === 'project'" class="font-medium text-blue-600">{{ projectName || 'Project' }}</span>
        <span>{{ graphData.nodes.length }} nodes</span>
        <span>{{ graphData.edges.length }} edges</span>
        <span v-if="mode === 'project'" class="text-gray-300">{{ graphData.edges.filter(e => e.label === 'mentions').length }} blocks</span>
      </div>

      <!-- Vue Flow graph -->
      <div class="flex-1">
        <VueFlow
          :nodes="vfNodes"
          :edges="vfEdges"
          :default-viewport="{ x: 0, y: 0, zoom: 1 }"
          :min-zoom="0.3"
          :max-zoom="2"
          fit-view-on-init
        >
          <Background :gap="20" :size="1" />
        </VueFlow>
      </div>

      <!-- Detail footer -->
      <div class="border-t border-gray-100 px-3 py-2 shrink-0">
        <div v-if="mode === 'block'" class="text-[11px] text-gray-400 truncate">
          {{ graphData.block.content.slice(0, 80) }}
        </div>
        <div class="flex gap-2 mt-1 flex-wrap">
          <span
            v-for="node in graphData.nodes"
            :key="node.id"
            class="text-[10px] px-1.5 py-0.5 rounded-full border"
            :class="{
              'border-blue-200 bg-blue-50 text-blue-700': node.type === 'project',
              'border-green-200 bg-green-50 text-green-700': node.type === 'person',
              'border-orange-200 bg-orange-50 text-orange-700': node.type === 'standard',
            }"
          >
            {{ node.type === 'person' ? '@' : '#' }}{{ node.name }}
          </span>
        </div>
      </div>
    </div>
  </aside>
</template>
