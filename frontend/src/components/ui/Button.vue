<script setup lang="ts">
import Spinner from "./Spinner.vue";

interface Props {
  variant?: "primary" | "outline" | "ghost";
  size?: "sm" | "md" | "lg";
  disabled?: boolean;
  isLoading?: boolean;
  type?: "button" | "submit" | "reset";
}

withDefaults(defineProps<Props>(), {
  variant: "primary",
  size: "md",
  disabled: false,
  isLoading: false,
  type: "button",
});
</script>

<template>
  <button
    :type="type"
    :disabled="disabled || isLoading"
    :class="[
      'inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 dark:focus:ring-gray-400 dark:focus:ring-offset-dark',
      {
        'px-3 py-1.5 text-xs': size === 'sm',
        'px-4 py-2 text-sm': size === 'md',
        'px-6 py-3 text-base': size === 'lg',
      },
      {
        'bg-blue-600 text-white hover:bg-blue-700 dark:bg-white dark:text-dark dark:hover:bg-gray-200': variant === 'primary',
        'border border-gray-300 bg-white text-gray-700 hover:bg-gray-50 dark:border-dark-border dark:bg-dark-card dark:text-gray-200 dark:hover:bg-dark-elevated':
          variant === 'outline',
        'text-gray-700 hover:bg-gray-100 dark:text-gray-200 dark:hover:bg-dark-elevated':
          variant === 'ghost',
      },
    ]"
  >
    <Spinner
      v-if="isLoading"
      size="sm"
      :class="variant === 'primary' ? 'text-white dark:text-dark' : 'text-blue-600 dark:text-white'"
    />
    <slot />
  </button>
</template>
