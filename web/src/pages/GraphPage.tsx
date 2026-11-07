import { useEffect, useMemo, useState } from 'react';
import { Button } from '../components/common/Button';
import { EmptyState } from '../components/common/EmptyState';
import { ErrorState } from '../components/common/ErrorState';
import { LoadingState } from '../components/common/LoadingState';
import { GraphControls } from '../components/graph/GraphControls';
import { GraphInspector } from '../components/graph/GraphInspector';
import { KnowledgeGraph } from '../components/graph/KnowledgeGraph';
import {
  fetchNoteBacklinks,
  fetchNoteLinks,
  type Backlink,
  type Link,
  type Note,
} from '../lib/graphql';
import {
  buildKnowledgeGraphLayout,
  buildRealGraphEdges,
} from '../lib/utils/graphLayout';

export interface GraphPageProps {
  notes: Note[];
  activeNoteId: string | null;
  isLoading: boolean;
  errorMessage: string | null;
  onFocusNote: (noteId: string | null) => void;
  onOpenNote: (noteId: string) => void;
  onSelectTag?: (tag: string) => void;
  onCreateNote: () => void;
  onRetry: () => void;
}

export function GraphPage({
  notes,
  activeNoteId,
  isLoading,
  errorMessage,
  onFocusNote,
  onOpenNote,
  onSelectTag,
  onCreateNote,
  onRetry,
}: GraphPageProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedTag, setSelectedTag] = useState<string | null>(null);
  const [showConnectedOnly, setShowConnectedOnly] = useState(false);
  const [zoom, setZoom] = useState(1);
  const [pan, setPan] = useState<{ x: number; y: number }>({ x: 0, y: 0 });

  const [linksByNoteId, setLinksByNoteId] = useState<Record<string, Link[]>>({});
  const [backlinksByNoteId, setBacklinksByNoteId] = useState<Record<string, Backlink[]>>({});
  const [isLoadingConnections, setIsLoadingConnections] = useState(false);

  const selectedNodeId = useMemo(() => {
    if (activeNoteId && notes.some((n) => n.id === activeNoteId)) {
      return activeNoteId;
    }
    return notes[0]?.id ?? null;
  }, [activeNoteId, notes]);

  useEffect(() => {
    if (!selectedNodeId) return;
    const controller = new AbortController();
    setIsLoadingConnections(true);

    Promise.all([
      fetchNoteLinks(selectedNodeId, controller.signal),
      fetchNoteBacklinks(selectedNodeId, controller.signal),
    ])
      .then(([freshLinks, freshBacklinks]) => {
        setLinksByNoteId((prev) => ({ ...prev, [selectedNodeId]: freshLinks }));
        setBacklinksByNoteId((prev) => ({
          ...prev,
          [selectedNodeId]: freshBacklinks,
        }));
      })
      .catch(() => {
        // Fall back to note's embedded links/backlinks if query fails
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          setIsLoadingConnections(false);
        }
      });

    return () => controller.abort();
  }, [selectedNodeId]);

  const availableTags = useMemo(() => {
    const tagSet = new Set<string>();
    for (const note of notes) {
      for (const tag of note.tags) {
        tagSet.add(tag);
      }
    }
    return Array.from(tagSet).sort((a, b) => a.localeCompare(b));
  }, [notes]);

  const allRealEdges = useMemo(
    () => buildRealGraphEdges(notes, linksByNoteId, backlinksByNoteId),
    [notes, linksByNoteId, backlinksByNoteId],
  );

  const connectedNoteIdSet = useMemo(() => {
    const set = new Set<string>();
    for (const edge of allRealEdges) {
      set.add(edge.sourceId);
      set.add(edge.targetId);
    }
    return set;
  }, [allRealEdges]);

  const filteredNotes = useMemo(() => {
    const q = searchQuery.trim().toLowerCase();
    return notes.filter((note) => {
      if (showConnectedOnly && !connectedNoteIdSet.has(note.id)) {
        return false;
      }
      if (selectedTag && !note.tags.includes(selectedTag)) {
        return false;
      }
      if (!q) return true;
      const inTitle = note.title.toLowerCase().includes(q);
      const inContent = note.content.toLowerCase().includes(q);
      const inTags = note.tags.some((t) => t.toLowerCase().includes(q));
      return inTitle || inContent || inTags;
    });
  }, [notes, searchQuery, selectedTag, showConnectedOnly, connectedNoteIdSet]);

  const visibleEdges = useMemo(() => {
    const visibleIds = new Set(filteredNotes.map((n) => n.id));
    return allRealEdges.filter(
      (edge) => visibleIds.has(edge.sourceId) && visibleIds.has(edge.targetId),
    );
  }, [allRealEdges, filteredNotes]);

  const graphData = useMemo(
    () => buildKnowledgeGraphLayout(filteredNotes, visibleEdges),
    [filteredNotes, visibleEdges],
  );

  const selectedNote = useMemo(
    () => notes.find((n) => n.id === selectedNodeId) ?? null,
    [notes, selectedNodeId],
  );

  const outgoingForSelected = useMemo(() => {
    if (!selectedNote) return [];
    return linksByNoteId[selectedNote.id] ?? selectedNote.links;
  }, [selectedNote, linksByNoteId]);

  const backlinksForSelected = useMemo<Backlink[]>(() => {
    if (!selectedNote) return [];
    const fetched = backlinksByNoteId[selectedNote.id];
    if (fetched) return fetched;
    const noteMap = new Map(notes.map((n) => [n.id, n]));
    return selectedNote.backlinks.map((bl) => {
      const source = noteMap.get(bl.sourceNoteId);
      return {
        id: bl.id,
        description: bl.description,
        createdAt: bl.createdAt,
        sourceNote: {
          id: bl.sourceNoteId,
          title: source?.title ?? 'Linked Note',
          updatedAt: source?.updatedAt ?? bl.createdAt,
        },
      };
    });
  }, [selectedNote, backlinksByNoteId, notes]);

  const adjustZoom = (delta: number) => {
    setZoom((prev) => Math.max(0.5, Math.min(2.25, Number((prev + delta).toFixed(2)))));
  };

  return (
    <section className="graph-page" aria-labelledby="graph-page-heading">
      <div className="page-header">
        <div>
          <h1 id="graph-page-heading" className="page-title">
            Knowledge Graph
          </h1>
          <p className="page-description">
            Explore real note-linking relationships across your personal library.
          </p>
        </div>

        <Button variant="primary" onClick={onCreateNote}>
          New Note
        </Button>
      </div>

      {errorMessage && (
        <ErrorState
          title="Could not load knowledge graph"
          message={errorMessage}
          onRetry={onRetry}
        />
      )}

      {isLoading ? (
        <LoadingState label="Building knowledge graph from your notes and links…" count={2} />
      ) : notes.length === 0 ? (
        <EmptyState
          title="No notes to visualize yet"
          description="Create notes and link them together to see your knowledge graph take shape."
          action={
            <Button variant="primary" onClick={onCreateNote}>
              Create First Note
            </Button>
          }
        />
      ) : (
        <div className="graph-workspace">
          <GraphControls
            searchQuery={searchQuery}
            selectedTag={selectedTag}
            availableTags={availableTags}
            showConnectedOnly={showConnectedOnly}
            zoom={zoom}
            totalNodes={notes.length}
            visibleNodes={graphData.nodes.length}
            visibleEdges={graphData.edges.length}
            onSearchQueryChange={setSearchQuery}
            onSelectTag={setSelectedTag}
            onToggleConnectedOnly={() => setShowConnectedOnly((prev) => !prev)}
            onZoomIn={() => adjustZoom(0.15)}
            onZoomOut={() => adjustZoom(-0.15)}
            onResetView={() => {
              setZoom(1);
              setPan({ x: 0, y: 0 });
            }}
          />

          {graphData.nodes.length === 0 ? (
            <EmptyState
              title="No nodes match your graph filters"
              description="Try clearing your search query, tag filter, or connected-only toggle."
              action={
                <Button
                  variant="secondary"
                  onClick={() => {
                    setSearchQuery('');
                    setSelectedTag(null);
                    setShowConnectedOnly(false);
                  }}
                >
                  Reset Graph Filters
                </Button>
              }
            />
          ) : (
            <KnowledgeGraph
              nodes={graphData.nodes}
              edges={graphData.edges}
              selectedNodeId={selectedNodeId}
              zoom={zoom}
              pan={pan}
              onPanChange={setPan}
              onZoomDelta={adjustZoom}
              onSelectNode={(noteId) => onFocusNote(noteId)}
              onOpenNote={onOpenNote}
            />
          )}

          <GraphInspector
            selectedNote={selectedNote}
            allNotes={notes}
            outgoingLinks={outgoingForSelected}
            incomingBacklinks={backlinksForSelected}
            isLoadingConnections={isLoadingConnections}
            onOpenNote={onOpenNote}
            onFocusNode={(noteId) => onFocusNote(noteId)}
            onSelectTag={onSelectTag}
          />
        </div>
      )}
    </section>
  );
}
