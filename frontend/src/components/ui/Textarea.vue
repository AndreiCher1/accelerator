<script setup lang="ts">
import { computed, useAttrs } from "vue";
import {
  fieldControlClass,
  fieldLabelClass,
  fieldTextareaSizeClass,
  fieldWrapperClass,
} from "./fieldStyles";

defineOptions({ inheritAttrs: false });

const props = withDefaults(
  defineProps<{
    modelValue: string;
    label?: string;
    placeholder?: string;
    rows?: number;
    disabled?: boolean;
    error?: boolean;
    hideLabel?: boolean;
  }>(),
  {
    rows: 4,
    disabled: false,
    error: false,
    hideLabel: false,
  },
);

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

const attrs = useAttrs();

const wrapperClass = computed(() =>
  props.label && !props.hideLabel ? fieldWrapperClass : "flex w-full flex-col",
);

const textareaClass = computed(() => {
  const extra = typeof attrs.class === "string" ? attrs.class : "";
  return [
    fieldControlClass({ error: props.error, disabled: props.disabled }),
    fieldTextareaSizeClass,
    extra,
  ]
    .filter(Boolean)
    .join(" ");
});

const handleInput = (event: Event) => {
  const target = event.target as HTMLTextAreaElement;
  emit("update:modelValue", target.value);
};
</script>

<template>
  <div :class="wrapperClass">
    <label v-if="label && !hideLabel" :class="fieldLabelClass">
      {{ label }}
    </label>
    <textarea
      :value="modelValue"
      :placeholder="placeholder"
      :rows="rows"
      :disabled="disabled"
      :class="textareaClass"
      v-bind="{ ...attrs, class: undefined }"
      @input="handleInput"
    />
  </div>
</template>
