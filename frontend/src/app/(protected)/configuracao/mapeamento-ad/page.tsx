import { ActiveDirectoryMappingManager } from "@/components/configuracao/ActiveDirectoryMappingManager";
import { ErrorState } from "@/components/ui/ErrorState";
import { Section } from "@/components/ui/Section";
import { ApiError } from "@/lib/api/client";
import { serverApiGet } from "@/lib/api/server";
import type { Localidade, Perfil } from "@/types/api";

export default async function MapeamentoADPage() {
  let perfis: Perfil[] = [];
  let localidades: Localidade[] = [];
  let errorMessage: string | null = null;

  try {
    const [resPerfis, resLocs] = await Promise.all([
      serverApiGet<Perfil[]>("v1/organizacao/perfis"),
      serverApiGet<Localidade[]>("v1/organizacao/localidades?all=true"),
    ]);
    perfis = resPerfis.data || [];
    localidades = resLocs.data || [];
  } catch (err) {
    errorMessage = err instanceof ApiError ? err.message : "Falha ao carregar dados de mapeamento do AD";
  }

  return (
    <Section
      title="Mapeamento Active Directory (AD)"
      description="Gerenciamento centralizado de correspondência entre grupos de segurança do Active Directory e os Perfis, Unidades e Setores do sistema socioassistencial."
    >
      <div className="flex flex-col gap-6">
        {errorMessage && <ErrorState message={errorMessage} />}
        <ActiveDirectoryMappingManager
          initialPerfis={perfis}
          initialLocalidades={localidades}
        />
      </div>
    </Section>
  );
}
