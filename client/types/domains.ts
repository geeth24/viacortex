export interface Domain {
  id: number;
  name: string;
  target_url: string;
  ssl_enabled: boolean;
  force_https: boolean;
  http2_enabled: boolean;
  hsts_enabled: boolean;
  hsts_max_age: number;
  websocket_enabled: boolean;
  custom_headers: Record<string, string>;
  basic_auth_enabled: boolean;
  basic_auth_username?: string;
  basic_auth_password?: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface IPRule {
  id: number;
  domain_id: number;
  ip_range: string;
  rule_type: "whitelist" | "blacklist";
  description: string;
  enabled: boolean;
  created_at: string;
}

export interface RateLimit {
  id: number;
  domain_id: number;
  requests_per_second: number;
  burst_size: number;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface Redirect {
  id: number;
  domain_id: number;
  source_path: string;
  target_url: string;
  redirect_type: 301 | 302 | 307 | 308;
  preserve_query: boolean;
  enabled: boolean;
  created_at: string;
}

export interface AccessLog {
  id: number;
  domain_id: number;
  client_ip: string;
  method: string;
  path: string;
  status_code: number;
  response_time_ms: number;
  user_agent: string;
  created_at: string;
}

export interface User {
  id: number;
  email: string;
  name: string;
  role: "admin" | "user";
  active: boolean;
  last_login?: string;
  created_at: string;
  updated_at: string;
}

export interface AuditLog {
  id: number;
  user_id: number;
  action: string;
  entity_type: string;
  entity_id: number;
  changes: Record<string, unknown>;
  ip_address: string;
  created_at: string;
}
