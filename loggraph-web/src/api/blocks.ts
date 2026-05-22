import type { Block, CursorPage, GraphData } from '../types'
import { api } from './client'

export interface BlockListParams {
  cursor?: string
  limit?: number
  project?: string
  person?: string
  status?: string
  q?: string
  since?: string
  until?: string
}

export const blocksApi = {
  list(params: BlockListParams = {}) {
    const qs = new URLSearchParams()
    if (params.cursor) qs.set('cursor', params.cursor)
    if (params.limit) qs.set('limit', String(params.limit))
    if (params.project) qs.set('project', params.project)
    if (params.person) qs.set('person', params.person)
    if (params.status) qs.set('status', params.status)
    if (params.q) qs.set('q', params.q)
    if (params.since) qs.set('since', params.since)
    if (params.until) qs.set('until', params.until)
    const q = qs.toString()
    return api.get<CursorPage<Block>>(`/blocks${q ? `?${q}` : ''}`)
  },

  create(data: { user_id?: string; content: string; metadata?: Record<string, any> }) {
    return api.post<Block>('/blocks', data)
  },

  get(id: string) {
    return api.get<Block>(`/blocks/${id}`)
  },

  update(id: string, data: { content?: string; status?: string; metadata?: Record<string, any> }) {
    return api.patch<Block>(`/blocks/${id}`, data)
  },

  delete(id: string) {
    return api.delete(`/blocks/${id}`)
  },

  graph(id: string) {
    return api.get<GraphData>(`/blocks/${id}/graph`)
  },

  nodeGraph(nodeId: string) {
    return api.get<GraphData>(`/nodes/${nodeId}/graph`)
  },
}
