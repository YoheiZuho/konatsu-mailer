// SPDX-License-Identifier: Apache-2.0
//
// The authenticated mail workspace: top bar + sidebar + list + reading pane
// (design doc layout 案B), responsive to a single column on narrow screens.

import { useCallback, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import clsx from 'clsx';

import { useUI } from '@/stores/ui';
import { useAppearance } from '@/stores/appearance';
import { usePreferences } from '@/hooks/queries';
import { useWebSocket } from '@/hooks/useWebSocket';
import { useMediaQuery } from '@/hooks/useMediaQuery';

import { TopBar } from '@/components/layout/TopBar';
import { Sidebar } from '@/components/layout/Sidebar';
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
  const composeOpen = useUI((s) => s.compose.open);
  const settingsOpen = useUI((s) => s.settingsOpen);

  // Live updates + server-stored appearance preferences.
  useWebSocket();
  const hydrateFromServer = useAppearance((s) => s.hydrateFromServer);
  const prefs = usePreferences();
  useEffect(() => {
    if (prefs.data) hydrateFromServer(prefs.data);
  }, [prefs.data, hydrateFromServer]);

  const selectEmail = useCallback(
    (id: string | null) => navigate(id ? `/mail/${id}` : '/', { replace: false }),
    [navigate],
  );

  // On narrow screens, show either the list or the reading pane (not both).
  const showList = isWide || !selectedId;
  const showReading = isWide || !!selectedId;

  return (
    <div className="flex h-full flex-col bg-bg">
      <TopBar />
      <div className="flex min-h-0 flex-1">
        {sidebarOpen && <Sidebar />}
        {showList && (
          <MailListPane
            selectedId={selectedId}
            onSelect={selectEmail}
            className={clsx(isWide ? 'w-[430px] flex-none' : 'flex-1')}
          />
        )}
        {showReading && (
          <ReadingPane
            emailId={selectedId}
            onBack={() => selectEmail(null)}
            showBack={!isWide}
            className="min-w-0 flex-1"
          />
        )}
      </div>

      {composeOpen && <ComposeDialog />}
      {settingsOpen && <SettingsDialog />}
    </div>
  );
}
