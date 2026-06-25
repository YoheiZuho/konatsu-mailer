// SPDX-License-Identifier: Apache-2.0
//
// AI draft generation over Server-Sent Events (design doc §5.7 / §7.3).
// POST /api/ai/draft returns text/event-stream; we surface incremental chunks.

import { API_BASE } from '@/lib/api';
import { authStore } from '@/stores/auth';
import type { DraftInput } from '@/lib/types';

export interface DraftStreamHandlers {
  onChunk: (text: string) => void;
  onDone?: () => void;
  onError?: (err: Error) => void;
  signal?: AbortSignal;
}

/**
 * Stream an AI-generated draft. Supports both SSE (`data:` framed) responses
 * and plain chunked text, so it works regardless of how the backend frames it.
 */
export async function streamDraft(input: DraftInput, handlers: DraftStreamHandlers): Promise<void> {
  const token = authStore.getState().accessToken;
  try {
    const res = await fetch(`${API_BASE}/ai/draft`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'text/event-stream',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify(input),
      signal: handlers.signal,
    });

    if (!res.ok || !res.body) {
      throw new Error(`AI draft request failed (${res.status})`);
    }

    const reader = res.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';

    for (;;) {
      const { value, done } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });

      // Split on SSE event boundaries (double newline).
      const events = buffer.split(/\n\n/);
      buffer = events.pop() ?? '';
      for (const evt of events) {
        const text = parseSseEvent(evt);
        if (text === '[DONE]') {
          handlers.onDone?.();
          return;
        }
        if (text) handlers.onChunk(text);
      }
    }

    // Flush any trailing buffered content.
    const tail = parseSseEvent(buffer);
    if (tail && tail !== '[DONE]') handlers.onChunk(tail);
    handlers.onDone?.();
  } catch (err) {
    if ((err as Error).name === 'AbortError') return;
    handlers.onError?.(err as Error);
  }
}

/** Extract the text payload from one SSE event block (or raw line). */
function parseSseEvent(block: string): string {
  const lines = block.split('\n');
  const dataLines = lines
    .filter((l) => l.startsWith('data:'))
    .map((l) => l.slice(5).replace(/^ /, ''));
  if (dataLines.length > 0) {
    const joined = dataLines.join('\n');
    // Some servers send JSON-encoded deltas; try to unwrap a `.text`/`.content`.
    try {
      const obj = JSON.parse(joined);
      if (typeof obj === 'string') return obj;
      return obj.text ?? obj.content ?? obj.delta ?? '';
    } catch {
      return joined;
    }
  }
  // Not SSE-framed: treat the whole block as text.
  return block.trim() ? block : '';
}
