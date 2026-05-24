import { ref } from 'vue'
import type { Node } from '../types'
import { nodesApi } from '../api/nodes'

// Module-level singletons — all callers share the same reactive refs (Bug 3 fix)
const projects = ref<Node[]>([])
const people = ref<Node[]>([])

export function useNodes() {
  async function fetchProjects() {
    try {
      projects.value = await nodesApi.list('project')
    } catch (e) {
      console.error('Failed to fetch projects:', e)
    }
  }

  async function fetchPeople() {
    try {
      people.value = await nodesApi.list('person')
    } catch (e) {
      console.error('Failed to fetch people:', e)
    }
  }

  async function suggest(query: string, type?: string, limit = 5): Promise<Node[]> {
    return nodesApi.suggest(query, type, limit)
  }

  async function suggestBlocks(query: string): Promise<any[]> {
    const res = await fetch(`/api/v1/blocks?q=${encodeURIComponent(query)}&limit=5`)
    if (!res.ok) return []
    const page = await res.json()
    return page.data || []
  }

  return { projects, people, fetchProjects, fetchPeople, suggest, suggestBlocks }
}
