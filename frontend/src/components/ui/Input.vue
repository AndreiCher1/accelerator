<script setup lang="ts">
import { computed, useAttrs } from "vue";
import {
  fieldControlClass,
  fieldControlSizeClass,
  fieldLabelClass,
  fieldWrapperClass,
} from "./fieldStyles";

defineOptions({ inheritAttrs: false });

const props = withDefaults(
  defineProps<{
    modelValue: string;
    label?: string;
    type?: "text" | "password" | "email" | "date" | "number";
    placeholder?: string;
    maxlength?: string | number;
    disabled?: boolean;
    error?: boolean;
    hideLabel?: boolean;
  }>(),
  {
    type: "text",
    disabled: false,
    error: false,
    hideLabel: false,
  },
);

const emit = defineEmits<{
  (e: "update:modelValue", value: string): void;
  (e: "input", event: Event): void;
}>();

const attrs = useAttrs();

const wrapperClass = computed(() =>
  props.label && !props.hideLabel ? fieldWrapperClass : "flex w-full flex-col",
);

const inputClass = computed(() => {
  const extra = typeof attrs.class === "string" ? attrs.class : "";
  return [fieldControlClass({ error: props.error, disabled: props.disabled }), fieldControlSizeClass, extra]
    .filter(Boolean)
    .join(" ");
});

const handleInput = (event: Event) => {
  const target = event.target as HTMLInputElement;
  emit("update:modelValue", target.value);
  emit("input", event);
};
</script>

<template>
  <div :class="wrapperClass">
    <label v-if="label && !hideLabel" :class="fieldLabelClass">
      {{ label }}
    </label>
    <input
      :type="type"
      :value="modelValue"
      :placeholder="placeholder"
      :maxlength="maxlength"
      :disabled="disabled"
      :class="inputClass"
      v-bind="{ ...attrs, class: undefined }"
      @input="handleInput"
    />
  </div>
</template>
