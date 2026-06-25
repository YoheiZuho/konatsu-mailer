// SPDX-License-Identifier: Apache-2.0
//
// Live updates via WebSocket (design doc §8). Maps server events onto the
// TanStack Query cache and the sync-status badge. Reconnects with backoff and
// fails quietly so a missing/unimplemented WS endpoint never breaks the UI.

import { useEffect, useRef } from 'react';
import { useQueryClient, type InfiniteData } from '@tanstack/react-query';
import { apiUrl } from '@/lib/api';
import { authStore } from '@/stores/auth';
import { useUI } from '@/stores/ui';
import type { EmailListResponse, ServerEvent } from '@/lib/types';

function wsUrl(): string {
  const httpUrl = apiUrl('/ws');
  return httpUrl.replace(/^http/, 'ws');
}

export function useWebSocket() {
  const qc = useQueryClient();
  const setSyncStatus = useUI((s) => s.setSyncStatus);
  const retryRef = useRef(0);
  const socketRef = useRef<WebSocket | null>(null);
  const closedRef = useRef(false);

  useEffect(() => {
    closedRef.current = false;
    let reconnectTimer: ReturnType<typeof setTimeout>;

    const connect = () => {
      const token = authStore.getState().accessToken;
      if (!token || closedRef.current) return;

      let ws: WebSocket;
      try {
        // Token is passed via the Sec-WebSocket-Protocol header (design doc §8).
        ws = new WebSocket(wsUrl(), ['bearer', token]);
      } catch {
        scheduleReconnect();
        return;
      }
      socketRef.current = ws;

      ws.onopen = () => {
        retryRef.current = 0;
        setSyncStatus('connected');
      };

      ws.onmessage = (event) => {
        let msg: ServerEvent;
        try {
          msg = JSON.parse(event.data);
        } catch {
          return;
        }
        handleEvent(msg);
      };

      ws.onclose = () => {
        socketRef.current = null;
        if (!closedRef.current) scheduleReconnect();
      };

      ws.onerror = () => ws.close();
    };

    const scheduleReconnect = () => {
      if (closedRef.current) return;
      setSyncStatus(retryRef.current === 0 ? 'reconnecting' : 'down');
      // Exponential backoff capped at 30s (design doc §3.4).
      const delay = Math.min(1000 * 2 ** retryRef.current, 30_000);
      retryRef.current += 1;
      reconnectTimer = setTimeout(connect, delay);
    };

    const handleEvent = (msg: ServerEvent) => {
      switch (msg.type) {
        case 'NEW_MAIL':
          qc.invalidateQueries({ queryKey: ['emails'] });
          break;
        case 'MAIL_ANALYZED': {
          const { email_id, ai_summary, ai_priority, labels } = msg.payload;
          // Pinpoint cache update — no refetch, avoids list flicker (§9.4).
          qc.setQueriesData<InfiniteData<EmailListResponse>>(
            { queryKey: ['emails'] },
            (data) =>
              data && {
                ...data,
                pages: data.pages.map((page) => ({
                  ...page,
                  items: page.items.map((item) =>
                    item.id === email_id
                      ? { ...item, ai_summary, ai_priority, labels, analysis_status: 'done' }
                      : item,
                  ),
                })),
              },
          );
          qc.invalidateQueries({ queryKey: ['email', email_id] });
          break;
        }
        case 'SYNC_STATUS':
          setSyncStatus(msg.payload.state);
          break;
      }
    };

    connect();

    return () => {
      closedRef.current = true;
      clearTimeout(reconnectTimer);
      socketRef.current?.close();
    };
  }, [qc, setSyncStatus]);
}
