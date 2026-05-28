import { api } from "./api";
import { cleanPayload } from "./utils";
import type {
  CreateGroupRequest,
  EditGroupRequest,
  GetGroupMembersResponse,
  GetGroupsResponse,
  GroupDto,
} from "./group.types";

const ADMIN_GROUPS = "/api/v1/admin/groups";

export const groupService = {
  getGroups(): Promise<GetGroupsResponse> {
    return api.get<GetGroupsResponse>(`${ADMIN_GROUPS}/`).then((r) => r.data);
  },

  getGroupMembers(groupId: string): Promise<GetGroupMembersResponse> {
    return api
      .get<GetGroupMembersResponse>(`${ADMIN_GROUPS}/${groupId}`)
      .then((r) => r.data);
  },

  createGroup(payload: CreateGroupRequest): Promise<GroupDto> {
    const body = cleanPayload(payload);
    return api.post<GroupDto>(`${ADMIN_GROUPS}/`, body).then((r) => r.data);
  },

  updateGroup(groupId: string, payload: EditGroupRequest): Promise<GroupDto> {
    const body = cleanPayload(payload, { stripEmptyStrings: true });
    return api.put<GroupDto>(`${ADMIN_GROUPS}/${groupId}`, body).then((r) => r.data);
  },

  deleteGroup(groupId: string): Promise<void> {
    return api.delete<void>(`${ADMIN_GROUPS}/${groupId}`).then(() => undefined);
  },

  addMember(groupId: string, userId: string): Promise<void> {
    return api
      .post<void>(`${ADMIN_GROUPS}/${groupId}/members/${userId}`)
      .then(() => undefined);
  },

  removeMember(groupId: string, userId: string): Promise<void> {
    return api
      .delete<void>(`${ADMIN_GROUPS}/${groupId}/members/${userId}`)
      .then(() => undefined);
  },
};
