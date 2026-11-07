import type { KeyboardEvent } from 'react';

export interface NoteTitleProps {
  value: string;
  onChange: (nextTitle: string) => void;
  onEnterKey?: () => void;
  disabled?: boolean;
}

export function NoteTitle({
  value,
  onChange,
  onEnterKey,
  disabled = false,
}: NoteTitleProps) {
  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter') {
      event.preventDefault();
      onEnterKey?.();
    }
  };

  return (
    <div className="note-title-wrapper">
      <label htmlFor="note-editor-title" className="sr-only">
        Note Title
      </label>
      <input
        id="note-editor-title"
        type="text"
        className="note-title-input"
        value={value}
        disabled={disabled}
        placeholder="Untitled note…"
        onChange={(event) => onChange(event.target.value)}
        onKeyDown={handleKeyDown}
        autoComplete="off"
      />
    </div>
  );
}
