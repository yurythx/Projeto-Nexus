"use client";

import { useMemo } from "react";

import { Select } from "@/components/ui/Select";
import { useApiQuery } from "@/lib/api/swr";
import type { OrgTree, ScopeRef } from "@/lib/nexus/types";

/** Escopo organizacional em cascata (Entidade → Unidade → Departamento)
 * a partir de GET /iam/org-tree. Campo vazio = escopo mais amplo. */
export function ScopePicker({
  value,
  onChange,
  idPrefix,
  required = false,
}: {
  value: ScopeRef;
  onChange: (v: ScopeRef) => void;
  idPrefix: string;
  /** Exige ao menos a Entidade. */
  required?: boolean;
}) {
  const { data: tree } = useApiQuery<OrgTree[]>("v1/iam/org-tree");

  const entidade = useMemo(() => tree?.find((e) => e.id === value.entidade_id), [tree, value.entidade_id]);
  const unidade = useMemo(() => entidade?.unidades.find((u) => u.id === value.unidade_id), [entidade, value.unidade_id]);

  return (
    <fieldset className="grid gap-3 sm:grid-cols-3">
      <legend className="sr-only">Escopo organizacional</legend>
      <Select
        id={`${idPrefix}-entidade`}
        label={required ? "Entidade *" : "Entidade"}
        placeholder={required ? "Selecione…" : "Todas (global)"}
        required={required}
        value={value.entidade_id ?? ""}
        onChange={(e) => onChange({ entidade_id: e.target.value || undefined })}
        options={(tree ?? []).map((e) => ({ value: e.id, label: e.sigla ? `${e.sigla} — ${e.nome}` : e.nome }))}
      />
      <Select
        id={`${idPrefix}-unidade`}
        label="Unidade"
        placeholder="Todas da entidade"
        disabled={!entidade}
        value={value.unidade_id ?? ""}
        onChange={(e) => onChange({ entidade_id: value.entidade_id, unidade_id: e.target.value || undefined })}
        options={(entidade?.unidades ?? []).map((u) => ({ value: u.id, label: u.sigla ? `${u.sigla} — ${u.nome}` : u.nome }))}
      />
      <Select
        id={`${idPrefix}-departamento`}
        label="Departamento"
        placeholder="Todos da unidade"
        disabled={!unidade}
        value={value.departamento_id ?? ""}
        onChange={(e) =>
          onChange({ entidade_id: value.entidade_id, unidade_id: value.unidade_id, departamento_id: e.target.value || undefined })
        }
        options={(unidade?.departamentos ?? []).map((d) => ({ value: d.id, label: d.sigla ? `${d.sigla} — ${d.nome}` : d.nome }))}
      />
    </fieldset>
  );
}
