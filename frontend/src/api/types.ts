// DTO mirror of the Go BFF structs. snake_case throughout, matching the JSON
// the backend emits (PLAN.md §4 + locked §0 reconciliations).

export type Role = 'admin' | 'viewer';
export type ProviderId = 'aws' | 'gcp' | 'azure';
export type LogLevel = 'info' | 'warn' | 'error';

export interface Meta {
  request_id: string;
  timestamp: string; // ISO 8601
}
export interface ApiOk<T> {
  data: T;
  meta: Meta;
}
export interface ApiErr {
  error: {
    code: string; // INVALID_CREDENTIALS, TOKEN_EXPIRED, PROVIDER_NOT_ALLOWED, ...
    message: string;
    request_id: string;
  };
}

// Per locked §0: the login response carries `user.allowed_providers`, so the
// SPA never needs to decode the JWT (jwt-decode dependency dropped).
export interface AuthUser {
  id: string;
  email: string;
  role: Role;
  allowed_providers: string[] | null; // null = all providers (admin)
}
export interface LoginReq {
  email: string;
  password: string;
}
export interface LoginData {
  token: string;
  token_type: 'Bearer';
  expires_in: number; // seconds
  user: AuthUser;
}

// Per locked §0: Azure containers are normalized server-side into `buckets[]`,
// so the SPA is provider-uniform (no `containers` fallback).
export interface ProviderInfo {
  id: ProviderId;
  name: string;
  icon?: string;
  buckets: string[];
  healthy: boolean;
  region?: string;
  error?: string; // error code when healthy=false
}
export interface ListProvidersData {
  providers: ProviderInfo[];
}

// Mirrors the Go `ObjectInfo` exactly (PLAN.md §3.1) — the server returns ONLY
// these fields. filename / size_human / level are CLIENT-DERIVED (see LogFileView).
export interface LogFile {
  key: string;
  size: number; // bytes
  last_modified: string; // ISO 8601
  etag: string;
  content_type: string;
}
export interface ListLogsData {
  files: LogFile[];
  next_cursor: string | null;
  count: number;
}

export interface LogFileView extends LogFile {
  filename: string;
  size_human: string;
  level: LogLevel;
}

// Locked §0: request body is snake_case `{ key, ttl_seconds }`.
export interface PresignReq {
  key: string;
  ttl_seconds: number;
}
export interface PresignData {
  url: string;
  expires_at: string; // ISO 8601
  ttl_seconds: number; // actual, post-clamp value
}
