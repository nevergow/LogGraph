<script setup lang="ts">
import { useToast } from '../composables/useToast'

const { toasts, dismissToast } = useToast()
</script>

<template>
  <Teleport to="body">
    <div class="fixed bottom-20 left-1/2 -translate-x-1/2 z-50 flex flex-col gap-2 pointer-events-none">
      <TransitionGroup name="fade">
        <div
          v-for="t in toasts"
          :key="t.id"
          class="pointer-events-auto bg-white rounded-lg shadow-lg border border-slate-200 px-4 py-3 flex items-center gap-3 text-sm"
        >
          <span class="text-slate-700">{{ t.message }}</span>
          <button
            v-if="t.action"
            class="text-blue-600 font-medium hover:text-blue-800 shrink-0"
            @click="t.action.handler(); dismissToast(t.id)"
          >
            {{ t.action.label }}
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>
