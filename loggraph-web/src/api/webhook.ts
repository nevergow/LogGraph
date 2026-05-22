import { api } from './client'
import type { WebhookToken } from '../types'

export interface TokenCreated {
  name: string
  token: string
}

export const webhookApi = {
  listTokens() {
    return api.get<WebhookToken[]>('/webhook/tokens')
  },

  generateToken(name: string) {
    return api.post<TokenCreated>('/webhook/tokens', { name })
  },

  deleteToken(id: string) {
    return api.delete(`/webhook/tokens/${id}`)
  },
}
