<script setup lang="ts">
interface Props {
  variant?: "success" | "warning" | "error" | "info" | "default";
  size?: "sm" | "md";
}

withDefaults(defineProps<Props>(), {
  variant: "default",
  size: "md",
});
</script>

<template>
  <span
    :class="[
      'inline-flex items-center gap-1.5 rounded-full font-medium',
      {
        'px-2.5 py-0.5 text-xs': size === 'sm',
        'px-3 py-1 text-sm': size === 'md',
      },
      {
        'bg-green-100 text-green-700 dark:bg-green-500/20 dark:text-green-400':
          variant === 'success',
        'bg-yellow-100 text-yellow-700 dark:bg-yellow-500/20 dark:text-yellow-400':
          variant === 'warning',
        'bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-400': variant === 'error',
        'bg-blue-100 text-blue-700 dark:bg-white/10 dark:text-gray-300': variant === 'info',
        'bg-gray-100 text-gray-700 dark:bg-white/10 dark:text-gray-400':
          variant === 'default',
      },
    ]"
  >
    <slot />
  </span>
</template>
