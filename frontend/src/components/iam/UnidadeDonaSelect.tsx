"use client";

import { Select } from "@/components/ui/Select";
import { useApiQuery } from "@/lib/api/swr";
import { useNexus } from "@/lib/nexus/NexusProvider";
import { unidadesGeridas } from "@/lib/nexus/permissions";
import type { OrgTree } from "@/lib/nexus/types";

/** Unidade dona de um conteúdo (ADR 013): oferece só as unidades em que a
 * pessoa tem a permissão de gestão; "institucional" (sem dono) só para a
 * gestão global — para os demais o campo é obrigatório. A API confere de
 * novo. */
export function UnidadeDonaSelect({
  id,
  name = "unidade_id",
  permission,
  defaultValue,
  label = "Unidade dona",
}: {
  id: string;
  name?: string;
  permission: string;
  defaultValue?: string | null;
  label?: string;
}) {
  const { me } = useNexus();
  const tree = useApiQuery<OrgTree[]>("v1/iam/org-tree");
  const geridas = unidadesGeridas(me, tree.data, permission);
  const options = (tree.data ?? []).flatMap((e) =>
    e.unidades.filter((u) => geridas.unidades.has(u.id)).map((u) => ({ value: u.id, label: u.sigla ? `${u.sigla} — ${u.nome}` : u.nome })),
  );
  return (
    <Select
      id={id}
      name={name}
      label={geridas.global ? label : `${label} *`}
      required={!geridas.global}
      placeholder={geridas.global ? "Institucional (sem unidade)" : "Selecione…"}
      defaultValue={defaultValue ?? ""}
      options={options}
    />
  );
}
