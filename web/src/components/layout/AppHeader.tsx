import type { RefObject } from 'react';
import { NAVIGATION_METADATA, type NavigationSection } from '../../types/navigation';

export interface AppHeaderProps {
  activeSection: NavigationSection;
  activeNoteTitle?: string | null;
  selectedTag?: string | null;
  isMobileSidebarOpen?: boolean;
  isTabletSidebarCollapsed?: boolean;
  isContextPanelOpen: boolean;
  menuButtonRef?: RefObject<HTMLButtonElement | null>;
  onSelectSection?: (section: NavigationSection) => void;
  onToggleMobileSidebar: () => void;
  onToggleTabletSidebar?: () => void;
  onToggleContextPanel: () => void;
}

export function AppHeader({
  activeSection,
  activeNoteTitle = null,
  selectedTag = null,
  isMobileSidebarOpen = false,
  isTabletSidebarCollapsed = false,
  isContextPanelOpen,
  menuButtonRef,
  onSelectSection,
  onToggleMobileSidebar,
  onToggleTabletSidebar,
  onToggleContextPanel,
}: AppHeaderProps) {
  const sectionMeta = NAVIGATION_METADATA[activeSection];

  const subContextLabel =
    activeSection === 'editor'
      ? activeNoteTitle || 'New Note'
      : activeSection === 'tags' && selectedTag
        ? `#${selectedTag}`
        : null;

  return (
    <header className="app-header">
      <div className="header-left">
        <button
          ref={menuButtonRef}
          type="button"
          className="icon-btn mobile-menu-btn"
          onClick={onToggleMobileSidebar}
          aria-expanded={isMobileSidebarOpen}
          aria-controls="app-primary-sidebar"
          aria-label={
            isMobileSidebarOpen
              ? 'Close navigation drawer'
              : 'Open navigation drawer'
          }
        >
          <svg
            width="18"
            height="18"
            viewBox="0 0 20 20"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.6"
            aria-hidden="true"
          >
            <path strokeLinecap="round" d="M3.5 5.5h13M3.5 10h13M3.5 14.5h13" />
          </svg>
          <span className="mobile-menu-label">Menu</span>
        </button>

        {onToggleTabletSidebar && (
          <button
            type="button"
            className={`icon-btn tablet-sidebar-btn ${!isTabletSidebarCollapsed ? 'is-active' : ''}`}
            onClick={onToggleTabletSidebar}
            aria-expanded={!isTabletSidebarCollapsed}
            aria-controls="app-primary-sidebar"
            aria-label={
              isTabletSidebarCollapsed
                ? 'Expand navigation sidebar'
                : 'Collapse navigation sidebar'
            }
          >
            <svg
              width="16"
              height="16"
              viewBox="0 0 20 20"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
              aria-hidden="true"
            >
              <rect x="3" y="3.5" width="14" height="13" rx="1.5" />
              <path d="M7.5 3.5v13" />
            </svg>
            <span>Sidebar</span>
          </button>
        )}

        <nav className="header-breadcrumb" aria-label="Breadcrumb">
          <button
            type="button"
            className="breadcrumb-link"
            onClick={() => onSelectSection?.('notes')}
          >
            Library
          </button>
          <span aria-hidden="true">/</span>
          {subContextLabel ? (
            <>
              <button
                type="button"
                className="breadcrumb-link"
                onClick={() =>
                  onSelectSection?.(activeSection === 'editor' ? 'notes' : activeSection)
                }
              >
                {activeSection === 'editor' ? 'All Notes' : sectionMeta.label}
              </button>
              <span aria-hidden="true">/</span>
              <span className="header-breadcrumb-current">{subContextLabel}</span>
            </>
          ) : (
            <span className="header-breadcrumb-current">{sectionMeta.label}</span>
          )}
        </nav>
      </div>

      <div className="header-right">
        <button
          type="button"
          className={`icon-btn ${isContextPanelOpen ? 'is-active' : ''}`}
          onClick={onToggleContextPanel}
          aria-expanded={isContextPanelOpen}
          aria-controls="note-context-panel"
          aria-label={isContextPanelOpen ? 'Hide context panel' : 'Show context panel'}
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 20 20"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.5"
            aria-hidden="true"
          >
            <rect x="3" y="3.5" width="14" height="13" rx="1.5" />
            <path d="M12.5 3.5v13" />
          </svg>
          <span>Context</span>
        </button>
      </div>
    </header>
  );
}
