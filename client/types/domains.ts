export interface Domain {
  id: number;
  name: string;
  target_url: string;
  ssl_enabled: boolean;
  health_check_enabled: boolean;
  health_check_interval: number;
  custom_error_pages?: Record<string, string>;
  created_at: string;
  updated_at: string;
}

export interface BackendServer {
  id: number;
  domain_id: number;
  name?: string;
  scheme: 'http' | 'https' | 'tcp';
  ip: string;
  port: number;
  weight: number;
  is_active: boolean;
  last_health_check?: string;
  health_status?: string;
  created_at?: string;
  updated_at?: string;
}

export interface DomainWithBackends {
  domain: Domain;
  backend_servers: BackendServer[];
}

export interface DomainsResponse {
  [index: number]: DomainWithBackends;
} 