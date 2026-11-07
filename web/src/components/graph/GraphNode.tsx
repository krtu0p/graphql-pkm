import type { KeyboardEvent } from 'react';
import type { GraphNodePosition } from '../../lib/utils/graphLayout';

export interface GraphNodeProps {
  node: GraphNodePosition;
  isSelected: boolean;
  isConnectedToSelected: boolean;
  isDimmed: boolean;
  onSelect: (noteId: string) => void;
  onOpenNote: (noteId: string) => void;
}

function truncateLabel(title: string, maxChars = 22): string {
  const clean = title.trim() || 'Untitled Note';
  if (clean.length <= maxChars) return clean;
  return `${clean.slice(0, maxChars - 1).trimEnd()}…`;
}

export function GraphNode({
  node,
  isSelected,
  isConnectedToSelected,
  isDimmed,
  onSelect,
  onOpenNote,
}: GraphNodeProps) {
  const baseRadius = Math.min(24, 14 + node.degree * 2);
  const label = truncateLabel(node.title);

  const handleKeyDown = (event: KeyboardEvent<SVGGElement>) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      event.stopPropagation();
      onSelect(node.id);
    }
  };

  const stateClass = isSelected
    ? 'is-selected'
    : isConnectedToSelected
      ? 'is-neighbor'
      : isDimmed
        ? 'is-dimmed'
        : '';

  return (
    <g
      className={`graph-node-group ${stateClass}`.trim()}
      transform={`translate(${node.x}, ${node.y})`}
      role="button"
      tabIndex={0}
      aria-pressed={isSelected}
      aria-label={`${node.title}, ${node.degree} ${node.degree === 1 ? 'connection' : 'connections'}. Click to inspect relationships, double-click to open note.`}
      onClick={(event) => {
        event.stopPropagation();
        onSelect(node.id);
      }}
      onDoubleClick={(event) => {
        event.stopPropagation();
        onOpenNote(node.id);
      }}
      onKeyDown={handleKeyDown}
    >
      {isSelected && (
        <circle
          className="graph-node-halo"
          r={baseRadius + 6}
          aria-hidden="true"
        />
      )}
      <circle
        className="graph-node-circle"
        r={baseRadius}
      />
      {node.degree > 0 && (
        <text
          className="graph-node-degree"
          textAnchor="middle"
          dy="0.35em"
          aria-hidden="true"
        >
          {node.degree}
        </text>
      )}
      <g transform={`translate(0, ${baseRadius + 16})`} aria-hidden="true">
        <rect
          className="graph-node-label-bg"
          x={-Math.max(42, label.length * 3.7)}
          y={-11}
          width={Math.max(84, label.length * 7.4)}
          height={20}
          rx={4}
        />
        <text className="graph-node-label" textAnchor="middle" dy="0.25em">
          {label}
        </text>
      </g>
    </g>
  );
}
