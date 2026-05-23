<script setup lang="ts">
defineProps<{ count?: number }>()
</script>

<template>
  <div class="space-y-2 px-4 py-3">
    <div
      v-for="i in (count || 3)"
      :key="i"
      class="bg-white rounded-md px-4 py-3 border-l-[3px] border-l-gray-200 animate-pulse"
    >
      <div class="flex justify-end mb-2">
        <div class="w-5 h-5 bg-gray-100 rounded-sm" />
      </div>
      <div class="space-y-2">
        <div class="h-3 bg-gray-100 rounded-sm w-3/4" />
        <div class="h-3 bg-gray-100 rounded-sm w-full" />
        <div class="h-3 bg-gray-100 rounded-sm w-1/2" />
      </div>
    </div>
  </div>
</template>
