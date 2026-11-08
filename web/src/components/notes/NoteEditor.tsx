import { useEffect, useRef, type KeyboardEvent } from 'react';

export interface NoteEditorProps {
  content: string;
  onChange: (nextContent: string) => void;
  onSaveShortcut?: () => void;
  disabled?: boolean;
  textareaRef?: React.RefObject<HTMLTextAreaElement | null>;
}

export function NoteEditor({
  content,
  onChange,
  onSaveShortcut,
  disabled = false,
  textareaRef: externalRef,
}: NoteEditorProps) {
  const internalRef = useRef<HTMLTextAreaElement | null>(null);
  const activeRef = externalRef ?? internalRef;

  useEffect(() => {
    const el = activeRef.current;
    if (!el) return;
    el.style.height = 'auto';
    el.style.height = `${Math.max(el.scrollHeight, 360)}px`;
  }, [content, activeRef]);

  const handleKeyDown = (event: KeyboardEvent<HTMLTextAreaElement>) => {
    if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 's') {
      event.preventDefault();
      onSaveShortcut?.();
    }
  };

  return (
    <div className="note-editor-canvas">
      <label htmlFor="note-editor-body" className="sr-only">
        Note Content
      </label>
      <textarea
        ref={activeRef}
        id="note-editor-body"
        className="note-editor-textarea"
        value={content}
        disabled={disabled}
        placeholder="Begin writing your note…"
        onChange={(event) => onChange(event.target.value)}
        onKeyDown={handleKeyDown}
        rows={14}
      />
    </div>
  );
}
