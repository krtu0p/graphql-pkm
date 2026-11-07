import type { Note } from '../../lib/graphql/types';
import { NoteCard } from './NoteCard';

export type NoteLayoutView = 'list' | 'grid';

export interface NoteListProps {
  notes: Note[];
  activeNoteId?: string | null;
  selectedTag?: string | null;
  viewMode?: NoteLayoutView;
  onSelectNote: (noteId: string) => void;
  onSelectTag?: (tag: string) => void;
  onRequestDelete?: (note: Note) => void;
}

export function NoteList({
  notes,
  activeNoteId = null,
  selectedTag = null,
  viewMode = 'list',
  onSelectNote,
  onSelectTag,
  onRequestDelete,
}: NoteListProps) {
  return (
    <div
      className={`note-list ${viewMode === 'grid' ? 'note-list-grid' : 'note-list-stack'}`}
      role="list"
    >
      {notes.map((note) => (
        <div key={note.id} role="listitem" className="note-list-item">
          <NoteCard
            note={note}
            isActive={activeNoteId === note.id}
            selectedTag={selectedTag}
            onSelectNote={onSelectNote}
            onSelectTag={onSelectTag}
            onRequestDelete={onRequestDelete}
          />
        </div>
      ))}
    </div>
  );
}
