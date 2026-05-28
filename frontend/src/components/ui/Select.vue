<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, useAttrs } from "vue";
import { ChevronDown, Check } from "@lucide/vue";
import {
  fieldControlClass,
  fieldControlSizeClass,
  fieldLabelClass,
  fieldSelectMenuClass,
  fieldSelectMenuScrollClass,
  fieldSelectOptionClass,
  fieldWrapperClass,
} from "./fieldStyles";

defineOptions({ inheritAttrs: false });

interface Option {
  value: string;
  label: string;
}

const props = withDefaults(
  defineProps<{
    modelValue: string;
    label?: string;
    options: (string | Option)[];
    disabled?: boolean;
    error?: boolean;
    placeholder?: string;
    hideLabel?: boolean;
  }>(),
  {
    disabled: false,
    error: false,
    placeholder: "Выберите...",
    hideLabel: false,
  },
);

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

const attrs = useAttrs();

const isOpen = ref(false);
const dropdownRef = ref<HTMLElement | null>(null);

const wrapperClass = computed(() => {
  const extra = typeof attrs.class === "string" ? attrs.class : "";
  const base = props.label && !props.hideLabel ? fieldWrapperClass : "flex w-full flex-col";
  return [base, extra].filter(Boolean).join(" ");
});

const normalizedOptions = computed(() =>
  props.options.map((opt) =>
    typeof opt === "string" ? { value: opt, label: opt } : opt,
  ),
);

const selectedLabel = computed(() => {
  const option = normalizedOptions.value.find((opt) => opt.value === props.modelValue);
  return option?.label ?? props.placeholder;
});

const isPlaceholder = computed(
  () => !normalizedOptions.value.some((opt) => opt.value === props.modelValue),
);

const triggerClass = computed(() =>
  [
    fieldControlClass({ error: props.error, disabled: props.disabled }),
    fieldControlSizeClass,
    "flex items-center justify-between gap-2 text-left",
    isPlaceholder.value ? "text-gray-400 dark:text-gray-500" : "",
  ]
    .filter(Boolean)
    .join(" "),
);

const selectOption = (value: string) => {
  emit("update:modelValue", value);
  isOpen.value = false;
};

const handleClickOutside = (event: MouseEvent) => {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    isOpen.value = false;
  }
};

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
});
</script>

<template>
  <div ref="dropdownRef" :class="wrapperClass">
    <label v-if="label && !hideLabel" :class="fieldLabelClass">
      {{ label }}
    </label>

    <div class="relative w-full">
      <button
        type="button"
        :disabled="disabled"
        :class="triggerClass"
        @click="isOpen = !isOpen"
      >
        <span class="min-w-0 flex-1 truncate">{{ selectedLabel }}</span>
        <ChevronDown
          :size="16"
          class="shrink-0 text-gray-400 transition-transform dark:text-gray-500"
          :class="{ 'rotate-180': isOpen }"
        />
      </button>

      <Transition
        enter-active-class="transition duration-150 ease-out"
        enter-from-class="-translate-y-1 opacity-0"
        enter-to-class="translate-y-0 opacity-100"
        leave-active-class="transition duration-100 ease-in"
        leave-from-class="translate-y-0 opacity-100"
        leave-to-class="-translate-y-1 opacity-0"
      >
        <div v-if="isOpen && !disabled" :class="fieldSelectMenuClass">
          <div :class="fieldSelectMenuScrollClass">
            <template v-if="normalizedOptions.length > 0">
              <button
                v-for="option in normalizedOptions"
                :key="option.value"
                type="button"
                :class="[
                  fieldSelectOptionClass,
                  option.value === modelValue
                    ? 'bg-blue-50 font-medium text-blue-600 dark:bg-white/10 dark:text-white'
                    : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-white/5',
                ]"
                @click="selectOption(option.value)"
              >
                <span class="truncate">{{ option.label }}</span>
                <Check
                  v-if="option.value === modelValue"
                  :size="16"
                  class="ml-2 shrink-0"
                />
              </button>
            </template>
            <div
              v-else
              class="px-3.5 py-2.5 text-center text-sm text-gray-400 dark:text-gray-500"
            >
              Список пуст
            </div>
          </div>
        </div>
      </Transition>
    </div>
  </div>
</template>
