import { useRef, useState, type KeyboardEvent, type PointerEvent, type WheelEvent } from 'react';
import {
  GRAPH_CANVAS_HEIGHT,
  GRAPH_CANVAS_WIDTH,
  type GraphEdge,
  type GraphNodePosition,
} from '../../lib/utils/graphLayout';
import { GraphNode } from './GraphNode';

export interface KnowledgeGraphProps {
  nodes: GraphNodePosition[];
  edges: GraphEdge[];
  selectedNodeId: string | null;
  zoom: number;
  pan: { x: number; y: number };
  onPanChange: (nextPan: { x: number; y: number }) => void;
  onZoomDelta: (delta: number) => void;
  onSelectNode: (noteId: string) => void;
  onOpenNote: (noteId: string) => void;
}

export function KnowledgeGraph({
  nodes,
  edges,
  selectedNodeId,
  zoom,
  pan,
  onPanChange,
  onZoomDelta,
  onSelectNode,
  onOpenNote,
}: KnowledgeGraphProps) {
  const [isDragging, setIsDragging] = useState(false);
  const dragOriginRef = useRef<{ startX: number; startY: number; panX: number; panY: number } | null>(
    null,
  );

  const nodeMap = new Map<string, GraphNodePosition>();
  for (const node of nodes) {
    nodeMap.set(node.id, node);
  }

  const connectedNodeIds = new Set<string>();
  if (selectedNodeId) {
    for (const edge of edges) {
      if (edge.sourceId === selectedNodeId) {
        connectedNodeIds.add(edge.targetId);
      } else if (edge.targetId === selectedNodeId) {
        connectedNodeIds.add(edge.sourceId);
      }
    }
  }

  const handlePointerDown = (event: PointerEvent<SVGSVGElement>) => {
    if (event.button !== 0) return;
    setIsDragging(true);
    dragOriginRef.current = {
      startX: event.clientX,
      startY: event.clientY,
      panX: pan.x,
      panY: pan.y,
    };
  };

  const handlePointerMove = (event: PointerEvent<SVGSVGElement>) => {
    if (!isDragging || !dragOriginRef.current) return;
    const dx = event.clientX - dragOriginRef.current.startX;
    const dy = event.clientY - dragOriginRef.current.startY;
    onPanChange({
      x: dragOriginRef.current.panX + dx,
      y: dragOriginRef.current.panY + dy,
    });
  };

  const handlePointerUp = () => {
    setIsDragging(false);
    dragOriginRef.current = null;
  };

  const handleWheel = (event: WheelEvent<SVGSVGElement>) => {
    if (event.ctrlKey || event.metaKey) {
      event.preventDefault();
      onZoomDelta(event.deltaY < 0 ? 0.12 : -0.12);
    }
  };

  const handleCanvasKeyDown = (event: KeyboardEvent<SVGSVGElement>) => {
    const step = 28;
    if (event.key === 'ArrowLeft') {
      event.preventDefault();
      onPanChange({ x: pan.x + step, y: pan.y });
    } else if (event.key === 'ArrowRight') {
      event.preventDefault();
      onPanChange({ x: pan.x - step, y: pan.y });
    } else if (event.key === 'ArrowUp') {
      event.preventDefault();
      onPanChange({ x: pan.x, y: pan.y + step });
    } else if (event.key === 'ArrowDown') {
      event.preventDefault();
      onPanChange({ x: pan.x, y: pan.y - step });
    } else if (event.key === '+' || event.key === '=') {
      event.preventDefault();
      onZoomDelta(0.15);
    } else if (event.key === '-') {
      event.preventDefault();
      onZoomDelta(-0.15);
    }
  };

  const centerX = GRAPH_CANVAS_WIDTH / 2;
  const centerY = GRAPH_CANVAS_HEIGHT / 2;
  const transformAttr = `translate(${centerX + pan.x}, ${centerY + pan.y}) scale(${zoom}) translate(${-centerX}, ${-centerY})`;

  return (
    <div className="knowledge-graph-canvas-wrap">
      <svg
        className={`knowledge-graph-svg ${isDragging ? 'is-dragging' : ''}`}
        viewBox={`0 0 ${GRAPH_CANVAS_WIDTH} ${GRAPH_CANVAS_HEIGHT}`}
        role="group"
        tabIndex={0}
        aria-label="Interactive note knowledge graph canvas. Use arrow keys or drag to pan, plus or minus to zoom, Tab to select nodes."
        onPointerDown={handlePointerDown}
        onPointerMove={handlePointerMove}
        onPointerUp={handlePointerUp}
        onPointerLeave={handlePointerUp}
        onWheel={handleWheel}
        onKeyDown={handleCanvasKeyDown}
      >
        <defs>
          <marker
            id="graph-edge-arrow"
            viewBox="0 0 10 10"
            refX="22"
            refY="5"
            markerWidth="6"
            markerHeight="6"
            orient="auto-start-reverse"
          >
            <path d="M 0 1.5 L 8.5 5 L 0 8.5 z" className="graph-arrow-path" />
          </marker>
          <marker
            id="graph-edge-arrow-active"
            viewBox="0 0 10 10"
            refX="22"
            refY="5"
            markerWidth="6.5"
            markerHeight="6.5"
            orient="auto-start-reverse"
          >
            <path d="M 0 1.5 L 8.5 5 L 0 8.5 z" className="graph-arrow-path-active" />
          </marker>
        </defs>

        <g transform={transformAttr}>
          {edges.map((edge) => {
            const source = nodeMap.get(edge.sourceId);
            const target = nodeMap.get(edge.targetId);
            if (!source || !target) return null;

            const isHighlighted =
              selectedNodeId === edge.sourceId || selectedNodeId === edge.targetId;
            const isDimmed = Boolean(selectedNodeId && !isHighlighted);
            const midX = (source.x + target.x) / 2;
            const midY = (source.y + target.y) / 2;

            return (
              <g key={edge.id} className="graph-edge-group">
                <line
                  x1={source.x}
                  y1={source.y}
                  x2={target.x}
                  y2={target.y}
                  className={`graph-edge-line ${isHighlighted ? 'is-highlighted' : ''} ${isDimmed ? 'is-dimmed' : ''}`.trim()}
                  markerEnd={
                    isHighlighted
                      ? 'url(#graph-edge-arrow-active)'
                      : 'url(#graph-edge-arrow)'
                  }
                />
                {isHighlighted && edge.description && (
                  <text
                    x={midX}
                    y={midY - 6}
                    textAnchor="middle"
                    className="graph-edge-caption"
                  >
                    {edge.description}
                  </text>
                )}
              </g>
            );
          })}

          {nodes.map((node) => {
            const isSelected = selectedNodeId === node.id;
            const isConnected = connectedNodeIds.has(node.id);
            const isDimmed = Boolean(
              selectedNodeId && !isSelected && !isConnected,
            );
            return (
              <GraphNode
                key={node.id}
                node={node}
                isSelected={isSelected}
                isConnectedToSelected={isConnected}
                isDimmed={isDimmed}
                onSelect={onSelectNode}
                onOpenNote={onOpenNote}
              />
            );
          })}
        </g>
      </svg>

      <div className="graph-canvas-legend" aria-hidden="true">
        <span>Drag or Arrow keys to pan</span>
        <span>·</span>
        <span>Click node to inspect</span>
        <span>·</span>
        <span>Double-click or &ldquo;Open Note&rdquo; to edit</span>
      </div>
    </div>
  );
}
