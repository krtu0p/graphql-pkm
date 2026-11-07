import { useEffect, type ReactNode } from 'react';
import { Button } from './Button';

export interface DialogProps {
  isOpen: boolean;
  title: string;
  description?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  isBusy?: boolean;
  onConfirm: () => void;
  onClose: () => void;
  children?: ReactNode;
}

export function Dialog({
  isOpen,
  title,
  description,
  confirmLabel = 'Confirm',
  cancelLabel = 'Cancel',
  isBusy = false,
  onConfirm,
  onClose,
  children,
}: DialogProps) {
  useEffect(() => {
    if (!isOpen) return;
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && !isBusy) {
        onClose();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, isBusy, onClose]);

  if (!isOpen) return null;

  return (
    <div className="dialog-overlay" role="presentation" onClick={!isBusy ? onClose : undefined}>
      <div
        className="dialog-surface"
        role="dialog"
        aria-modal="true"
        aria-labelledby="dialog-title"
        aria-describedby={description ? 'dialog-description' : undefined}
        onClick={(event) => event.stopPropagation()}
      >
        <h2 id="dialog-title" className="dialog-title">
          {title}
        </h2>
        {description && (
          <p id="dialog-description" className="dialog-description">
            {description}
          </p>
        )}
        {children && <div className="dialog-body">{children}</div>}
        <div className="dialog-actions">
          <Button variant="secondary" onClick={onClose} disabled={isBusy}>
            {cancelLabel}
          </Button>
          <Button variant="primary" onClick={onConfirm} disabled={isBusy}>
            {isBusy ? 'Working…' : confirmLabel}
          </Button>
        </div>
      </div>
    </div>
  );
}
