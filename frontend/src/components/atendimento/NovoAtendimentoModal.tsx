"use client";

import React, { useState } from "react";
import { 
  Users, 
  Search, 
  CheckCircle2, 
  AlertCircle, 
  X, 
  ShieldAlert, 
  FileText, 
  Sparkles,
  MapPin,
  Phone,
  UserCheck
} from "lucide-react";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Textarea } from "@/components/ui/Textarea";
import { apiClient, ApiError } from "@/lib/api/client";
import type { CitizenSearchResult } from "./types";

interface NovoAtendimentoModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
  currentService: string;
  currentUnit: string;
  isReceptionist?: boolean;
}

const COMMON_DEMANDS = [
  "Acolhida e Atendimento Geral",
  "Auxílio Alimentação / Cesta Básica",
  "Atualização / Inscrição no Cadastro Único",
  "Auxílio Natalidade",
  "Auxílio Funeral",
  "Segunda Via de Documentos Civis",
  "Acolhimento / Pernoite / Higiene",
  "Conflito Familiar / Orientação Psicossocial",
  "Denúncia / Suspeita de Violação de Direitos",
];

const ACCESS_FORMS = [
  "Demanda Espontânea",
  "Busca Ativa",
  "Encaminhamento CRAS / CREAS",
  "Encaminhamento Unidade de Saúde (SUS)",
  "Encaminhamento Conselho Tutelar",
  "Encaminhamento Poder Judiciário / Defensoria",
];

const VULNERABILITY_OPTIONS = [
  { id: "extrema_pobreza", label: "Extrema Pobreza / Sem Renda Fixa" },
  { id: "bolsa_familia", label: "Beneficiário Programa Bolsa Família" },
  { id: "bpc_loas", label: "Beneficiário BPC / LOAS" },
  { id: "situacao_rua", label: "População em Situação de Rua" },
  { id: "sem_documentos", label: "Ausência de Documentação Civil" },
  { id: "trabalho_infantil", label: "Suspeita ou Indício de Trabalho Infantil" },
  { id: "violencia_domestica", label: "Vítima de Violência Doméstica" },
  { id: "substancias_psicoativas", label: "Uso Abusivo de Álcool / Substâncias" },
  { id: "idoso_vulneravel", label: "Idoso em Situação de Vulnerabilidade" },
  { id: "pcd_sem_apoio", label: "PCD sem Rede de Apoio" },
];

export function NovoAtendimentoModal({
  isOpen,
  onClose,
  onSuccess,
  currentService,
  currentUnit,
  isReceptionist = false,
}: NovoAtendimentoModalProps) {
  const [cpfSearch, setCpfSearch] = useState("");
  const [searchingCpf, setSearchingCpf] = useState(false);
  const [searchHistoryNotice, setSearchHistoryNotice] = useState<string | null>(null);

  const [citizenName, setCitizenName] = useState("");
  const [citizenCpf, setCitizenCpf] = useState("");
  const [citizenRg, setCitizenRg] = useState("");
  const [citizenPhone, setCitizenPhone] = useState("");
  const [citizenNeighborhood, setCitizenNeighborhood] = useState("");
  const [citizenAddress, setCitizenAddress] = useState("");
  
  const [accessForm, setAccessForm] = useState("Demanda Espontânea");
  const [demandType, setDemandType] = useState("Acolhida e Atendimento Geral");
  const [riskLevel, setRiskLevel] = useState("Baixo");
  const [technicalNotes, setTechnicalNotes] = useState("");

  const [selectedVulnerabilities, setSelectedVulnerabilities] = useState<Record<string, boolean>>({});
  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleCpfLookup = async () => {
    const clean = cpfSearch.replace(/\D/g, "");
    if (!clean) return;

    setSearchingCpf(true);
    setSearchHistoryNotice(null);
    setErrorMessage(null);

    try {
      const res = await apiClient.get<CitizenSearchResult>(`v1/atendimentos/busca-cidadao/${clean}`);
      if (res.data) {
        const last = res.data.last_record;
        const total = res.data.history?.length ?? res.data.total_attendances ?? (last ? 1 : 0);
        if (last) {
          setCitizenCpf(last.citizen_cpf || clean);
          setCitizenName(last.citizen_name || "");
          setCitizenRg(last.citizen_rg || "");
          setCitizenPhone(last.citizen_phone || "");
          setCitizenNeighborhood(last.citizen_neighborhood || "");
          setCitizenAddress(last.citizen_address || "");
        } else if (res.data.citizen_name) {
          setCitizenCpf(res.data.citizen_cpf || clean);
          setCitizenName(res.data.citizen_name || "");
          setCitizenRg(res.data.citizen_rg || "");
          setCitizenPhone(res.data.citizen_phone || "");
          setCitizenNeighborhood(res.data.citizen_neighborhood || "");
          setCitizenAddress(res.data.citizen_address || "");
        } else {
          setCitizenCpf(clean);
        }

        if (total > 0) {
          setSearchHistoryNotice(
            `Cidadão localizado no histórico: ${total} atendimento(s) anterior(es) registrado(s).`
          );
        } else {
          setSearchHistoryNotice("Nenhum atendimento anterior encontrado. Cadastrando novo acolhimento.");
        }
      }
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        setSearchHistoryNotice("Cidadão não encontrado no histórico. Preencha os dados abaixo.");
        setCitizenCpf(clean);
      } else {
        setErrorMessage("Erro ao consultar base de dados do cidadão.");
      }
    } finally {
      setSearchingCpf(false);
    }
  };

  const toggleVulnerability = (id: string) => {
    setSelectedVulnerabilities((prev) => ({
      ...prev,
      [id]: !prev[id],
    }));
  };

  const handleSubmit = async (targetStatus: "aguardando" | "em_atendimento") => {
    if (!citizenName.trim()) {
      setErrorMessage("O Nome do Cidadão é obrigatório para registrar o atendimento.");
      return;
    }

    setLoading(true);
    setErrorMessage(null);

    try {
      await apiClient.post("v1/atendimentos", {
        service_slug: currentService,
        unit: currentUnit,
        citizen_name: citizenName.trim(),
        citizen_cpf: citizenCpf.trim(),
        citizen_rg: citizenRg.trim(),
        citizen_phone: citizenPhone.trim(),
        citizen_neighborhood: citizenNeighborhood.trim(),
        citizen_address: citizenAddress.trim(),
        access_form: accessForm,
        demand_type: demandType,
        risk_level: riskLevel,
        vulnerabilities: selectedVulnerabilities,
        technical_notes: technicalNotes.trim(),
        status: targetStatus,
      });

      onSuccess();
      onClose();
    } catch (err) {
      setErrorMessage(err instanceof ApiError ? err.message : "Falha ao registrar atendimento.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm">
      <div 
        role="dialog" 
        aria-modal="true" 
        aria-labelledby="modal-title"
        className="flex max-h-[92vh] w-full max-w-3xl flex-col rounded-2xl border border-surface-border bg-surface shadow-2xl overflow-hidden"
      >
        {/* Header */}
        <div className="flex items-center justify-between border-b border-surface-border px-6 py-4 bg-surface-hover/30">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary/10 text-primary">
              <Users size={22} />
            </div>
            <div>
              <h2 id="modal-title" className="text-lg font-bold text-foreground">
                Registrar Novo Atendimento Socioassistencial
              </h2>
              <p className="text-xs text-muted">
                Unidade: <strong className="text-foreground">{currentUnit || "Geral"}</strong> · Serviço: <strong className="text-foreground uppercase">{currentService}</strong>
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            aria-label="Fechar formulário de atendimento"
            className="rounded-lg p-2 text-muted hover:bg-surface-border/50 hover:text-foreground transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        {/* Form Body */}
        <div className="flex-1 overflow-y-auto px-6 py-5 space-y-6">
          {errorMessage && (
            <div className="flex items-center gap-2 rounded-xl border border-danger/40 bg-danger/10 p-3 text-xs text-danger">
              <AlertCircle size={16} className="shrink-0" />
              <span>{errorMessage}</span>
            </div>
          )}

          {/* Busca Rápida por CPF */}
          <div className="rounded-xl border border-primary/20 bg-primary/5 p-4">
            <label className="text-xs font-bold uppercase tracking-wider text-primary flex items-center gap-2 mb-2">
              <Search size={14} /> Localização de Cidadão por CPF (Histórico na Rede)
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                placeholder="Digite o CPF para autocompletar (somente números)..."
                value={cpfSearch}
                onChange={(e) => setCpfSearch(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && (e.preventDefault(), handleCpfLookup())}
                className="flex-1 rounded-lg border border-surface-border bg-surface px-3 py-2 text-sm text-foreground placeholder:text-muted focus:outline-none focus:ring-2 focus:ring-primary/40"
              />
              <Button 
                type="button" 
                size="sm" 
                onClick={handleCpfLookup} 
                disabled={searchingCpf}
                className="gap-2 shrink-0"
              >
                {searchingCpf ? "Buscando..." : "Consultar"}
              </Button>
            </div>
            {searchHistoryNotice && (
              <p className="mt-2 text-xs font-medium text-primary flex items-center gap-1.5">
                <CheckCircle2 size={14} className="text-success" />
                {searchHistoryNotice}
              </p>
            )}
          </div>

          {/* Dados Pessoais do Cidadão */}
          <div className="space-y-4">
            <h3 className="text-xs font-bold uppercase tracking-wider text-muted flex items-center gap-2">
              <UserCheck size={16} className="text-primary" /> Identificação do Usuário
            </h3>

            <div className="grid gap-3 sm:grid-cols-2">
              <div className="sm:col-span-2">
                <Input
                  label="Nome Completo *"
                  required
                  placeholder="Nome do cidadão atendido"
                  value={citizenName}
                  onChange={(e) => setCitizenName(e.target.value)}
                />
              </div>

              <div>
                <Input
                  label="CPF"
                  placeholder="000.000.000-00"
                  value={citizenCpf}
                  onChange={(e) => setCitizenCpf(e.target.value)}
                />
              </div>

              <div>
                <Input
                  label="RG / Documento"
                  placeholder="Número do documento"
                  value={citizenRg}
                  onChange={(e) => setCitizenRg(e.target.value)}
                />
              </div>

              <div>
                <Input
                  label="Telefone / WhatsApp"
                  placeholder="(66) 90000-0000"
                  value={citizenPhone}
                  onChange={(e) => setCitizenPhone(e.target.value)}
                />
              </div>

              <div>
                <Input
                  label="Bairro"
                  placeholder="Ex: Vila Operária, Jardim Atlântico"
                  value={citizenNeighborhood}
                  onChange={(e) => setCitizenNeighborhood(e.target.value)}
                />
              </div>

              <div className="sm:col-span-2">
                <Input
                  label="Endereço / Ponto de Referência"
                  placeholder="Rua, número, complemento ou ponto de referência"
                  value={citizenAddress}
                  onChange={(e) => setCitizenAddress(e.target.value)}
                />
              </div>
            </div>
          </div>

          {/* Parâmetros do Atendimento */}
          <div className="space-y-4">
            <h3 className="text-xs font-bold uppercase tracking-wider text-muted flex items-center gap-2">
              <FileText size={16} className="text-primary" /> Demanda e Classificação do Atendimento
            </h3>

            <div className="grid gap-3 sm:grid-cols-3">
              <div>
                <label className="text-xs font-semibold text-foreground mb-1.5 block">
                  Forma de Acesso
                </label>
                <select
                  value={accessForm}
                  onChange={(e) => setAccessForm(e.target.value)}
                  className="w-full rounded-lg border border-surface-border bg-surface px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40"
                >
                  {ACCESS_FORMS.map((form) => (
                    <option key={form} value={form}>{form}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="text-xs font-semibold text-foreground mb-1.5 block">
                  Demanda Principal
                </label>
                <select
                  value={demandType}
                  onChange={(e) => setDemandType(e.target.value)}
                  className="w-full rounded-lg border border-surface-border bg-surface px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40"
                >
                  {COMMON_DEMANDS.map((demand) => (
                    <option key={demand} value={demand}>{demand}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="text-xs font-semibold text-foreground mb-1.5 block">
                  Nível de Risco / Prioridade
                </label>
                <select
                  value={riskLevel}
                  onChange={(e) => setRiskLevel(e.target.value)}
                  className="w-full rounded-lg border border-surface-border bg-surface px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40"
                >
                  <option value="Baixo">Baixo (Atendimento Padrão)</option>
                  <option value="Médio">Médio (Prioridade Lei/Idoso/Gestante)</option>
                  <option value="Alto">Alto (Situação Crítica)</option>
                  <option value="Urgente">Urgente (Risco Iminente / Violação)</option>
                </select>
              </div>
            </div>
          </div>

          {/* Marcadores de Vulnerabilidade */}
          <div className="space-y-3">
            <h3 className="text-xs font-bold uppercase tracking-wider text-muted flex items-center gap-2">
              <ShieldAlert size={16} className="text-warning" /> Marcadores de Vulnerabilidade e Risco
            </h3>
            <div className="grid gap-2 sm:grid-cols-2">
              {VULNERABILITY_OPTIONS.map((item) => (
                <label
                  key={item.id}
                  className={`flex items-center gap-2.5 rounded-lg border p-2.5 text-xs font-medium cursor-pointer transition-colors ${
                    selectedVulnerabilities[item.id]
                      ? "border-primary bg-primary/10 text-primary"
                      : "border-surface-border bg-surface-hover/30 text-foreground hover:bg-surface-hover"
                  }`}
                >
                  <input
                    type="checkbox"
                    checked={!!selectedVulnerabilities[item.id]}
                    onChange={() => toggleVulnerability(item.id)}
                    className="h-4 w-4 rounded border-surface-border text-primary focus:ring-primary"
                  />
                  <span>{item.label}</span>
                </label>
              ))}
            </div>
          </div>

          {/* Relato Inicial / Parecer */}
          <div className="space-y-2">
            <label className="text-xs font-semibold text-foreground block">
              Relato Inicial do Cidadão / Parecer Técnico
            </label>
            <Textarea
              rows={3}
              placeholder="Descreva o motivo informado pelo cidadão, observações do acolhimento ou encaminhamentos prévios..."
              value={technicalNotes}
              onChange={(e) => setTechnicalNotes(e.target.value)}
            />
          </div>
        </div>

        {/* Footer Actions */}
        <div className="flex flex-col-reverse sm:flex-row sm:items-center sm:justify-end gap-2 border-t border-surface-border px-6 py-4 bg-surface-hover/30">
          <Button
            type="button"
            variant="secondary"
            onClick={onClose}
            disabled={loading}
          >
            Cancelar
          </Button>

          {isReceptionist ? (
            <Button
              type="button"
              onClick={() => handleSubmit("aguardando")}
              disabled={loading}
              className="gap-2 font-semibold"
            >
              <CheckCircle2 size={16} />
              {loading ? "Encaminhando..." : "Encaminhar para Fila de Espera"}
            </Button>
          ) : (
            <>
              <Button
                type="button"
                variant="secondary"
                onClick={() => handleSubmit("aguardando")}
                disabled={loading}
                className="gap-2"
              >
                Adicionar à Fila de Espera
              </Button>

              <Button
                type="button"
                onClick={() => handleSubmit("em_atendimento")}
                disabled={loading}
                className="gap-2"
              >
                <Sparkles size={16} />
                {loading ? "Registrando..." : "Iniciar Atendimento Agora"}
              </Button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
