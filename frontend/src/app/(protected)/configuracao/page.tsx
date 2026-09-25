import { BrandingSettingsForm } from "@/components/settings/BrandingSettingsForm";

export default function SistemaIdentidadePage() {
  return (
    <div className="flex flex-col gap-6">
      {/* 1. Branding & Identidade Visual Governamental (SEMPRAS / Prefeitura de Rondonópolis) */}
      <BrandingSettingsForm />
    </div>
  );
}
