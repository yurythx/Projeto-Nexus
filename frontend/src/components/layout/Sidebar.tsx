"use client";

import {
  LayoutDashboard,
  Plug,
  Settings,
  Activity,
  Box,
  ClipboardList,
  HeartHandshake,
  Home,
  Shield,
  ShieldAlert,
  Scale,
  FileSpreadsheet,
  Gift,
  Building,
} from "lucide-react";
import Link from "next/link";
import { usePathname, useSearchParams } from "next/navigation";
import { useEffect } from "react";
import { useSession } from "next-auth/react";

import useSWR from "swr";
import { apiClient } from "@/lib/api/client";
import { useBranding } from "@/components/branding/BrandingContext";
import { getUserAssignedServices, type UserServiceInfo } from "@/lib/auth/adMapping";
import type { FeatureFlag } from "@/types/api";
import packageJson from "../../../package.json";

const FRONTEND_VERSION = packageJson.version;

const systemLinks: { href: string; label: string; icon: typeof LayoutDashboard; flag?: string; match?: string[] }[] = [
  { href: "/dashboard", label: "Visão Geral", icon: LayoutDashboard },
  { href: "/exemplos", label: "Módulo Modelo", icon: Box, flag: "module_exemplos_enabled" },
  { href: "/monitoramento", label: "Monitoramento", icon: Activity },
  { href: "/integracoes", label: "Integrações", icon: Plug },
  { href: "/configuracao", label: "Configurações", icon: Settings },
];

function getServiceIcon(slug: string) {
  switch (slug) {
    case "cras":
      return Home;
    case "centro-pop":
    case "centro_pop":
      return HeartHandshake;
    case "creas":
      return ShieldAlert;
    case "casa-mulher":
    case "casa-da-mulher":
      return Shield;
    case "conselho-tutelar":
      return Scale;
    case "cadunico-bolsa-familia":
      return FileSpreadsheet;
    case "beneficios-eventuais":
      return Gift;
    default:
      return Building;
  }
}

function matchesPath(pathname: string, prefix: string): boolean {
  return pathname === prefix || pathname.startsWith(prefix + "/");
}

interface SidebarProps {
  collapsed: boolean;
  mobileOpen: boolean;
  onCloseMobile: () => void;
}

export function Sidebar({ collapsed, mobileOpen, onCloseMobile }: SidebarProps) {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const currentServico = searchParams.get("servico");
  const { branding } = useBranding();
  const { data: session } = useSession();

  const { data: featureFlags } = useSWR<FeatureFlag[]>(
    "v1/admin/feature-flags",
    () => apiClient.get<FeatureFlag[]>("v1/admin/feature-flags").then((res) => res.data),
    { revalidateOnFocus: false, shouldRetryOnError: false }
  );

  const disabledFlags = new Set(
    (featureFlags ?? []).filter((f) => f.enabled === false).map((f) => f.key)
  );

  const userRoles = session?.user?.roles ?? [];
  const userGroups = session?.user?.groups ?? [];

  const { services: assignedServices, activeUnit, isAdmin, isTechnician, isReceptionist } = getUserAssignedServices(
    userGroups,
    userRoles,
    disabledFlags
  );

  const visibleSystemLinks = systemLinks.filter((l) => !l.flag || !disabledFlags.has(l.flag));

  let activeHref: string | null = null;
  let bestLen = -1;
  for (const l of visibleSystemLinks) {
    for (const prefix of l.match ?? [l.href]) {
      if (matchesPath(pathname, prefix) && prefix.length > bestLen) {
        bestLen = prefix.length;
        activeHref = l.href;
      }
    }
  }

  useEffect(() => {
    if (!mobileOpen) return;
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") onCloseMobile();
    }
    document.addEventListener("keydown", onKeyDown);
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.removeEventListener("keydown", onKeyDown);
      document.body.style.overflow = previousOverflow;
    };
  }, [mobileOpen, onCloseMobile]);

  return (
    <>
      {mobileOpen && (
        <div
          className="fixed inset-x-0 bottom-0 top-[var(--topbar-h)] z-40 bg-black/40 md:hidden"
          onClick={onCloseMobile}
          aria-hidden="true"
        />
      )}

      <nav
        aria-label="Principal"
        className={`fixed left-0 bottom-0 top-[var(--topbar-h)] z-40 flex flex-col overflow-y-auto overflow-x-hidden border-r border-surface-border bg-surface
          transition-[transform,width] duration-[var(--shell-motion)] ease-[var(--shell-ease)] md:translate-x-0
          ${collapsed ? "md:w-[var(--sidebar-w-collapsed)]" : "md:w-[var(--sidebar-w)]"}
          ${mobileOpen ? "translate-x-0" : "-translate-x-full"} w-72`}
      >
        {/* Painel de Lotação do Usuário no AD */}
        {!collapsed && activeUnit && (
          <div className="mx-3 mt-3 mb-1 rounded-md border border-primary/20 bg-primary/5 px-2.5 py-1.5 text-xs">
            <div className="flex items-center justify-between">
              <span className="block font-semibold uppercase tracking-wider text-[10px] text-primary">Lotação / AD</span>
              {isAdmin ? (
                <span className="text-[9px] bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 font-bold px-1.5 py-0.5 rounded">Gestão</span>
              ) : isReceptionist ? (
                <span className="text-[9px] bg-amber-500/10 text-amber-600 dark:text-amber-400 font-bold px-1.5 py-0.5 rounded">Recepção</span>
              ) : (
                <span className="text-[9px] bg-primary/10 text-primary font-bold px-1.5 py-0.5 rounded">Técnico Social</span>
              )}
            </div>
            <span className="font-medium text-foreground truncate block mt-0.5">{activeUnit}</span>
          </div>
        )}

        {/* MÓDULOS DE ATENDIMENTO SOCIOASSISTENCIAL */}
        {assignedServices.length > 0 && (
          <div className="mt-2">
            {!collapsed && (
              <div className="px-3 pt-2 pb-1 text-[11px] font-semibold uppercase tracking-wider text-muted flex items-center justify-between">
                <span>Atendimento Socioassistencial</span>
                {isAdmin && <span className="text-[10px] bg-primary/10 text-primary px-1.5 py-0.5 rounded font-normal">Todas Unidades</span>}
              </div>
            )}
            <ul className="flex flex-col gap-0.5 px-2">
              {isAdmin && (
                <li>
                  <Link
                    href="/atendimento"
                    onClick={onCloseMobile}
                    title={collapsed ? "Todos os Atendimentos" : undefined}
                    aria-label={collapsed ? "Todos os Atendimentos" : undefined}
                    className={`flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors
                      ${pathname === "/atendimento" && !currentServico ? "bg-primary/10 text-primary" : "text-foreground hover:bg-black/5 dark:hover:bg-white/5"}
                      ${collapsed ? "md:justify-center" : ""}`}
                  >
                    <ClipboardList size={18} aria-hidden="true" className="shrink-0" />
                    <span className={collapsed ? "md:hidden" : ""}>Fila & Atendimentos</span>
                  </Link>
                </li>
              )}

              {assignedServices.map((srv: UserServiceInfo) => {
                const Icon = getServiceIcon(srv.slug);
                const srvHref = `/atendimento?servico=${srv.slug}`;
                const isSelected = pathname === "/atendimento" && currentServico === srv.slug;
                return (
                  <li key={srv.slug}>
                    <Link
                      href={srvHref}
                      onClick={onCloseMobile}
                      title={collapsed ? srv.name : undefined}
                      aria-label={collapsed ? srv.name : undefined}
                      className={`flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors
                        ${isSelected ? "bg-primary/10 text-primary font-semibold" : "text-foreground hover:bg-black/5 dark:hover:bg-white/5"}
                        ${collapsed ? "md:justify-center" : ""}`}
                    >
                      <Icon size={18} aria-hidden="true" className="shrink-0 text-primary/80" />
                      <span className={collapsed ? "md:hidden" : ""}>{srv.name}</span>
                    </Link>
                  </li>
                );
              })}
            </ul>
          </div>
        )}

        {/* MENU PRINCIPAL & SISTEMA */}
        <div className="mt-2">
          {!collapsed && (
            <div className="px-3 pt-3 pb-1 text-[11px] font-semibold uppercase tracking-wider text-muted">
              Sistema & Gestão
            </div>
          )}
          <ul className={`flex flex-col gap-0.5 px-2 pb-3 ${collapsed ? "pt-2" : "pt-1"}`}>
            {visibleSystemLinks.map((link) => {
              const Icon = link.icon;
              const active = link.href === activeHref;
              return (
                <li key={link.href}>
                  <Link
                    href={link.href}
                    onClick={onCloseMobile}
                    title={collapsed ? link.label : undefined}
                    aria-label={collapsed ? link.label : undefined}
                    aria-current={active ? "page" : undefined}
                    className={`flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors
                      ${active ? "bg-primary/10 text-primary" : "text-foreground hover:bg-black/5 dark:hover:bg-white/5"}
                      ${collapsed ? "md:justify-center" : ""}`}
                  >
                    <Icon size={18} aria-hidden="true" className="shrink-0" />
                    <span className={collapsed ? "md:hidden" : ""}>{link.label}</span>
                  </Link>
                </li>
              );
            })}
          </ul>
        </div>

        {/* Rodapé informativo */}
        <div
          className={`mt-auto shrink-0 border-t border-surface-border px-3 py-3 text-[11px] leading-tight text-muted
            ${collapsed ? "md:px-0 md:text-center" : ""}`}
          title={`${branding.appName} · v${FRONTEND_VERSION}`}
        >
          <p className={`truncate font-medium text-foreground/70 ${collapsed ? "md:hidden" : ""}`}>
            {branding.appName}
          </p>
          <p className={collapsed ? "md:hidden" : ""}>v{FRONTEND_VERSION}</p>
        </div>
      </nav>
    </>
  );
}
