export interface LoadingStateProps {
  label?: string;
  count?: number;
}

export function LoadingState({
  label = 'Loading notes…',
  count = 3,
}: LoadingStateProps) {
  const placeholders = Array.from({ length: count }, (_, idx) => idx);

  return (
    <div className="loading-state" role="status" aria-live="polite" aria-label={label}>
      <span className="loading-state-label">{label}</span>
      <div className="skeleton-list" aria-hidden="true">
        {placeholders.map((key) => (
          <div key={key} className="skeleton-card">
            <div className="skeleton-line skeleton-title" />
            <div className="skeleton-line skeleton-body" />
            <div className="skeleton-line skeleton-meta" />
          </div>
        ))}
      </div>
    </div>
  );
}
