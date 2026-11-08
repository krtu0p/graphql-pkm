import type { InputHTMLAttributes, ReactNode } from 'react';

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  hideLabel?: boolean;
  startIcon?: ReactNode;
  endAction?: ReactNode;
}

export function Input({
  id,
  label,
  hideLabel = false,
  startIcon,
  endAction,
  className = '',
  ...rest
}: InputProps) {
  return (
    <div className={`input-field-group ${className}`.trim()}>
      {label && (
        <label htmlFor={id} className={hideLabel ? 'sr-only' : 'input-label'}>
          {label}
        </label>
      )}
      <div className="input-control-wrapper">
        {startIcon && (
          <span className="input-icon-prefix" aria-hidden="true">
            {startIcon}
          </span>
        )}
        <input
          id={id}
          className={`text-input ${startIcon ? 'has-prefix' : ''} ${endAction ? 'has-suffix' : ''}`.trim()}
          {...rest}
        />
        {endAction && <div className="input-action-suffix">{endAction}</div>}
      </div>
    </div>
  );
}
