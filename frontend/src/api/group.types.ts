import type { UserRole } from "./api.types";

export interface GroupDto {
  group_id: string;
  name: string;
  description: string;
  member_count?: number;
  owner_id: string;
  created_at: string;
  can_edit: boolean;
  can_delete: boolean;
}

export interface GetGroupsResponse {
  groups: GroupDto[];
}

export interface GroupMemberDto {
  user_id: string;
  login: string;
  full_name: string;
  position: string;
  role: UserRole;
  added_at: string;
}

export interface GetGroupMembersResponse {
  group_id: string;
  name: string;
  description: string;
  members: GroupMemberDto[];
  owner_id: string;
  created_at: string;
  can_edit: boolean;
  can_delete: boolean;
}

export interface CreateGroupRequest {
  name: string;
  description: string;
  owner_id?: string;
}

export interface EditGroupRequest {
  name?: string;
  description?: string;
  owner_id?: string;
}
