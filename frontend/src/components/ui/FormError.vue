<script setup lang="ts">
import { CircleX } from "@lucide/vue";

interface Props {
  message?: string | null;
}

defineProps<Props>();
</script>

<template>
  <Transition
    enter-active-class="transition duration-150 ease-out"
    enter-from-class="opacity-0 -translate-y-1"
    enter-to-class="opacity-100 translate-y-0"
    leave-active-class="transition duration-100 ease-in"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div
      v-if="message"
      class="flex items-start gap-2 rounded-xl bg-red-50 p-3 text-sm text-red-600 dark:bg-red-500/10 dark:text-red-400"
      role="alert"
    >
      <CircleX :size="16" class="mt-0.5 flex-shrink-0" />
      <span class="break-words">{{ message }}</span>
    </div>
  </Transition>
</template>
