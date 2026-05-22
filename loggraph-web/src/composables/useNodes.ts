import { ref } from 'vue'
import type { Node } from '../types'
import { nodesApi } from '../api/nodes'

export function useNodes() {
  const projects = ref<Node[]>([])
  const people = ref<Node[]>([])

  async function fetchProjects() {
    const all = await nodesApi.list('project')
    const std = await nodesApi.list('standard')
    projects.value = [...all, ...std]
  }

  async function fetchPeople() {
    people.value = await nodesApi.list('person')
  }

  async function suggest(query: string, type: 'project' | 'person'): Promise<Node[]> {
    return nodesApi.suggest(query, type)
  }

  async function suggestBlocks(query: string): Promise<any[]> {
    const res = await fetch(`/api/v1/blocks?q=${encodeURIComponent(query)}&limit=5`)
    if (!res.ok) return []
    const page = await res.json()
    return page.data || []
  }

  return { projects, people, fetchProjects, fetchPeople, suggest, suggestBlocks }
}
