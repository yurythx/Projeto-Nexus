"use client";

import { useState } from "react";
import { Building, Check, ChevronRight, Edit3, Shield, UserCheck, X } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { ModalShell } from "@/components/ui/ModalShell";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "@/components/ui/Table";
import { apiClient, ApiError } from "@/lib/api/client";
import type { Localidade, Perfil, User, UsuarioLotacao } from "@/types/api";

interface UsuariosLotacaoTableProps {
  users: User[];
  allPerfis: Perfil[];
  allLocalidades: Localidade[];
}

export function UsuariosLotacaoTable({ users, allPerfis, allLocalidades }: UsuariosLotacaoTableProps) {
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [loadingLotacao, setLoadingLotacao] = useState(false);
  const [savingLotacao, setSavingLotacao] = useState(false);
  const [feedback, setFeedback] = useState<{ type: "success" | "error"; text: string } | null>(null);

  // Seleções do modal
  const [selectedPerfilIds, setSelectedPerfilIds] = useState<string[]>([]);
  const [selectedLocalidadeIds, setSelectedLocalidadeIds] = useState<string[]>([]);
  const [selectedSetorIds, setSelectedSetorIds] = useState<string[]>([]);

  async function openLotacaoModal(user: User) {
    setSelectedUser(user);
    setLoadingLotacao(true);
    setFeedback(null);
    try {
      const res = await apiClient.get<UsuarioLotacao>(`v1/organizacao/usuarios/${user.id}/lotacao`);
      const lot = res.data;
      setSelectedPerfilIds(lot?.perfis?.map((p) => p.id) || []);
      setSelectedLocalidadeIds(lot?.localidades?.map((l) => l.id) || []);
      setSelectedSetorIds(lot?.setores?.map((s) => s.id) || []);
    } catch (err) {
      console.error("Erro ao carregar lotação:", err);
      setSelectedPerfilIds([]);
      setSelectedLocalidadeIds([]);
      setSelectedSetorIds([]);
    } finally {
      setLoadingLotacao(false);
    }
  }

  function toggleLocalidade(locId: string) {
    setSelectedLocalidadeIds((prev) => {
      const isSelected = prev.includes(locId);
      if (isSelected) {
        // Remover localidade e também remover os setores dessa localidade
        const loc = allLocalidades.find((l) => l.id === locId);
        const locSetorIds = new Set(loc?.setores?.map((s) => s.id) || []);
        setSelectedSetorIds((sPrev) => sPrev.filter((sid) => !locSetorIds.has(sid)));
        return prev.filter((id) => id !== locId);
      } else {
        return [...prev, locId];
      }
    });
  }

  function toggleSetor(setorId: string) {
    setSelectedSetorIds((prev) =>
      prev.includes(setorId) ? prev.filter((id) => id !== setorId) : [...prev, setorId]
    );
  }

  function togglePerfil(perfilId: string) {
    setSelectedPerfilIds((prev) =>
      prev.includes(perfilId) ? prev.filter((id) => id !== perfilId) : [...prev, perfilId]
    );
  }

  async function handleSaveLotacao(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedUser) return;
    setSavingLotacao(true);
    setFeedback(null);
    try {
      await apiClient.put(`v1/organizacao/usuarios/${selectedUser.id}/lotacao`, {
        perfil_ids: selectedPerfilIds,
        localidade_ids: selectedLocalidadeIds,
        setor_ids: selectedSetorIds,
      });
      setFeedback({ type: "success", text: "Lotação e perfis do servidor atualizados com sucesso!" });
      setTimeout(() => {
        setSelectedUser(null);
        setFeedback(null);
      }, 1200);
    } catch (err: any) {
      const msg = err instanceof ApiError ? err.message : "Erro ao salvar lotação do usuário.";
      setFeedback({ type: "error", text: msg });
    } finally {
      setSavingLotacao(false);
    }
  }

  return (
    <>
      <Table caption="Usuários do sistema com status de acesso e gerenciamento de perfis e unidades.">
        <TableHead>
          <TableRow>
            <TableHeaderCell>Nome do Servidor</TableHeaderCell>
            <TableHeaderCell>E-mail Institucional</TableHeaderCell>
            <TableHeaderCell>Status</TableHeaderCell>
            <TableHeaderCell>Último Acesso</TableHeaderCell>
            <TableHeaderCell className="text-right">Ação de Lotação</TableHeaderCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {users.map((user) => (
            <TableRow key={user.id}>
              <TableCell className="font-medium text-foreground">
                {user.display_name || user.username}
                <div className="text-xs text-muted font-mono">{user.username}</div>
              </TableCell>
              <TableCell>{user.email || "—"}</TableCell>
              <TableCell>
                {user.active ? (
                  <span className="rounded-full bg-emerald-500/10 px-2 py-0.5 text-xs font-semibold text-emerald-600 dark:text-emerald-400">
                    Ativo
                  </span>
                ) : (
                  <span className="rounded-full bg-danger/10 px-2 py-0.5 text-xs font-semibold text-danger">
                    Inativo
                  </span>
                )}
              </TableCell>
              <TableCell className="text-xs text-muted">
                {user.last_seen_at ? new Date(user.last_seen_at).toLocaleString("pt-BR") : "—"}
              </TableCell>
              <TableCell className="text-right">
                <Button
                  size="sm"
                  variant="secondary"
                  onClick={() => openLotacaoModal(user)}
                  className="flex items-center gap-1.5 text-xs ml-auto"
                >
                  <Building size={14} className="text-primary" />
                  Perfil & Lotação
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>

      {/* MODAL DE VÍNCULO DE LOTAÇÃO & PERFIL */}
      {selectedUser && (
        <ModalShell
          open={!!selectedUser}
          size="lg"
          onClose={() => setSelectedUser(null)}
        >
          <div className="flex flex-col gap-4 p-5">
            <h2 className="text-lg font-semibold text-foreground">
              Lotação e Perfis: {selectedUser.display_name || selectedUser.username}
            </h2>
          {loadingLotacao ? (
            <div className="p-8 text-center text-sm text-muted">Carregando lotações e vínculos do servidor...</div>
          ) : (
            <>
              {feedback && (
                <div
                  className={`rounded-lg p-3 text-xs font-medium ${
                    feedback.type === "success"
                      ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20"
                      : "bg-danger/10 text-danger border border-danger/20"
                  }`}
                >
                  {feedback.text}
                </div>
              )}

              <form onSubmit={handleSaveLotacao} className="flex flex-col gap-6 max-h-[75vh] overflow-y-auto pr-1">
              {/* 1. SELEÇÃO DE PERFIL */}
              <div>
                <label className="text-xs font-bold uppercase tracking-wider text-muted flex items-center gap-1.5">
                  <Shield size={14} className="text-primary" />
                  1. Perfis de Acesso no SUAS (RBAC / Active Directory)
                </label>
                <p className="text-xs text-muted mt-0.5 mb-2.5">
                  Selecione o perfil que define o nível de atuação do servidor na plataforma:
                </p>

                <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                  {allPerfis.map((p) => {
                    const isChecked = selectedPerfilIds.includes(p.id);
                    return (
                      <button
                        type="button"
                        key={p.id}
                        onClick={() => togglePerfil(p.id)}
                        className={`flex items-start gap-2.5 rounded-lg border p-3 text-left transition-all ${
                          isChecked
                            ? "border-primary bg-primary/5 dark:bg-primary/10"
                            : "border-surface-border bg-surface hover:bg-black/5 dark:hover:bg-white/5"
                        }`}
                      >
                        <div
                          className={`mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded border ${
                            isChecked ? "border-primary bg-primary text-white" : "border-surface-border"
                          }`}
                        >
                          {isChecked && <Check size={12} />}
                        </div>
                        <div>
                          <div className="text-sm font-semibold text-foreground">{p.nome}</div>
                          <div className="text-[11px] text-muted">{p.descricao}</div>
                          <div className="mt-1 font-mono text-[10px] text-primary/80">AD: {p.grupo_ad}</div>
                        </div>
                      </button>
                    );
                  })}
                </div>
              </div>

              {/* 2. SELEÇÃO DE LOCALIDADES (MÚLTIPLA SELEÇÃO) */}
              <div className="border-t border-surface-border pt-4">
                <label className="text-xs font-bold uppercase tracking-wider text-muted flex items-center gap-1.5">
                  <Building size={14} className="text-primary" />
                  2. Localidades Autorizadas ({selectedLocalidadeIds.length} selecionadas)
                </label>
                <p className="text-xs text-muted mt-0.5 mb-2.5">
                  Marque as unidades socioassistenciais em que o servidor tem permissão de acesso e atuação:
                </p>

                <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                  {allLocalidades.map((loc) => {
                    const isChecked = selectedLocalidadeIds.includes(loc.id);
                    return (
                      <button
                        type="button"
                        key={loc.id}
                        onClick={() => toggleLocalidade(loc.id)}
                        className={`flex items-start gap-2.5 rounded-lg border p-3 text-left transition-all ${
                          isChecked
                            ? "border-primary bg-primary/5 dark:bg-primary/10"
                            : "border-surface-border bg-surface hover:bg-black/5 dark:hover:bg-white/5"
                        }`}
                      >
                        <div
                          className={`mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded border ${
                            isChecked ? "border-primary bg-primary text-white" : "border-surface-border"
                          }`}
                        >
                          {isChecked && <Check size={12} />}
                        </div>
                        <div className="flex-1">
                          <div className="flex items-center justify-between">
                            <span className="text-sm font-semibold text-foreground">{loc.nome}</span>
                            <span className="rounded bg-surface-border/50 px-1.5 py-0.2 text-[10px] font-mono text-muted">
                              {loc.tipo}
                            </span>
                          </div>
                          <div className="text-[11px] text-muted">{loc.bairro || loc.slug}</div>
                          <div className="mt-0.5 font-mono text-[10px] text-primary/70">AD: {loc.grupo_ad}</div>
                        </div>
                      </button>
                    );
                  })}
                </div>
              </div>

              {/* 3. SELEÇÃO DE SETORES DAS LOCALIDADES SELECIONADAS */}
              {selectedLocalidadeIds.length > 0 && (
                <div className="border-t border-surface-border pt-4">
                  <label className="text-xs font-bold uppercase tracking-wider text-muted flex items-center gap-1.5">
                    <UserCheck size={14} className="text-primary" />
                    3. Setores / Departamentos Vinculados
                  </label>
                  <p className="text-xs text-muted mt-0.5 mb-2.5">
                    Defina em quais setores das unidades selecionadas o servidor está alocado (Técnico, Recepção, Gerência):
                  </p>

                  <div className="flex flex-col gap-3">
                    {allLocalidades
                      .filter((loc) => selectedLocalidadeIds.includes(loc.id))
                      .map((loc) => (
                        <div key={loc.id} className="rounded-lg border border-surface-border bg-surface p-3">
                          <div className="text-xs font-bold text-foreground mb-2 flex items-center gap-1.5">
                            <Building size={12} className="text-primary" />
                            {loc.nome}
                          </div>

                          <div className="grid grid-cols-1 gap-1.5 sm:grid-cols-3">
                            {(loc.setores || []).map((setor) => {
                              const isChecked = selectedSetorIds.includes(setor.id);
                              return (
                                <button
                                  type="button"
                                  key={setor.id}
                                  onClick={() => toggleSetor(setor.id)}
                                  className={`flex items-center gap-2 rounded border p-2 text-left text-xs transition-colors ${
                                    isChecked
                                      ? "border-primary bg-primary/10 text-primary font-semibold"
                                      : "border-surface-border bg-surface text-foreground hover:bg-black/5 dark:hover:bg-white/5"
                                  }`}
                                >
                                  <div
                                    className={`flex h-3.5 w-3.5 shrink-0 items-center justify-center rounded border ${
                                      isChecked ? "border-primary bg-primary text-white" : "border-surface-border"
                                    }`}
                                  >
                                    {isChecked && <Check size={10} />}
                                  </div>
                                  <span className="truncate">{setor.nome.replace(` - ${loc.nome}`, "")}</span>
                                </button>
                              );
                            })}
                          </div>
                        </div>
                      ))}
                  </div>
                </div>
              )}

              <div className="mt-4 flex justify-end gap-2 border-t border-surface-border pt-4">
                <Button type="button" variant="secondary" onClick={() => setSelectedUser(null)}>
                  Cancelar
                </Button>
                <Button type="submit" loading={savingLotacao}>
                  Salvar Lotação e Perfis
                </Button>
              </div>
            </form>
          </>
          )}
        </div>
      </ModalShell>
      )}
    </>
  );
}
