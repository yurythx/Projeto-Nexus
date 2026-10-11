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
  if (cat.includes("comunicação") || cat.includes("comunicacao") || cat.includes("mensageria")) {
    return {
      badge: "bg-emerald-50 border border-emerald-200 text-emerald-700 dark:bg-emerald-950/60 dark:border-emerald-800/60 dark:text-emerald-300",
      icon: "text-emerald-700 dark:text-emerald-300",
      tag: "bg-emerald-50 text-emerald-800 border-emerald-200 dark:bg-emerald-950/60 dark:text-emerald-300 dark:border-emerald-800/60",
    };
  }
  if (cat.includes("automação") || cat.includes("automacao") || cat.includes("integração")) {
    return {
      badge: "bg-purple-50 border border-purple-200 text-purple-700 dark:bg-purple-950/60 dark:border-purple-800/60 dark:text-purple-300",
      icon: "text-purple-700 dark:text-purple-300",
      tag: "bg-purple-50 text-purple-800 border-purple-200 dark:bg-purple-950/60 dark:text-purple-300 dark:border-purple-800/60",
    };
  }
  if (cat.includes("segurança") || cat.includes("seguranca") || cat.includes("identidade")) {
    return {
      badge: "bg-amber-50 border border-amber-200 text-amber-800 dark:bg-amber-950/60 dark:border-amber-800/60 dark:text-amber-300",
      icon: "text-amber-800 dark:text-amber-300",
      tag: "bg-amber-50 text-amber-900 border-amber-200 dark:bg-amber-950/60 dark:text-amber-300 dark:border-amber-800/60",
    };
  }
  if (cat.includes("gestão") || cat.includes("gestao") || cat.includes("monitoramento")) {
    return {
      badge: "bg-indigo-50 border border-indigo-200 text-indigo-700 dark:bg-indigo-950/60 dark:border-indigo-800/60 dark:text-indigo-300",
      icon: "text-indigo-700 dark:text-indigo-300",
      tag: "bg-indigo-50 text-indigo-800 border-indigo-200 dark:bg-indigo-950/60 dark:text-indigo-300 dark:border-indigo-800/60",
    };
  }
  // Padrão: Infraestrutura / Nuvem / Redes
  return {
    badge: "bg-sky-50 border border-sky-200 text-sky-700 dark:bg-sky-950/60 dark:border-sky-800/60 dark:text-sky-300",
    icon: "text-sky-700 dark:text-sky-300",
    tag: "bg-sky-50 text-sky-800 border-sky-200 dark:bg-sky-950/60 dark:text-sky-300 dark:border-sky-800/60",
  };
}
