import type { Node } from '../types'
import { api } from './client'

export const nodesApi = {
  list(type?: string) {
    const qs = type ? `?type=${encodeURIComponent(type)}` : ''
    return api.get<Node[]>(`/nodes${qs}`)
  },

  suggest(query: string, type?: string, limit = 8) {
    const qs = new URLSearchParams({ q: query, limit: String(limit) })
    if (type) qs.set('type', type)
    return api.get<Node[]>(`/nodes/suggest?${qs}`)
  },

  create(name: string, type: string) {
    return api.post<Node>('/nodes', { name, type })
  },

  update(id: string, name: string, cascade = false) {
    return api.patch<Node>(`/nodes/${encodeURIComponent(id)}`, { name, cascade })
  },

  delete(id: string) {
    return api.delete(`/nodes/${encodeURIComponent(id)}`)
  },
}
