<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";

interface Props {
  align?: "left" | "right";
  empty?: boolean;
  emptyText?: string;
}

withDefaults(defineProps<Props>(), {
  align: "left",
  empty: false,
  emptyText: "Список пуст",
});

const isOpen = ref(false);
const dropdownRef = ref<HTMLElement | null>(null);

const toggle = () => {
  isOpen.value = !isOpen.value;
};

const close = () => {
  isOpen.value = false;
};

const handleClickOutside = (event: MouseEvent) => {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    close();
  }
};

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
});

defineExpose({ close });
</script>

<template>
  <div ref="dropdownRef" class="relative inline-block">
    <div @click="toggle">
      <slot name="trigger" :is-open="isOpen" />
    </div>

    <Transition
      enter-active-class="transition duration-100 ease-out"
      enter-from-class="transform scale-95 opacity-0"
      enter-to-class="transform scale-100 opacity-100"
      leave-active-class="transition duration-75 ease-in"
      leave-from-class="transform scale-100 opacity-100"
      leave-to-class="transform scale-95 opacity-0"
    >
      <div
        v-if="isOpen"
        :class="[
          'absolute z-50 mt-2 min-w-[12rem] rounded-lg border border-gray-200 bg-white py-1 shadow-lg dark:border-dark-border dark:bg-dark-card',
          {
            'left-0': align === 'left',
            'right-0': align === 'right',
          },
        ]"
      >
        <slot v-if="!empty" name="content" :close="close" />
        <slot v-else name="empty">
          <div
            class="px-4 py-3 text-center text-sm text-gray-400 dark:text-gray-500"
          >
            {{ emptyText }}
          </div>
        </slot>
      </div>
    </Transition>
  </div>
</template>
