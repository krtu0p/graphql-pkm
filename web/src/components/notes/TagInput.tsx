import { useState, type KeyboardEvent } from 'react';
import { normalizeTag } from '../../lib/utils/format';
import { TagList } from './TagList';

export interface TagInputProps {
  tags: string[];
  onChange: (nextTags: string[]) => void;
  disabled?: boolean;
}

export function TagInput({ tags, onChange, disabled = false }: TagInputProps) {
  const [inputValue, setInputValue] = useState('');

  const addTag = (raw: string) => {
    const clean = normalizeTag(raw);
    if (!clean) return;
    if (!tags.includes(clean)) {
      onChange([...tags, clean]);
    }
    setInputValue('');
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter' || event.key === ',') {
      event.preventDefault();
      addTag(inputValue);
    } else if (event.key === 'Backspace' && !inputValue && tags.length > 0) {
      onChange(tags.slice(0, -1));
    }
  };

  const handleRemoveTag = (tagToRemove: string) => {
    if (disabled) return;
    onChange(tags.filter((tag) => tag !== tagToRemove));
  };

  return (
    <div className="tag-input-container">
      <TagList tags={tags} onRemoveTag={!disabled ? handleRemoveTag : undefined} />
      <div className="tag-input-add">
        <label htmlFor="note-tag-input" className="sr-only">
          Add tag
        </label>
        <input
          id="note-tag-input"
          type="text"
          className="tag-inline-input"
          value={inputValue}
          disabled={disabled}
          placeholder={tags.length === 0 ? 'Add tags (press Enter)…' : '+ Add tag…'}
          onChange={(event) => setInputValue(event.target.value)}
          onKeyDown={handleKeyDown}
          onBlur={() => {
            if (inputValue.trim()) {
              addTag(inputValue);
            }
          }}
        />
      </div>
    </div>
  );
}
