"use client";

import Link from "next/link";
import { useSession } from "next-auth/react";
import { Suspense, useEffect, useState, type ReactNode } from "react";

import { AuthFlashToast } from "@/components/layout/AuthFlashToast";
import { GovFooter } from "@/components/layout/GovFooter";
import { GovHeader } from "@/components/layout/GovHeader";
import { LGPDConsentModal } from "@/components/layout/LGPDConsentModal";
import { ToastProvider } from "@/components/notifications/ToastProvider";
import { Button } from "@/components/ui/Button";
import { ThemeToggle } from "@/components/ui/ThemeToggle";
import { publicFetch } from "@/lib/api/publicClient";
import type { PublicModule } from "@/lib/nexus/types";

const PUBLIC_LINKS: { href: string; label: string; module?: string }[] = [
  { href: "/servicos", label: "Serviços", module: "catalog" },
  { href: "/setores", label: "Setores", module: "directory" },
  { href: "/eventos", label: "Agenda", module: "calendar" },
  { href: "/contato", label: "Contato", module: "contact" },
  { href: "/sobre", label: "Sobre" },
];

/** Shell do site institucional (visitante anônimo). Os links de plugins
 * desativados no Kernel somem do menu. */
export function PublicShell({ children }: { children: ReactNode }) {
  const { status } = useSession();
  const [enabled, setEnabled] = useState<Record<string, boolean> | null>(null);

  useEffect(() => {
    publicFetch<PublicModule[]>("v1/system/public-modules")
      .then((mods) => setEnabled(Object.fromEntries(mods.map((m) => [m.key, m.enabled]))))
      .catch(() => setEnabled({}));
  }, []);

  const links = PUBLIC_LINKS.filter((l) => !l.module || enabled === null || enabled[l.module]);

  return (
    <ToastProvider>
      <div className="flex min-h-screen flex-col pt-[var(--topbar-h)]">
        <Suspense fallback={null}>
          <AuthFlashToast />
        </Suspense>
        <GovHeader
          homeHref="/"
          nav={
            <nav id="main-menu" tabIndex={-1} aria-label="Menu principal" className="ml-4 hidden outline-none lg:block">
              <ul className="flex items-center gap-5 text-sm">
                {links.map((l) => (
                  <li key={l.href}>
                    <Link href={l.href} className="text-muted transition-colors hover:text-foreground">
                      {l.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </nav>
          }
          actions={
            <>
              <ThemeToggle />
              <Link href={status === "authenticated" ? "/dashboard" : "/login"}>
                <Button size="sm">{status === "authenticated" ? "Acessar o painel" : "Entrar"}</Button>
              </Link>
            </>
          }
        />
        <main id="main-content" tabIndex={-1} className="flex-1 outline-none">
          {children}
        </main>
        <GovFooter />
        <LGPDConsentModal />
      </div>
    </ToastProvider>
  );
}
