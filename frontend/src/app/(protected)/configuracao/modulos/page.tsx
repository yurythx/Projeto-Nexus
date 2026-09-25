"use client";

import { Lock } from "lucide-react";

import { ModuleIcon } from "@/components/layout/ModuleIcon";
import { DataState } from "@/components/nexus/DataState";
import { useAction } from "@/components/nexus/useAction";
import { Badge } from "@/components/ui/Badge";
import { Card } from "@/components/ui/Card";
import { Toggle } from "@/components/ui/Toggle";
import { apiClient } from "@/lib/api/client";
import { useNexus } from "@/lib/nexus/NexusProvider";
import type { ModuleStatus } from "@/lib/nexus/types";

/**
 * Ciclo de vida dos plug-ins (Kernel): ativar/desativar em runtime, sem
 * recompilar nem reiniciar. Um módulo inativo desliga rotas HTTP, workers,
 * consumidores de fila e tópicos WebSocket; o núcleo (IAM, Auditoria)
 * nunca é desativável. Dependências são validadas pelo backend (409).
 */
export default function ModulosPage() {
  const { modules, loading, can, refreshModules } = useNexus();
  const { run, pending } = useAction();
  const canManage = can("modules:manage");
  const byKey = new Map(modules.map((m) => [m.key, m]));

  async function toggle(m: ModuleStatus, enabled: boolean) {
    await run(
      () => apiClient.patch<ModuleStatus>(`v1/admin/modules/${m.key}`, { enabled }),
      enabled ? `${m.name} ativado` : `${m.name} desativado`,
    );
    refreshModules();
  }

  return (
    <DataState loading={loading && modules.length === 0} error={null} empty={modules.length === 0} emptyTitle="Nenhum módulo registrado">
      <ul className="grid gap-4 lg:grid-cols-2">
        {modules.map((m) => {
          const dependents = modules.filter((o) => o.enabled && o.depends_on.includes(m.key));
          return (
            <li key={m.key}>
              <Card className="flex h-full flex-col gap-3 p-5">
                <div className="flex items-start gap-3">
                  <span className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
                    <ModuleIcon name={m.icon} size={20} aria-hidden="true" />
                  </span>
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <h2 className="text-base font-semibold">{m.name}</h2>
                      <code className="text-xs text-muted">{m.key}</code>
                      {m.core && <Badge tone="info">Núcleo</Badge>}
                      {m.public && <Badge>Superfície pública</Badge>}
                    </div>
                    <p className="mt-1 text-sm text-muted">{m.description}</p>
                  </div>
                  {m.core ? (
                    <span className="inline-flex items-center gap-1 text-xs text-muted" title="Módulo do núcleo — nunca desativável">
                      <Lock size={14} aria-hidden="true" /> Sempre ativo
                    </span>
                  ) : (
                    <Toggle
                      label={`${m.enabled ? "Desativar" : "Ativar"} ${m.name}`}
                      checked={m.enabled}
                      disabled={!canManage || pending}
                      onChange={(v) => void toggle(m, v)}
                    />
                  )}
                </div>
                {(m.depends_on.length > 0 || dependents.length > 0) && (
                  <dl className="flex flex-col gap-1 text-xs text-muted">
                    {m.depends_on.length > 0 && (
                      <div className="flex flex-wrap gap-1">
                        <dt className="font-medium">Depende de:</dt>
                        {m.depends_on.map((d) => (
                          <dd key={d}>
                            <Badge tone={byKey.get(d)?.enabled ? "success" : "warning"}>{byKey.get(d)?.name ?? d}</Badge>
                          </dd>
                        ))}
                      </div>
                    )}
                    {dependents.length > 0 && (
                      <div className="flex flex-wrap gap-1">
                        <dt className="font-medium">Usado por:</dt>
                        {dependents.map((d) => (
                          <dd key={d.key}>
                            <Badge>{d.name}</Badge>
                          </dd>
                        ))}
                      </div>
                    )}
                  </dl>
                )}
                {m.permissions.length > 0 && (
                  <details className="text-xs">
                    <summary className="cursor-pointer text-muted hover:text-foreground">{m.permissions.length} permissões declaradas</summary>
                    <ul className="mt-2 flex flex-col gap-1">
                      {m.permissions.map((p) => (
                        <li key={p.key}>
                          <code className="font-mono">{p.key}</code> — <span className="text-muted">{p.description}</span>
                        </li>
                      ))}
                    </ul>
                  </details>
                )}
              </Card>
            </li>
          );
        })}
      </ul>
    </DataState>
  );
}
