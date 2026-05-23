<script setup lang="ts">
import { useToast } from '../composables/useToast'

const { toasts, dismissToast } = useToast()
</script>

<template>
  <Teleport to="body">
    <div class="fixed top-4 right-4 z-50 flex flex-col gap-2 pointer-events-none max-w-sm">
      <TransitionGroup name="toast">
        <div
          v-for="t in toasts"
          :key="t.id"
          class="pointer-events-auto bg-white rounded-md shadow-elevated border border-gray-200 px-4 py-3 flex items-center gap-3 text-sm"
        >
          <span class="text-gray-700">{{ t.message }}</span>
          <button
            v-if="t.action"
            class="text-brand-600 font-medium hover:text-brand-800 shrink-0"
            @click="t.action.handler(); dismissToast(t.id)"
          >
            {{ t.action.label }}
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>
