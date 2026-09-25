"use client";

import { useState } from "react";
import {
  Building2,
  ChevronDown,
  ChevronRight,
  Edit3,
  FolderPlus,
  MapPin,
  Phone,
  Plus,
  Search,
  ShieldCheck,
} from "lucide-react";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { ModalShell } from "@/components/ui/ModalShell";
import { apiClient, ApiError } from "@/lib/api/client";
import { useToast } from "@/components/notifications/ToastProvider";
import type { Localidade, Setor } from "@/types/api";

interface LocalidadesManagerProps {
  initialLocalidades: Localidade[];
}

export function LocalidadesManager({ initialLocalidades }: LocalidadesManagerProps) {
  const { showToast } = useToast();
  const [localidades, setLocalidades] = useState<Localidade[]>(initialLocalidades);
  const [expanded, setExpanded] = useState<Record<string, boolean>>({});
  const [searchTerm, setSearchTerm] = useState("");
  const [filterType, setFilterType] = useState<string>("ALL");

  // Modal: Nova Localidade
  const [isNewLocModalOpen, setIsNewLocModalOpen] = useState(false);
  const [newLocName, setNewLocName] = useState("");
  const [newLocSlug, setNewLocSlug] = useState("");
  const [newLocTipo, setNewLocTipo] = useState<Localidade["tipo"]>("CRAS");
  const [newLocGrupoAD, setNewLocGrupoAD] = useState("");
  const [newLocEndereco, setNewLocEndereco] = useState("");
  const [newLocTelefone, setNewLocTelefone] = useState("");
  const [newLocBairro, setNewLocBairro] = useState("");
  const [savingLoc, setSavingLoc] = useState(false);

  // Modal: Editar Localidade
  const [editingLoc, setEditingLoc] = useState<Localidade | null>(null);
  const [editLocName, setEditLocName] = useState("");
  const [editLocSlug, setEditLocSlug] = useState("");
  const [editLocTipo, setEditLocTipo] = useState<Localidade["tipo"]>("CRAS");
  const [editLocGrupoAD, setEditLocGrupoAD] = useState("");
  const [editLocEndereco, setEditLocEndereco] = useState("");
  const [editLocTelefone, setEditLocTelefone] = useState("");
  const [editLocBairro, setEditLocBairro] = useState("");
  const [editLocAtivo, setEditLocAtivo] = useState(true);
  const [savingEditLoc, setSavingEditLoc] = useState(false);

  // Modal: Novo Setor
  const [selectedLocForSetor, setSelectedLocForSetor] = useState<Localidade | null>(null);
  const [newSetorName, setNewSetorName] = useState("");
  const [newSetorSlug, setNewSetorSlug] = useState("");
  const [newSetorTipo, setNewSetorTipo] = useState<Setor["tipo"]>("TECNICO");
  const [newSetorGrupoAD, setNewSetorGrupoAD] = useState("");
  const [newSetorDescricao, setNewSetorDescricao] = useState("");
  const [savingSetor, setSavingSetor] = useState(false);

  // Modal: Editar Setor
  const [editingSetor, setEditingSetor] = useState<Setor | null>(null);
  const [editingSetorLocId, setEditingSetorLocId] = useState<string>("");
  const [editSetorName, setEditSetorName] = useState("");
  const [editSetorSlug, setEditSetorSlug] = useState("");
  const [editSetorTipo, setEditSetorTipo] = useState<Setor["tipo"]>("TECNICO");
  const [editSetorGrupoAD, setEditSetorGrupoAD] = useState("");
  const [editSetorDescricao, setEditSetorDescricao] = useState("");
  const [editSetorAtivo, setEditSetorAtivo] = useState(true);
  const [savingEditSetor, setSavingEditSetor] = useState(false);

  const toggleExpand = (id: string) => {
    setExpanded((prev) => ({ ...prev, [id]: !prev[id] }));
  };

  const filteredLocalidades = localidades.filter((l) => {
    const matchesSearch =
      l.nome.toLowerCase().includes(searchTerm.toLowerCase()) ||
      l.slug.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (l.bairro && l.bairro.toLowerCase().includes(searchTerm.toLowerCase())) ||
      (l.grupo_ad && l.grupo_ad.toLowerCase().includes(searchTerm.toLowerCase()));
    const matchesType = filterType === "ALL" || l.tipo === filterType;
    return matchesSearch && matchesType;
  });

  // Funções de Criação
  async function handleCreateLocalidade(e: React.FormEvent) {
    e.preventDefault();
    setSavingLoc(true);
    try {
      const res = await apiClient.post<Localidade>("v1/organizacao/localidades", {
        nome: newLocName,
        slug: newLocSlug,
        tipo: newLocTipo,
        grupo_ad: newLocGrupoAD,
        endereco: newLocEndereco,
        telefone: newLocTelefone,
        bairro: newLocBairro,
      });
      const created = res.data;
      setLocalidades((prev) => [...prev, { ...created, setores: [] }]);
      showToast({
        title: "Unidade cadastrada",
        description: `Localidade ${created.nome} cadastrada com sucesso.`,
        tone: "success",
      });
      setIsNewLocModalOpen(false);
      setNewLocName("");
      setNewLocSlug("");
      setNewLocGrupoAD("");
      setNewLocEndereco("");
      setNewLocTelefone("");
      setNewLocBairro("");
    } catch (err: any) {
      showToast({
        title: "Erro ao cadastrar localidade",
        description: err instanceof ApiError ? err.message : "Erro ao cadastrar localidade.",
        tone: "danger",
      });
    } finally {
      setSavingLoc(false);
    }
  }

  async function handleCreateSetor(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedLocForSetor) return;
    setSavingSetor(true);
    try {
      const res = await apiClient.post<Setor>(
        `v1/organizacao/localidades/${selectedLocForSetor.id}/setores`,
        {
          nome: newSetorName,
          slug: newSetorSlug,
          tipo: newSetorTipo,
          grupo_ad: newSetorGrupoAD,
          descricao: newSetorDescricao,
        }
      );
      const created = res.data;
      setLocalidades((prev) =>
        prev.map((loc) => {
          if (loc.id === selectedLocForSetor.id) {
            return {
              ...loc,
              setores: [...(loc.setores || []), created],
            };
          }
          return loc;
        })
      );
      showToast({
        title: "Setor adicionado",
        description: `Setor ${created.nome} vinculado com sucesso à unidade.`,
        tone: "success",
      });
      setSelectedLocForSetor(null);
      setNewSetorName("");
      setNewSetorSlug("");
      setNewSetorGrupoAD("");
      setNewSetorDescricao("");
    } catch (err: any) {
      showToast({
        title: "Erro ao adicionar setor",
        description: err instanceof ApiError ? err.message : "Erro ao adicionar setor à localidade.",
        tone: "danger",
      });
    } finally {
      setSavingSetor(false);
    }
  }

  // Funções de Edição
  function openEditLocalidade(loc: Localidade) {
    setEditingLoc(loc);
    setEditLocName(loc.nome);
    setEditLocSlug(loc.slug);
    setEditLocTipo(loc.tipo);
    setEditLocGrupoAD(loc.grupo_ad || "");
    setEditLocEndereco(loc.endereco || "");
    setEditLocTelefone(loc.telefone || "");
    setEditLocBairro(loc.bairro || "");
    setEditLocAtivo(loc.ativo);
  }

  async function handleUpdateLocalidade(e: React.FormEvent) {
    e.preventDefault();
    if (!editingLoc) return;
    setSavingEditLoc(true);
    try {
      const res = await apiClient.put<Localidade>(
        `v1/organizacao/localidades/${editingLoc.id}`,
        {
          nome: editLocName,
          slug: editLocSlug,
          tipo: editLocTipo,
          grupo_ad: editLocGrupoAD,
          endereco: editLocEndereco,
          telefone: editLocTelefone,
          bairro: editLocBairro,
          ativo: editLocAtivo,
        }
      );
      const updated = res.data;
      setLocalidades((prev) =>
        prev.map((loc) => (loc.id === updated.id ? { ...updated, setores: loc.setores } : loc))
      );
      showToast({
        title: "Localidade atualizada",
        description: `Dados de ${updated.nome} salvos com sucesso.`,
        tone: "success",
      });
      setEditingLoc(null);
    } catch (err: any) {
      showToast({
        title: "Erro ao atualizar localidade",
        description: err instanceof ApiError ? err.message : "Erro ao atualizar localidade.",
        tone: "danger",
      });
    } finally {
      setSavingEditLoc(false);
    }
  }

  function openEditSetor(locId: string, setor: Setor) {
    setEditingSetor(setor);
    setEditingSetorLocId(locId);
    setEditSetorName(setor.nome);
    setEditSetorSlug(setor.slug);
    setEditSetorTipo(setor.tipo);
    setEditSetorGrupoAD(setor.grupo_ad || "");
    setEditSetorDescricao(setor.descricao || "");
    setEditSetorAtivo(setor.ativo);
  }

  async function handleUpdateSetor(e: React.FormEvent) {
    e.preventDefault();
    if (!editingSetor) return;
    setSavingEditSetor(true);
    try {
      const res = await apiClient.put<Setor>(
        `v1/organizacao/setores/${editingSetor.id}`,
        {
          nome: editSetorName,
          slug: editSetorSlug,
          tipo: editSetorTipo,
          grupo_ad: editSetorGrupoAD,
          descricao: editSetorDescricao,
          ativo: editSetorAtivo,
        }
      );
      const updated = res.data;
      setLocalidades((prev) =>
        prev.map((loc) => {
          if (loc.id === editingSetorLocId) {
            return {
              ...loc,
              setores: loc.setores?.map((s) => (s.id === updated.id ? updated : s)),
            };
          }
          return loc;
        })
      );
      showToast({
        title: "Setor atualizado",
        description: `Dados de ${updated.nome} salvos com sucesso.`,
        tone: "success",
      });
      setEditingSetor(null);
    } catch (err: any) {
      showToast({
        title: "Erro ao atualizar setor",
        description: err instanceof ApiError ? err.message : "Erro ao atualizar setor.",
        tone: "danger",
      });
    } finally {
      setSavingEditSetor(false);
    }
  }

  return (
    <div className="flex flex-col gap-6">
      {/* Controles de Busca e Criação */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-1 items-center gap-3">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted" aria-hidden="true" />
            <input
              type="text"
              placeholder="Buscar por unidade, bairro, slug ou grupo AD..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full rounded-md border border-surface-border bg-surface py-2 pl-9 pr-3 text-sm text-foreground focus:border-primary focus:outline-none"
            />
          </div>
          <select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value)}
            className="rounded-md border border-surface-border bg-surface px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-none"
          >
            <option value="ALL">Todos os Tipos</option>
            <option value="CRAS">CRAS (9 Unidades)</option>
            <option value="CREAS">CREAS</option>
            <option value="CENTRO_POP">Centro POP</option>
            <option value="CASA_MULHER">Casa da Mulher</option>
            <option value="CASA_ABRIGO">Casa Abrigo</option>
            <option value="CONSELHO_TUTELAR">Conselhos Tutelares</option>
            <option value="GESTAO_SEMPRAS">Sede / Gestão</option>
          </select>
        </div>

        <Button onClick={() => setIsNewLocModalOpen(true)} className="flex items-center gap-2">
          <Plus size={16} aria-hidden="true" />
          Nova Localidade
        </Button>
      </div>

      {/* Lista de Localidades & Setores */}
      <div className="flex flex-col gap-3">
        {filteredLocalidades.length === 0 ? (
          <div className="rounded-lg border border-dashed border-surface-border p-8 text-center text-sm text-muted">
            Nenhuma localidade encontrada com os filtros informados.
          </div>
        ) : (
          filteredLocalidades.map((loc) => {
            const isExp = expanded[loc.id] ?? false;
            const setores = loc.setores || [];

            return (
              <div
                key={loc.id}
                className="overflow-hidden rounded-lg border border-surface-border bg-surface transition-shadow hover:shadow-xs"
              >
                {/* Linha Principal da Localidade */}
                <div className="flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between">
                  <div className="flex items-start gap-3">
                    <button
                      type="button"
                      onClick={() => toggleExpand(loc.id)}
                      className="mt-1 rounded p-1 text-muted hover:bg-black/5 hover:text-foreground dark:hover:bg-white/5"
                      title={isExp ? "Recolher setores" : "Expandir setores"}
                    >
                      {isExp ? <ChevronDown size={18} /> : <ChevronRight size={18} />}
                    </button>

                    <Building2 className="mt-1 h-5 w-5 text-primary shrink-0" />

                    <div>
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="font-semibold text-foreground text-base">{loc.nome}</span>
                        <span className="rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-medium text-primary">
                          {loc.tipo}
                        </span>
                        <span className="rounded-full bg-surface-border/60 px-2 py-0.5 text-xs text-muted font-mono">
                          {loc.slug}
                        </span>
                        {!loc.ativo && (
                          <span className="rounded-full bg-danger/10 px-2 py-0.5 text-xs text-danger font-medium">
                            Inativo
                          </span>
                        )}
                      </div>

                      <div className="mt-1 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted">
                        <span className="flex items-center gap-1">
                          <ShieldCheck size={13} className="text-primary/70" />
                          AD: <code className="font-mono text-foreground font-semibold">{loc.grupo_ad || "—"}</code>
                        </span>
                        {loc.bairro && (
                          <span className="flex items-center gap-1">
                            <MapPin size={13} />
                            {loc.bairro} {loc.endereco && `· ${loc.endereco}`}
                          </span>
                        )}
                        {loc.telefone && (
                          <span className="flex items-center gap-1">
                            <Phone size={13} />
                            {loc.telefone}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-2 self-end sm:self-center">
                    <span className="text-xs text-muted mr-1">
                      {setores.length} {setores.length === 1 ? "setor" : "setores"}
                    </span>
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={() => openEditLocalidade(loc)}
                      className="flex items-center gap-1 text-xs"
                      title="Editar dados da unidade"
                    >
                      <Edit3 size={13} />
                      Editar
                    </Button>
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={() => {
                        setSelectedLocForSetor(loc);
                        setExpanded((p) => ({ ...p, [loc.id]: true }));
                      }}
                      className="flex items-center gap-1 text-xs"
                    >
                      <FolderPlus size={14} />
                      + Setor
                    </Button>
                  </div>
                </div>

                {/* Área Expandida com os Setores da Unidade */}
                {isExp && (
                  <div className="border-t border-surface-border bg-black/[0.02] dark:bg-white/[0.02] p-4">
                    <p className="text-xs font-semibold uppercase tracking-wider text-muted mb-3">
                      Setores e Departamentos Vinculados ({setores.length})
                    </p>

                    {setores.length === 0 ? (
                      <p className="text-xs text-muted italic">Nenhum setor cadastrado para esta localidade.</p>
                    ) : (
                      <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-2 lg:grid-cols-3">
                        {setores.map((setor) => (
                          <div
                            key={setor.id}
                            className="flex flex-col justify-between rounded-lg border border-surface-border/70 bg-surface p-3.5 shadow-xs"
                          >
                            <div>
                              <div className="flex items-center justify-between gap-2">
                                <span className="text-sm font-semibold text-foreground">{setor.nome}</span>
                                <span
                                  className={`rounded px-1.5 py-0.5 text-[10px] font-bold uppercase ${
                                    setor.tipo === "TECNICO"
                                      ? "bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-300"
                                      : setor.tipo === "RECEPCAO"
                                      ? "bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300"
                                      : setor.tipo === "GERENCIA"
                                      ? "bg-purple-100 text-purple-800 dark:bg-purple-900/40 dark:text-purple-300"
                                      : "bg-surface-border text-foreground"
                                  }`}
                                >
                                  {setor.tipo}
                                </span>
                              </div>
                              {setor.descricao && (
                                <p className="mt-1 text-xs text-muted line-clamp-2">{setor.descricao}</p>
                              )}
                            </div>

                            <div className="mt-3 flex items-center justify-between border-t border-surface-border/40 pt-2 text-[11px] text-muted">
                              <span className="font-mono text-primary/90 text-[10px]">{setor.grupo_ad || "—"}</span>
                              <Button
                                size="sm"
                                variant="secondary"
                                onClick={() => openEditSetor(loc.id, setor)}
                                className="h-6 px-2 text-[11px] flex items-center gap-1"
                              >
                                <Edit3 size={11} />
                                Editar
                              </Button>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>

      {/* MODAL: NOVA LOCALIDADE */}
      {isNewLocModalOpen && (
        <ModalShell open={isNewLocModalOpen} onClose={() => setIsNewLocModalOpen(false)}>
          <div className="flex flex-col gap-4 p-5">
            <h2 className="text-lg font-semibold text-foreground">Cadastrar Nova Localidade</h2>
            <form onSubmit={handleCreateLocalidade} className="flex flex-col gap-4">
              <Input
                label="Nome da Unidade"
                placeholder="Ex: CRAS - Sagrada Família ou Casa Abrigo"
                value={newLocName}
                onChange={(e) => {
                  setNewLocName(e.target.value);
                  if (!newLocSlug) {
                    setNewLocSlug(
                      e.target.value
                        .toLowerCase()
                        .normalize("NFD")
                        .replace(/[\u0300-\u036f]/g, "")
                        .replace(/[^a-z0-9]+/g, "-")
                        .replace(/^-+|-+$/g, "")
                    );
                  }
                }}
                required
              />

              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <Input
                  label="Identificador (Slug único)"
                  placeholder="Ex: cras-sagrada-familia"
                  value={newLocSlug}
                  onChange={(e) => setNewLocSlug(e.target.value)}
                  required
                />

                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-semibold text-foreground">Tipo de Polo</label>
                  <select
                    value={newLocTipo}
                    onChange={(e) => setNewLocTipo(e.target.value as any)}
                    className="rounded-md border border-surface-border bg-surface px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-none"
                    required
                  >
                    <option value="CRAS">CRAS (Centro de Referência de Assistência Social)</option>
                    <option value="CREAS">CREAS (Centro Especializado de Assistência Social)</option>
                    <option value="CENTRO_POP">Centro POP (População em Situação de Rua)</option>
                    <option value="CASA_MULHER">Casa da Mulher</option>
                    <option value="CASA_ABRIGO">Casa Abrigo</option>
                    <option value="CONSELHO_TUTELAR">Conselho Tutelar</option>
                    <option value="GESTAO_SEMPRAS">Sede SEMPRAS / Gestão</option>
                    <option value="OUTROS">Outra Unidade Municipal</option>
                  </select>
                </div>
              </div>

              <Input
                label="Grupo Vinculado no Active Directory (AD)"
                placeholder="Ex: Grupo_CRAS_Sagrada_Familia"
                value={newLocGrupoAD}
                onChange={(e) => setNewLocGrupoAD(e.target.value)}
                required
              />

              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <Input
                  label="Bairro"
                  placeholder="Ex: Vila Operária ou Centro"
                  value={newLocBairro}
                  onChange={(e) => setNewLocBairro(e.target.value)}
                />
                <Input
                  label="Telefone de Contato"
                  placeholder="Ex: (66) 3411-5000"
                  value={newLocTelefone}
                  onChange={(e) => setNewLocTelefone(e.target.value)}
                />
              </div>

              <Input
                label="Endereço Completo"
                placeholder="Ex: Rua das Flores, 123"
                value={newLocEndereco}
                onChange={(e) => setNewLocEndereco(e.target.value)}
              />

              <div className="mt-4 flex justify-end gap-2 border-t border-surface-border pt-4">
                <Button type="button" variant="secondary" onClick={() => setIsNewLocModalOpen(false)}>
                  Cancelar
                </Button>
                <Button type="submit" loading={savingLoc}>
                  Salvar Localidade
                </Button>
              </div>
            </form>
          </div>
        </ModalShell>
      )}

      {/* MODAL: EDITAR LOCALIDADE */}
      {editingLoc && (
        <ModalShell open={!!editingLoc} onClose={() => setEditingLoc(null)}>
          <div className="flex flex-col gap-4 p-5">
            <h2 className="text-lg font-semibold text-foreground">Editar Localidade: {editingLoc.nome}</h2>
            <form onSubmit={handleUpdateLocalidade} className="flex flex-col gap-4">
              <Input
                label="Nome da Unidade"
                value={editLocName}
                onChange={(e) => setEditLocName(e.target.value)}
                required
              />

              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <Input
                  label="Identificador (Slug único)"
                  value={editLocSlug}
                  onChange={(e) => setEditLocSlug(e.target.value)}
                  required
                />

                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-semibold text-foreground">Tipo de Polo</label>
                  <select
                    value={editLocTipo}
                    onChange={(e) => setEditLocTipo(e.target.value as any)}
                    className="rounded-md border border-surface-border bg-surface px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-none"
                    required
                  >
                    <option value="CRAS">CRAS (Centro de Referência de Assistência Social)</option>
                    <option value="CREAS">CREAS (Centro Especializado de Assistência Social)</option>
                    <option value="CENTRO_POP">Centro POP (População em Situação de Rua)</option>
                    <option value="CASA_MULHER">Casa da Mulher</option>
                    <option value="CASA_ABRIGO">Casa Abrigo</option>
                    <option value="CONSELHO_TUTELAR">Conselho Tutelar</option>
                    <option value="GESTAO_SEMPRAS">Sede SEMPRAS / Gestão</option>
                    <option value="OUTROS">Outra Unidade Municipal</option>
                  </select>
                </div>
              </div>

              <Input
                label="Grupo Vinculado no Active Directory (AD)"
                value={editLocGrupoAD}
                onChange={(e) => setEditLocGrupoAD(e.target.value)}
                required
              />

              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <Input
                  label="Bairro"
                  value={editLocBairro}
                  onChange={(e) => setEditLocBairro(e.target.value)}
                />
                <Input
                  label="Telefone de Contato"
                  value={editLocTelefone}
                  onChange={(e) => setEditLocTelefone(e.target.value)}
                />
              </div>

              <Input
                label="Endereço Completo"
                value={editLocEndereco}
                onChange={(e) => setEditLocEndereco(e.target.value)}
              />

              <label className="flex items-center gap-2 text-xs font-medium text-foreground cursor-pointer">
                <input
                  type="checkbox"
                  checked={editLocAtivo}
                  onChange={(e) => setEditLocAtivo(e.target.checked)}
                  className="rounded border-surface-border text-primary focus:ring-primary"
                />
                Unidade ativa no sistema
              </label>

              <div className="mt-4 flex justify-end gap-2 border-t border-surface-border pt-4">
                <Button type="button" variant="secondary" onClick={() => setEditingLoc(null)}>
                  Cancelar
                </Button>
                <Button type="submit" loading={savingEditLoc}>
                  Salvar Alterações
                </Button>
              </div>
            </form>
          </div>
        </ModalShell>
      )}

      {/* MODAL: NOVO SETOR */}
      {selectedLocForSetor && (
        <ModalShell
          open={!!selectedLocForSetor}
          onClose={() => setSelectedLocForSetor(null)}
        >
          <div className="flex flex-col gap-4 p-5">
            <h2 className="text-lg font-semibold text-foreground">Novo Setor em: {selectedLocForSetor.nome}</h2>
            <form onSubmit={handleCreateSetor} className="flex flex-col gap-4">
              <Input
                label="Nome do Setor / Departamento"
                placeholder="Ex: Atendimento Especializado PAIF"
                value={newSetorName}
                onChange={(e) => {
                  setNewSetorName(e.target.value);
                  if (!newSetorSlug) {
                    setNewSetorSlug(
                      e.target.value
                        .toLowerCase()
                        .normalize("NFD")
                        .replace(/[\u0300-\u036f]/g, "")
                        .replace(/[^a-z0-9]+/g, "-")
                        .replace(/^-+|-+$/g, "")
                    );
                  }
                }}
                required
              />

              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <Input
                  label="Slug do Setor"
                  placeholder="Ex: paif ou recepcao"
                  value={newSetorSlug}
                  onChange={(e) => setNewSetorSlug(e.target.value)}
                  required
                />

                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-semibold text-foreground">Tipo de Atuação</label>
                  <select
                    value={newSetorTipo}
                    onChange={(e) => setNewSetorTipo(e.target.value as any)}
                    className="rounded-md border border-surface-border bg-surface px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-none"
                    required
                  >
                    <option value="TECNICO">Atendimento Técnico (Assistente Social / Psicólogo)</option>
                    <option value="RECEPCAO">Recepção / Triagem e Entrada</option>
                    <option value="GERENCIA">Gerência e Coordenação da Unidade</option>
                    <option value="ADMINISTRATIVO">Apoio Administrativo</option>
                  </select>
                </div>
              </div>

              <Input
                label="Grupo do Active Directory para este Setor"
                placeholder="Ex: Grupo_CRAS_Ana_Carla_Tecnico"
                value={newSetorGrupoAD}
                onChange={(e) => setNewSetorGrupoAD(e.target.value)}
              />

              <Input
                label="Descrição das Atividades do Setor"
                placeholder="Ex: Condução de planos de acompanhamento familiar e PAIF..."
                value={newSetorDescricao}
                onChange={(e) => setNewSetorDescricao(e.target.value)}
              />

              <div className="mt-4 flex justify-end gap-2 border-t border-surface-border pt-4">
                <Button type="button" variant="secondary" onClick={() => setSelectedLocForSetor(null)}>
                  Cancelar
                </Button>
                <Button type="submit" loading={savingSetor}>
                  Adicionar Setor
                </Button>
              </div>
            </form>
          </div>
        </ModalShell>
      )}

      {/* MODAL: EDITAR SETOR */}
      {editingSetor && (
        <ModalShell open={!!editingSetor} onClose={() => setEditingSetor(null)}>
          <div className="flex flex-col gap-4 p-5">
            <h2 className="text-lg font-semibold text-foreground">Editar Setor: {editingSetor.nome}</h2>
            <form onSubmit={handleUpdateSetor} className="flex flex-col gap-4">
              <Input
                label="Nome do Setor / Departamento"
                value={editSetorName}
                onChange={(e) => setEditSetorName(e.target.value)}
                required
              />

              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <Input
                  label="Slug do Setor"
                  value={editSetorSlug}
                  onChange={(e) => setEditSetorSlug(e.target.value)}
                  required
                />

                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-semibold text-foreground">Tipo de Atuação</label>
                  <select
                    value={editSetorTipo}
                    onChange={(e) => setEditSetorTipo(e.target.value as any)}
                    className="rounded-md border border-surface-border bg-surface px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-none"
                    required
                  >
                    <option value="TECNICO">Atendimento Técnico (Assistente Social / Psicólogo)</option>
                    <option value="RECEPCAO">Recepção / Triagem e Entrada</option>
                    <option value="GERENCIA">Gerência e Coordenação da Unidade</option>
                    <option value="ADMINISTRATIVO">Apoio Administrativo</option>
                  </select>
                </div>
              </div>

              <Input
                label="Grupo do Active Directory para este Setor"
                value={editSetorGrupoAD}
                onChange={(e) => setEditSetorGrupoAD(e.target.value)}
              />

              <Input
                label="Descrição das Atividades do Setor"
                value={editSetorDescricao}
                onChange={(e) => setEditSetorDescricao(e.target.value)}
              />

              <label className="flex items-center gap-2 text-xs font-medium text-foreground cursor-pointer">
                <input
                  type="checkbox"
                  checked={editSetorAtivo}
                  onChange={(e) => setEditSetorAtivo(e.target.checked)}
                  className="rounded border-surface-border text-primary focus:ring-primary"
                />
                Setor ativo
              </label>

              <div className="mt-4 flex justify-end gap-2 border-t border-surface-border pt-4">
                <Button type="button" variant="secondary" onClick={() => setEditingSetor(null)}>
                  Cancelar
                </Button>
                <Button type="submit" loading={savingEditSetor}>
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
