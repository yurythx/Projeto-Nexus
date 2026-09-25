import { PerfisManager } from "@/components/configuracao/PerfisManager";
import { ErrorState } from "@/components/ui/ErrorState";
import { Section } from "@/components/ui/Section";
import { ApiError } from "@/lib/api/client";
import { serverApiGet } from "@/lib/api/server";
import type { Perfil } from "@/types/api";

export default async function PerfisPage() {
  let perfis: Perfil[] = [];
  let errorMessage: string | null = null;

  try {
    const { data } = await serverApiGet<Perfil[]>("v1/organizacao/perfis");
    perfis = data || [];
  } catch (err) {
    errorMessage = err instanceof ApiError ? err.message : "Falha ao carregar perfis";
  }

  return (
    <Section
      title="Perfis de Acesso & RBAC do SUAS"
      description="Gerenciamento de funções, perfis institucionais e permissões vinculadas aos grupos do Active Directory da Prefeitura Municipal."
    >
      <div className="flex flex-col gap-6">
        {errorMessage && <ErrorState message={errorMessage} />}
        <PerfisManager initialPerfis={perfis} />
      </div>
    </Section>
  );
}
