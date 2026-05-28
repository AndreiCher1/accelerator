import { defineStore } from "pinia";
import { computed, ref } from "vue";
import {
  ApiError,
  extractApiErrorMessage,
  groupService,
  type CreateGroupRequest,
  type EditGroupRequest,
  type GetGroupMembersResponse,
  type GroupDto,
} from "../api";

const CACHE_TTL_MS = 5 * 60 * 1000;
const ACTIVE_GROUP_STORAGE_KEY = "active_group_id";

interface MembersCacheEntry {
  data: GetGroupMembersResponse;
  fetchedAt: number;
}

export const GROUP_COLOR_PRESETS = [
  "bg-blue-100 text-blue-600 dark:bg-white/10 dark:text-gray-300",
  "bg-orange-100 text-orange-600 dark:bg-orange-500/20 dark:text-orange-400",
  "bg-green-100 text-green-600 dark:bg-green-500/20 dark:text-green-400",
  "bg-purple-100 text-purple-600 dark:bg-purple-500/20 dark:text-purple-400",
  "bg-pink-100 text-pink-600 dark:bg-pink-500/20 dark:text-pink-400",
  "bg-red-100 text-red-600 dark:bg-red-500/20 dark:text-red-400",
] as const;

function hashString(value: string): number {
  let hash = 0;
  for (let i = 0; i < value.length; i++) {
    hash = (hash << 5) - hash + value.charCodeAt(i);
    hash |= 0;
  }
  return Math.abs(hash);
}

export function groupColorClass(groupId: string): string {
  return GROUP_COLOR_PRESETS[hashString(groupId) % GROUP_COLOR_PRESETS.length];
}

export function groupPrefix(name: string): string {
  return name.trim().charAt(0).toUpperCase() || "?";
}

export function memberConflictMessage(err: unknown): string {
  if (err instanceof ApiError && err.isConflict) {
    return err.message;
  }
  return extractApiErrorMessage(err, "Не удалось изменить состав группы");
}

export const useGroupsStore = defineStore("groups", () => {
  const groups = ref<GroupDto[]>([]);
  const groupsFetchedAt = ref<number | null>(null);

  const activeGroupId = ref<string | null>(localStorage.getItem(ACTIVE_GROUP_STORAGE_KEY));
  const activeGroupDetails = ref<GetGroupMembersResponse | null>(null);

  const membersCache = ref<Map<string, MembersCacheEntry>>(new Map());

  const isLoading = ref(false);
  const isMembersLoading = ref(false);
  const isMutating = ref(false);

  const error = ref<string | null>(null);
  const membersError = ref<string | null>(null);
  const forbidden = ref(false);

  const activeGroup = computed(() => {
    if (!activeGroupId.value) return null;
    return groups.value.find((g) => g.group_id === activeGroupId.value) ?? null;
  });

  const isGroupsCacheValid = computed(() => {
    if (groupsFetchedAt.value === null) return false;
    return Date.now() - groupsFetchedAt.value < CACHE_TTL_MS;
  });

  function invalidateCache(): void {
    groupsFetchedAt.value = null;
    membersCache.value.clear();
  }

  function invalidateMembersCache(groupId?: string): void {
    if (groupId) {
      membersCache.value.delete(groupId);
    } else {
      membersCache.value.clear();
    }
  }

  async function fetchGroups(options?: { force?: boolean }): Promise<void> {
    const force = options?.force ?? false;

    if (!force && isGroupsCacheValid.value && groups.value.length > 0) {
      error.value = null;
      forbidden.value = false;
      return;
    }

    isLoading.value = true;
    error.value = null;
    forbidden.value = false;

    try {
      const data = await groupService.getGroups();
      groups.value = data.groups;
      groupsFetchedAt.value = Date.now();

      if (groups.value.length === 0) {
        activeGroupId.value = null;
        localStorage.removeItem(ACTIVE_GROUP_STORAGE_KEY);
      } else if (
        !activeGroupId.value ||
        !groups.value.some((g) => g.group_id === activeGroupId.value)
      ) {
        activeGroupId.value = groups.value[0].group_id;
      }

      if (activeGroupId.value) {
        localStorage.setItem(ACTIVE_GROUP_STORAGE_KEY, activeGroupId.value);
      }
    } catch (err: unknown) {
      if (err instanceof ApiError && err.isForbidden) {
        forbidden.value = true;
        error.value = "Недостаточно прав для просмотра групп";
      } else {
        error.value = extractApiErrorMessage(err, "Не удалось загрузить группы");
      }
      throw err;
    } finally {
      isLoading.value = false;
    }
  }

  async function fetchGroupMembers(
    groupId: string,
    options?: { force?: boolean },
  ): Promise<void> {
    const force = options?.force ?? false;
    persistActiveGroup(groupId);

    const cached = membersCache.value.get(groupId);
    if (!force && cached && Date.now() - cached.fetchedAt < CACHE_TTL_MS) {
      activeGroupDetails.value = cached.data;
      membersError.value = null;
      return;
    }

    isMembersLoading.value = true;
    membersError.value = null;

    try {
      const data = await groupService.getGroupMembers(groupId);
      membersCache.value.set(groupId, { data, fetchedAt: Date.now() });
      activeGroupDetails.value = data;
    } catch (err: unknown) {
      activeGroupDetails.value = null;
      if (err instanceof ApiError && err.isNotFound) {
        membersError.value = "Группа не найдена или недоступна";
      } else if (err instanceof ApiError && err.isForbidden) {
        membersError.value = "Недостаточно прав для просмотра участников";
      } else {
        membersError.value = extractApiErrorMessage(
          err,
          "Не удалось загрузить участников группы",
        );
      }
      throw err;
    } finally {
      isMembersLoading.value = false;
    }
  }

  function persistActiveGroup(groupId: string | null): void {
    activeGroupId.value = groupId;
    if (groupId) {
      localStorage.setItem(ACTIVE_GROUP_STORAGE_KEY, groupId);
    } else {
      localStorage.removeItem(ACTIVE_GROUP_STORAGE_KEY);
    }
  }

  async function selectGroup(groupId: string): Promise<void> {
    persistActiveGroup(groupId);

    if (activeGroupId.value === groupId && activeGroupDetails.value) {
      const cached = membersCache.value.get(groupId);
      if (cached && Date.now() - cached.fetchedAt < CACHE_TTL_MS) {
        return;
      }
    }
    await fetchGroupMembers(groupId);
  }

  async function createGroup(payload: CreateGroupRequest): Promise<GroupDto> {
    isMutating.value = true;
    try {
      const created = await groupService.createGroup(payload);
      invalidateCache();
      await fetchGroups({ force: true });
      activeGroupId.value = created.group_id;
      await fetchGroupMembers(created.group_id, { force: true });
      return created;
    } finally {
      isMutating.value = false;
    }
  }

  async function updateGroup(groupId: string, payload: EditGroupRequest): Promise<GroupDto> {
    isMutating.value = true;
    try {
      const updated = await groupService.updateGroup(groupId, payload);
      invalidateCache();
      await fetchGroups({ force: true });
      await fetchGroupMembers(groupId, { force: true });
      return updated;
    } finally {
      isMutating.value = false;
    }
  }

  async function deleteGroup(groupId: string): Promise<void> {
    isMutating.value = true;
    try {
      await groupService.deleteGroup(groupId);
      invalidateCache();
      if (activeGroupId.value === groupId) {
        persistActiveGroup(null);
        activeGroupDetails.value = null;
      }
      await fetchGroups({ force: true });
      if (activeGroupId.value) {
        await fetchGroupMembers(activeGroupId.value, { force: true });
      }
    } finally {
      isMutating.value = false;
    }
  }

  async function addMember(groupId: string, userId: string): Promise<void> {
    isMutating.value = true;
    try {
      await groupService.addMember(groupId, userId);
      invalidateCache();
      await fetchGroups({ force: true });
      await fetchGroupMembers(groupId, { force: true });
    } finally {
      isMutating.value = false;
    }
  }

  async function removeMember(groupId: string, userId: string): Promise<void> {
    isMutating.value = true;
    try {
      await groupService.removeMember(groupId, userId);
      invalidateCache();
      await fetchGroups({ force: true });
      await fetchGroupMembers(groupId, { force: true });
    } finally {
      isMutating.value = false;
    }
  }

  return {
    groups,
    activeGroupId,
    activeGroup,
    activeGroupDetails,
    isLoading,
    isMembersLoading,
    isMutating,
    error,
    membersError,
    forbidden,
    isGroupsCacheValid,
    fetchGroups,
    fetchGroupMembers,
    selectGroup,
    invalidateCache,
    invalidateMembersCache,
    createGroup,
    updateGroup,
    deleteGroup,
    addMember,
    removeMember,
  };
});
