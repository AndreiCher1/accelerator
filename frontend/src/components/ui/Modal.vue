<script setup lang="ts">
import { onMounted, onUnmounted } from "vue";
import { X } from "@lucide/vue";

interface Props {
  modelValue: boolean;
  title?: string;
  size?: "sm" | "md" | "lg" | "xl";
  closeOnClickOutside?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  size: "md",
  closeOnClickOutside: true,
});

const emit = defineEmits<{
  "update:modelValue": [value: boolean];
}>();

const close = () => {
  emit("update:modelValue", false);
};

const handleEscape = (e: KeyboardEvent) => {
  if (e.key === "Escape" && props.modelValue) {
    close();
  }
};

onMounted(() => {
  document.addEventListener("keydown", handleEscape);
});

onUnmounted(() => {
  document.removeEventListener("keydown", handleEscape);
});
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="modelValue"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
        @click.self="closeOnClickOutside && close()"
      >
        <Transition
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="transform scale-95 opacity-0"
          enter-to-class="transform scale-100 opacity-100"
          leave-active-class="transition duration-150 ease-in"
          leave-from-class="transform scale-100 opacity-100"
          leave-to-class="transform scale-95 opacity-0"
        >
          <div
            v-if="modelValue"
            :class="[
              'relative w-full rounded-lg bg-white shadow-xl dark:bg-dark-card',
              {
                'max-w-sm': size === 'sm',
                'max-w-md': size === 'md',
                'max-w-2xl': size === 'lg',
                'max-w-4xl': size === 'xl',
              },
            ]"
          >
            <!-- Header -->
            <div
              v-if="title || $slots.header"
              class="flex items-center justify-between border-b border-gray-200 px-6 py-4 dark:border-dark-border"
            >
              <slot name="header">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ title }}
                </h3>
              </slot>
              <button
                type="button"
                class="rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:hover:bg-dark-elevated dark:hover:text-white"
                @click="close"
              >
                <X :size="20" />
              </button>
            </div>

            <!-- Body -->
            <div class="px-6 py-4">
              <slot />
            </div>

            <!-- Footer -->
            <div
              v-if="$slots.footer"
              class="flex items-center justify-end gap-3 border-t border-gray-200 px-6 py-4 dark:border-dark-border"
            >
              <slot name="footer" :close="close" />
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>
