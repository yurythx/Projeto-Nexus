import {
  Archive,
  Baby,
  BookOpen,
  Boxes,
  Briefcase,
  Building2,
  Cloud,
  Code,
  Cpu,
  Database,
  Eye,
  FileCode,
  FileText,
  Filter,
  FolderOpen,
  Globe,
  HandHeart,
  HardDrive,
  HeartPulse,
  Home,
  Info,
  KeyRound,
  Layers,
  LayoutDashboard,
  LayoutGrid,
  MessageCircle,
  MessageSquare,
  Monitor,
  Network,
  Scale,
  Server,
  Shield,
  ShieldCheck,
  Terminal,
  Users,
  Workflow,
  Wrench,
  type LucideProps,
} from "lucide-react";
import type { ComponentType } from "react";

// Ícone de um serviço do catálogo (campo "icon", nome lucide livre
// escolhido pelo gestor). Nome desconhecido cai no ícone genérico.
const ICONS: Record<string, ComponentType<LucideProps>> = {
  archive: Archive,
  baby: Baby,
  "book-open": BookOpen,
  boxes: Boxes,
  briefcase: Briefcase,
  building: Building2,
  cloud: Cloud,
  code: Code,
  cpu: Cpu,
  database: Database,
  eye: Eye,
  "file-code": FileCode,
  "file-text": FileText,
  filter: Filter,
  "folder-open": FolderOpen,
  globe: Globe,
  "hand-heart": HandHeart,
  "hard-drive": HardDrive,
  "heart-pulse": HeartPulse,
  home: Home,
  info: Info,
  "key-round": KeyRound,
  layers: Layers,
  "layout-dashboard": LayoutDashboard,
  "message-circle": MessageCircle,
  "message-square": MessageSquare,
  monitor: Monitor,
  network: Network,
  scale: Scale,
  server: Server,
  shield: Shield,
  "shield-check": ShieldCheck,
  terminal: Terminal,
  users: Users,
  workflow: Workflow,
  wrench: Wrench,
};

export const SERVICE_ICON_NAMES = Object.keys(ICONS);

export function ServiceIcon({ name, ...props }: { name?: string } & LucideProps) {
  const Icon = (name && ICONS[name]) || LayoutGrid;
  return <Icon {...props} />;
}

export function getCategoryIconStyles(category?: string | null): {
  badge: string;
  icon: string;
  tag: string;
} {
  const cat = (category || "").toLowerCase();

  // 1. Comunicação & Mensageria (WhatsApp, Chatwoot, Evolution)
  if (cat.includes("comunicação") || cat.includes("comunicacao") || cat.includes("mensageria")) {
    return {
      badge: "bg-emerald-500/15 border border-emerald-600/30 text-emerald-800 dark:bg-emerald-500/20 dark:border-emerald-400/50 dark:text-emerald-300 shadow-xs",
      icon: "text-emerald-700 dark:text-emerald-300",
      tag: "bg-emerald-50 text-emerald-900 border-emerald-300 dark:bg-emerald-950/80 dark:text-emerald-300 dark:border-emerald-500/50 font-bold",
    };
  }

  // 2. Automação & DevOps (n8n, Ansible)
  if (cat.includes("automação") || cat.includes("automacao") || cat.includes("integração")) {
    return {
      badge: "bg-violet-500/15 border border-violet-600/30 text-violet-800 dark:bg-violet-500/20 dark:border-violet-400/50 dark:text-violet-300 shadow-xs",
      icon: "text-violet-700 dark:text-violet-300",
      tag: "bg-violet-50 text-violet-900 border-violet-300 dark:bg-violet-950/80 dark:text-violet-300 dark:border-violet-500/50 font-bold",
    };
  }

  // 3. Segurança & Identidade (Keycloak, Tailscale, CrowdSec)
  if (cat.includes("segurança") || cat.includes("seguranca") || cat.includes("identidade")) {
    return {
      badge: "bg-rose-500/15 border border-rose-600/30 text-rose-800 dark:bg-rose-500/20 dark:border-rose-400/50 dark:text-rose-300 shadow-xs",
      icon: "text-rose-700 dark:text-rose-300",
      tag: "bg-rose-50 text-rose-900 border-rose-300 dark:bg-rose-950/80 dark:text-rose-300 dark:border-rose-500/50 font-bold",
    };
  }

  // 4. Gestão de TI & Monitoramento (GLPI, Zabbix)
  if (cat.includes("gestão") || cat.includes("gestao") || cat.includes("monitoramento")) {
    return {
      badge: "bg-amber-500/15 border border-amber-600/30 text-amber-900 dark:bg-amber-500/20 dark:border-amber-400/50 dark:text-amber-300 shadow-xs",
      icon: "text-amber-800 dark:text-amber-300",
      tag: "bg-amber-50 text-amber-950 border-amber-300 dark:bg-amber-950/80 dark:text-amber-300 dark:border-amber-500/50 font-bold",
    };
  }

  // 5. Acesso Remoto, Redes & Física (RustDesk, Samba, Windows Server, Redes, Cabeamento)
  if (cat.includes("acesso") || cat.includes("redes") || cat.includes("física") || cat.includes("fisica") || cat.includes("hardware")) {
    return {
      badge: "bg-cyan-500/15 border border-cyan-600/30 text-cyan-800 dark:bg-cyan-500/20 dark:border-cyan-400/50 dark:text-cyan-300 shadow-xs",
      icon: "text-cyan-700 dark:text-cyan-300",
      tag: "bg-cyan-50 text-cyan-900 border-cyan-300 dark:bg-cyan-950/80 dark:text-cyan-300 dark:border-cyan-500/50 font-bold",
    };
  }

  // 6. Produtividade & CRM (Twenty CRM, Papermark, Plone)
  if (cat.includes("produtividade") || cat.includes("crm") || cat.includes("documentos")) {
    return {
      badge: "bg-indigo-500/15 border border-indigo-600/30 text-indigo-800 dark:bg-indigo-500/20 dark:border-indigo-400/50 dark:text-indigo-300 shadow-xs",
      icon: "text-indigo-700 dark:text-indigo-300",
      tag: "bg-indigo-50 text-indigo-900 border-indigo-300 dark:bg-indigo-950/80 dark:text-indigo-300 dark:border-indigo-500/50 font-bold",
    };
  }

  // 7. Padrão: Infraestrutura, Nuvem & Storages (Proxmox, TrueNAS, Nextcloud, aaPanel, Terraform)
  return {
    badge: "bg-blue-500/15 border border-blue-600/30 text-blue-900 dark:bg-blue-500/20 dark:border-blue-400/50 dark:text-blue-300 shadow-xs",
    icon: "text-blue-700 dark:text-blue-300",
    tag: "bg-blue-50 text-blue-900 border-blue-300 dark:bg-blue-950/80 dark:text-blue-300 dark:border-blue-500/50 font-bold",
  };
}
