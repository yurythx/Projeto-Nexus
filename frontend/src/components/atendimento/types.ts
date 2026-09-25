export interface Atendimento {
  id: string;
  protocol?: string;
  protocol_number?: string;
  service_slug: string;
  unit: string;
  citizen_name: string;
  citizen_cpf: string;
  citizen_rg: string;
  citizen_phone: string;
  citizen_neighborhood: string;
  citizen_address: string;
  access_form: string;
  demand_type: string;
  risk_level: string;
  vulnerabilities: Record<string, any> | null;
  technical_notes: string;
  referrals: Record<string, any> | null;
  status: "aguardando" | "em_atendimento" | "concluido" | "cancelado";
  attendant_id?: string | null;
  attendant_name?: string | null;
  created_at: string;
  updated_at: string;
  finished_at?: string | null;
}

export interface AtendimentoStats {
  total: number;
  aguardando: number;
  em_atendimento: number;
  concluido: number;
}

export interface CitizenSearchResult {
  found?: boolean;
  cpf?: string;
  citizen_cpf?: string;
  citizen_name?: string;
  citizen_rg?: string;
  citizen_phone?: string;
  citizen_neighborhood?: string;
  citizen_address?: string;
  total_attendances?: number;
  recent_attendances?: Atendimento[];
  history?: Atendimento[];
  last_record?: Atendimento;
}

export interface CentroPopProntuario {
  id: string;
  legacy_id?: number | null;
  name: string;
  preferred_name: string;
  nickname: string;
  cpf: string;
  rg: string;
  date_birth?: string | null;
  mother_name: string;
  father_name: string;
  sex: string;
  gender: string;
  situation: string;
  time_homelessness?: number | null;
  reason_homelessness: string;
  address: string;
  date_opened: string;
  health_notes: Record<string, any> | null;
  photo?: string;
  created_at: string;
  updated_at: string;
}

export interface ServicePermissionInfo {
  service_slug?: string;
  name?: string;
  unit?: string;
  default_unit?: string;
  is_admin: boolean;
  is_technician?: boolean;
  is_receptionist?: boolean;
  roles?: string[];
  groups?: string[];
  allowed_services?: string[];
  services?: Array<{
    slug: string;
    name: string;
    category: string;
    unit_default: string;
    can_change_unit: boolean;
  }>;
}
