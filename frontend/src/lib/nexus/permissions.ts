import type { Me, OrgTree } from "@/lib/nexus/types";

// Mesma regra de curingas do backend (internal/platform/auth/rbac.go):
// "*" concede tudo, "recurso:*" concede toda ação do recurso. No frontend
// isto só decide o que MOSTRAR — a autorização efetiva é sempre do
// backend (middleware RequirePermission, A01).
export function hasPermission(granted: readonly string[] | undefined, want: string): boolean {
  if (!granted) return false;
  const [resource] = want.split(":");
  return granted.some((g) => g === "*" || g === want || g === `${resource}:*`);
}

/** Onde uma permissão com escopo vale (ADR 013) — espelho de
 * auth.Scope.Covers no backend, só para decidir o que OFERECER nos
 * formulários (a API confere de novo):
 *   - concessão global (ou o papel nexus-admin): todas as unidades e o
 *     "institucional" (sem dono);
 *   - entidade: todas as unidades dela;
 *   - unidade: ela e as subunidades (árvore parent_id);
 *   - departamento: nenhuma unidade inteira. */
export function unidadesGeridas(
  me: Pick<Me, "roles" | "scopes"> | undefined,
  tree: readonly OrgTree[] | undefined,
  permission: string,
): { global: boolean; unidades: Set<string> } {
  const all = (tree ?? []).flatMap((e) => e.unidades);
  if (me?.roles?.includes("nexus-admin")) return { global: true, unidades: new Set(all.map((u) => u.id)) };
  const filhas = new Map<string, string[]>();
  for (const u of all) if (u.parent_id) filhas.set(u.parent_id, [...(filhas.get(u.parent_id) ?? []), u.id]);
  const unidades = new Set<string>();
  const arvore = (id: string) => {
    if (unidades.has(id)) return;
    unidades.add(id);
    for (const f of filhas.get(id) ?? []) arvore(f);
  };
  let global = false;
  for (const s of me?.scopes ?? []) {
    if (!hasPermission(s.permissions, permission)) continue;
    if (!s.entidade_id && !s.unidade_id && !s.departamento_id) global = true;
    else if (s.departamento_id) continue;
    else if (s.unidade_id) arvore(s.unidade_id);
    else for (const u of all.filter((x) => x.entidade_id === s.entidade_id)) unidades.add(u.id);
  }
  return { global, unidades: global ? new Set(all.map((u) => u.id)) : unidades };
}
