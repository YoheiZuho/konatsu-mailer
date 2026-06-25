// SPDX-License-Identifier: Apache-2.0
//
// API data-transfer types, mirroring the backend contracts (design doc §7/§8).

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  expires_at: string;
}

/** Public auth settings used by the login UI (GET /api/auth/config). */
export interface AuthConfig {
  allow_registration: boolean;
}

/** Optional IMAP/SMTP account provisioned during registration. */
export interface MailAccountSetup {
  email: string;
  imap_host: string;
  imap_port: number;
  imap_use_tls: boolean;
  smtp_host: string;
  smtp_port: number;
  smtp_use_starttls: boolean;
  auth_user?: string;
  password: string;
}

export interface RegisterInput {
  email: string;
  password: string;
  display_name?: string;
  mail_account?: MailAccountSetup;
}

export interface ApiError {
  error: { code: string; message: string };
}

/** Label as attached to an email row or returned by /api/labels. */
export interface Label {
  id?: string;
  name: string;
  /** UI display color (oklch string or HEX). */
  color: string;
  is_system?: boolean;
}

export type AnalysisStatus = 'pending' | 'skipped' | 'done' | 'error';

/** A row in the mail list — GET /api/emails. */
export interface EmailListItem {
  id: string;
  sender_name: string | null;
  sender_addr: string;
  subject: string | null;
  body_preview: string | null;
  ai_summary: string | null;
  ai_priority: number | null; // 1..5
  labels: Label[];
  is_read: boolean;
  is_starred: boolean;
  has_attachment: boolean;
  date_sent: string;
  analysis_status: AnalysisStatus;
  thread_id?: string | null;
}

export interface EmailListResponse {
  items: EmailListItem[];
  next_cursor: string | null;
}

export interface Address {
  name?: string;
  addr: string;
}

export interface AttachmentMeta {
  id: string;
  filename: string;
  content_type?: string;
  size?: number;
}

/** A single message within a thread — part of GET /api/emails/:id. */
export interface ThreadMessage {
  id: string;
  from: Address;
  to: Address[];
  cc?: Address[];
  date: string;
  subject?: string | null;
  html?: string | null;
  text?: string | null;
  ai_summary?: string | null;
  ai_priority?: number | null;
  attachments?: AttachmentMeta[];
  is_read?: boolean;
}

export interface ThreadDetail {
  thread_id: string;
  subject?: string | null;
  labels?: Label[];
  is_starred?: boolean;
  messages: ThreadMessage[];
}

export interface Account {
  id: string;
  email: string;
  imap_host: string;
  imap_port: number;
  imap_use_tls: boolean;
  smtp_host: string;
  smtp_port: number;
  smtp_use_starttls: boolean;
  auth_user: string;
  is_active: boolean;
}

/** Payload for creating/updating an account. Password is write-only. */
export interface AccountInput {
  email: string;
  imap_host: string;
  imap_port: number;
  imap_use_tls: boolean;
  smtp_host: string;
  smtp_port: number;
  smtp_use_starttls: boolean;
  auth_user: string;
  password?: string;
  is_active?: boolean;
}

export interface LLMConfig {
  id: string;
  name: string;
  base_url: string;
  model: string;
  temperature: number;
  max_tokens: number;
  supports_json_schema: boolean;
  request_timeout_ms: number;
  is_default: boolean;
  is_active: boolean;
  /** True when an API key is stored; the key itself is never returned. */
  has_api_key?: boolean;
}

export interface LLMConfigInput {
  name: string;
  base_url: string;
  model: string;
  api_key?: string;
  temperature: number;
  max_tokens: number;
  supports_json_schema: boolean;
  request_timeout_ms: number;
  is_default: boolean;
  is_active: boolean;
}

export interface LLMTestResult {
  ok: boolean;
  models?: string[];
  error?: string;
}

export type ThemePref = 'system' | 'light' | 'dark';
export type Density = 'comfortable' | 'compact';

// --- Translation (LibreTranslate) ---

export interface TranslateConfig {
  enabled: boolean;
  default_target: string;
}

export interface TranslateLanguage {
  code: string;
  name: string;
}

export interface TranslateResult {
  translated_text: string;
  target: string;
  detected_source?: string;
}

export interface Preferences {
  theme: ThemePref;
  brand_color: string;
  density: Density;
  ai_summaries: boolean;
}

export interface SendEmailInput {
  account_id?: string;
  to: string[];
  cc?: string[];
  bcc?: string[];
  subject: string;
  text: string;
  html?: string;
  in_reply_to?: string;
  thread_id?: string;
}

export interface DraftInput {
  /** "reply" to an existing thread, or "compose" from scratch. */
  mode: 'reply' | 'compose';
  thread_id?: string;
  email_id?: string;
  instruction?: string;
  context?: string;
}

// --- WebSocket events (design doc §8) ---

export type SyncState = 'connected' | 'reconnecting' | 'down';

export type ServerEvent =
  | { type: 'NEW_MAIL'; payload: { account_id: string; email_id: string; summary?: string } }
  | {
      type: 'MAIL_ANALYZED';
      payload: {
        email_id: string;
        ai_summary: string;
        ai_priority: number;
        labels: Label[];
      };
    }
  | { type: 'SYNC_STATUS'; payload: { account_id: string; state: SyncState } };
