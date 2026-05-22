export interface Block {
  id: string
  user_id: string
  content: string
  status: 'active' | 'completed' | 'blocked'
  metadata?: Record<string, any>
  created_at: string
  updated_at: string
}

export interface Node {
  id: string
  name: string
  type: 'project' | 'person' | 'standard' | 'custom'
  metadata?: Record<string, any>
  created_at: string
}

export interface GraphEdge {
  source: string
  target: string
  label: string
}

export interface GraphData {
  block: Block
  nodes: Node[]
  edges: GraphEdge[]
}

export interface CursorPage<T> {
  data: T[]
  next_cursor?: string
  has_more: boolean
}

export interface WebhookToken {
  id: string
  name: string
  created_at: string
  expires_at?: string
}
