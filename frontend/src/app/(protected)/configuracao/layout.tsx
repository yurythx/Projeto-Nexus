"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";

// Sub-navegação de /configuracao: Sistema (index, feature flags) e
// Usuários. Integrações teve um menu próprio na barra lateral (§
// Integrações como menu próprio) — deixou de ser uma aba daqui porque
// virou grande o bastante pra merecer navegação de primeiro nível (lista
// + página de detalhe por integração). Mesmo padrão de estado ativo por
// pathname já usado em Sidebar.tsx, só que como uma tira de abas
// horizontal em vez de itens verticais.
const tabs = [
  { href: "/configuracao", label: "Sistema & Identidade" },
  { href: "/configuracao/keycloak", label: "Keycloak (IAM)" },
  { href: "/configuracao/modulos", label: "Módulos do Sistema" },
  { href: "/configuracao/localidades", label: "Localidades & Setores" },
  { href: "/configuracao/perfis", label: "Perfis de Acesso" },
  { href: "/configuracao/mapeamento-ad", label: "Mapeamento Active Directory" },
  { href: "/configuracao/usuarios", label: "Usuários & Lotações" },
];

export default function ConfiguracaoLayout({ children }: { children: ReactNode }) {
  const pathname = usePathname();

  return (
    <div className="flex flex-col gap-6">
      <div>
        <p className="dateline">Administração do sistema</p>
        <h1 className="mt-2 text-2xl font-semibold">Configurações</h1>
        <p className="mt-1 text-sm text-muted">
          Identidade visual, integração Keycloak (IAM), módulos ativos, unidades e controle de acesso.
        </p>
      </div>

      <nav aria-label="Configurações" className="border-b border-surface-border">
        <ul className="-mb-px flex gap-4 overflow-x-auto pb-1">
          {tabs.map((tab) => {
            const active =
              pathname === tab.href || (tab.href !== "/configuracao" && pathname.startsWith(tab.href + "/"));
            return (
              <li key={tab.href} className="shrink-0">
                <Link
                  href={tab.href}
                  aria-current={active ? "page" : undefined}
                  className={`inline-block border-b-2 px-1 pb-3 text-sm font-medium transition-colors whitespace-nowrap
                    ${active ? "border-primary text-primary font-semibold" : "border-transparent text-muted hover:text-foreground"}`}
                >
                  {tab.label}
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>

      {children}
    </div>
  );
}
