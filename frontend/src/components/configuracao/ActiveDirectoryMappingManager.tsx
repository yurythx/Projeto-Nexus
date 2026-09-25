"use client";

import { useState } from "react";
import {
  Building2,
  Check,
  Edit3,
  HelpCircle,
  KeyRound,
  Layers,
  Search,
  Shield,
  ShieldAlert,
  ShieldCheck,
  Users,
} from "lucide-react";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
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
import { useToast } from "@/components/notifications/ToastProvider";
import type { Localidade, Perfil, Setor } from "@/types/api";

interface ActiveDirectoryMappingManagerProps {
  initialPerfis: Perfil[];
  initialLocalidades: Localidade[];
}

export function ActiveDirectoryMappingManager({
  initialPerfis,
  initialLocalidades,
}: ActiveDirectoryMappingManagerProps) {
  const { showToast } = useToast();
  const [perfis, setPerfis] = useState<Perfil[]>(initialPerfis);
  const [localidades, setLocalidades] = useState<Localidade[]>(initialLocalidades);

  const [activeTab, setActiveTab] = useState<"perfis" | "localidades" | "setores">("perfis");
  const [searchTerm, setSearchTerm] = useState("");

  // Modal para editar grupo AD de Perfil
  const [editingPerfil, setEditingPerfil] = useState<Perfil | null>(null);
  const [editPerfilGrupoAD, setEditPerfilGrupoAD] = useState("");
  const [savingPerfil, setSavingPerfil] = useState(false);

  // Modal para editar grupo AD de Localidade
  const [editingLoc, setEditingLoc] = useState<Localidade | null>(null);
  const [editLocGrupoAD, setEditLocGrupoAD] = useState("");
  const [savingLoc, setSavingLoc] = useState(false);

  // Modal para editar grupo AD de Setor
  const [editingSetor, setEditingSetor] = useState<{
    setor: Setor;
    locId: string;
    localidadeNome: string;
  } | null>(null);
  const [editSetorGrupoAD, setEditSetorGrupoAD] = useState("");
  const [savingSetor, setSavingSetor] = useState(false);

  // Todos os setores planificados
  const allSetores = localidades.flatMap((loc) =>
    (loc.setores || []).map((s) => ({
      ...s,
      localidadeNome: loc.nome,
      localidadeTipo: loc.tipo,
      locId: loc.id,
    }))
  );

  // Ações de Atualização de Grupo AD
  async function handleSavePerfilGrupoAD(e: React.FormEvent) {
    e.preventDefault();
    if (!editingPerfil) return;
    setSavingPerfil(true);
    try {
      const res = await apiClient.put<Perfil>(`v1/organizacao/perfis/${editingPerfil.id}`, {
        nome: editingPerfil.nome,
        slug: editingPerfil.slug,
        descricao: editingPerfil.descricao,
        nivel: editingPerfil.nivel,
        permissoes: editingPerfil.permissoes,
        ativo: editingPerfil.ativo,
        grupo_ad: editPerfilGrupoAD.trim(),
      });
      const updated = res.data;
      setPerfis((prev) => prev.map((p) => (p.id === updated.id ? updated : p)));
      showToast({
        title: "Mapeamento salvo",
        description: `Grupo do AD vinculado ao perfil ${updated.nome} com sucesso.`,
        tone: "success",
      });
      setEditingPerfil(null);
    } catch (err: any) {
      showToast({
        title: "Erro ao atualizar grupo",
        description: err instanceof ApiError ? err.message : "Erro ao atualizar grupo do AD no perfil.",
        tone: "danger",
      });
    } finally {
      setSavingPerfil(false);
    }
  }

  async function handleSaveLocGrupoAD(e: React.FormEvent) {
    e.preventDefault();
    if (!editingLoc) return;
    setSavingLoc(true);
    try {
      const res = await apiClient.put<Localidade>(`v1/organizacao/localidades/${editingLoc.id}`, {
        nome: editingLoc.nome,
        slug: editingLoc.slug,
        tipo: editingLoc.tipo,
        endereco: editingLoc.endereco,
        telefone: editingLoc.telefone,
        bairro: editingLoc.bairro,
        ativo: editingLoc.ativo,
        grupo_ad: editLocGrupoAD.trim(),
      });
      const updated = res.data;
      setLocalidades((prev) =>
        prev.map((l) => (l.id === updated.id ? { ...updated, setores: l.setores } : l))
      );
      showToast({
        title: "Mapeamento salvo",
        description: `Grupo do AD vinculado à localidade ${updated.nome} com sucesso.`,
        tone: "success",
      });
      setEditingLoc(null);
    } catch (err: any) {
      showToast({
        title: "Erro ao atualizar grupo",
        description: err instanceof ApiError ? err.message : "Erro ao atualizar grupo do AD na localidade.",
        tone: "danger",
      });
    } finally {
      setSavingLoc(false);
    }
  }

  async function handleSaveSetorGrupoAD(e: React.FormEvent) {
    e.preventDefault();
    if (!editingSetor) return;
    setSavingSetor(true);
    try {
      const res = await apiClient.put<Setor>(`v1/organizacao/setores/${editingSetor.setor.id}`, {
        nome: editingSetor.setor.nome,
        slug: editingSetor.setor.slug,
        tipo: editingSetor.setor.tipo,
        descricao: editingSetor.setor.descricao,
        ativo: editingSetor.setor.ativo,
        grupo_ad: editSetorGrupoAD.trim(),
      });
      const updated = res.data;
      setLocalidades((prev) =>
        prev.map((l) => {
          if (l.id === editingSetor.locId) {
            return {
              ...l,
              setores: l.setores?.map((s) => (s.id === updated.id ? updated : s)),
            };
          }
          return l;
        })
      );
      showToast({
        title: "Mapeamento salvo",
        description: `Grupo do AD vinculado ao setor ${updated.nome} com sucesso.`,
        tone: "success",
      });
      setEditingSetor(null);
    } catch (err: any) {
      showToast({
        title: "Erro ao atualizar grupo",
        description: err instanceof ApiError ? err.message : "Erro ao atualizar grupo do AD no setor.",
        tone: "danger",
      });
    } finally {
      setSavingSetor(false);
    }
  }

  return (
    <div className="flex flex-col gap-6">
      {/* PAINEL DIDÁTICO / EXPLICATIVO DO ACTIVE DIRECTORY */}
      <div className="rounded-xl border border-primary/20 bg-primary/5 p-5 dark:border-primary/30 dark:bg-primary/10">
        <div className="flex items-start gap-3">
          <div className="rounded-lg bg-primary p-2 text-white shrink-0">
            <ShieldCheck size={22} />
          </div>
          <div className="flex flex-col gap-2">
            <h2 className="text-base font-semibold text-foreground">
              Guia de Integração com o Active Directory (AD Municipal)
            </h2>
            <p className="text-xs text-muted leading-relaxed">
              O sistema de atendimento socioassistencial (SEMPRAS) sincroniza as permissões e lotações dos servidores públicos diretamente a partir dos grupos de segurança gerenciados no Active Directory da Prefeitura Municipal de Rondonópolis.
            </p>

            <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-4 mt-2">
              <div className="rounded-lg border border-surface-border bg-surface p-3 text-xs">
                <div className="font-semibold text-foreground flex items-center gap-1.5 mb-1 text-danger">
                  <ShieldAlert size={14} /> 1. Chave de Entrada
                </div>
                <p className="text-muted leading-relaxed">
                  Todo servidor <strong className="text-foreground">obrigatoriamente</strong> precisa do grupo <code className="font-mono bg-black/5 dark:bg-white/5 px-1 rounded font-semibold text-primary">Grupo_Assistencia_Social</code> para autenticar. Caso não pertença, o login é bloqueado.
                </p>
              </div>

              <div className="rounded-lg border border-surface-border bg-surface p-3 text-xs">
                <div className="font-semibold text-foreground flex items-center gap-1.5 mb-1 text-primary">
                  <Shield size={14} /> 2. Perfis de Função (RBAC)
                </div>
                <p className="text-muted leading-relaxed">
                  O nível de acesso (Técnico Social, Recepção, Gerência, Gestão) é concedido se o usuário pertencer ao grupo correspondente do AD listado na aba <strong>Perfis</strong>.
                </p>
              </div>

              <div className="rounded-lg border border-surface-border bg-surface p-3 text-xs">
                <div className="font-semibold text-foreground flex items-center gap-1.5 mb-1 text-emerald-600 dark:text-emerald-400">
                  <Building2 size={14} /> 3. Polos Autorizados
                </div>
                <p className="text-muted leading-relaxed">
                  O servidor pode acessar uma ou mais Unidades (CRAS, CREAS, Centro POP, Casa da Mulher) com base nos grupos de unidade cadastrados na aba <strong>Localidades</strong>.
                </p>
              </div>

              <div className="rounded-lg border border-surface-border bg-surface p-3 text-xs">
                <div className="font-semibold text-foreground flex items-center gap-1.5 mb-1 text-purple-600 dark:text-purple-400">
                  <Layers size={14} /> 4. Setores e Departamentos
                </div>
                <p className="text-muted leading-relaxed">
                  Departamentos específicos (Equipe Técnica, Recepção, Coordenação) podem ser amarrados a grupos individuais na aba <strong>Setores</strong>.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* CONTROLE DE SUB-ABAS */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between border-b border-surface-border pb-3">
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => {
              setActiveTab("perfis");
              setSearchTerm("");
            }}
            className={`rounded-lg px-3.5 py-2 text-xs font-semibold transition-colors flex items-center gap-1.5 ${
              activeTab === "perfis"
                ? "bg-primary text-white"
                : "bg-surface text-muted hover:text-foreground border border-surface-border"
            }`}
          >
            <Shield size={14} />
            Grupos AD por Perfil ({perfis.length})
          </button>

          <button
            type="button"
            onClick={() => {
              setActiveTab("localidades");
              setSearchTerm("");
            }}
            className={`rounded-lg px-3.5 py-2 text-xs font-semibold transition-colors flex items-center gap-1.5 ${
              activeTab === "localidades"
                ? "bg-primary text-white"
                : "bg-surface text-muted hover:text-foreground border border-surface-border"
            }`}
          >
            <Building2 size={14} />
            Grupos AD por Localidade ({localidades.length})
          </button>

          <button
            type="button"
            onClick={() => {
              setActiveTab("setores");
              setSearchTerm("");
            }}
            className={`rounded-lg px-3.5 py-2 text-xs font-semibold transition-colors flex items-center gap-1.5 ${
              activeTab === "setores"
                ? "bg-primary text-white"
                : "bg-surface text-muted hover:text-foreground border border-surface-border"
            }`}
          >
            <Layers size={14} />
            Grupos AD por Setor ({allSetores.length})
          </button>
        </div>

        <div className="relative max-w-xs w-full">
          <Search className="absolute left-3 top-2.5 h-3.5 w-3.5 text-muted" aria-hidden="true" />
          <input
            type="text"
            placeholder="Buscar por nome ou grupo AD..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full rounded-md border border-surface-border bg-surface py-1.5 pl-8 pr-3 text-xs text-foreground focus:border-primary focus:outline-none"
          />
        </div>
      </div>

      {/* 1. TABELA DE GRUPOS AD DOS PERFIS */}
      {activeTab === "perfis" && (
        <Table caption="Mapeamento de Perfis de Acesso ao Active Directory">
          <TableHead>
            <TableRow>
              <TableHeaderCell>Perfil no Sistema</TableHeaderCell>
              <TableHeaderCell>Nível Hierárquico</TableHeaderCell>
              <TableHeaderCell>Descrição das Atribuições</TableHeaderCell>
              <TableHeaderCell>Grupo no Active Directory (AD)</TableHeaderCell>
              <TableHeaderCell className="text-right">Ação</TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {perfis
              .filter(
                (p) =>
                  p.nome.toLowerCase().includes(searchTerm.toLowerCase()) ||
                  p.grupo_ad.toLowerCase().includes(searchTerm.toLowerCase()) ||
                  p.slug.toLowerCase().includes(searchTerm.toLowerCase())
              )
              .map((p) => (
                <TableRow key={p.id}>
                  <TableCell className="font-semibold text-foreground">
                    <div className="flex items-center gap-2">
                      <Shield size={16} className="text-primary" />
                      <span>{p.nome}</span>
                    </div>
                    <code className="text-[11px] text-muted font-mono">{p.slug}</code>
                  </TableCell>
                  <TableCell>
                    <span className="rounded-full bg-primary/10 px-2 py-0.5 text-xs font-semibold text-primary uppercase">
                      {p.nivel}
                    </span>
                  </TableCell>
                  <TableCell className="text-xs text-muted max-w-sm">{p.descricao || "—"}</TableCell>
                  <TableCell>
                    <span className="inline-flex items-center gap-1 rounded bg-black/5 dark:bg-white/5 px-2.5 py-1 text-xs font-mono font-bold text-foreground border border-surface-border">
                      <Users size={12} className="text-primary" />
                      {p.grupo_ad || "—"}
                    </span>
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={() => {
                        setEditingPerfil(p);
                        setEditPerfilGrupoAD(p.grupo_ad || "");
                      }}
                      className="flex items-center gap-1.5 text-xs ml-auto"
                    >
                      <Edit3 size={13} />
                      Alterar Grupo AD
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
          </TableBody>
        </Table>
      )}

      {/* 2. TABELA DE GRUPOS AD DAS LOCALIDADES */}
      {activeTab === "localidades" && (
        <Table caption="Mapeamento de Localidades e Polos ao Active Directory">
          <TableHead>
            <TableRow>
              <TableHeaderCell>Unidade / Polo</TableHeaderCell>
              <TableHeaderCell>Tipo</TableHeaderCell>
              <TableHeaderCell>Bairro / Endereço</TableHeaderCell>
              <TableHeaderCell>Grupo no Active Directory (AD)</TableHeaderCell>
              <TableHeaderCell className="text-right">Ação</TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {localidades
              .filter(
                (l) =>
                  l.nome.toLowerCase().includes(searchTerm.toLowerCase()) ||
                  (l.grupo_ad && l.grupo_ad.toLowerCase().includes(searchTerm.toLowerCase())) ||
                  (l.bairro && l.bairro.toLowerCase().includes(searchTerm.toLowerCase()))
              )
              .map((l) => (
                <TableRow key={l.id}>
                  <TableCell className="font-semibold text-foreground">
                    <div className="flex items-center gap-2">
                      <Building2 size={16} className="text-primary" />
                      <span>{l.nome}</span>
                    </div>
                    <code className="text-[11px] text-muted font-mono">{l.slug}</code>
                  </TableCell>
                  <TableCell>
                    <span className="rounded-full bg-surface-border/60 px-2 py-0.5 text-xs font-semibold text-foreground">
                      {l.tipo}
                    </span>
                  </TableCell>
                  <TableCell className="text-xs text-muted">
                    {l.bairro} {l.endereco && `· ${l.endereco}`}
                  </TableCell>
                  <TableCell>
                    <span className="inline-flex items-center gap-1 rounded bg-black/5 dark:bg-white/5 px-2.5 py-1 text-xs font-mono font-bold text-foreground border border-surface-border">
                      <Users size={12} className="text-primary" />
                      {l.grupo_ad || "—"}
                    </span>
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={() => {
                        setEditingLoc(l);
                        setEditLocGrupoAD(l.grupo_ad || "");
                      }}
                      className="flex items-center gap-1.5 text-xs ml-auto"
                    >
                      <Edit3 size={13} />
                      Alterar Grupo AD
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
          </TableBody>
        </Table>
      )}

      {/* 3. TABELA DE GRUPOS AD DOS SETORES */}
      {activeTab === "setores" && (
        <Table caption="Mapeamento de Setores e Departamentos ao Active Directory">
          <TableHead>
            <TableRow>
              <TableHeaderCell>Unidade / Polo</TableHeaderCell>
              <TableHeaderCell>Nome do Setor / Departamento</TableHeaderCell>
              <TableHeaderCell>Tipo de Atuação</TableHeaderCell>
              <TableHeaderCell>Grupo no Active Directory (AD)</TableHeaderCell>
              <TableHeaderCell className="text-right">Ação</TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {allSetores
              .filter(
                (s) =>
                  s.nome.toLowerCase().includes(searchTerm.toLowerCase()) ||
                  s.localidadeNome.toLowerCase().includes(searchTerm.toLowerCase()) ||
                  (s.grupo_ad && s.grupo_ad.toLowerCase().includes(searchTerm.toLowerCase()))
              )
              .map((s) => (
                <TableRow key={s.id}>
                  <TableCell className="font-medium text-foreground">
                    <div className="flex items-center gap-1.5 text-xs">
                      <Building2 size={13} className="text-muted" />
                      {s.localidadeNome}
                    </div>
                  </TableCell>
                  <TableCell className="font-semibold text-foreground">
                    <div>{s.nome}</div>
                    <code className="text-[10px] text-muted font-mono">{s.slug}</code>
                  </TableCell>
                  <TableCell>
                    <span
                      className={`rounded px-1.5 py-0.5 text-[10px] font-bold uppercase ${
                        s.tipo === "TECNICO"
                          ? "bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-300"
                          : s.tipo === "RECEPCAO"
                          ? "bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300"
                          : s.tipo === "GERENCIA"
                          ? "bg-purple-100 text-purple-800 dark:bg-purple-900/40 dark:text-purple-300"
                          : "bg-surface-border text-foreground"
                      }`}
                    >
                      {s.tipo}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className="inline-flex items-center gap-1 rounded bg-black/5 dark:bg-white/5 px-2.5 py-1 text-xs font-mono font-bold text-foreground border border-surface-border">
                      <Users size={12} className="text-primary" />
                      {s.grupo_ad || "—"}
                    </span>
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={() => {
                        setEditingSetor({ setor: s, locId: s.locId, localidadeNome: s.localidadeNome });
                        setEditSetorGrupoAD(s.grupo_ad || "");
                      }}
                      className="flex items-center gap-1.5 text-xs ml-auto"
                    >
                      <Edit3 size={13} />
                      Alterar Grupo AD
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
          </TableBody>
        </Table>
      )}

      {/* MODAL: EDITAR GRUPO AD DE PERFIL */}
      {editingPerfil && (
        <ModalShell open={!!editingPerfil} onClose={() => setEditingPerfil(null)}>
          <div className="flex flex-col gap-4 p-5">
            <h2 className="text-lg font-semibold text-foreground">
              Vincular Grupo AD ao Perfil: {editingPerfil.nome}
            </h2>
            <p className="text-xs text-muted">
              Ao informar o nome exato do grupo do Active Directory abaixo, qualquer usuário pertencente a esse grupo receberá este perfil automaticamente ao autenticar.
            </p>

            <form onSubmit={handleSavePerfilGrupoAD} className="flex flex-col gap-4">
              <Input
                label="Nome do Grupo no Active Directory (AD)"
                placeholder="Ex: Grupo_Tecnico_Social ou Grupo_AS_Gestao"
                value={editPerfilGrupoAD}
                onChange={(e) => setEditPerfilGrupoAD(e.target.value)}
                required
              />

              <div className="mt-4 flex justify-end gap-2 border-t border-surface-border pt-4">
                <Button type="button" variant="secondary" onClick={() => setEditingPerfil(null)}>
                  Cancelar
                </Button>
                <Button type="submit" loading={savingPerfil}>
                  Salvar Vínculo
                </Button>
              </div>
            </form>
          </div>
        </ModalShell>
      )}

      {/* MODAL: EDITAR GRUPO AD DE LOCALIDADE */}
      {editingLoc && (
        <ModalShell open={!!editingLoc} onClose={() => setEditingLoc(null)}>
          <div className="flex flex-col gap-4 p-5">
            <h2 className="text-lg font-semibold text-foreground">
              Vincular Grupo AD à Unidade: {editingLoc.nome}
            </h2>
            <p className="text-xs text-muted">
              Ao vincular este grupo do Active Directory, os servidores pertencentes a ele terão acesso liberado para visualizar e operar nesta localidade.
            </p>

            <form onSubmit={handleSaveLocGrupoAD} className="flex flex-col gap-4">
              <Input
                label="Nome do Grupo no Active Directory (AD)"
                placeholder="Ex: Grupo_CRAS_Ana_Carla"
                value={editLocGrupoAD}
                onChange={(e) => setEditLocGrupoAD(e.target.value)}
                required
              />

              <div className="mt-4 flex justify-end gap-2 border-t border-surface-border pt-4">
                <Button type="button" variant="secondary" onClick={() => setEditingLoc(null)}>
                  Cancelar
                </Button>
                <Button type="submit" loading={savingLoc}>
                  Salvar Vínculo
                </Button>
              </div>
            </form>
          </div>
        </ModalShell>
      )}

      {/* MODAL: EDITAR GRUPO AD DE SETOR */}
      {editingSetor && (
        <ModalShell open={!!editingSetor} onClose={() => setEditingSetor(null)}>
          <div className="flex flex-col gap-4 p-5">
            <h2 className="text-lg font-semibold text-foreground">
              Vincular Grupo AD ao Setor: {editingSetor.setor.nome}
            </h2>
            <p className="text-xs text-muted">
              Localidade: <strong className="text-foreground">{editingSetor.localidadeNome}</strong>. Servidores com esse grupo no AD serão automaticamente alocados neste setor específico.
            </p>

            <form onSubmit={handleSaveSetorGrupoAD} className="flex flex-col gap-4">
              <Input
                label="Nome do Grupo no Active Directory (AD)"
                placeholder="Ex: Grupo_CRAS_Ana_Carla_Tecnico"
                value={editSetorGrupoAD}
                onChange={(e) => setEditSetorGrupoAD(e.target.value)}
                required
              />

              <div className="mt-4 flex justify-end gap-2 border-t border-surface-border pt-4">
                <Button type="button" variant="secondary" onClick={() => setEditingSetor(null)}>
                  Cancelar
                </Button>
                <Button type="submit" loading={savingSetor}>
                  Salvar Vínculo
                </Button>
              </div>
            </form>
          </div>
        </ModalShell>
      )}
    </div>
  );
}
