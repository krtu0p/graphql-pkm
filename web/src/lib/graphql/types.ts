export interface Link {
  id: string;
  sourceNoteId: string;
  targetNoteId: string;
  description: string | null;
  createdAt: string;
}

export interface Note {
  id: string;
  title: string;
  content: string;
  tags: string[];
  createdAt: string;
  updatedAt: string;
  summary: string | null;
  keyConcepts: string[];
  links: Link[];
  backlinks: Link[];
}

export interface Backlink {
  id: string;
  sourceNote: {
    id: string;
    title: string;
    updatedAt: string;
  };
  description: string | null;
  createdAt: string;
}

export interface SearchResult {
  note: Note;
  score: number;
  reason: string;
  matchType: string;
}

export interface SmartSearchResponse {
  results: SearchResult[];
  explanation: string;
  gaps: string[];
  connections: string[];
}

export interface AiSettings {
  enabled: boolean;
  provider: string;
  baseUrl: string;
  model: string;
  hasApiKey: boolean;
  embedBaseUrl: string;
  embedModel: string;
}

export interface SaveAiSettingsInput {
  enabled: boolean;
  provider: string;
  baseUrl: string;
  model: string;
  apiKey?: string;
  embedBaseUrl: string;
  embedModel: string;
}

export interface ListAiModelsInput {
  provider?: string;
  baseUrl: string;
  apiKey?: string;
}

export interface AiModelList {
  models: string[];
  error: string | null;
}

export interface CreateNoteInput {
  title: string;
  content: string;
  tags?: string[];
}

export interface UpdateNoteInput {
  title?: string;
  content?: string;
  tags?: string[];
}

export interface GraphQLErrorItem {
  message: string;
  path?: Array<string | number>;
}

export interface GraphQLResponse<T> {
  data?: T | null;
  errors?: GraphQLErrorItem[];
}
