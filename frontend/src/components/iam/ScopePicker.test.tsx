import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";

import type { OrgTree, ScopeRef } from "@/lib/nexus/types";

const tree: OrgTree[] = [
  {
    id: "e1", nome: "Secretaria", sigla: "SEC", slug: "sec", documento: "", ativo: true,
    unidades: [
      { id: "u1", entidade_id: "e1", nome: "Protocolo", sigla: "PRT", slug: "", ad_group: "", email: "", telefone: "", endereco: "", ativo: true,
        departamentos: [{ id: "d1", unidade_id: "u1", nome: "Triagem", sigla: "", slug: "", ad_group: "", email: "", telefone: "", ativo: true }] },
    ],
  },
];
vi.mock("@/lib/api/swr", () => ({ useApiQuery: () => ({ data: tree }) }));

import { ScopePicker } from "./ScopePicker";

function Harness({ onChange }: { onChange: (v: ScopeRef) => void }) {
  const [v, setV] = useState<ScopeRef>({});
  return (
    <ScopePicker
      idPrefix="t"
      value={v}
      onChange={(n) => {
        setV(n);
        onChange(n);
      }}
    />
  );
}

describe("ScopePicker", () => {
  it("cascata: unidade e departamento só habilitam após o nível acima; trocar o nível acima limpa os de baixo", async () => {
    const onChange = vi.fn();
    render(<Harness onChange={onChange} />);
    expect(screen.getByLabelText("Unidade")).toBeDisabled();
    expect(screen.getByLabelText("Departamento")).toBeDisabled();

    await userEvent.selectOptions(screen.getByLabelText("Entidade"), "e1");
    await userEvent.selectOptions(screen.getByLabelText("Unidade"), "u1");
    await userEvent.selectOptions(screen.getByLabelText("Departamento"), "d1");
    expect(onChange).toHaveBeenLastCalledWith({ entidade_id: "e1", unidade_id: "u1", departamento_id: "d1" });

    await userEvent.selectOptions(screen.getByLabelText("Entidade"), "");
    expect(onChange).toHaveBeenLastCalledWith({ entidade_id: undefined });
    expect(screen.getByLabelText("Unidade")).toBeDisabled();
  });
});
