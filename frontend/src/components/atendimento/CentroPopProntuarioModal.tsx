"use client";

import React, { useState } from "react";
import { 
  HeartHandshake, 
  Search, 
  CheckCircle2, 
  AlertCircle, 
  X, 
  User, 
  MapPin, 
  Calendar,
  Activity,
  Plus
} from "lucide-react";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Textarea } from "@/components/ui/Textarea";
import { apiClient, ApiError } from "@/lib/api/client";
import type { CentroPopProntuario } from "./types";

interface CentroPopProntuarioModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export function CentroPopProntuarioModal({
  isOpen,
  onClose,
  onSuccess,
}: CentroPopProntuarioModalProps) {
  const [name, setName] = useState("");
  const [preferredName, setPreferredName] = useState("");
  const [nickname, setNickname] = useState("");
  const [cpf, setCpf] = useState("");
  const [rg, setRg] = useState("");
  const [dateBirth, setDateBirth] = useState("");
  const [motherName, setMotherName] = useState("");
  const [fatherName, setFatherName] = useState("");
  const [sex, setSex] = useState("M");
  const [gender, setGender] = useState("Cisgênero");
  const [situation, setSituation] = useState("rua");
  const [timeHomelessness, setTimeHomelessness] = useState<string>("12");
  const [reasonHomelessness, setReasonHomelessness] = useState("");
  const [address, setAddress] = useState("");

  // Health notes
  const [substanceUse, setSubstanceUse] = useState("Nenhum");
  const [chronicCondition, setChronicCondition] = useState("");
  const [mentalHealth, setMentalHealth] = useState("");

  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setErrorMessage("O Nome do usuário é obrigatório.");
      return;
    }

    setLoading(true);
    setErrorMessage(null);

    try {
      await apiClient.post("v1/centro-pop/prontuarios", {
        name: name.trim(),
        preferred_name: preferredName.trim(),
        nickname: nickname.trim(),
        cpf: cpf.trim(),
        rg: rg.trim(),
        date_birth: dateBirth ? dateBirth : null,
        mother_name: motherName.trim(),
        father_name: fatherName.trim(),
        sex,
        gender,
        situation,
        time_homelessness: timeHomelessness ? parseFloat(timeHomelessness) : null,
        reason_homelessness: reasonHomelessness.trim(),
        address: address.trim(),
        health_notes: {
          substance_use: substanceUse,
          chronic_condition: chronicCondition,
          mental_health: mentalHealth,
        },
      });

      onSuccess();
      onClose();
    } catch (err) {
      setErrorMessage(err instanceof ApiError ? err.message : "Erro ao abrir prontuário Centro POP.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm">
      <div 
        role="dialog" 
        aria-modal="true" 
        className="flex max-h-[92vh] w-full max-w-3xl flex-col rounded-2xl border border-surface-border bg-surface shadow-2xl overflow-hidden"
      >
        <div className="flex items-center justify-between border-b border-surface-border px-6 py-4 bg-surface-hover/30">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary/10 text-primary">
              <HeartHandshake size={22} />
            </div>
            <div>
              <h2 className="text-lg font-bold text-foreground">
                Novo Prontuário Especializado — Centro POP
              </h2>
              <p className="text-xs text-muted">
                Acolhimento da População em Situação de Rua · SEMPRAS
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            aria-label="Fechar"
            className="rounded-lg p-2 text-muted hover:bg-surface-border/50 hover:text-foreground transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto px-6 py-5 space-y-6">
          {errorMessage && (
            <div className="flex items-center gap-2 rounded-xl border border-danger/40 bg-danger/10 p-3 text-xs text-danger">
              <AlertCircle size={16} className="shrink-0" />
              <span>{errorMessage}</span>
            </div>
          )}

          {/* Identificação Geral */}
          <div className="space-y-4">
            <h3 className="text-xs font-bold uppercase tracking-wider text-muted flex items-center gap-2">
              <User size={16} className="text-primary" /> Identificação e Vínculo
            </h3>

            <div className="grid gap-3 sm:grid-cols-3">
              <div className="sm:col-span-2">
                <Input
                  label="Nome de Registro *"
                  required
                  placeholder="Nome civil completo"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              </div>

              <div>
                <Input
                  label="Nome Social / Preferido"
                  placeholder="Nome que prefere ser chamado"
                  value={preferredName}
                  onChange={(e) => setPreferredName(e.target.value)}
                />
              </div>

              <div>
                <Input
                  label="Apelido / Vulgo de Rua"
                  placeholder="Ex: Baiano, Índio, Paulista"
                  value={nickname}
                  onChange={(e) => setNickname(e.target.value)}
                />
              </div>

              <div>
                <Input
                  label="CPF"
                  placeholder="000.000.000-00 (opcional)"
                  value={cpf}
                  onChange={(e) => setCpf(e.target.value)}
                />
              </div>

              <div>
                <Input
                  label="RG"
                  placeholder="Número de identidade"
                  value={rg}
                  onChange={(e) => setRg(e.target.value)}
                />
              </div>

              <div>
                <Input
                  label="Data de Nascimento"
                  type="date"
                  value={dateBirth}
                  onChange={(e) => setDateBirth(e.target.value)}
                />
              </div>

              <div className="sm:col-span-2">
                <Input
                  label="Nome da Mãe"
                  placeholder="Nome materno para busca CadÚnico/SINAN"
                  value={motherName}
                  onChange={(e) => setMotherName(e.target.value)}
                />
              </div>
            </div>
          </div>

          {/* Histórico na Rua */}
          <div className="space-y-4">
            <h3 className="text-xs font-bold uppercase tracking-wider text-muted flex items-center gap-2">
              <MapPin size={16} className="text-primary" /> Condição de Moradia e Tempo de Rua
            </h3>

            <div className="grid gap-3 sm:grid-cols-3">
              <div>
                <label className="text-xs font-semibold text-foreground mb-1.5 block">
                  Situação Habitacional
                </label>
                <select
                  value={situation}
                  onChange={(e) => setSituation(e.target.value)}
                  className="w-full rounded-lg border border-surface-border bg-surface px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40"
                >
                  <option value="rua">Totalmente em Situação de Rua</option>
                  <option value="abrigo">Acolhimento Institucional / Albergue</option>
                  <option value="ocupacao">Ocupação / Barraco Provisório</option>
                  <option value="itinerante">Trecheiro / Itinerante / Migrante</option>
                </select>
              </div>

              <div>
                <Input
                  label="Tempo em Situação de Rua (meses)"
                  type="number"
                  placeholder="Ex: 6, 12, 24"
                  value={timeHomelessness}
                  onChange={(e) => setTimeHomelessness(e.target.value)}
                />
              </div>

              <div>
                <Input
                  label="Local / Ponto de Permanência"
                  placeholder="Ex: Praça Brasil, Rodoviária, Cascalheira"
                  value={address}
                  onChange={(e) => setAddress(e.target.value)}
                />
              </div>

              <div className="sm:col-span-3">
                <label className="text-xs font-semibold text-foreground mb-1.5 block">
                  Motivo Principal da Ida para a Rua
                </label>
                <input
                  type="text"
                  placeholder="Ex: Ruptura de vínculos familiares, desemprego extremo, dependência química..."
                  value={reasonHomelessness}
                  onChange={(e) => setReasonHomelessness(e.target.value)}
                  className="w-full rounded-lg border border-surface-border bg-surface px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40"
                />
              </div>
            </div>
          </div>

          {/* Fatores de Saúde */}
          <div className="space-y-4">
            <h3 className="text-xs font-bold uppercase tracking-wider text-muted flex items-center gap-2">
              <Activity size={16} className="text-primary" /> Fatores de Saúde e Substâncias
            </h3>

            <div className="grid gap-3 sm:grid-cols-3">
              <div>
                <label className="text-xs font-semibold text-foreground mb-1.5 block">
                  Uso de Substâncias Psicoativas
                </label>
                <select
                  value={substanceUse}
                  onChange={(e) => setSubstanceUse(e.target.value)}
                  className="w-full rounded-lg border border-surface-border bg-surface px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40"
                >
                  <option value="Nenhum">Nenhum / Nega uso</option>
                  <option value="Alcool">Álcool</option>
                  <option value="Tabaco">Tabaco</option>
                  <option value="Crack / Cocaina">Crack / Cocaína</option>
                  <option value="Maconha">Maconha</option>
                  <option value="Poliusuario">Poliusuário (Múltiplas substâncias)</option>
                </select>
              </div>

              <div>
                <Input
                  label="Condição Crônica / Doenças"
                  placeholder="Ex: Hipertensão, Diabetes, Tuberculose"
                  value={chronicCondition}
                  onChange={(e) => setChronicCondition(e.target.value)}
                />
              </div>

              <div>
                <Input
                  label="Saúde Mental / Acompanhamento CAPS"
                  placeholder="Ex: CAPS AD III, Depressão, Nega"
                  value={mentalHealth}
                  onChange={(e) => setMentalHealth(e.target.value)}
                />
              </div>
            </div>
          </div>

          <div className="flex items-center justify-end gap-2 border-t border-surface-border pt-4">
            <Button
              type="button"
              variant="secondary"
              onClick={onClose}
              disabled={loading}
            >
              Cancelar
            </Button>
            <Button
              type="submit"
              disabled={loading}
              className="gap-2"
            >
              <CheckCircle2 size={16} />
              {loading ? "Salvando..." : "Abrir Prontuário Centro POP"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
