"use client";

import { useState } from "react";
import { Edit3, KeyRound, Plus, Shield, Users } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { ModalShell } from "@/components/ui/ModalShell";
import { apiClient, ApiError } from "@/lib/api/client";
import { useToast } from "@/components/notifications/ToastProvider";
import type { Perfil } from "@/types/api";

interface PerfisManagerProps {
  initialPerfis: Perfil[];
}

export function PerfisManager({ initialPerfis }: PerfisManagerProps) {
  const { showToast } = useToast();
  const [perfis, setPerfis] = useState<Perfil[]>(initialPerfis);

  // Modal: Novo Perfil
  const [isNewModalOpen, setIsNewModalOpen] = useState(false);
  const [nome, setNome] = useState("");
  const [slug, setSlug] = useState("");
  const [descricao, setDescricao] = useState("");
  const [grupoAD, setGrupoAD] = useState("");
  const [nivel, setNivel] = useState<Perfil["nivel"]>("tecnico");
  const [permissoesStr, setPermissoesStr] = useState("");
  const [saving, setSaving] = useState(false);

  // Modal: Editar Perfil
  const [editingPerfil, setEditingPerfil] = useState<Perfil | null>(null);
  const [editNome, setEditNome] = useState("");
  const [editSlug, setEditSlug] = useState("");
  const [editDescricao, setEditDescricao] = useState("");
  const [editGrupoAD, setEditGrupoAD] = useState("");
  const [editNivel, setEditNivel] = useState<Perfil["nivel"]>("tecnico");
  const [editPermissoesStr, setEditPermissoesStr] = useState("");
  const [editAtivo, setEditAtivo] = useState(true);
  const [savingEdit, setSavingEdit] = useState(false);

  async function handleCreatePerfil(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    try {
      const permissoes = permissoesStr
        .split(",")
        .map((p) => p.trim())
        .filter(Boolean);

      const res = await apiClient.post<Perfil>("v1/organizacao/perfis", {
        nome,
        slug,
        descricao,
        grupo_ad: grupoAD,
        nivel,
        permissoes,
      });

      setPerfis((prev) => [...prev, res.data]);
      showToast({
        title: "Perfil criado",
        description: `Perfil ${res.data.nome} cadastrado com sucesso.`,
        tone: "success",
      });
      setIsNewModalOpen(false);
      setNome("");
      setSlug("");
      setDescricao("");
      setGrupoAD("");
      setPermissoesStr("");
    } catch (err: any) {
      showToast({
        title: "Erro ao cadastrar perfil",
        description: err instanceof ApiError ? err.message : "Erro ao cadastrar perfil.",
        tone: "danger",
      });
    } finally {
      setSaving(false);
    }
  }

  function openEditPerfil(p: Perfil) {
    setEditingPerfil(p);
    setEditNome(p.nome);
    setEditSlug(p.slug);
    setEditDescricao(p.descricao || "");
    setEditGrupoAD(p.grupo_ad || "");
    setEditNivel(p.nivel);
    setEditPermissoesStr((p.permissoes || []).join(", "));
    setEditAtivo(p.ativo);
  }

  async function handleUpdatePerfil(e: React.FormEvent) {
    e.preventDefault();
    if (!editingPerfil) return;
    setSavingEdit(true);
    try {
      const permissoes = editPermissoesStr
        .split(",")
        .map((p) => p.trim())
        .filter(Boolean);

      const res = await apiClient.put<Perfil>(`v1/organizacao/perfis/${editingPerfil.id}`, {
        nome: editNome,
        slug: editSlug,
        descricao: editDescricao,
        grupo_ad: editGrupoAD,
        nivel: editNivel,
        permissoes,
        ativo: editAtivo,
      });

      const updated = res.data;
      setPerfis((prev) => prev.map((p) => (p.id === updated.id ? updated : p)));
      showToast({
        title: "Perfil atualizado",
        description: `Perfil ${updated.nome} atualizado com sucesso.`,
        tone: "success",
      });
      setEditingPerfil(null);
    } catch (err: any) {
      showToast({
        title: "Erro ao atualizar perfil",
        description: err instanceof ApiError ? err.message : "Erro ao atualizar perfil.",
        tone: "danger",
      });
    } finally {
      setSavingEdit(false);
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-foreground">Perfis Socioassistenciais (RBAC)</h2>
          <p className="text-xs text-muted mt-0.5">
            Perfis de função definem permissões de sistema e são mapeados diretamente aos grupos de segurança do Active Directory.
          </p>
        </div>

        <Button onClick={() => setIsNewModalOpen(true)} className="flex items-center gap-2">
          <Plus size={16} />
          Novo Perfil
        </Button>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {perfis.map((p) => (
          <div
            key={p.id}
            className="flex flex-col justify-between rounded-lg border border-surface-border bg-surface p-5 shadow-xs transition-shadow hover:shadow-sm"
          >
            <div>
              <div className="flex items-start justify-between gap-3">
                <div className="flex items-center gap-2">
                  <div className="rounded-md bg-primary/10 p-2 text-primary">
                    <Shield size={20} />
                  </div>
                  <div>
                    <h3 className="font-semibold text-foreground text-sm">{p.nome}</h3>
                    <code className="text-[11px] text-muted font-mono">{p.slug}</code>
                  </div>
                </div>

                <div className="flex items-center gap-1.5">
                  <span
                    className={`rounded px-2 py-0.5 text-xs font-semibold uppercase ${
                      p.nivel === "admin"
                        ? "bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-300"
                        : p.nivel === "gestao"
                        ? "bg-purple-100 text-purple-800 dark:bg-purple-900/40 dark:text-purple-300"
                        : p.nivel === "gerencia"
                        ? "bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300"
                        : p.nivel === "tecnico"
                        ? "bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-300"
                        : "bg-surface-border text-foreground"
                    }`}
                  >
                    {p.nivel}
                  </span>
                  {!p.ativo && (
                    <span className="rounded bg-danger/10 px-1.5 py-0.5 text-[10px] font-semibold text-danger">
                      Inativo
                    </span>
                  )}
                </div>
              </div>

              <p className="mt-3 text-xs text-muted leading-relaxed line-clamp-3">
                {p.descricao || "Sem descrição cadastrada para este perfil."}
              </p>

              <div className="mt-4 flex flex-col gap-2 rounded-md bg-black/[0.02] dark:bg-white/[0.02] p-3 border border-surface-border/50 text-xs">
                <div className="flex items-center justify-between">
                  <span className="text-muted flex items-center gap-1">
                    <Users size={12} /> Grupo AD:
                  </span>
                  <code className="font-mono font-semibold text-foreground text-[11px]">
                    {p.grupo_ad || "—"}
                  </code>
                </div>

                <div className="flex flex-col gap-1 mt-1 border-t border-surface-border/40 pt-2">
                  <span className="text-[11px] text-muted flex items-center gap-1">
                    <KeyRound size={12} /> Permissões Granulares:
                  </span>
                  <div className="flex flex-wrap gap-1 mt-0.5">
                    {p.permissoes && p.permissoes.length > 0 ? (
                      p.permissoes.map((perm) => (
                        <span
                          key={perm}
                          className="rounded bg-surface px-1.5 py-0.5 text-[10px] font-mono text-muted border border-surface-border"
                        >
                          {perm}
                        </span>
                      ))
                    ) : (
                      <span className="text-[10px] text-muted italic">Acesso padrão de leitura</span>
                    )}
                  </div>
                </div>
              </div>
            </div>

            <div className="mt-4 pt-3 border-t border-surface-border/50 flex items-center justify-end">
              <Button
                size="sm"
                variant="secondary"
                onClick={() => openEditPerfil(p)}
                className="flex items-center gap-1.5 text-xs"
              >
                <Edit3 size={13} />
                Editar Perfil
              </Button>
            </div>
          </div>
        ))}
      </div>

      {/* MODAL: NOVO PERFIL */}
      {isNewModalOpen && (
        <ModalShell open={isNewModalOpen} onClose={() => setIsNewModalOpen(false)}>
          <div className="flex flex-col gap-4 p-5">
            <h2 className="text-lg font-semibold text-foreground">Criar Novo Perfil de Acesso</h2>
            <form onSubmit={handleCreatePerfil} className="flex flex-col gap-4">
              <Input
                label="Nome do Perfil"
                placeholder="Ex: Coordenador de Políticas Públicas"
                value={nome}
                onChange={(e) => {
                  setNome(e.target.value);
                  if (!slug) {
                    setSlug(
                      e.target.value
                        .toLowerCase()
                        .normalize("NFD")
                        .replace(/[\u0300-\u036f]/g, "")
                        .replace(/[^a-z0-9]+/g, "_")
                        .replace(/^_+|_+$/g, "")
                    );
                  }
                }}
                required
              />

              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <Input
                  label="Identificador (Slug único)"
                  placeholder="Ex: coordenador_politicas"
                  value={slug}
                  onChange={(e) => setSlug(e.target.value)}
                  required
                />

                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-semibold text-foreground">Nível Hierárquico</label>
                  <select
                    value={nivel}
                    onChange={(e) => setNivel(e.target.value as any)}
                    className="rounded-md border border-surface-border bg-surface px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-none"
                    required
                  >
                    <option value="tecnico">Técnico Social (Assistente Social / Psicólogo)</option>
                    <option value="recepcao">Recepção / Triagem</option>
                    <option value="gerencia">Gerência / Coordenação da Unidade</option>
                    <option value="gestao">Gestão SEMPRAS</option>
                    <option value="admin">Administrador Geral</option>
                  </select>
                </div>
              </div>

              <Input
                label="Grupo Correspondente no Active Directory"
                placeholder="Ex: Grupo_AS_Gestao ou Grupo_Tecnico_Social"
                value={grupoAD}
                onChange={(e) => setGrupoAD(e.target.value)}
                required
              />

              <Input
                label="Descrição do Perfil"
                placeholder="Ex: Responsável por acolhimento e elaboração de relatórios..."
                value={descricao}
                onChange={(e) => setDescricao(e.target.value)}
              />

              <Input
                label="Permissões (separadas por vírgula)"
                placeholder="Ex: atendimento:chamar, atendimento:triagem, prontuario:sigiloso"
                value={permissoesStr}
                onChange={(e) => setPermissoesStr(e.target.value)}
              />

              <div className="mt-4 flex justify-end gap-2 border-t border-surface-border pt-4">
                <Button type="button" variant="secondary" onClick={() => setIsNewModalOpen(false)}>
                  Cancelar
                </Button>
                <Button type="submit" loading={saving}>
                  Criar Perfil
                </Button>
              </div>
            </form>
          </div>
        </ModalShell>
      )}

      {/* MODAL: EDITAR PERFIL */}
      {editingPerfil && (
        <ModalShell open={!!editingPerfil} onClose={() => setEditingPerfil(null)}>
          <div className="flex flex-col gap-4 p-5">
            <h2 className="text-lg font-semibold text-foreground">Editar Perfil: {editingPerfil.nome}</h2>
            <form onSubmit={handleUpdatePerfil} className="flex flex-col gap-4">
              <Input
                label="Nome do Perfil"
                value={editNome}
                onChange={(e) => setEditNome(e.target.value)}
                required
              />

              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <Input
                  label="Identificador (Slug único)"
                  value={editSlug}
                  onChange={(e) => setEditSlug(e.target.value)}
                  required
                />

                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-semibold text-foreground">Nível Hierárquico</label>
                  <select
                    value={editNivel}
                    onChange={(e) => setEditNivel(e.target.value as any)}
                    className="rounded-md border border-surface-border bg-surface px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-none"
                    required
                  >
                    <option value="tecnico">Técnico Social (Assistente Social / Psicólogo)</option>
                    <option value="recepcao">Recepção / Triagem</option>
                    <option value="gerencia">Gerência / Coordenação da Unidade</option>
                    <option value="gestao">Gestão SEMPRAS</option>
                    <option value="admin">Administrador Geral</option>
                  </select>
                </div>
              </div>

              <Input
                label="Grupo Correspondente no Active Directory"
                value={editGrupoAD}
                onChange={(e) => setEditGrupoAD(e.target.value)}
                required
              />

              <Input
                label="Descrição do Perfil"
                value={editDescricao}
                onChange={(e) => setEditDescricao(e.target.value)}
              />

              <Input
                label="Permissões (separadas por vírgula)"
                value={editPermissoesStr}
                onChange={(e) => setEditPermissoesStr(e.target.value)}
              />

              <label className="flex items-center gap-2 text-xs font-medium text-foreground cursor-pointer">
                <input
                  type="checkbox"
                  checked={editAtivo}
                  onChange={(e) => setEditAtivo(e.target.checked)}
                  className="rounded border-surface-border text-primary focus:ring-primary"
                />
                Perfil ativo no sistema
              </label>

              <div className="mt-4 flex justify-end gap-2 border-t border-surface-border pt-4">
                <Button type="button" variant="secondary" onClick={() => setEditingPerfil(null)}>
                  Cancelar
                </Button>
                <Button type="submit" loading={savingEdit}>
                  Salvar Alterações
                </Button>
              </div>
            </form>
          </div>
        </ModalShell>
      )}
    </div>
  );
}
