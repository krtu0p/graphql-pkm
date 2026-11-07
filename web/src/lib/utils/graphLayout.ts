import type { Backlink, Link, Note } from '../graphql/types';

export interface GraphNodePosition {
  id: string;
  title: string;
  tags: string[];
  degree: number;
  outgoingCount: number;
  incomingCount: number;
  x: number;
  y: number;
}

export interface GraphEdge {
  id: string;
  sourceId: string;
  targetId: string;
  description: string | null;
}

export interface GraphData {
  nodes: GraphNodePosition[];
  edges: GraphEdge[];
}

export const GRAPH_CANVAS_WIDTH = 880;
export const GRAPH_CANVAS_HEIGHT = 540;

/**
 * Extracts real, deduplicated directed edges strictly from existing note links
 * and any explicitly fetched links/backlinks queries. Never invents relationships.
 */
export function buildRealGraphEdges(
  notes: Note[],
  extraLinksByNoteId?: Record<string, Link[]>,
  extraBacklinksByNoteId?: Record<string, Backlink[]>,
): GraphEdge[] {
  const validNoteIds = new Set(notes.map((n) => n.id));
  const edgeMap = new Map<string, GraphEdge>();

  const registerEdge = (
    id: string,
    sourceId: string,
    targetId: string,
    description: string | null,
  ) => {
    if (!validNoteIds.has(sourceId) || !validNoteIds.has(targetId)) {
      return;
    }
    if (sourceId === targetId) {
      return;
    }
    const pairKey = `${sourceId}->${targetId}`;
    const existing = edgeMap.get(pairKey);
    if (!existing) {
      edgeMap.set(pairKey, {
        id: id || pairKey,
        sourceId,
        targetId,
        description: description ?? null,
      });
    } else if (!existing.description && description) {
      edgeMap.set(pairKey, {
        ...existing,
        description,
      });
    }
  };

  for (const note of notes) {
    for (const link of note.links) {
      registerEdge(link.id, link.sourceNoteId, link.targetNoteId, link.description);
    }
    for (const backlink of note.backlinks) {
      registerEdge(
        backlink.id,
        backlink.sourceNoteId,
        backlink.targetNoteId,
        backlink.description,
      );
    }
  }

  if (extraLinksByNoteId) {
    for (const links of Object.values(extraLinksByNoteId)) {
      for (const link of links) {
        registerEdge(link.id, link.sourceNoteId, link.targetNoteId, link.description);
      }
    }
  }

  if (extraBacklinksByNoteId) {
    for (const [targetNoteId, backlinks] of Object.entries(extraBacklinksByNoteId)) {
      for (const bl of backlinks) {
        registerEdge(bl.id, bl.sourceNote.id, targetNoteId, bl.description);
      }
    }
  }

  return Array.from(edgeMap.values());
}

/**
 * Computes a deterministic, readable 2D node layout using radial seeding
 * followed by force-directed relaxation within the canvas bounds.
 */
export function buildKnowledgeGraphLayout(
  notes: Note[],
  edges: GraphEdge[],
  width = GRAPH_CANVAS_WIDTH,
  height = GRAPH_CANVAS_HEIGHT,
): GraphData {
  if (notes.length === 0) {
    return { nodes: [], edges: [] };
  }

  const outgoingCount = new Map<string, number>();
  const incomingCount = new Map<string, number>();

  for (const note of notes) {
    outgoingCount.set(note.id, 0);
    incomingCount.set(note.id, 0);
  }

  for (const edge of edges) {
    outgoingCount.set(edge.sourceId, (outgoingCount.get(edge.sourceId) ?? 0) + 1);
    incomingCount.set(edge.targetId, (incomingCount.get(edge.targetId) ?? 0) + 1);
  }

  const centerX = width / 2;
  const centerY = height / 2;
  const radius = Math.min(width, height) * (notes.length === 1 ? 0 : 0.34);

  const sortedNotes = [...notes].sort((a, b) => {
    const degA = (outgoingCount.get(a.id) ?? 0) + (incomingCount.get(a.id) ?? 0);
    const degB = (outgoingCount.get(b.id) ?? 0) + (incomingCount.get(b.id) ?? 0);
    if (degB !== degA) return degB - degA;
    return a.id.localeCompare(b.id);
  });

  const nodes: GraphNodePosition[] = sortedNotes.map((note, idx) => {
    const out = outgoingCount.get(note.id) ?? 0;
    const inc = incomingCount.get(note.id) ?? 0;
    const deg = out + inc;

    if (sortedNotes.length === 1) {
      return {
        id: note.id,
        title: note.title || 'Untitled Note',
        tags: note.tags,
        degree: deg,
        outgoingCount: out,
        incomingCount: inc,
        x: centerX,
        y: centerY,
      };
    }

    const angle = (2 * Math.PI * idx) / sortedNotes.length - Math.PI / 2;
    const ringFactor = deg > 0 && idx === 0 && sortedNotes.length > 3 ? 0.35 : 1;
    return {
      id: note.id,
      title: note.title || 'Untitled Note',
      tags: note.tags,
      degree: deg,
      outgoingCount: out,
      incomingCount: inc,
      x: centerX + Math.cos(angle) * radius * ringFactor,
      y: centerY + Math.sin(angle) * radius * ringFactor,
    };
  });

  const indexById = new Map<string, number>();
  nodes.forEach((n, idx) => indexById.set(n.id, idx));

  const iterations = notes.length > 1 ? 45 : 0;
  const idealDistance = Math.min(width, height) * 0.28;
  const paddingX = 76;
  const paddingY = 56;

  for (let step = 0; step < iterations; step += 1) {
    const dx = new Array<number>(nodes.length).fill(0);
    const dy = new Array<number>(nodes.length).fill(0);

    for (let i = 0; i < nodes.length; i += 1) {
      for (let j = i + 1; j < nodes.length; j += 1) {
        const diffX = nodes[i].x - nodes[j].x;
        const diffY = nodes[i].y - nodes[j].y;
        const distSq = Math.max(100, diffX * diffX + diffY * diffY);
        const dist = Math.sqrt(distSq);
        const repulse = (idealDistance * idealDistance * 0.42) / distSq;
        const fx = (diffX / dist) * repulse;
        const fy = (diffY / dist) * repulse;
        dx[i] += fx;
        dy[i] += fy;
        dx[j] -= fx;
        dy[j] -= fy;
      }
    }

    for (const edge of edges) {
      const sIdx = indexById.get(edge.sourceId);
      const tIdx = indexById.get(edge.targetId);
      if (sIdx === undefined || tIdx === undefined) continue;
      const diffX = nodes[tIdx].x - nodes[sIdx].x;
      const diffY = nodes[tIdx].y - nodes[sIdx].y;
      const dist = Math.max(1, Math.sqrt(diffX * diffX + diffY * diffY));
      const attract = (dist - idealDistance * 0.72) * 0.08;
      const fx = (diffX / dist) * attract;
      const fy = (diffY / dist) * attract;
      dx[sIdx] += fx;
      dy[sIdx] += fy;
      dx[tIdx] -= fx;
      dy[tIdx] -= fy;
    }

    for (let i = 0; i < nodes.length; i += 1) {
      dx[i] += (centerX - nodes[i].x) * 0.03;
      dy[i] += (centerY - nodes[i].y) * 0.03;

      const maxStep = 14;
      const moveX = Math.max(-maxStep, Math.min(maxStep, dx[i]));
      const moveY = Math.max(-maxStep, Math.min(maxStep, dy[i]));

      nodes[i].x = Math.max(paddingX, Math.min(width - paddingX, nodes[i].x + moveX));
      nodes[i].y = Math.max(paddingY, Math.min(height - paddingY, nodes[i].y + moveY));
    }
  }

  return { nodes, edges };
}
