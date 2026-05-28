<script setup lang="ts">
defineProps<{
  modelValue: boolean;
}>();

defineEmits<{
  (e: "update:modelValue", value: boolean): void;
}>();
</script>

<template>
  <label class="flex cursor-pointer items-center gap-2">
    <div class="relative flex items-center">
      <input
        type="checkbox"
        :checked="modelValue"
        @change="$emit('update:modelValue', ($event.target as HTMLInputElement).checked)"
        class="peer focus:ring-primary checked:bg-primary checked:border-primary h-5 w-5 cursor-pointer appearance-none rounded-md border border-gray-300 transition-colors dark:border-dark-border dark:bg-dark-card dark:checked:bg-white dark:checked:border-white"
      />
      <svg
        class="pointer-events-none absolute top-0.5 left-0.5 h-4 w-4 text-white opacity-0 peer-checked:opacity-100 dark:text-dark"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        stroke-width="3"
      >
        <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
      </svg>
    </div>
    <span class="text-sm font-medium text-[#606060] select-none dark:text-gray-400">
      <slot></slot>
    </span>
  </label>
</template>
