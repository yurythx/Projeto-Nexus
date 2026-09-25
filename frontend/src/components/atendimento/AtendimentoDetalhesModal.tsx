"use client";

import React, { useState, useEffect } from "react";
import { 
  Users, 
  CheckCircle2, 
  AlertCircle, 
  X, 
  ShieldAlert, 
  FileText, 
  MapPin,
  Phone,
  Clock,
  Send,
  UserCheck,
  Building2,
  Printer
} from "lucide-react";
import { Button } from "@/components/ui/Button";
import { Textarea } from "@/components/ui/Textarea";
import { apiClient, ApiError } from "@/lib/api/client";
import type { Atendimento } from "./types";

interface AtendimentoDetalhesModalProps {
  atendimento: Atendimento | null;
  isOpen: boolean;
  onClose: () => void;
  onUpdated: () => void;
  isReceptionist?: boolean;
}

const REFERRAL_OPTIONS = [
  { id: "saude", label: "Rede de Saúde (UBS / UPA / CAPS)" },
  { id: "habitacao", label: "Habitação / Regularização Fundiária" },
  { id: "trabalho_sine", label: "SINE / Qualificação Profissional" },
  { id: "educacao_creche", label: "Educação / Vaga em Creche ou Escola" },
  { id: "conselho_tutelar", label: "Conselho Tutelar" },
  { id: "creas_paefi", label: "CREAS / Acompanhamento PAEFI" },
  { id: "cadunico", label: "Setor Cadastro Único / Bolsa Família" },
  { id: "defensoria_publica", label: "Defensoria Pública / Judiciário" },
];

export function AtendimentoDetalhesModal({
  atendimento,
  isOpen,
  onClose,
  onUpdated,
  isReceptionist = false,
}: AtendimentoDetalhesModalProps) {
  const [technicalNotes, setTechnicalNotes] = useState("");
  const [selectedReferrals, setSelectedReferrals] = useState<Record<string, boolean>>({});
  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  useEffect(() => {
    if (atendimento) {
      setTechnicalNotes(atendimento.technical_notes || "");
      setSelectedReferrals(atendimento.referrals || {});
      setErrorMessage(null);
      setSuccessMessage(null);
    }
  }, [atendimento]);

  if (!isOpen || !atendimento) return null;

  const toggleReferral = (id: string) => {
    setSelectedReferrals((prev) => ({
      ...prev,
      [id]: !prev[id],
    }));
  };

  const handleUpdateStatus = async (newStatus: string) => {
    setLoading(true);
    setErrorMessage(null);
    setSuccessMessage(null);

    try {
      await apiClient.put(`v1/atendimentos/${atendimento.id}`, {
        status: newStatus,
        technical_notes: technicalNotes.trim(),
        referrals: selectedReferrals,
      });

      setSuccessMessage(`Status atualizado para "${newStatus}" com sucesso.`);
      onUpdated();
      if (newStatus === "concluido") {
        setTimeout(() => {
          onClose();
        }, 1200);
      }
    } catch (err) {
      setErrorMessage(err instanceof ApiError ? err.message : "Erro ao atualizar atendimento.");
    } finally {
      setLoading(false);
    }
  };

  const handleSaveNotes = async () => {
    setLoading(true);
    setErrorMessage(null);
    setSuccessMessage(null);

    try {
      await apiClient.put(`v1/atendimentos/${atendimento.id}`, {
        status: atendimento.status,
        technical_notes: technicalNotes.trim(),
        referrals: selectedReferrals,
      });

      setSuccessMessage("Parecer técnico e encaminhamentos salvos com sucesso.");
      onUpdated();
    } catch (err) {
      setErrorMessage(err instanceof ApiError ? err.message : "Erro ao salvar anotações técnicas.");
    } finally {
      setLoading(false);
    }
  };

  const handlePrint = () => {
    window.print();
  };

  const formatDateTime = (iso: string) => {
    try {
      const d = new Date(iso);
      return d.toLocaleString("pt-BR", {
        day: "2-digit",
        month: "2-digit",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      });
    } catch {
      return iso;
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm">
      <div 
        role="dialog" 
        aria-modal="true" 
        className="flex max-h-[92vh] w-full max-w-4xl flex-col rounded-2xl border border-surface-border bg-surface shadow-2xl overflow-hidden"
      >
        {/* Header */}
        <div className="flex items-center justify-between border-b border-surface-border px-6 py-4 bg-surface-hover/30">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary/10 text-primary">
              <FileText size={22} />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-lg font-bold text-foreground">
                  Atendimento {atendimento.protocol}
                </h2>
                <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-bold uppercase tracking-wider ${
                  atendimento.status === "concluido"
                    ? "bg-success/10 text-success border border-success/30"
                    : atendimento.status === "em_atendimento"
                    ? "bg-primary/10 text-primary border border-primary/30"
                    : "bg-warning/10 text-warning border border-warning/30"
                }`}>
                  {atendimento.status.replace("_", " ")}
                </span>
              </div>
              <p className="text-xs text-muted flex items-center gap-2 mt-0.5">
                <span>Unidade: <strong>{atendimento.unit}</strong></span>
                <span>•</span>
                <span>Serviço: <strong className="uppercase">{atendimento.service_slug}</strong></span>
                <span>•</span>
                <span>Registrado em: {formatDateTime(atendimento.created_at)}</span>
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Button
              type="button"
              variant="secondary"
              size="sm"
              onClick={handlePrint}
              className="gap-1.5 text-xs"
            >
              <Printer size={14} /> Imprimir Ficha
            </Button>
            <button
              onClick={onClose}
              aria-label="Fechar"
              className="rounded-lg p-2 text-muted hover:bg-surface-border/50 hover:text-foreground transition-colors"
            >
              <X size={20} />
            </button>
          </div>
        </div>

        {/* Modal Body */}
        <div className="flex-1 overflow-y-auto px-6 py-5 space-y-6">
          {errorMessage && (
            <div className="flex items-center gap-2 rounded-xl border border-danger/40 bg-danger/10 p-3 text-xs text-danger">
              <AlertCircle size={16} className="shrink-0" />
              <span>{errorMessage}</span>
            </div>
          )}

          {successMessage && (
            <div className="flex items-center gap-2 rounded-xl border border-success/40 bg-success/10 p-3 text-xs text-success">
              <CheckCircle2 size={16} className="shrink-0" />
              <span>{successMessage}</span>
            </div>
          )}

          {/* Dados do Cidadão */}
          <div className="rounded-xl border border-surface-border bg-surface-hover/20 p-4">
            <h3 className="text-xs font-bold uppercase tracking-wider text-muted flex items-center gap-2 mb-3">
              <UserCheck size={16} className="text-primary" /> Ficha do Cidadão
            </h3>
            <div className="grid gap-3 sm:grid-cols-3 text-xs">
              <div>
                <span className="text-muted block">Nome Completo:</span>
                <span className="font-semibold text-foreground text-sm">{atendimento.citizen_name}</span>
              </div>
              <div>
                <span className="text-muted block">CPF:</span>
                <span className="font-mono font-medium text-foreground">{atendimento.citizen_cpf || "Não informado"}</span>
              </div>
              <div>
                <span className="text-muted block">RG:</span>
                <span className="font-mono text-foreground">{atendimento.citizen_rg || "Não informado"}</span>
              </div>
              <div>
                <span className="text-muted block">Telefone:</span>
                <span className="text-foreground">{atendimento.citizen_phone || "Não informado"}</span>
              </div>
              <div>
                <span className="text-muted block">Bairro:</span>
                <span className="text-foreground">{atendimento.citizen_neighborhood || "Não informado"}</span>
              </div>
              <div>
                <span className="text-muted block">Endereço:</span>
                <span className="text-foreground">{atendimento.citizen_address || "Não informado"}</span>
              </div>
            </div>
          </div>

          {/* Dados da Demanda */}
          <div className="grid gap-4 sm:grid-cols-3">
            <div className="rounded-xl border border-surface-border p-3 bg-surface">
              <span className="text-[11px] font-semibold uppercase text-muted block">Forma de Acesso</span>
              <span className="text-xs font-bold text-foreground mt-0.5 block">{atendimento.access_form}</span>
            </div>
            <div className="rounded-xl border border-surface-border p-3 bg-surface">
              <span className="text-[11px] font-semibold uppercase text-muted block">Demanda Principal</span>
              <span className="text-xs font-bold text-foreground mt-0.5 block">{atendimento.demand_type}</span>
            </div>
            <div className="rounded-xl border border-surface-border p-3 bg-surface">
              <span className="text-[11px] font-semibold uppercase text-muted block">Classificação de Risco</span>
              <span className="text-xs font-bold text-foreground mt-0.5 block">{atendimento.risk_level}</span>
            </div>
          </div>

          {/* Marcadores de Vulnerabilidade Registrados */}
          {atendimento.vulnerabilities && Object.keys(atendimento.vulnerabilities).length > 0 && (
            <div>
              <h4 className="text-xs font-bold uppercase tracking-wider text-muted flex items-center gap-1.5 mb-2">
                <ShieldAlert size={14} className="text-warning" /> Vulnerabilidades Identificadas
              </h4>
              <div className="flex flex-wrap gap-1.5">
                {Object.entries(atendimento.vulnerabilities).map(([key, val]) => {
                  if (!val) return null;
                  return (
                    <span 
                      key={key}
                      className="rounded-lg border border-warning/30 bg-warning/10 px-2.5 py-1 text-[11px] font-medium text-warning"
                    >
                      {key.replace(/_/g, " ").toUpperCase()}
                    </span>
                  );
                })}
              </div>
            </div>
          )}

          {/* Parecer Técnico e Evolução Social / Restrição Recepcionista */}
          {isReceptionist ? (
            <div className="rounded-xl border border-amber-500/30 bg-amber-500/10 p-4 text-xs text-amber-900 dark:text-amber-200 space-y-2">
              <div className="flex items-center gap-2 font-bold text-sm text-amber-800 dark:text-amber-300">
                <ShieldAlert size={18} />
                <span>Sigilo Socioassistencial & LGPD (Acesso Restrito)</span>
              </div>
              <p className="leading-relaxed">
                O parecer técnico, anotações de evolução e encaminhamentos especializados são restritos aos Técnicos de Nível Superior (Assistentes Sociais e Psicólogos credenciados). Como profissional de recepção/triagem, seus registros estão focados na acolhida inicial e encaminhamento à fila.
              </p>
              <div className="flex items-center gap-2 pt-1 font-semibold text-[11px] text-muted">
                <span>Status Atual na Fila:</span>
                <span className="font-mono uppercase text-foreground">{atendimento.status}</span>
              </div>
            </div>
          ) : (
            <>
              {/* Parecer Técnico e Evolução Social */}
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <h4 className="text-xs font-bold uppercase tracking-wider text-foreground flex items-center gap-1.5">
                    <FileText size={16} className="text-primary" /> Parecer Técnico & Evolução do Atendimento
                  </h4>
                  <span className="text-xs text-muted">
                    Profissional: <strong>{atendimento.attendant_name || "Aguardando acolhimento"}</strong>
                  </span>
                </div>
                <Textarea
                  rows={4}
                  placeholder="Registre a escuta qualificada, orientações fornecidas, benefícios concedidos ou avaliação técnica social..."
                  value={technicalNotes}
                  onChange={(e) => setTechnicalNotes(e.target.value)}
                  disabled={atendimento.status === "concluido"}
                />
              </div>

              {/* Encaminhamentos Intersetoriais */}
              <div className="space-y-3">
                <h4 className="text-xs font-bold uppercase tracking-wider text-muted flex items-center gap-1.5">
                  <Send size={14} className="text-primary" /> Encaminhamentos da Rede Socioassistencial
                </h4>
                <div className="grid gap-2 sm:grid-cols-2">
                  {REFERRAL_OPTIONS.map((item) => (
                    <label
                      key={item.id}
                      className={`flex items-center gap-2.5 rounded-lg border p-2 text-xs font-medium cursor-pointer transition-colors ${
                        selectedReferrals[item.id]
                          ? "border-primary bg-primary/10 text-primary"
                          : "border-surface-border bg-surface-hover/30 text-foreground hover:bg-surface-hover"
                      }`}
                    >
                      <input
                        type="checkbox"
                        checked={!!selectedReferrals[item.id]}
                        onChange={() => toggleReferral(item.id)}
                        disabled={atendimento.status === "concluido"}
                        className="h-4 w-4 rounded border-surface-border text-primary focus:ring-primary"
                      />
                      <span>{item.label}</span>
                    </label>
                  ))}
                </div>
              </div>
            </>
          )}
        </div>

        {/* Actions Footer */}
        <div className="flex flex-col-reverse sm:flex-row sm:items-center sm:justify-between gap-2 border-t border-surface-border px-6 py-4 bg-surface-hover/30">
          <div className="flex items-center gap-2">
            <Button
              type="button"
              variant="secondary"
              onClick={onClose}
            >
              Fechar
            </Button>
            <Button
              type="button"
              variant="secondary"
              onClick={handlePrint}
              className="gap-1.5"
            >
              <Printer size={15} />
              Imprimir {isReceptionist ? "Comprovante" : "Ficha Completa"}
            </Button>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            {!isReceptionist && atendimento.status === "aguardando" && (
              <Button
                type="button"
                onClick={() => handleUpdateStatus("em_atendimento")}
                disabled={loading}
                className="gap-2"
              >
                <Clock size={16} />
                Chamar / Iniciar Atendimento
              </Button>
            )}

            {!isReceptionist && atendimento.status === "em_atendimento" && (
              <>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={handleSaveNotes}
                  disabled={loading}
                >
                  Salvar Evolução
                </Button>
                <Button
                  type="button"
                  onClick={() => handleUpdateStatus("concluido")}
                  disabled={loading}
                  className="gap-2 bg-success hover:bg-success/90 text-white"
                >
                  <CheckCircle2 size={16} />
                  Finalizar Atendimento com Parecer
                </Button>
              </>
            )}

            {atendimento.status === "concluido" && (
              <span className="text-xs font-semibold text-success flex items-center gap-1.5">
                <CheckCircle2 size={16} /> Atendimento Finalizado e Arquivado no Histórico
              </span>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
