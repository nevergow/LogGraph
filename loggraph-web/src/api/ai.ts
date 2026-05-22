import { api } from './client'

export interface ReportResult {
  project: string
  period: string
  markdown: string
  block_count: number
}

export interface ReminderResult {
  has_reminder: boolean
  reminder_text: string | null
  due_at: string | null
}

export interface AISettings {
  base_url: string
  api_key: string
  model: string
}

export const aiApi = {
  generateReport(project: string, since?: string, until?: string) {
    return api.post<ReportResult>('/ai/report', { project, since, until })
  },

  parseReminder(content: string) {
    return api.post<ReminderResult>('/ai/reminders', { content })
  },

  getSettings() {
    return api.get<AISettings>('/ai/settings')
  },

  updateSettings(settings: AISettings) {
    return api.put<AISettings>('/ai/settings', settings)
  },
}
