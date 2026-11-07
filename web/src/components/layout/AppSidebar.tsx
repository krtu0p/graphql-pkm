import { useEffect, useMemo, useRef, type ReactNode } from 'react';
import type { Note } from '../../lib/graphql/types';
import type { NavigationSection } from '../../types/navigation';

export interface AppSidebarProps {
  activeSection: NavigationSection;
  isOpenMobile: boolean;
  isCollapsedTablet?: boolean;
  notes?: Note[];
  selectedTag?: string | null;
  onSelectSection: (section: NavigationSection) => void;
  onSelectTagFromSidebar?: (tag: string) => void;
  onNewNote: () => void;
  onCloseMobile: () => void;
  onToggleTabletSidebar?: () => void;
}

interface SidebarNavItem {
  id: NavigationSection;
  label: string;
  icon: ReactNode;
}

const PRIMARY_NAV_ITEMS: SidebarNavItem[] = [
  {
    id: 'notes',
    label: 'All Notes',
    icon: (
      <svg
        className="nav-item-icon"
        viewBox="0 0 20 20"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.6"
        aria-hidden="true"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M4.5 3.5h8l3 3v10a1 1 0 0 1-1 1h-10a1 1 0 0 1-1-1v-12a1 1 0 0 1 1-1Z"
        />
        <path strokeLinecap="round" d="M7 9h6M7 12.5h4" />
      </svg>
    ),
  },
  {
    id: 'tags',
    label: 'Tags',
    icon: (
      <svg
        className="nav-item-icon"
        viewBox="0 0 20 20"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.6"
        aria-hidden="true"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M3.5 10.25V4.5a1 1 0 0 1 1-1h5.75l6.25 6.25-6.75 6.75-6.25-6.25Z"
        />
        <circle cx="7" cy="7" r="1" fill="currentColor" />
      </svg>
    ),
  },
  {
    id: 'graph',
    label: 'Knowledge Graph',
    icon: (
      <svg
        className="nav-item-icon"
        viewBox="0 0 20 20"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.6"
        aria-hidden="true"
      >
        <circle cx="5" cy="6" r="2" />
        <circle cx="15" cy="5" r="2" />
        <circle cx="11" cy="14.5" r="2" />
        <path strokeLinecap="round" d="M6.8 6.8 9.6 13M13.2 5.6l-6.2.2M14 6.8l-2 6" />
      </svg>
    ),
  },
  {
    id: 'search',
    label: 'AI Search',
    icon: (
      <svg
        className="nav-item-icon"
        viewBox="0 0 20 20"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.6"
        aria-hidden="true"
      >
        <circle cx="9" cy="9" r="5" />
        <path strokeLinecap="round" d="m13 13 3.5 3.5" />
      </svg>
    ),
  },
];

export function AppSidebar({
  activeSection,
  isOpenMobile,
  isCollapsedTablet = false,
  notes = [],
  selectedTag = null,
  onSelectSection,
  onSelectTagFromSidebar,
  onNewNote,
  onCloseMobile,
  onToggleTabletSidebar,
}: AppSidebarProps) {
  const sidebarRef = useRef<HTMLElement | null>(null);
  const closeBtnRef = useRef<HTMLButtonElement | null>(null);

  useEffect(() => {
    if (!isOpenMobile) return;

    closeBtnRef.current?.focus();

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.preventDefault();
        onCloseMobile();
        return;
      }

      if (event.key === 'Tab' && sidebarRef.current) {
        const focusable = sidebarRef.current.querySelectorAll<HTMLElement>(
          'button:not([disabled]), [href], input:not([disabled]), [tabindex]:not([tabindex="-1"])',
        );
        if (focusable.length === 0) return;
        const first = focusable[0];
        const last = focusable[focusable.length - 1];

        if (event.shiftKey && document.activeElement === first) {
          event.preventDefault();
          last.focus();
        } else if (!event.shiftKey && document.activeElement === last) {
          event.preventDefault();
          first.focus();
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpenMobile, onCloseMobile]);

  const topTags = useMemo(() => {
    const counts = new Map<string, number>();
    for (const note of notes) {
      for (const tag of note.tags) {
        counts.set(tag, (counts.get(tag) ?? 0) + 1);
      }
    }
    return Array.from(counts.entries())
      .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
      .slice(0, 6)
      .map(([tag, count]) => ({ tag, count }));
  }, [notes]);

  const totalTagCount = useMemo(() => {
    const unique = new Set<string>();
    for (const note of notes) {
      for (const tag of note.tags) {
        unique.add(tag);
      }
    }
    return unique.size;
  }, [notes]);

  const handleSelect = (section: NavigationSection) => {
    onSelectSection(section);
    onCloseMobile();
  };

  const handleCreateNote = () => {
    onNewNote();
    onCloseMobile();
  };

  const handleTagShortcut = (tag: string) => {
    onSelectTagFromSidebar?.(tag);
    onCloseMobile();
  };

  const sidebarClasses = [
    'app-sidebar',
    isOpenMobile ? 'is-open' : '',
    isCollapsedTablet ? 'is-collapsed-tablet' : '',
  ]
    .filter(Boolean)
    .join(' ');

  return (
    <>
      <div
        className={`sidebar-backdrop ${isOpenMobile ? 'is-open' : ''}`}
        onClick={onCloseMobile}
        aria-hidden="true"
      />
      <aside
        id="app-primary-sidebar"
        ref={sidebarRef}
        className={sidebarClasses}
        aria-label="Primary Sidebar"
      >
        <div className="sidebar-header">
          <button
            type="button"
            className="brand-identity"
            onClick={() => handleSelect('notes')}
            aria-label="Go to All Notes"
          >
            <span className="brand-logo" aria-hidden="true">
              F
            </span>
            <div className="brand-text">
              <span className="brand-name">Folio</span>
              <span className="brand-subtitle">Knowledge Library</span>
            </div>
          </button>

          <button
            ref={closeBtnRef}
            type="button"
            className="icon-btn sidebar-close-btn"
            onClick={() => {
              if (isOpenMobile) {
                onCloseMobile();
              } else {
                onToggleTabletSidebar?.();
              }
            }}
            aria-label="Close navigation sidebar"
          >
            <svg
              width="16"
              height="16"
              viewBox="0 0 20 20"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.7"
              aria-hidden="true"
            >
              <path strokeLinecap="round" d="m5 5 10 10M15 5 5 15" />
            </svg>
          </button>
        </div>

        <div className="sidebar-action-area">
          <button
            type="button"
            className="btn-primary btn-full"
            onClick={handleCreateNote}
          >
            <svg
              width="15"
              height="15"
              viewBox="0 0 20 20"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.8"
              aria-hidden="true"
            >
              <path strokeLinecap="round" d="M10 4v12M4 10h12" />
            </svg>
            <span>New Note</span>
          </button>
        </div>

        <nav className="sidebar-nav" aria-label="Workspace Sections">
          <span className="sidebar-section-label">Library</span>
          {PRIMARY_NAV_ITEMS.map((item) => {
            const isActive =
              activeSection === item.id ||
              (item.id === 'notes' && activeSection === 'editor');
            const badgeCount =
              item.id === 'notes'
                ? notes.length
                : item.id === 'tags'
                  ? totalTagCount
                  : null;

            return (
              <button
                key={item.id}
                type="button"
                className={`nav-item ${isActive ? 'is-active' : ''}`}
                aria-current={isActive ? 'page' : undefined}
                onClick={() => handleSelect(item.id)}
              >
                {item.icon}
                <span className="nav-item-label">{item.label}</span>
                {badgeCount !== null && badgeCount > 0 && (
                  <span className="nav-item-count">{badgeCount}</span>
                )}
              </button>
            );
          })}

          {topTags.length > 0 && onSelectTagFromSidebar && (
            <div className="sidebar-tags-group" role="group" aria-label="Quick tag filters">
              <span className="sidebar-section-label">Topics</span>
              {topTags.map(({ tag, count }) => {
                const isSelectedTag =
                  activeSection === 'tags' && selectedTag === tag;
                return (
                  <button
                    key={tag}
                    type="button"
                    className={`nav-item nav-tag-item ${isSelectedTag ? 'is-active' : ''}`}
                    onClick={() => handleTagShortcut(tag)}
                  >
                    <span className="nav-tag-hash" aria-hidden="true">
                      #
                    </span>
                    <span className="nav-item-label">{tag}</span>
                    <span className="nav-item-count">{count}</span>
                  </button>
                );
              })}
            </div>
          )}
        </nav>

        <div className="sidebar-footer">
          <button
            type="button"
            className={`nav-item ${activeSection === 'settings' ? 'is-active' : ''}`}
            aria-current={activeSection === 'settings' ? 'page' : undefined}
            onClick={() => handleSelect('settings')}
          >
            <svg
              className="nav-item-icon"
              viewBox="0 0 20 20"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.6"
              aria-hidden="true"
            >
              <circle cx="10" cy="10" r="2.5" />
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M10 3.5v1.75M10 14.75v1.75M3.5 10h1.75M14.75 10h1.75M5.4 5.4l1.25 1.25M13.35 13.35l1.25 1.25M14.6 5.4l-1.25 1.25M6.65 13.35 5.4 14.6"
              />
            </svg>
            <span className="nav-item-label">Settings</span>
          </button>
        </div>
      </aside>
    </>
  );
}
