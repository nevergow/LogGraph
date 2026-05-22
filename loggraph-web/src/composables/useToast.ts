import { ref } from 'vue'

export interface Toast {
  id: number
  message: string
  action?: { label: string; handler: () => void }
}

const toasts = ref<Toast[]>([])
const timers = new Map<number, ReturnType<typeof setTimeout>>()
let nextId = 0

export function useToast() {
  function showToast(message: string, action?: { label: string; handler: () => void }, duration = 3000) {
    const id = nextId++
    toasts.value.push({ id, message, action })
    const timer = setTimeout(() => dismissToast(id), duration)
    timers.set(id, timer)
  }

  function dismissToast(id: number) {
    const timer = timers.get(id)
    if (timer) {
      clearTimeout(timer)
      timers.delete(id)
    }
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  return { toasts, showToast, dismissToast }
}
