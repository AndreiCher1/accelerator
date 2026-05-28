import { api } from "./api";
import { cleanPayload } from "./utils";
import type {
  AllGroupPatternsResponse,
  CreatePatternRequest,
  EditPatternRequest,
  GlobalPatternsResponse,
  GroupPatternsResponse,
  PatternDto,
} from "./patterns.types";

const PATTERNS = "/api/v1/patterns";

export const patternsService = {
  createPattern(payload: CreatePatternRequest): Promise<PatternDto> {
    const body = cleanPayload(payload);
    return api.post<PatternDto>(`${PATTERNS}/`, body).then((r) => r.data);
  },

  getPatternById(patternId: string): Promise<PatternDto> {
    return api.get<PatternDto>(`${PATTERNS}/${patternId}`).then((r) => r.data);
  },

  fetchGlobalPatterns(): Promise<GlobalPatternsResponse> {
    return api.get<GlobalPatternsResponse>(`${PATTERNS}/global`).then((r) => r.data);
  },

  fetchGroupPatterns(groupId: string): Promise<GroupPatternsResponse> {
    return api
      .get<GroupPatternsResponse>(`${PATTERNS}/all/${groupId}`)
      .then((r) => r.data);
  },

  fetchAllPatternsByGroups(): Promise<AllGroupPatternsResponse> {
    return api.get<AllGroupPatternsResponse>(`${PATTERNS}/all`).then((r) => r.data);
  },

  updatePattern(patternId: string, payload: EditPatternRequest): Promise<PatternDto> {
    const body = cleanPayload(payload, { stripEmptyStrings: true });
    return api.put<PatternDto>(`${PATTERNS}/${patternId}`, body).then((r) => r.data);
  },

  deletePattern(patternId: string): Promise<void> {
    return api.delete<void>(`${PATTERNS}/${patternId}`).then(() => undefined);
  },

  /** @deprecated Используйте createPattern */
  create: (payload: CreatePatternRequest) => patternsService.createPattern(payload),
  /** @deprecated Используйте getPatternById */
  getById: (patternId: string) => patternsService.getPatternById(patternId),
  /** @deprecated Используйте fetchGroupPatterns */
  getByGroup: (groupId: string) => patternsService.fetchGroupPatterns(groupId),
  /** @deprecated Используйте fetchGlobalPatterns */
  getGlobal: () => patternsService.fetchGlobalPatterns(),
  /** @deprecated Используйте fetchAllPatternsByGroups */
  getAllByGroups: () => patternsService.fetchAllPatternsByGroups(),
  /** @deprecated Используйте updatePattern */
  update: (patternId: string, payload: EditPatternRequest) =>
    patternsService.updatePattern(patternId, payload),
  /** @deprecated Используйте deletePattern */
  delete: (patternId: string) => patternsService.deletePattern(patternId),
};
