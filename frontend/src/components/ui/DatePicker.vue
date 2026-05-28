<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { Calendar, ChevronLeft, ChevronRight } from "@lucide/vue";
import {
  fieldControlClass,
  fieldControlSizeClass,
  fieldLabelClass,
  fieldSelectMenuClass,
  fieldWrapperClass,
} from "./fieldStyles";

const props = withDefaults(
  defineProps<{
    modelValue: string;
    label?: string;
    placeholder?: string;
    disabled?: boolean;
    error?: boolean;
    hideLabel?: boolean;
    min?: string;
    max?: string;
  }>(),
  {
    placeholder: "дд.мм.гггг",
    disabled: false,
    error: false,
    hideLabel: false,
  },
);

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

const MONTHS = [
  "Январь",
  "Февраль",
  "Март",
  "Апрель",
  "Май",
  "Июнь",
  "Июль",
  "Август",
  "Сентябрь",
  "Октябрь",
  "Ноябрь",
  "Декабрь",
];
const WEEKDAYS = ["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"];

const isOpen = ref(false);
const rootRef = ref<HTMLElement | null>(null);
const view = ref<"days" | "months" | "years">("days");

const today = new Date();
today.setHours(0, 0, 0, 0);

const parseISO = (value: string): Date | null => {
  if (!value) return null;
  const match = /^(\d{4})-(\d{2})-(\d{2})/.exec(value);
  if (!match) return null;
  const d = new Date(Number(match[1]), Number(match[2]) - 1, Number(match[3]));
  d.setHours(0, 0, 0, 0);
  return Number.isNaN(d.getTime()) ? null : d;
};

const formatISO = (date: Date): string => {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
};

const formatDisplay = (date: Date): string => {
  const d = String(date.getDate()).padStart(2, "0");
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const y = date.getFullYear();
  return `${d}.${m}.${y}`;
};

const sameDay = (a: Date, b: Date) =>
  a.getFullYear() === b.getFullYear() &&
  a.getMonth() === b.getMonth() &&
  a.getDate() === b.getDate();

const minDate = computed(() => parseISO(props.min ?? ""));
const maxDate = computed(() => parseISO(props.max ?? ""));

const isDisabledDate = (date: Date) => {
  if (minDate.value && date < minDate.value) return true;
  if (maxDate.value && date > maxDate.value) return true;
  return false;
};

const selectedDate = computed(() => parseISO(props.modelValue));

const displayValue = computed(() =>
  selectedDate.value ? formatDisplay(selectedDate.value) : "",
);

const cursor = ref(new Date(today.getFullYear(), today.getMonth(), 1));

watch(
  () => props.modelValue,
  (v) => {
    const d = parseISO(v);
    if (d) {
      cursor.value = new Date(d.getFullYear(), d.getMonth(), 1);
    }
  },
  { immediate: true },
);

interface DayCell {
  date: Date;
  inMonth: boolean;
  isToday: boolean;
  isSelected: boolean;
  isWeekend: boolean;
  disabled: boolean;
}

const calendarDays = computed<DayCell[]>(() => {
  const year = cursor.value.getFullYear();
  const month = cursor.value.getMonth();

  const firstOfMonth = new Date(year, month, 1);
  // ISO weekday: Mon = 0 ... Sun = 6
  const startWeekday = (firstOfMonth.getDay() + 6) % 7;

  const start = new Date(year, month, 1 - startWeekday);

  const cells: DayCell[] = [];
  for (let i = 0; i < 42; i++) {
    const date = new Date(start.getFullYear(), start.getMonth(), start.getDate() + i);
    const weekday = (date.getDay() + 6) % 7; // 0..6, 5/6 -> сб/вс
    cells.push({
      date,
      inMonth: date.getMonth() === month,
      isToday: sameDay(date, today),
      isSelected: !!selectedDate.value && sameDay(date, selectedDate.value),
      isWeekend: weekday >= 5,
      disabled: isDisabledDate(date),
    });
  }
  return cells;
});

const monthYearLabel = computed(
  () => `${MONTHS[cursor.value.getMonth()]} ${cursor.value.getFullYear()}`,
);

const yearGridStart = computed(() => {
  const y = cursor.value.getFullYear();
  return y - (y % 12);
});

const prev = () => {
  if (view.value === "days") {
    cursor.value = new Date(cursor.value.getFullYear(), cursor.value.getMonth() - 1, 1);
  } else if (view.value === "months") {
    cursor.value = new Date(cursor.value.getFullYear() - 1, cursor.value.getMonth(), 1);
  } else {
    cursor.value = new Date(cursor.value.getFullYear() - 12, cursor.value.getMonth(), 1);
  }
};

const next = () => {
  if (view.value === "days") {
    cursor.value = new Date(cursor.value.getFullYear(), cursor.value.getMonth() + 1, 1);
  } else if (view.value === "months") {
    cursor.value = new Date(cursor.value.getFullYear() + 1, cursor.value.getMonth(), 1);
  } else {
    cursor.value = new Date(cursor.value.getFullYear() + 12, cursor.value.getMonth(), 1);
  }
};

const toggleHeader = () => {
  view.value = view.value === "days" ? "months" : view.value === "months" ? "years" : "days";
};

const pickMonth = (m: number) => {
  cursor.value = new Date(cursor.value.getFullYear(), m, 1);
  view.value = "days";
};

const pickYear = (y: number) => {
  cursor.value = new Date(y, cursor.value.getMonth(), 1);
  view.value = "months";
};

const pickDay = (cell: DayCell) => {
  if (cell.disabled) return;
  emit("update:modelValue", formatISO(cell.date));
  isOpen.value = false;
};

const pickToday = () => {
  if (isDisabledDate(today)) return;
  cursor.value = new Date(today.getFullYear(), today.getMonth(), 1);
  emit("update:modelValue", formatISO(today));
  isOpen.value = false;
};

const clear = () => {
  emit("update:modelValue", "");
};

const toggleOpen = () => {
  if (props.disabled) return;
  isOpen.value = !isOpen.value;
  if (isOpen.value) view.value = "days";
};

const handleClickOutside = (event: MouseEvent) => {
  if (rootRef.value && !rootRef.value.contains(event.target as Node)) {
    isOpen.value = false;
  }
};

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === "Escape" && isOpen.value) {
    isOpen.value = false;
  }
};

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
  document.addEventListener("keydown", handleKeydown);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
  document.removeEventListener("keydown", handleKeydown);
});

const wrapperClass = computed(() =>
  props.label && !props.hideLabel ? fieldWrapperClass : "flex w-full flex-col",
);

const triggerClass = computed(() =>
  [
    fieldControlClass({ error: props.error, disabled: props.disabled }),
    fieldControlSizeClass,
    "flex items-center gap-2 text-left",
    !displayValue.value ? "text-gray-400 dark:text-gray-500" : "",
  ]
    .filter(Boolean)
    .join(" "),
);
</script>

<template>
  <div ref="rootRef" :class="wrapperClass">
    <label v-if="label && !hideLabel" :class="fieldLabelClass">
      {{ label }}
    </label>

    <div class="relative w-full">
      <button type="button" :disabled="disabled" :class="triggerClass" @click="toggleOpen">
        <span class="min-w-0 flex-1 truncate">
          {{ displayValue || placeholder }}
        </span>
        <Calendar
          :size="16"
          class="shrink-0 text-gray-400 dark:text-gray-500"
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
        <div
          v-if="isOpen && !disabled"
          :class="[fieldSelectMenuClass, 'w-[20rem] p-3']"
        >
          <!-- header -->
          <div class="mb-2 flex items-center justify-between gap-2">
            <button
              type="button"
              class="flex h-8 w-8 items-center justify-center rounded-lg text-gray-500 transition-colors hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-white/10"
              @click="prev"
            >
              <ChevronLeft :size="16" />
            </button>

            <button
              type="button"
              class="rounded-lg px-3 py-1 text-sm font-semibold text-gray-900 transition-colors hover:bg-gray-100 dark:text-white dark:hover:bg-white/10"
              @click="toggleHeader"
            >
              <template v-if="view === 'days'">{{ monthYearLabel }}</template>
              <template v-else-if="view === 'months'">{{ cursor.getFullYear() }}</template>
              <template v-else>{{ yearGridStart }} – {{ yearGridStart + 11 }}</template>
            </button>

            <button
              type="button"
              class="flex h-8 w-8 items-center justify-center rounded-lg text-gray-500 transition-colors hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-white/10"
              @click="next"
            >
              <ChevronRight :size="16" />
            </button>
          </div>

          <!-- days view -->
          <div v-if="view === 'days'">
            <div class="mb-1 grid grid-cols-7 text-center text-[11px] font-medium uppercase tracking-wide">
              <div
                v-for="(d, i) in WEEKDAYS"
                :key="d"
                :class="
                  i >= 5
                    ? 'text-rose-500 dark:text-rose-400'
                    : 'text-gray-500 dark:text-gray-400'
                "
              >
                {{ d }}
              </div>
            </div>

            <div class="grid grid-cols-7 gap-y-0.5">
              <button
                v-for="(cell, idx) in calendarDays"
                :key="idx"
                type="button"
                :disabled="cell.disabled"
                :class="[
                  'relative mx-auto flex h-9 w-9 items-center justify-center rounded-lg text-sm font-medium transition-colors',
                  cell.disabled ? 'cursor-not-allowed opacity-40' : '',
                  cell.isSelected
                    ? 'bg-blue-600 text-white shadow-sm hover:bg-blue-600 dark:bg-white dark:text-dark dark:hover:bg-white'
                    : !cell.inMonth
                      ? cell.isWeekend
                        ? 'text-rose-300 hover:bg-gray-100 dark:text-rose-400/40 dark:hover:bg-white/5'
                        : 'text-gray-400 hover:bg-gray-100 dark:text-gray-500 dark:hover:bg-white/5'
                      : cell.isWeekend
                        ? 'text-rose-500 hover:bg-rose-50 dark:text-rose-400 dark:hover:bg-white/10'
                        : 'text-gray-800 hover:bg-gray-100 dark:text-gray-100 dark:hover:bg-white/10',
                  cell.isToday && !cell.isSelected
                    ? 'ring-1 ring-blue-500/60 dark:ring-white/40'
                    : '',
                ]"
                @click="pickDay(cell)"
              >
                {{ cell.date.getDate() }}
              </button>
            </div>
          </div>

          <!-- months view -->
          <div v-else-if="view === 'months'" class="grid grid-cols-3 gap-2">
            <button
              v-for="(m, i) in MONTHS"
              :key="m"
              type="button"
              :class="[
                'rounded-lg px-2 py-2 text-sm font-medium transition-colors',
                i === cursor.getMonth() &&
                selectedDate &&
                selectedDate.getFullYear() === cursor.getFullYear()
                  ? 'bg-blue-600 text-white dark:bg-white dark:text-dark'
                  : 'text-gray-700 hover:bg-gray-100 dark:text-gray-200 dark:hover:bg-white/10',
              ]"
              @click="pickMonth(i)"
            >
              {{ m.slice(0, 3) }}
            </button>
          </div>

          <!-- years view -->
          <div v-else class="grid grid-cols-3 gap-2">
            <button
              v-for="i in 12"
              :key="i"
              type="button"
              :class="[
                'rounded-lg px-2 py-2 text-sm font-medium transition-colors',
                yearGridStart + i - 1 === cursor.getFullYear()
                  ? 'bg-blue-600 text-white dark:bg-white dark:text-dark'
                  : 'text-gray-700 hover:bg-gray-100 dark:text-gray-200 dark:hover:bg-white/10',
              ]"
              @click="pickYear(yearGridStart + i - 1)"
            >
              {{ yearGridStart + i - 1 }}
            </button>
          </div>

          <!-- footer -->
          <div
            class="mt-3 flex items-center justify-between border-t border-gray-100 pt-2 dark:border-white/10"
          >
            <button
              type="button"
              class="rounded-lg px-2 py-1 text-xs font-medium text-blue-600 transition-colors hover:bg-blue-50 dark:text-white dark:hover:bg-white/10"
              @click="pickToday"
            >
              Сегодня
            </button>
            <button
              type="button"
              class="rounded-lg px-2 py-1 text-xs font-medium text-gray-500 transition-colors hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-white/10"
              @click="clear"
            >
              Очистить
            </button>
          </div>
        </div>
      </Transition>
    </div>
  </div>
</template>
