import { Button } from './Button';

export interface ErrorStateProps {
  title?: string;
  message: string;
  onRetry?: () => void;
}

export function ErrorState({
  title = 'Unable to complete request',
  message,
  onRetry,
}: ErrorStateProps) {
  return (
    <div className="error-banner" role="alert">
      <div className="error-banner-icon" aria-hidden="true">
        <svg
          width="18"
          height="18"
          viewBox="0 0 20 20"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.7"
        >
          <circle cx="10" cy="10" r="7.5" />
          <path strokeLinecap="round" d="M10 6.5v4.25" />
          <circle cx="10" cy="13.5" r="0.85" fill="currentColor" stroke="none" />
        </svg>
      </div>
      <div className="error-banner-content">
        <strong className="error-banner-title">{title}</strong>
        <p className="error-banner-message">{message}</p>
      </div>
      {onRetry && (
        <div className="error-banner-action">
          <Button variant="secondary" onClick={onRetry}>
            Retry
          </Button>
        </div>
      )}
    </div>
  );
}
