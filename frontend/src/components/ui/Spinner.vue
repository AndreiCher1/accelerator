<script setup lang="ts">
interface Props {
  size?: "sm" | "md" | "lg";
}

withDefaults(defineProps<Props>(), {
  size: "md",
});
</script>

<template>
  <div
    :class="[
      'inline-block animate-spin rounded-full border-2 border-current border-t-transparent',
      {
        'h-4 w-4': size === 'sm',
        'h-6 w-6': size === 'md',
        'h-8 w-8': size === 'lg',
      },
    ]"
    role="status"
    aria-label="Загрузка"
  >
    <span class="sr-only">Загрузка...</span>
  </div>
</template>
