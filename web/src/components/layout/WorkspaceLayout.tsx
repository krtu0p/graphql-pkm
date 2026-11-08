import { useCallback, useRef, useState, type ReactNode } from 'react';
import type { AiStatus } from '../../hooks/useAiSettings';
import type { Backlink, Note } from '../../lib/graphql/types';
import type { NavigationSection } from '../../types/navigation';
import { AppHeader } from './AppHeader';
import { AppSidebar } from './AppSidebar';
import { ContextPanel } from './ContextPanel';

export interface WorkspaceLayoutProps {
  activeSection: NavigationSection;
  activeNote?: Note | null;
  allNotes?: Note[];
  backlinks?: Backlink[];
  selectedTag?: string | null;
  isEnriching?: boolean;
  aiStatus?: AiStatus;
  onSelectSection: (section: NavigationSection) => void;
  onSelectTagFromSidebar?: (tag: string) => void;
  onNewNote: () => void;
  onSelectNote?: (noteId: string) => void;
  onRemoveLink?: (linkId: string) => Promise<void>;
  onRefreshMetadata?: () => void;
  children: ReactNode;
}

export function WorkspaceLayout({
  activeSection,
  activeNote = null,
  allNotes = [],
  backlinks = [],
  selectedTag = null,
  isEnriching = false,
  aiStatus = 'unavailable',
  onSelectSection,
  onSelectTagFromSidebar,
  onNewNote,
  onSelectNote,
  onRemoveLink,
  onRefreshMetadata,
  children,
}: WorkspaceLayoutProps) {
  const [isMobileSidebarOpen, setIsMobileSidebarOpen] = useState(false);
  const [isTabletSidebarCollapsed, setIsTabletSidebarCollapsed] = useState(false);
  const [isContextPanelOpen, setIsContextPanelOpen] = useState(() =>
    typeof window !== 'undefined' ? window.innerWidth > 1024 : true,
  );

  const mobileMenuBtnRef = useRef<HTMLButtonElement | null>(null);

  const handleCloseMobileSidebar = useCallback(() => {
    setIsMobileSidebarOpen((wasOpen) => {
      if (wasOpen) {
        window.requestAnimationFrame(() => {
          mobileMenuBtnRef.current?.focus();
        });
      }
      return false;
    });
  }, []);

  return (
    <div className="workspace-shell">
      <AppSidebar
        activeSection={activeSection}
        isOpenMobile={isMobileSidebarOpen}
        isCollapsedTablet={isTabletSidebarCollapsed}
        notes={allNotes}
        selectedTag={selectedTag}
        aiStatus={aiStatus}
        onSelectSection={onSelectSection}
        onSelectTagFromSidebar={onSelectTagFromSidebar}
        onNewNote={onNewNote}
        onCloseMobile={handleCloseMobileSidebar}
        onToggleTabletSidebar={() => setIsTabletSidebarCollapsed((prev) => !prev)}
      />

      <div className="workspace-center">
        <AppHeader
          activeSection={activeSection}
          activeNoteTitle={activeNote?.title ?? null}
          selectedTag={selectedTag}
          isMobileSidebarOpen={isMobileSidebarOpen}
          isTabletSidebarCollapsed={isTabletSidebarCollapsed}
          isContextPanelOpen={isContextPanelOpen}
          menuButtonRef={mobileMenuBtnRef}
          onSelectSection={onSelectSection}
          onToggleMobileSidebar={() => setIsMobileSidebarOpen((prev) => !prev)}
          onToggleTabletSidebar={() => setIsTabletSidebarCollapsed((prev) => !prev)}
          onToggleContextPanel={() => setIsContextPanelOpen((prev) => !prev)}
        />
        <main className="workspace-main">{children}</main>
      </div>

      <ContextPanel
        isOpen={isContextPanelOpen}
        activeNote={activeNote}
        allNotes={allNotes}
        backlinks={backlinks}
        isEnriching={isEnriching}
        onClose={() => setIsContextPanelOpen(false)}
        onToggle={() => setIsContextPanelOpen((prev) => !prev)}
        onSelectNote={onSelectNote}
        onRemoveLink={onRemoveLink}
        onRefreshMetadata={onRefreshMetadata}
      />
    </div>
  );
}
