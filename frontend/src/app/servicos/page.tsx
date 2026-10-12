import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { PublicShell } from "@/components/layout/PublicShell";
import { PublicApiError } from "@/lib/api/publicServer";
import { listPublishedServices } from "@/lib/catalog/catalog";
import { APP_URL } from "@/lib/env";
import { ServicosCatalogClient } from "./ServicosCatalogClient";

export const dynamic = "force-dynamic";

const description =
  "Catálogo completo de soluções tecnológicas e serviços especializados de infraestrutura, comunicação, automação e segurança.";

export const metadata: Metadata = {
  title: "Catálogo de Serviços",
  description,
  alternates: { canonical: `${APP_URL}/servicos` },
  openGraph: { title: "Catálogo de Serviços", description, type: "website", url: `${APP_URL}/servicos` },
};

export default async function ServicosPage() {
  let services: Awaited<ReturnType<typeof listPublishedServices>> = [];
  try {
    services = await listPublishedServices();
  } catch (err) {
    if (err instanceof PublicApiError && err.status === 404) notFound();
    throw err;
  }

  return (
    <PublicShell>
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-10 px-6 py-12">
        <div className="flex flex-col gap-3">
          <p className="dateline">Catálogo Especializado</p>
          <h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
            Soluções & Serviços de TI
          </h1>
          <p className="max-w-3xl text-base text-muted">{description}</p>
        </div>

        <ServicosCatalogClient initialServices={services} />
      </div>
    </PublicShell>
  );
}
