import { LocalidadesManager } from "@/components/configuracao/LocalidadesManager";
import { ErrorState } from "@/components/ui/ErrorState";
import { Section } from "@/components/ui/Section";
import { ApiError } from "@/lib/api/client";
import { serverApiGet } from "@/lib/api/server";
import type { Localidade } from "@/types/api";

export default async function LocalidadesPage() {
  let localidades: Localidade[] = [];
  let errorMessage: string | null = null;

  try {
    const { data } = await serverApiGet<Localidade[]>("v1/organizacao/localidades?all=true");
    localidades = data || [];
  } catch (err) {
    errorMessage = err instanceof ApiError ? err.message : "Falha ao carregar localidades";
  }

  return (
    <Section
      title="Localidades & Setores da Assistência Social"
      description="Gerenciamento de polos de atendimento (9 CRAS, Centro POP, CREAS, Casa da Mulher, etc.) e seus respectivos setores/departamentos vinculados ao Active Directory."
    >
      <div className="flex flex-col gap-6">
        {errorMessage && <ErrorState message={errorMessage} />}
        <LocalidadesManager initialLocalidades={localidades} />
      </div>
    </Section>
  );
}
