// SPDX-License-Identifier: Apache-2.0

import { useEffect } from 'react';
import { Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { useAuth } from '@/stores/auth';
import { useAppearance } from '@/stores/appearance';
import { applyTheme, watchSystemTheme } from '@/lib/theme';
import { LoginPage } from '@/features/auth/LoginPage';
import { MailApp } from '@/features/mail/MailApp';

/** Re-apply the resolved palette when the OS theme changes (while in 'system'). */
function useSystemThemeSync() {
  const theme = useAppearance((s) => s.theme);
  useEffect(() => {
    if (theme !== 'system') return;
    return watchSystemTheme(() => applyTheme('system'));
  }, [theme]);
}

function RequireAuth({ children }: { children: React.ReactNode }) {
  const authed = useAuth((s) => !!s.accessToken);
  const location = useLocation();
  if (!authed) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }
  return <>{children}</>;
}

export function App() {
  useSystemThemeSync();
  const authed = useAuth((s) => !!s.accessToken);

  return (
    <Routes>
      <Route path="/login" element={authed ? <Navigate to="/" replace /> : <LoginPage />} />
      <Route
        path="/"
        element={
          <RequireAuth>
            <MailApp />
          </RequireAuth>
        }
      />
      <Route
        path="/mail/:emailId"
        element={
          <RequireAuth>
            <MailApp />
          </RequireAuth>
        }
      />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
