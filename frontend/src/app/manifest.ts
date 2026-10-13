import type { MetadataRoute } from "next";

import { getServerBranding } from "@/lib/branding/server";

export default async function manifest(): Promise<MetadataRoute.Manifest> {
  const brand = await getServerBranding().catch(() => null);
  const name = brand?.appName || "Projeto Nexus";
  const shortName = "Nexus";
  const description =
    brand?.appDescription ||
    "Plataforma completa de soluções corporativas de TI, nuvem privada, catálogo interativo e governança digital.";

  return {
    name,
    short_name: shortName,
    description,
    start_url: "/",
    display: "standalone",
    background_color: "#0f172a",
    theme_color: "#1b365d",
    orientation: "portrait-primary",
    icons: [
      {
        src: "/icon.svg",
        sizes: "any",
        type: "image/svg+xml",
        purpose: "any",
      },
      {
        src: "/apple-icon.png",
        sizes: "180x180",
        type: "image/png",
        purpose: "maskable",
      },
    ],
  };
}
