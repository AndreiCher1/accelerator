<script setup lang="ts">
interface Props {
  padding?: "none" | "sm" | "md" | "lg";
  hover?: boolean;
}

withDefaults(defineProps<Props>(), {
  padding: "md",
  hover: false,
});
</script>

<template>
  <div
    :class="[
      'rounded-lg border border-gray-200 bg-white dark:border-dark-border dark:bg-dark-card',
      {
        'p-0': padding === 'none',
        'p-3': padding === 'sm',
        'p-4': padding === 'md',
        'p-6': padding === 'lg',
      }
    ]"
  >
    <slot />
  </div>
</template>
