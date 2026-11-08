import { executeGraphQL } from './client';
import type {
  AiModelList,
  AiSettings,
  Backlink,
  CreateNoteInput,
  Link,
  ListAiModelsInput,
  Note,
  SaveAiSettingsInput,
  SearchResult,
  SmartSearchResponse,
  UpdateNoteInput,
} from './types';
import {
  buildListAiModelsInput,
  buildSaveAiSettingsInput,
  parseAiModelList,
  parseAiSettings,
  parseBacklinkList,
  parseLink,
  parseLinkList,
  parseNote,
  parseNoteList,
  parseSearchResultList,
  parseSmartSearchResponse,
} from './validators';

const NOTE_FIELDS = `
  id
  title
  content
  tags
  createdAt
  updatedAt
  summary
  keyConcepts
  links {
    id
    sourceNoteId
    targetNoteId
    description
    createdAt
  }
  backlinks {
    id
    sourceNoteId
    targetNoteId
    description
    createdAt
  }
`;

const NOTES_QUERY = `
  query GetNotes {
    notes {
      ${NOTE_FIELDS}
    }
  }
`;

const NOTE_QUERY = `
  query GetNote($id: ID!) {
    note(id: $id) {
      ${NOTE_FIELDS}
    }
  }
`;

const NOTES_BY_TAG_QUERY = `
  query GetNotesByTag($tag: String!) {
    notesByTag(tag: $tag) {
      ${NOTE_FIELDS}
    }
  }
`;

const SEARCH_NOTES_QUERY = `
  query SearchNotes($query: String!) {
    searchNotes(query: $query) {
      ${NOTE_FIELDS}
    }
  }
`;

const LINKS_QUERY = `
  query GetLinks($noteId: ID!) {
    links(noteId: $noteId) {
      id
      sourceNoteId
      targetNoteId
      description
      createdAt
    }
  }
`;

const BACKLINKS_QUERY = `
  query GetBacklinks($noteId: ID!) {
    backlinks(noteId: $noteId) {
      id
      description
      createdAt
      sourceNote {
        id
        title
        updatedAt
      }
    }
  }
`;

const SEMANTIC_SEARCH_QUERY = `
  query SemanticSearch($query: String!) {
    semanticSearch(query: $query) {
      score
      reason
      matchType
      note {
        ${NOTE_FIELDS}
      }
    }
  }
`;

const HYBRID_SEARCH_QUERY = `
  query HybridSearch($query: String!) {
    hybridSearch(query: $query) {
      score
      reason
      matchType
      note {
        ${NOTE_FIELDS}
      }
    }
  }
`;

const SMART_SEARCH_QUERY = `
  query SmartSearch($query: String!) {
    smartSearch(query: $query) {
      explanation
      gaps
      connections
      results {
        score
        reason
        matchType
        note {
          ${NOTE_FIELDS}
        }
      }
    }
  }
`;

const CHECK_RELATIONSHIP_QUERY = `
  query CheckRelationship($noteIDA: ID!, $noteIDB: ID!) {
    checkRelationship(noteIDA: $noteIDA, noteIDB: $noteIDB)
  }
`;

const CREATE_NOTE_MUTATION = `
  mutation CreateNote($input: CreateNoteInput!) {
    createNote(input: $input) {
      ${NOTE_FIELDS}
    }
  }
`;

const UPDATE_NOTE_MUTATION = `
  mutation UpdateNote($id: ID!, $input: UpdateNoteInput!) {
    updateNote(id: $id, input: $input) {
      ${NOTE_FIELDS}
    }
  }
`;

const DELETE_NOTE_MUTATION = `
  mutation DeleteNote($id: ID!) {
    deleteNote(id: $id)
  }
`;

const LINK_NOTES_MUTATION = `
  mutation LinkNotes($sourceId: ID!, $targetId: ID!, $description: String) {
    linkNotes(sourceId: $sourceId, targetId: $targetId, description: $description) {
      id
      sourceNoteId
      targetNoteId
      description
      createdAt
    }
  }
`;

const UNLINK_NOTES_MUTATION = `
  mutation UnlinkNotes($linkId: ID!) {
    unlinkNotes(linkId: $linkId)
  }
`;

const AI_SETTINGS_FIELDS = `
  enabled
  provider
  baseUrl
  model
  hasApiKey
  embedBaseUrl
  embedModel
`;

const AI_SETTINGS_QUERY = `
  query GetAiSettings {
    aiSettings {
      ${AI_SETTINGS_FIELDS}
    }
  }
`;

const SAVE_AI_SETTINGS_MUTATION = `
  mutation SaveAiSettings($input: SaveAISettingsInput!) {
    saveAISettings(input: $input) {
      ${AI_SETTINGS_FIELDS}
    }
  }
`;

const LIST_AI_MODELS_MUTATION = `
  mutation ListAiModels($input: ListAIModelsInput!) {
    listAIModels(input: $input) {
      models
      error
    }
  }
`;

export async function fetchAiSettings(signal?: AbortSignal): Promise<AiSettings> {
  const data = await executeGraphQL<{ aiSettings: unknown }>(
    AI_SETTINGS_QUERY,
    undefined,
    signal,
  );
  const parsed = parseAiSettings(data.aiSettings);
  if (!parsed) {
    throw new Error('Server returned an invalid AI settings payload.');
  }
  return parsed;
}

export async function saveAiSettings(input: SaveAiSettingsInput): Promise<AiSettings> {
  const cleanInput = buildSaveAiSettingsInput(input);
  const data = await executeGraphQL<{ saveAISettings: unknown }>(
    SAVE_AI_SETTINGS_MUTATION,
    { input: cleanInput },
  );
  const parsed = parseAiSettings(data.saveAISettings);
  if (!parsed) {
    throw new Error('Server returned an invalid AI settings payload after saving.');
  }
  return parsed;
}

export async function fetchAiModels(input: ListAiModelsInput): Promise<AiModelList> {
  try {
    const data = await executeGraphQL<{ listAIModels: unknown }>(
      LIST_AI_MODELS_MUTATION,
      { input: buildListAiModelsInput(input) },
    );
    return parseAiModelList(data.listAIModels);
  } catch (error) {
    return {
      models: [],
      error: error instanceof Error ? error.message : 'Unable to load AI models.',
    };
  }
}

export async function fetchNotes(signal?: AbortSignal): Promise<Note[]> {
  const data = await executeGraphQL<{ notes: unknown }>(NOTES_QUERY, undefined, signal);
  return parseNoteList(data.notes);
}

export async function fetchNoteById(id: string, signal?: AbortSignal): Promise<Note | null> {
  const data = await executeGraphQL<{ note: unknown }>(NOTE_QUERY, { id }, signal);
  return parseNote(data.note);
}

export async function fetchNotesByTag(tag: string, signal?: AbortSignal): Promise<Note[]> {
  const data = await executeGraphQL<{ notesByTag: unknown }>(
    NOTES_BY_TAG_QUERY,
    { tag },
    signal,
  );
  return parseNoteList(data.notesByTag);
}

export async function searchNotesKeyword(query: string, signal?: AbortSignal): Promise<Note[]> {
  const data = await executeGraphQL<{ searchNotes: unknown }>(
    SEARCH_NOTES_QUERY,
    { query },
    signal,
  );
  return parseNoteList(data.searchNotes);
}

export async function fetchNoteLinks(noteId: string, signal?: AbortSignal): Promise<Link[]> {
  const data = await executeGraphQL<{ links: unknown }>(LINKS_QUERY, { noteId }, signal);
  return parseLinkList(data.links);
}

export async function fetchNoteBacklinks(
  noteId: string,
  signal?: AbortSignal,
): Promise<Backlink[]> {
  const data = await executeGraphQL<{ backlinks: unknown }>(
    BACKLINKS_QUERY,
    { noteId },
    signal,
  );
  return parseBacklinkList(data.backlinks);
}

export async function fetchSemanticSearch(
  query: string,
  signal?: AbortSignal,
): Promise<SearchResult[]> {
  const data = await executeGraphQL<{ semanticSearch: unknown }>(
    SEMANTIC_SEARCH_QUERY,
    { query },
    signal,
  );
  return parseSearchResultList(data.semanticSearch);
}

export async function fetchHybridSearch(
  query: string,
  signal?: AbortSignal,
): Promise<SearchResult[]> {
  const data = await executeGraphQL<{ hybridSearch: unknown }>(
    HYBRID_SEARCH_QUERY,
    { query },
    signal,
  );
  return parseSearchResultList(data.hybridSearch);
}

export async function fetchSmartSearch(
  query: string,
  signal?: AbortSignal,
): Promise<SmartSearchResponse> {
  const data = await executeGraphQL<{ smartSearch: unknown }>(
    SMART_SEARCH_QUERY,
    { query },
    signal,
  );
  return parseSmartSearchResponse(data.smartSearch);
}

export async function fetchCheckRelationship(
  noteIDA: string,
  noteIDB: string,
  signal?: AbortSignal,
): Promise<string> {
  const data = await executeGraphQL<{ checkRelationship: unknown }>(
    CHECK_RELATIONSHIP_QUERY,
    { noteIDA, noteIDB },
    signal,
  );
  return typeof data.checkRelationship === 'string' ? data.checkRelationship : '';
}

export async function createNote(input: CreateNoteInput): Promise<Note> {
  const trimmedTitle = input.title.trim();
  if (!trimmedTitle) {
    throw new Error('Note title cannot be empty.');
  }
  const cleanInput: CreateNoteInput = {
    title: trimmedTitle,
    content: input.content,
    tags: input.tags ?? [],
  };
  const data = await executeGraphQL<{ createNote: unknown }>(CREATE_NOTE_MUTATION, {
    input: cleanInput,
  });
  const parsed = parseNote(data.createNote);
  if (!parsed) {
    throw new Error('Server returned an invalid note payload after creation.');
  }
  return parsed;
}

export async function updateNote(id: string, input: UpdateNoteInput): Promise<Note> {
  const cleanInput: UpdateNoteInput = {};
  if (input.title !== undefined) {
    const trimmed = input.title.trim();
    if (!trimmed) {
      throw new Error('Note title cannot be empty.');
    }
    cleanInput.title = trimmed;
  }
  if (input.content !== undefined) {
    cleanInput.content = input.content;
  }
  if (input.tags !== undefined) {
    cleanInput.tags = input.tags;
  }

  const data = await executeGraphQL<{ updateNote: unknown }>(UPDATE_NOTE_MUTATION, {
    id,
    input: cleanInput,
  });
  const parsed = parseNote(data.updateNote);
  if (!parsed) {
    throw new Error('Server returned an invalid note payload after update.');
  }
  return parsed;
}

export async function deleteNote(id: string): Promise<boolean> {
  const data = await executeGraphQL<{ deleteNote: unknown }>(DELETE_NOTE_MUTATION, { id });
  return Boolean(data.deleteNote);
}

export async function linkNotes(
  sourceId: string,
  targetId: string,
  description?: string,
): Promise<Link> {
  const data = await executeGraphQL<{ linkNotes: unknown }>(LINK_NOTES_MUTATION, {
    sourceId,
    targetId,
    description: description?.trim() ? description.trim() : null,
  });
  const parsed = parseLink(data.linkNotes);
  if (!parsed) {
    throw new Error('Server returned an invalid link payload.');
  }
  return parsed;
}

export async function unlinkNotes(linkId: string): Promise<boolean> {
  const data = await executeGraphQL<{ unlinkNotes: unknown }>(UNLINK_NOTES_MUTATION, {
    linkId,
  });
  return Boolean(data.unlinkNotes);
}
