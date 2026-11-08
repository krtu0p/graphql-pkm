import type { ReactNode } from 'react';

export interface TooltipProps {
  label: string;
  children: ReactNode;
}

export function Tooltip({ label, children }: TooltipProps) {
  return (
    <span className="tooltip-wrapper" title={label}>
      {children}
    </span>
  );
}
