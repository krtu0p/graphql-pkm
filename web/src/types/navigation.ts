export type NavigationSection =
  | 'notes'
  | 'editor'
  | 'tags'
  | 'graph'
  | 'search'
  | 'settings';

export interface NavigationMeta {
  id: NavigationSection;
  label: string;
  subtitle: string;
}

export const NAVIGATION_METADATA: Record<NavigationSection, NavigationMeta> = {
  notes: {
    id: 'notes',
    label: 'All Notes',
    subtitle: 'Browse, filter, and organize your personal knowledge library.',
  },
  editor: {
    id: 'editor',
    label: 'Note Editor',
    subtitle: 'Distraction-free writing workspace with linked notes and concepts.',
  },
  tags: {
    id: 'tags',
    label: 'Tags',
    subtitle: 'Explore notes categorized by topic tags.',
  },
  graph: {
    id: 'graph',
    label: 'Knowledge Graph',
    subtitle: 'Visualize bidirectional relationships between notes.',
  },
  search: {
    id: 'search',
    label: 'AI Search',
    subtitle: 'Search across your library using keyword, semantic, hybrid, or smart modes.',
  },
  settings: {
    id: 'settings',
    label: 'Settings',
    subtitle: 'Workspace preferences and local GraphQL connection configuration.',
  },
};
