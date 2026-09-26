import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { ShieldCheck } from "lucide-react";

import { PublicShell } from "@/components/layout/PublicShell";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";

export const metadata: Metadata = { title: "Verificar assinatura", description: "Confira a autenticidade de um documento assinado eletronicamente." };

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

async function verificar(formData: FormData) {
  "use server";
  const codigo = String(formData.get("codigo") ?? "").trim();
  if (UUID_RE.test(codigo)) redirect(`/verificar/${codigo.toLowerCase()}`);
  redirect("/verificar?erro=1");
}

export default async function VerificarIndex({ searchParams }: { searchParams: Promise<{ erro?: string }> }) {
  const { erro } = await searchParams;
  return (
    <PublicShell>
      <div className="mx-auto flex w-full max-w-xl flex-col gap-6 px-6 py-14">
        <div className="flex items-center gap-3">
          <ShieldCheck size={28} aria-hidden="true" className="text-primary" />
          <h1 className="text-2xl font-bold">Verificar assinatura eletrônica</h1>
        </div>
        <p className="text-muted">Informe o código de verificação impresso no documento assinado.</p>
        <form action={verificar} className="flex flex-col gap-3">
          <Input
            id="codigo"
            name="codigo"
            label="Código de verificação"
            required
            placeholder="00000000-0000-0000-0000-000000000000"
            error={erro ? "Código inválido — confira os 36 caracteres." : undefined}
            autoComplete="off"
          />
          <Button type="submit" className="self-start">
            Verificar
          </Button>
        </form>
      </div>
    </PublicShell>
  );
}
