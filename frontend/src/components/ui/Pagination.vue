<script setup lang="ts">
import { ChevronLeft, ChevronRight } from "@lucide/vue";
import Button from "./Button.vue";

interface Props {
  currentPage: number;
  totalPages: number;
  totalItems: number;
  itemsPerPage: number;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  "page-change": [page: number];
}>();

const goToPage = (page: number) => {
  if (page >= 1 && page <= props.totalPages && page !== props.currentPage) {
    emit("page-change", page);
  }
};
</script>

<template>
  <div class="flex items-center justify-between">
    <p class="text-sm text-gray-700 dark:text-gray-300">
      Страница {{ currentPage }} из {{ totalPages }} • Всего: {{ totalItems }}
    </p>

    <div class="flex items-center gap-2">
      <Button
        variant="outline"
        size="sm"
        :disabled="currentPage === 1"
        @click="goToPage(currentPage - 1)"
      >
        <ChevronLeft :size="16" />
        Назад
      </Button>

      <Button
        variant="outline"
        size="sm"
        :disabled="currentPage === totalPages"
        @click="goToPage(currentPage + 1)"
      >
        Вперед
        <ChevronRight :size="16" />
      </Button>
    </div>
  </div>
</template>
