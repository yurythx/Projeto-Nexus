"use client";

import { createContext, useCallback, useContext, useMemo, type ReactNode } from "react";

import { useApiQuery } from "@/lib/api/swr";
import { hasPermission } from "@/lib/nexus/permissions";
import type { Me, ModuleStatus } from "@/lib/nexus/types";

interface NexusContextValue {
  me: Me | undefined;
  modules: ModuleStatus[];
  loading: boolean;
  /** Permissão efetiva ("recurso:ação", com curingas) — só para exibição. */
  can: (permission: string) => boolean;
  /** Módulo ativo no Kernel neste instante. */
  enabled: (key: string) => boolean;
  refreshModules: () => void;
}

const NexusContext = createContext<NexusContextValue | null>(null);

/** Identidade efetiva (GET /me, resolvida pelo IAM) e estado dos módulos
 * (GET /system/modules). Os módulos revalidam a cada 30s: um plugin
 * desativado por outro administrador some do menu sem recarregar. */
export function NexusProvider({ children }: { children: ReactNode }) {
  const me = useApiQuery<Me>("v1/me", { revalidateOnFocus: true });
  const modules = useApiQuery<ModuleStatus[]>("v1/system/modules", { refreshInterval: 30_000 });

  const can = useCallback((p: string) => hasPermission(me.data?.permissions, p), [me.data]);
  const enabled = useCallback(
    (key: string) => Boolean(modules.data?.find((m) => m.key === key)?.enabled),
    [modules.data],
  );

  const value = useMemo<NexusContextValue>(
    () => ({
      me: me.data,
      modules: modules.data ?? [],
      loading: me.isLoading || modules.isLoading,
      can,
      enabled,
      refreshModules: () => void modules.mutate(),
    }),
    [me.data, me.isLoading, modules, can, enabled],
  );
  return <NexusContext.Provider value={value}>{children}</NexusContext.Provider>;
}

export function useNexus(): NexusContextValue {
  const ctx = useContext(NexusContext);
  if (!ctx) {
    // Fora do provider (testes/telas públicas): nada permitido, nada ativo.
    return { me: undefined, modules: [], loading: false, can: () => false, enabled: () => false, refreshModules: () => {} };
  }
  return ctx;
}
