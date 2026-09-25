"use client";

import React, { useEffect, useState, useMemo } from "react";
import { useSession } from "next-auth/react";
import { useSearchParams } from "next/navigation";
import { 
  Users, 
  Search, 
  Plus, 
  Clock, 
  CheckCircle2, 
  AlertCircle, 
  ShieldAlert, 
  Building2, 
  Sparkles, 
  RefreshCw, 
  HeartHandshake, 
  MapPin, 
  FileText,
  Filter,
  Eye,
  Phone,
  UserCheck,
  Calendar
} from "lucide-react";
import { Button } from "@/components/ui/Button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/Card";
import { Table, TableBody, TableCell, TableHead, TableHeaderCell, TableRow } from "@/components/ui/Table";
import { apiClient, ApiError } from "@/lib/api/client";
import { useToast } from "@/components/notifications/ToastProvider";
import { NovoAtendimentoModal } from "@/components/atendimento/NovoAtendimentoModal";
import { AtendimentoDetalhesModal } from "@/components/atendimento/AtendimentoDetalhesModal";
import { CentroPopProntuarioModal } from "@/components/atendimento/CentroPopProntuarioModal";
import type { 
  Atendimento, 
  AtendimentoStats, 
  CentroPopProntuario, 
  ServicePermissionInfo 
} from "@/components/atendimento/types";

const ALL_SERVICES = [
  { slug: "cras", label: "CRAS — Centro de Referência de Assistência Social" },
  { slug: "centro-pop", label: "Centro POP — População em Situação de Rua" },
  { slug: "creas", label: "CREAS — Centro Especializado de Assistência Social" },
  { slug: "casa-da-mulher", label: "Casa da Mulher — Atendimento Especializado à Mulher" },
  { slug: "conselho-tutelar", label: "Conselho Tutelar" },
  { slug: "cadunico-bolsa-familia", label: "Setor Cadastro Único & Bolsa Família" },
  { slug: "beneficios-eventuais", label: "Benefícios Eventuais" },
];

export default function AtendimentoWorkspacePage() {
  const { data: session } = useSession();
  const { showToast } = useToast();
  const searchParams = useSearchParams();
  const urlServiceParam = searchParams.get("servico") || searchParams.get("modulo");

  // User service permission info from backend
  const [serviceInfo, setServiceInfo] = useState<ServicePermissionInfo | null>(null);
  const [selectedService, setSelectedService] = useState<string>("cras");
  const [currentUnit, setCurrentUnit] = useState<string>("");

  // Views and Tabs
  const [activeView, setActiveView] = useState<"fila" | "prontuarios_pop">("fila");
  const [statusFilter, setStatusFilter] = useState<string>("todos");
  const [searchTerm, setSearchTerm] = useState<string>("");

  // Data states
  const [atendimentos, setAtendimentos] = useState<Atendimento[]>([]);
  const [stats, setStats] = useState<AtendimentoStats>({ total: 0, aguardando: 0, em_atendimento: 0, concluido: 0 });
  const [prontuarios, setProntuarios] = useState<CentroPopProntuario[]>([]);

  const [loading, setLoading] = useState<boolean>(true);
  const [refreshing, setRefreshing] = useState<boolean>(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Modals
  const [isNovoAtendimentoOpen, setIsNovoAtendimentoOpen] = useState(false);
  const [isNovoProntuarioOpen, setIsNovoProntuarioOpen] = useState(false);
  const [selectedAtendimento, setSelectedAtendimento] = useState<Atendimento | null>(null);

  // 1. Initial Load: Resolve User's Authorized Service from AD Groups
  useEffect(() => {
    async function loadUserServices() {
      try {
        const res = await apiClient.get<ServicePermissionInfo>("v1/servicos/meus-servicos");
        setServiceInfo(res.data);

        const availableSlugs = res.data.services?.map(s => s.slug) || res.data.allowed_services || [];
        const initialSlug = urlServiceParam && (res.data.is_admin || availableSlugs.includes(urlServiceParam))
          ? urlServiceParam
          : res.data.services?.[0]?.slug || res.data.service_slug || "cras";

        setSelectedService(initialSlug);
        setCurrentUnit(res.data.services?.[0]?.unit_default || res.data.default_unit || res.data.unit || "Unidade Central");
      } catch (err) {
        setErrorMessage("Erro ao identificar permissões do operador.");
      }
    }
    loadUserServices();
  }, [urlServiceParam]);

  // 2. Fetch Data whenever selectedService changes
  const fetchData = async () => {
    if (!selectedService) return;
    setRefreshing(true);
    setErrorMessage(null);

    try {
      // Fetch queue attendances
      const attendancesRes = await apiClient.get<Atendimento[]>(`v1/atendimentos?service=${selectedService}`);
      setAtendimentos(attendancesRes.data || []);

      // Fetch stats
      const statsRes = await apiClient.get<any>(`v1/atendimentos/stats?service=${selectedService}`);
      if (statsRes.data) {
        setStats({
          total: statsRes.data.today_total ?? statsRes.data.total ?? 0,
          aguardando: statsRes.data.today_waiting ?? statsRes.data.aguardando ?? 0,
          em_atendimento: statsRes.data.today_in_progress ?? statsRes.data.em_atendimento ?? 0,
          concluido: statsRes.data.today_completed ?? statsRes.data.concluido ?? 0,
        });
      }

      // If Centro POP, also fetch dossiers
      if (selectedService === "centro-pop" || selectedService === "centro_pop") {
        const prontuariosRes = await apiClient.get<CentroPopProntuario[]>("v1/centro-pop/prontuarios");
        setProntuarios(prontuariosRes.data || []);
      }
    } catch (err) {
      if (err instanceof ApiError && err.status === 403) {
        setErrorMessage("Acesso negado: seu perfil do Active Directory não possui acesso a este serviço socioassistencial.");
      } else {
        setErrorMessage("Falha ao carregar dados do atendimento.");
      }
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [selectedService]);

  // Filtered queue attendances
  const filteredAtendimentos = useMemo(() => {
    return atendimentos.filter((item) => {
      // Filter by status tab
      if (statusFilter !== "todos" && item.status !== statusFilter) {
        return false;
      }

      // Filter by search term (name, CPF, protocol)
      if (searchTerm.trim()) {
        const term = searchTerm.toLowerCase();
        const matchesName = (item.citizen_name || "").toLowerCase().includes(term);
        const matchesCpf = (item.citizen_cpf || "").toLowerCase().includes(term);
        const matchesProtocol = (item.protocol_number || item.protocol || "").toLowerCase().includes(term);
        const matchesBairro = (item.citizen_neighborhood || "").toLowerCase().includes(term);
        return matchesName || matchesCpf || matchesProtocol || matchesBairro;
      }

      return true;
    });
  }, [atendimentos, statusFilter, searchTerm]);

  // Filtered Centro POP dossiers
  const filteredProntuarios = useMemo(() => {
    if (!searchTerm.trim()) return prontuarios;
    const term = searchTerm.toLowerCase();
    return prontuarios.filter((p) => {
      return (
        (p.name || "").toLowerCase().includes(term) ||
        (p.preferred_name && p.preferred_name.toLowerCase().includes(term)) ||
        (p.nickname && p.nickname.toLowerCase().includes(term)) ||
        (p.cpf && p.cpf.toLowerCase().includes(term))
      );
    });
  }, [prontuarios, searchTerm]);

  // Quick Action: "Chamar" (Sets status to em_atendimento)
  const handleChamar = async (item: Atendimento) => {
    try {
      await apiClient.put(`v1/atendimentos/${item.id}`, {
        status: "em_atendimento",
        technical_notes: item.technical_notes,
        referrals: item.referrals,
      });
      showToast({
        title: "Atendimento iniciado",
        description: `Cidadão ${item.citizen_name || ""} chamado para atendimento.`,
        tone: "info",
      });
      fetchData();
    } catch (err: any) {
      showToast({
        title: "Erro ao chamar cidadão",
        description: err instanceof ApiError ? err.message : "Não foi possível alterar o status do atendimento.",
        tone: "danger",
      });
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "aguardando":
        return (
          <span className="inline-flex items-center gap-1.5 rounded-full bg-warning/10 px-2.5 py-0.5 text-xs font-semibold text-warning border border-warning/30">
            <Clock size={12} /> Aguardando
          </span>
        );
      case "em_atendimento":
        return (
          <span className="inline-flex items-center gap-1.5 rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-semibold text-primary border border-primary/30 animate-pulse">
            <Sparkles size={12} /> Em Atendimento
          </span>
        );
      case "concluido":
        return (
          <span className="inline-flex items-center gap-1.5 rounded-full bg-success/10 px-2.5 py-0.5 text-xs font-semibold text-success border border-success/30">
            <CheckCircle2 size={12} /> Concluído
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center rounded-full bg-surface-border px-2.5 py-0.5 text-xs font-medium text-muted">
            {status}
          </span>
        );
    }
  };

  const getRiskBadge = (risk: string) => {
    switch (risk?.toLowerCase()) {
      case "urgente":
        return <span className="text-[11px] font-bold text-danger uppercase tracking-wider">Urgente</span>;
      case "alto":
        return <span className="text-[11px] font-semibold text-danger">Alto</span>;
      case "médio":
      case "medio":
        return <span className="text-[11px] font-medium text-warning">Médio</span>;
      default:
        return <span className="text-[11px] text-muted">Baixo</span>;
    }
  };

  const formatTime = (iso: string) => {
    try {
      const d = new Date(iso);
      return d.toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" });
    } catch {
      return "--:--";
    }
  };

  return (
    <div className="flex flex-col gap-6 pb-12">
      {/* Header com Identificação de Unidade e Seletor de Serviço */}
      <header className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between border-b border-surface-border pb-6">
        <div className="flex flex-col gap-1">
          <div className="flex flex-wrap items-center gap-2">
            <span className="dateline">SEMPRAS · Prefeitura de Rondonópolis</span>
            <span className="inline-flex items-center gap-1 rounded-full bg-primary/10 px-2.5 py-0.5 text-[11px] font-semibold text-primary">
              <Building2 size={12} />
              {currentUnit || "Unidade Operacional"}
            </span>
            {serviceInfo?.is_admin ? (
              <span className="inline-flex items-center gap-1 rounded-full bg-emerald-500/10 px-2.5 py-0.5 text-[11px] font-bold text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
                🛡️ Gestão Geral (Acesso Total)
              </span>
            ) : serviceInfo?.is_receptionist ? (
              <span className="inline-flex items-center gap-1 rounded-full bg-amber-500/10 px-2.5 py-0.5 text-[11px] font-bold text-amber-700 dark:text-amber-400 border border-amber-500/20">
                📋 Recepção & Triagem
              </span>
            ) : (
              <span className="inline-flex items-center gap-1 rounded-full bg-primary/10 px-2.5 py-0.5 text-[11px] font-bold text-primary border border-primary/20">
                🩺 Técnico Social / Atendimento
              </span>
            )}
          </div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground">
            Central de Atendimento Socioassistencial
          </h1>
          <p className="text-xs sm:text-sm text-muted">
            Fila operacional, triagem técnica, histórico do cidadão e acompanhamento socioassistencial.
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2.5">
          {/* Seletor de Serviço (Habilitado para Administrador ou bloqueado para o grupo AD específico) */}
          <div className="flex items-center gap-2 bg-surface border border-surface-border rounded-xl px-3 py-1.5 shadow-sm">
            <span className="text-xs font-semibold text-muted">Serviço:</span>
            <select
              value={selectedService}
              onChange={(e) => {
                setSelectedService(e.target.value);
                setActiveView("fila");
              }}
              disabled={!serviceInfo?.is_admin && ((serviceInfo?.services?.length ?? serviceInfo?.allowed_services?.length ?? 0) <= 1)}
              className="bg-transparent text-xs font-bold text-foreground focus:outline-none cursor-pointer"
            >
              {ALL_SERVICES.map((srv) => {
                const allowedSlugs = serviceInfo?.services?.map(s => s.slug) || serviceInfo?.allowed_services;
                if (!serviceInfo?.is_admin && allowedSlugs && !allowedSlugs.includes(srv.slug)) {
                  return null;
                }
                return (
                  <option key={srv.slug} value={srv.slug}>
                    {srv.label.split("—")[0]?.trim() || srv.label}
                  </option>
                );
              })}
            </select>
          </div>

          <Button
            size="sm"
            variant="secondary"
            onClick={fetchData}
            disabled={refreshing}
            className="gap-1.5"
            title="Atualizar lista"
          >
            <RefreshCw size={14} className={refreshing ? "animate-spin" : ""} />
            Atualizar
          </Button>

          {(selectedService === "centro-pop" || selectedService === "centro_pop") && !serviceInfo?.is_receptionist && (
            <Button
              size="sm"
              variant="secondary"
              onClick={() => setIsNovoProntuarioOpen(true)}
              className="gap-1.5 border-primary/40 text-primary"
            >
              <HeartHandshake size={15} />
              Novo Prontuário POP
            </Button>
          )}

          <Button
            size="sm"
            onClick={() => setIsNovoAtendimentoOpen(true)}
            className="gap-2 shadow-sm font-semibold"
          >
            <Plus size={16} />
            {serviceInfo?.is_receptionist ? "Encaminhar Cidadão" : "Novo Atendimento"}
          </Button>
        </div>
      </header>

      {/* Alerta de Erro se houver */}
      {errorMessage && (
        <div className="flex items-center gap-2 rounded-xl border border-danger/40 bg-danger/10 p-4 text-xs text-danger">
          <AlertCircle size={18} className="shrink-0" />
          <span>{errorMessage}</span>
        </div>
      )}

      {/* Cards de Métricas do Dia */}
      <section className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div className="flex items-center gap-4 rounded-xl border border-surface-border bg-surface p-4 shadow-sm">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-primary/10 text-primary">
            <Users size={22} />
          </div>
          <div>
            <p className="text-xs font-semibold uppercase tracking-wider text-muted">Total Hoje</p>
            <p className="font-mono text-2xl font-bold text-foreground">{stats.total}</p>
            <p className="text-[11px] text-muted">Registros no serviço</p>
          </div>
        </div>

        <div className="flex items-center gap-4 rounded-xl border border-surface-border bg-surface p-4 shadow-sm">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-warning/10 text-warning">
            <Clock size={22} />
          </div>
          <div>
            <p className="text-xs font-semibold uppercase tracking-wider text-muted">Aguardando</p>
            <p className="font-mono text-2xl font-bold text-warning">{stats.aguardando}</p>
            <p className="text-[11px] text-muted">Na fila de espera</p>
          </div>
        </div>

        <div className="flex items-center gap-4 rounded-xl border border-surface-border bg-surface p-4 shadow-sm">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-primary/10 text-primary">
            <Sparkles size={22} />
          </div>
          <div>
            <p className="text-xs font-semibold uppercase tracking-wider text-muted">Em Atendimento</p>
            <p className="font-mono text-2xl font-bold text-primary">{stats.em_atendimento}</p>
            <p className="text-[11px] text-muted">Sendo acolhidos agora</p>
          </div>
        </div>

        <div className="flex items-center gap-4 rounded-xl border border-surface-border bg-surface p-4 shadow-sm">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-success/10 text-success">
            <CheckCircle2 size={22} />
          </div>
          <div>
            <p className="text-xs font-semibold uppercase tracking-wider text-muted">Concluídos</p>
            <p className="font-mono text-2xl font-bold text-success">{stats.concluido}</p>
            <p className="text-[11px] text-muted">Atendimentos finalizados</p>
          </div>
        </div>
      </section>

      {/* Alternância de Visualização para Centro POP (Apenas Técnicos e Gestores) */}
      {(selectedService === "centro-pop" || selectedService === "centro_pop") && !serviceInfo?.is_receptionist && (
        <div className="flex items-center gap-2 border-b border-surface-border pb-2">
          <button
            onClick={() => setActiveView("fila")}
            className={`flex items-center gap-2 px-4 py-2 text-xs font-bold rounded-lg transition-colors ${
              activeView === "fila"
                ? "bg-primary text-white"
                : "text-muted hover:bg-surface-hover hover:text-foreground"
            }`}
          >
            <Users size={15} />
            Fila de Atendimento do Dia ({atendimentos.length})
          </button>
          <button
            onClick={() => setActiveView("prontuarios_pop")}
            className={`flex items-center gap-2 px-4 py-2 text-xs font-bold rounded-lg transition-colors ${
              activeView === "prontuarios_pop"
                ? "bg-primary text-white"
                : "text-muted hover:bg-surface-hover hover:text-foreground"
            }`}
          >
            <HeartHandshake size={15} />
            Prontuários e Acompanhamento Centro POP ({prontuarios.length})
          </button>
        </div>
      )}

      {/* Visualização 1: Fila de Atendimento Operacional */}
      {activeView === "fila" && (
        <Card>
          <CardHeader className="pb-3 border-b border-surface-border">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              {/* Filtro por Abas de Status */}
              <div className="flex items-center gap-1.5">
                {[
                  { id: "todos", label: "Todos", count: atendimentos.length },
                  { id: "aguardando", label: "Aguardando", count: stats.aguardando },
                  { id: "em_atendimento", label: "Em Atendimento", count: stats.em_atendimento },
                  { id: "concluido", label: "Concluídos", count: stats.concluido },
                ].map((tab) => (
                  <button
                    key={tab.id}
                    onClick={() => setStatusFilter(tab.id)}
                    className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-lg transition-colors ${
                      statusFilter === tab.id
                        ? "bg-primary/15 text-primary border border-primary/30"
                        : "text-muted hover:bg-surface-hover hover:text-foreground"
                    }`}
                  >
                    <span>{tab.label}</span>
                    <span className="rounded-full bg-surface-border/80 px-1.5 py-0.2 text-[10px] font-mono">
                      {tab.count}
                    </span>
                  </button>
                ))}
              </div>

              {/* Barra de Busca */}
              <div className="relative min-w-[16rem]">
                <Search size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted" />
                <input
                  type="text"
                  placeholder="Buscar por cidadão, CPF ou protocolo..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-full rounded-lg border border-surface-border bg-surface pl-9 pr-3 py-1.5 text-xs text-foreground placeholder:text-muted focus:outline-none focus:ring-2 focus:ring-primary/40"
                />
              </div>
            </div>
          </CardHeader>

          <CardContent className="p-0">
            {loading ? (
              <div className="flex items-center justify-center p-12 text-xs text-muted">
                <RefreshCw size={18} className="animate-spin mr-2 text-primary" />
                Carregando atendimentos...
              </div>
            ) : filteredAtendimentos.length === 0 ? (
              <div className="flex flex-col items-center justify-center p-12 text-center">
                <Users size={36} className="text-muted/40 mb-2" />
                <p className="text-sm font-semibold text-foreground">Nenhum atendimento na fila com estes filtros</p>
                <p className="text-xs text-muted mt-1 max-w-sm">
                  Clique em &quot;Novo Atendimento&quot; para acolher um cidadão e registrar a demanda socioassistencial.
                </p>
                <Button 
                  size="sm" 
                  onClick={() => setIsNovoAtendimentoOpen(true)}
                  className="mt-4 gap-1.5"
                >
                  <Plus size={15} /> Registrar Atendimento
                </Button>
              </div>
            ) : (
              <>
                {/* Visualização Mobile em Cards Touch (Telas Pequenas / Celular) */}
                <div className="flex flex-col gap-3 md:hidden p-3.5">
                  {filteredAtendimentos.map((item) => (
                    <article 
                      key={item.id} 
                      className="rounded-xl border border-surface-border bg-surface p-3.5 shadow-sm space-y-2.5 transition-colors hover:border-primary/40"
                    >
                      <div className="flex items-center justify-between">
                        <span className="font-mono text-xs font-bold text-primary bg-primary/10 px-2 py-0.5 rounded">
                          {item.protocol_number || item.protocol}
                        </span>
                        <div className="flex items-center gap-1.5">
                          {getRiskBadge(item.risk_level)}
                          {getStatusBadge(item.status)}
                        </div>
                      </div>

                      <div>
                        <h4 className="font-semibold text-foreground text-sm leading-tight">{item.citizen_name}</h4>
                        <p className="text-[11px] font-mono text-muted mt-0.5">
                          {item.citizen_cpf ? `CPF: ${item.citizen_cpf}` : "CPF não informado"}
                          {item.citizen_neighborhood ? ` · ${item.citizen_neighborhood}` : ""}
                        </p>
                      </div>

                      <div className="rounded-lg bg-surface-hover/50 p-2.5 text-xs space-y-1">
                        <div className="flex items-center justify-between text-[11px]">
                          <span className="text-muted">Demanda:</span>
                          <span className="font-medium text-foreground truncate max-w-[200px]">{item.demand_type}</span>
                        </div>
                        <div className="flex items-center justify-between text-[11px]">
                          <span className="text-muted">Forma de Acesso:</span>
                          <span className="text-muted">{item.access_form}</span>
                        </div>
                        <div className="flex items-center justify-between text-[11px]">
                          <span className="text-muted">Horário:</span>
                          <span className="text-muted flex items-center gap-1">
                            <Clock size={11} /> {formatTime(item.created_at)}
                          </span>
                        </div>
                        {item.attendant_name && (
                          <div className="flex items-center justify-between text-[11px]">
                            <span className="text-muted">Técnico:</span>
                            <span className="text-foreground font-medium">{item.attendant_name}</span>
                          </div>
                        )}
                      </div>

                      <div className="flex items-center gap-2 pt-1">
                        {!serviceInfo?.is_receptionist && item.status === "aguardando" && (
                          <Button
                            size="sm"
                            variant="secondary"
                            onClick={() => handleChamar(item)}
                            className="flex-1 h-9 text-xs gap-1.5 text-primary border-primary/30"
                          >
                            <Sparkles size={13} /> Chamar
                          </Button>
                        )}
                        <Button
                          size="sm"
                          variant={!serviceInfo?.is_receptionist && item.status === "em_atendimento" ? "primary" : "secondary"}
                          onClick={() => setSelectedAtendimento(item)}
                          className="flex-1 h-9 text-xs gap-1.5"
                        >
                          <Eye size={13} />
                          {serviceInfo?.is_receptionist 
                            ? "Ficha de Triagem" 
                            : item.status === "em_atendimento" 
                              ? "Evolução Técnica" 
                              : "Ver Ficha"}
                        </Button>
                      </div>
                    </article>
                  ))}
                </div>

                {/* Visualização Desktop em Tabela Semântica */}
                <div className="hidden md:block overflow-x-auto">
                  <Table caption="Fila operacional de atendimento socioassistencial">
                  <TableHead>
                    <TableRow>
                      <TableHeaderCell>Protocolo</TableHeaderCell>
                      <TableHeaderCell>Cidadão / Usuário</TableHeaderCell>
                      <TableHeaderCell>Demanda & Forma de Acesso</TableHeaderCell>
                      <TableHeaderCell>Risco</TableHeaderCell>
                      <TableHeaderCell>Atendente / Horário</TableHeaderCell>
                      <TableHeaderCell>Status</TableHeaderCell>
                      <TableHeaderCell className="text-right">Ações</TableHeaderCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {filteredAtendimentos.map((item) => (
                      <TableRow key={item.id} className="hover:bg-surface-hover/50 transition-colors">
                        <TableCell className="font-mono text-xs font-bold text-primary">
                          {item.protocol_number || item.protocol}
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-col">
                            <span className="font-semibold text-foreground text-xs">{item.citizen_name}</span>
                            <span className="text-[11px] font-mono text-muted">
                              {item.citizen_cpf ? `CPF: ${item.citizen_cpf}` : "CPF não informado"}
                              {item.citizen_neighborhood ? ` · ${item.citizen_neighborhood}` : ""}
                            </span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-col">
                            <span className="text-xs font-medium text-foreground">{item.demand_type}</span>
                            <span className="text-[11px] text-muted">{item.access_form}</span>
                          </div>
                        </TableCell>
                        <TableCell>
                          {getRiskBadge(item.risk_level)}
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-col text-xs">
                            <span className="text-foreground font-medium">
                              {item.attendant_name || "Aguardando..."}
                            </span>
                            <span className="text-[11px] text-muted flex items-center gap-1">
                              <Clock size={11} /> {formatTime(item.created_at)}
                            </span>
                          </div>
                        </TableCell>
                        <TableCell>
                          {getStatusBadge(item.status)}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex items-center justify-end gap-1.5">
                            {item.status === "aguardando" && (
                              <Button
                                size="sm"
                                variant="secondary"
                                onClick={() => handleChamar(item)}
                                className="h-7 px-2 text-[11px] gap-1 text-primary border-primary/30 hover:bg-primary/10"
                                title="Chamar para atendimento"
                              >
                                <Sparkles size={12} /> Chamar
                              </Button>
                            )}

                            <Button
                              size="sm"
                              variant={item.status === "em_atendimento" ? "primary" : "secondary"}
                              onClick={() => setSelectedAtendimento(item)}
                              className="h-7 px-2.5 text-[11px] gap-1"
                            >
                              <Eye size={12} />
                              {item.status === "em_atendimento" ? "Evolução" : "Ficha"}
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </>
          )}
          </CardContent>
        </Card>
      )}

      {/* Visualização 2: Prontuários Especializados Centro POP */}
      {selectedService === "centro_pop" && activeView === "prontuarios_pop" && (
        <Card>
          <CardHeader className="pb-3 border-b border-surface-border">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <CardTitle className="text-base font-bold">
                  Prontuários e Histórico — Centro POP
                </CardTitle>
                <CardDescription className="text-xs text-muted">
                  Base histórica de acolhimento à população em situação de rua (dados integrados e migrados).
                </CardDescription>
              </div>

              <div className="relative min-w-[16rem]">
                <Search size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted" />
                <input
                  type="text"
                  placeholder="Buscar por nome, apelido, vulgo ou CPF..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-full rounded-lg border border-surface-border bg-surface pl-9 pr-3 py-1.5 text-xs text-foreground placeholder:text-muted focus:outline-none focus:ring-2 focus:ring-primary/40"
                />
              </div>
            </div>
          </CardHeader>

          <CardContent className="p-0">
            {filteredProntuarios.length === 0 ? (
              <div className="flex flex-col items-center justify-center p-12 text-center">
                <HeartHandshake size={36} className="text-muted/40 mb-2" />
                <p className="text-sm font-semibold text-foreground">Nenhum prontuário encontrado</p>
                <Button 
                  size="sm" 
                  onClick={() => setIsNovoProntuarioOpen(true)}
                  className="mt-4 gap-1.5"
                >
                  <Plus size={15} /> Abrir Prontuário POP
                </Button>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <Table caption="Prontuários especializados Centro POP">
                  <TableHead>
                    <TableRow>
                      <TableHeaderCell>Nome Civil / Registro</TableHeaderCell>
                      <TableHeaderCell>Nome Social / Apelido</TableHeaderCell>
                      <TableHeaderCell>Documento / CPF</TableHeaderCell>
                      <TableHeaderCell>Condição / Tempo de Rua</TableHeaderCell>
                      <TableHeaderCell>Ponto de Permanência</TableHeaderCell>
                      <TableHeaderCell>Fatores de Saúde</TableHeaderCell>
                      <TableHeaderCell className="text-right">Ação</TableHeaderCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {filteredProntuarios.map((p) => (
                      <TableRow key={p.id} className="hover:bg-surface-hover/50 transition-colors">
                        <TableCell>
                          <div className="flex flex-col">
                            <span className="font-semibold text-foreground text-xs">{p.name}</span>
                            {p.mother_name && (
                              <span className="text-[11px] text-muted">Mãe: {p.mother_name}</span>
                            )}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-col">
                            <span className="font-medium text-primary text-xs">
                              {p.preferred_name || p.nickname || "—"}
                            </span>
                            {p.nickname && p.preferred_name && (
                              <span className="text-[10px] text-muted">Apelido: {p.nickname}</span>
                            )}
                          </div>
                        </TableCell>
                        <TableCell className="font-mono text-xs text-foreground">
                          {p.cpf || "Sem CPF cadastrado"}
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-col text-xs">
                            <span className="font-medium capitalize text-foreground">{p.situation}</span>
                            <span className="text-[11px] text-muted">
                              {p.time_homelessness ? `${p.time_homelessness} meses` : "Não informado"}
                            </span>
                          </div>
                        </TableCell>
                        <TableCell className="text-xs text-muted max-w-[12rem] truncate">
                          {p.address || "Rondonópolis"}
                        </TableCell>
                        <TableCell>
                          {p.health_notes && (
                            <span className="text-[11px] font-medium text-warning bg-warning/10 px-2 py-0.5 rounded-md border border-warning/20">
                              {p.health_notes.substance_use || "Sem anotações"}
                            </span>
                          )}
                        </TableCell>
                        <TableCell className="text-right">
                          <Button
                            size="sm"
                            variant="secondary"
                            onClick={() => {
                              // Abre atendimento direto para este cidadão
                              setIsNovoAtendimentoOpen(true);
                            }}
                            className="h-7 px-2.5 text-[11px] gap-1"
                          >
                            <Users size={12} /> Atender
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Modais Operacionais */}
      <NovoAtendimentoModal
        isOpen={isNovoAtendimentoOpen}
        onClose={() => setIsNovoAtendimentoOpen(false)}
        onSuccess={fetchData}
        currentService={selectedService}
        currentUnit={currentUnit}
        isReceptionist={Boolean(serviceInfo?.is_receptionist)}
      />

      <AtendimentoDetalhesModal
        atendimento={selectedAtendimento}
        isOpen={!!selectedAtendimento}
        onClose={() => setSelectedAtendimento(null)}
        onUpdated={fetchData}
        isReceptionist={Boolean(serviceInfo?.is_receptionist)}
      />

      <CentroPopProntuarioModal
        isOpen={isNovoProntuarioOpen}
        onClose={() => setIsNovoProntuarioOpen(false)}
        onSuccess={fetchData}
      />
    </div>
  );
}
