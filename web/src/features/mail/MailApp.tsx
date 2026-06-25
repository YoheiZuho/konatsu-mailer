// SPDX-License-Identifier: Apache-2.0
//
// The authenticated mail workspace. Layout adapts to the selection:
//   - no email selected → 案A: full-width classic list
//   - email selected    → 案B: list column + reading pane
// Sidebar and list-column widths are mouse-resizable (design §9.1 / §9.4).

import { useCallback, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

import { useUI } from '@/stores/ui';
import { useAppearance } from '@/stores/appearance';
import { usePreferences } from '@/hooks/queries';
import { useWebSocket } from '@/hooks/useWebSocket';
import { useMediaQuery } from '@/hooks/useMediaQuery';

import { TopBar } from '@/components/layout/TopBar';
import { Sidebar } from '@/components/layout/Sidebar';
import { Resizer } from '@/components/common/Resizer';
import { MailListPane } from '@/components/mail/MailListPane';
import { ReadingPane } from '@/components/mail/ReadingPane';
import { ComposeDialog } from '@/components/compose/ComposeDialog';
import { SettingsDialog } from '@/components/settings/SettingsDialog';

export function MailApp() {
  const navigate = useNavigate();
  const { emailId } = useParams<{ emailId: string }>();
  const selectedId = emailId ?? null;

  const isWide = useMediaQuery('(min-width: 1024px)');
  const sidebarOpen = useUI((s) => s.sidebarOpen);
  const setSidebarOpen = useUI((s) => s.setSidebarOpen);
  const composeOpen = useUI((s) => s.compose.open);
  const settingsOpen = useUI((s) => s.settingsOpen);

  const sidebarWidth = useAppearance((s) => s.sidebarWidth);
  const setSidebarWidth = useAppearance((s) => s.setSidebarWidth);
  const listWidth = useAppearance((s) => s.listWidth);
  const setListWidth = useAppearance((s) => s.setListWidth);

  // Live updates + server-stored appearance preferences.
  useWebSocket();
  const hydrateFromServer = useAppearance((s) => s.hydrateFromServer);
  const prefs = usePreferences();
  useEffect(() => {
    if (prefs.data) hydrateFromServer(prefs.data);
  }, [prefs.data, hydrateFromServer]);

  // Collapse the sidebar into a drawer on narrow screens.
  useEffect(() => {
    if (!isWide) setSidebarOpen(false);
  }, [isWide, setSidebarOpen]);

  const selectEmail = useCallback(
    (id: string | null) => navigate(id ? `/mail/${id}` : '/', { replace: false }),
    [navigate],
  );

  return (
    <div className="flex h-full flex-col bg-bg">
      <TopBar />
      <div className="flex min-h-0 flex-1">
        {/* Sidebar — inline (desktop) or drawer (mobile). */}
        {isWide && sidebarOpen && (
          <>
            <Sidebar width={sidebarWidth} />
            <Resizer width={sidebarWidth} onWidthChange={setSidebarWidth} />
          </>
        )}

        {/* Main area: adaptive 案A / 案B. */}
        {isWide ? (
          selectedId ? (
            <>
              <MailListPane
                variant="column"
                selectedId={selectedId}
                onSelect={selectEmail}
                className="flex-none"
                style={{ width: listWidth }}
              />
              <Resizer width={listWidth} onWidthChange={setListWidth} />
              <ReadingPane emailId={selectedId} onBack={() => selectEmail(null)} showBack={false} className="min-w-0 flex-1" />
            </>
          ) : (
            <MailListPane variant="wide" selectedId={null} onSelect={selectEmail} className="flex-1" />
          )
        ) : selectedId ? (
          <ReadingPane emailId={selectedId} onBack={() => selectEmail(null)} showBack className="min-w-0 flex-1" />
        ) : (
          <MailListPane variant="wide" selectedId={null} onSelect={selectEmail} className="flex-1" />
        )}
      </div>

      {/* Mobile sidebar drawer */}
      {!isWide && sidebarOpen && (
        <div className="fixed inset-0 z-40 flex">
          <div className="w-[260px] max-w-[80%] overflow-y-auto bg-surface shadow-compose">
            <Sidebar />
          </div>
          <div className="flex-1 bg-black/40" onClick={() => setSidebarOpen(false)} aria-hidden="true" />
        </div>
      )}

      {composeOpen && <ComposeDialog />}
      {settingsOpen && <SettingsDialog />}
    </div>
  );
}
