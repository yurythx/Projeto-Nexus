"use client";

import { ArrowLeft, Printer } from "lucide-react";
import Link from "next/link";
import { useParams, useSearchParams } from "next/navigation";
import { Suspense } from "react";

import { FichaValidacao } from "@/components/atlas/FichaValidacao";
import {
  funcaoDaSecretaria,
  porFuncao,
  SITUACOES,
  useFuncoes,
  useProcedimentosDaSecretaria,
} from "@/components/atlas/secretaria";
import { DataState } from "@/components/nexus/DataState";
import { Button } from "@/components/ui/Button";
import { useApiQuery, withQuery } from "@/lib/api/swr";
import { useNexus } from "@/lib/nexus/NexusProvider";
import type { ProcedimentosOrgao, Workflow } from "@/lib/nexus/types";

/** Uma ficha do caderno: carrega o procedimento completo (com as etapas). */
function Ficha({ id }: { id: string }) {
  const detail = useApiQuery<Workflow>(`v1/atlas/admin/workflows/${encodeURIComponent(id)}`);
  return (
    <div className="break-before-page pt-6">
      <DataState
        loading={detail.isLoading}
        error={detail.error}
        onRetry={() => void detail.mutate()}
        empty={!detail.data}
      >
        {detail.data && <FichaValidacao wf={detail.data} aninhada />}
      </DataState>
    </div>
  );
}

import { useRouter } from "next/navigation";

/** Caderno da entrevista (ADR 028): capa e sumário da secretaria e a ficha de
 * validação de cada procedimento, uma por página, para imprimir de uma vez.
 * Suporta filtro de departamento único ou multi-funções para reuniões setoriais. */
function Caderno() {
  const router = useRouter();
  const { prefixo } = useParams<{ prefixo: string }>();
  const params = useSearchParams();
  const situacao = params.get("situacao") ?? "";

  // Suporte a multi-funções (múltiplos ?funcao= ou ?funcoes= separated by comma)
  const rawFuncaoParams = params.getAll("funcao");
  const rawCommaFuncoes = params.get("funcoes")?.split(",").map((s) => s.trim()) ?? [];
  const rawList = Array.from(new Set([...rawFuncaoParams, ...rawCommaFuncoes])).filter(Boolean);

  const funcoesSecretaria = useFuncoes(prefixo);
  const funcoesValidas = rawList
    .map((f) => funcaoDaSecretaria(prefixo, f))
    .filter(Boolean);

  const isMulti = funcoesValidas.length > 1;
  const funcaoUnica = funcoesValidas.length === 1 ? funcoesValidas[0]! : "";

  // Se for função única, usa na busca da API. Se for multi, busca pela secretaria e filtra em memória.
  const lista = useProcedimentosDaSecretaria(
    prefixo,
    true,
    situacao,
    isMulti ? "" : funcaoUnica,
  );

  const rawItems = lista.data?.items ?? [];
  const items = isMulti
    ? rawItems.filter((wf) => {
        const cod = wf.classificacao?.subfuncao?.funcao?.codigo;
        return cod && funcoesValidas.includes(cod);
      })
    : rawItems;

  const grupos = porFuncao(items);
  const filtro = SITUACOES.find((s) => s.value === situacao && s.value);
  const resumo = useApiQuery<ProcedimentosOrgao[]>("v1/atlas/admin/workflows/secretarias");
  const orgao = resumo.data?.find((o) => o.prefixo === prefixo);

  const nomeFuncaoUnica = funcoesSecretaria.find((f) => f.codigo === funcaoUnica)?.nome;

  function toggleFuncao(cod: string) {
    let next: string[];
    if (funcoesValidas.includes(cod)) {
      next = funcoesValidas.filter((c) => c !== cod);
    } else {
      next = [...funcoesValidas, cod];
    }
    const query = new URLSearchParams();
    if (situacao) query.set("situacao", situacao);
    for (const f of next) {
      query.append("funcao", f);
    }
    router.replace(
      `/atlas/secretarias/${encodeURIComponent(prefixo)}/caderno?${query.toString()}`,
    );
  }

  function selecionarTodas() {
    const query = new URLSearchParams();
    if (situacao) query.set("situacao", situacao);
    router.replace(
      `/atlas/secretarias/${encodeURIComponent(prefixo)}/caderno?${query.toString()}`,
    );
  }

  return (
    <div className="mx-auto flex max-w-4xl flex-col gap-5 text-foreground">
      <div className="flex flex-wrap items-center justify-between gap-2 print:hidden">
        <Link
          href={withQuery(`/atlas/secretarias/${encodeURIComponent(prefixo)}`, {
            situacao,
            funcao: isMulti ? undefined : funcaoUnica,
          })}
          className="inline-flex items-center gap-1 text-sm text-muted hover:text-foreground"
        >
          <ArrowLeft size={14} aria-hidden="true" /> Voltar à secretaria
        </Link>
        <Button size="sm" onClick={() => window.print()} disabled={items.length === 0}>
          <Printer size={14} aria-hidden="true" className="mr-1" /> Imprimir caderno
        </Button>
      </div>

      {/* Painel Interativo de Multi-Seleção de Funções (Visível apenas em tela) */}
      {funcoesSecretaria.length > 0 && (
        <div className="flex flex-col gap-2 rounded-xl border border-surface-border bg-surface-raised p-3 text-xs print:hidden">
          <div className="flex items-center justify-between">
            <span className="font-medium text-foreground">
              Departamentos da Reunião (Filtro Multi-função):
            </span>
            <button
              type="button"
              onClick={selecionarTodas}
              className="text-primary hover:underline"
            >
              Ver toda a secretaria
            </button>
          </div>
          <div className="flex flex-wrap gap-1.5">
            {funcoesSecretaria.map((f) => {
              const selected = funcoesValidas.includes(f.codigo);
              return (
                <button
                  key={f.codigo}
                  type="button"
                  onClick={() => toggleFuncao(f.codigo)}
                  className={`rounded-lg px-2.5 py-1 text-xs transition-colors ${
                    selected
                      ? "bg-primary text-primary-foreground font-medium"
                      : "border border-surface-border bg-surface text-muted hover:text-foreground"
                  }`}
                >
                  <span className="font-mono text-[10px] mr-1">{f.codigo}</span>
                  {f.nome}
                </button>
              );
            })}
          </div>
        </div>
      )}

      <DataState
        loading={lista.isLoading}
        error={lista.error}
        onRetry={() => void lista.mutate()}
        empty={items.length === 0}
        emptyTitle="Nenhum procedimento para o caderno"
      >
        <header className="flex flex-col gap-2 border-b-2 border-foreground pb-3">
          <p className="text-xs uppercase tracking-wide">Caderno de validação dos procedimentos</p>
          <h1 className="text-2xl font-bold">{orgao?.nome ?? `Órgão ${prefixo}`}</h1>
          {funcaoUnica && (
            <p className="text-lg font-semibold">
              Função {funcaoUnica}
              {nomeFuncaoUnica && ` — ${nomeFuncaoUnica}`}
            </p>
          )}
          {isMulti && (
            <div className="text-base font-semibold">
              <span className="block text-xs uppercase text-muted">Funções Selecionadas ({funcoesValidas.length})</span>
              <ul className="flex flex-wrap gap-2 pt-1 font-normal text-sm">
                {funcoesValidas.map((cod) => {
                  const nome = funcoesSecretaria.find((f) => f.codigo === cod)?.nome;
                  return (
                    <li key={cod} className="rounded bg-surface-subtle px-2 py-0.5 border border-surface-border">
                      <strong className="font-mono">{cod}</strong> {nome ? `— ${nome}` : ""}
                    </li>
                  );
                })}
              </ul>
            </div>
          )}
          <p className="text-sm">
            TTDD {prefixo} · {items.length} {items.length === 1 ? "procedimento" : "procedimentos"}
            {filtro && ` · situação: ${filtro.label.toLowerCase()}`}
          </p>
          <p className="text-sm">Data da entrevista: ____/____/________</p>
        </header>
        <section aria-labelledby="caderno-sumario" className="flex flex-col gap-2">
          <h2 id="caderno-sumario" className="font-semibold">
            Sumário
          </h2>
          {grupos.map((g) => (
            <div key={g.codigo || "outros"} className="text-sm">
              <p className="font-medium">
                {g.codigo && <span className="mr-1 font-mono">{g.codigo}</span>}
                {g.nome}
              </p>
              <ul className="ml-4 list-disc">
                {g.itens.map((wf) => (
                  <li key={wf.id}>
                    <span className="font-mono">{wf.codigo_processual}</span> — {wf.titulo}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </section>
        {grupos
          .flatMap((g) => g.itens)
          .map((wf) => (
            <Ficha key={wf.id} id={wf.id} />
          ))}
      </DataState>
    </div>
  );
}

export default function CadernoPage() {
  const { can } = useNexus();
  if (!can("atlas:manage")) {
    return (
      <p className="text-sm text-muted">
        O caderno da entrevista é da gestão do Atlas (permissão atlas:manage).
      </p>
    );
  }
  return (
    <Suspense>
      <Caderno />
    </Suspense>
  );
}
