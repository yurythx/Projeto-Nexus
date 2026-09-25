import type { Metadata } from "next";
import { Geist_Mono, Inter, Newsreader } from "next/font/google";
import { cookies } from "next/headers";
import type { ReactNode } from "react";
import "./globals.css";

import { BRANDING_COOKIE, mergeServerBranding, parseBrandingCookie, tokensToCss } from "@/components/branding/brandingConfig";
import { publicGetOr } from "@/lib/api/publicServer";
import type { Branding } from "@/lib/nexus/types";
import { Providers } from "./providers";

const inter = Inter({ variable: "--font-inter", subsets: ["latin"], display: "swap" });
const newsreader = Newsreader({ variable: "--font-newsreader", subsets: ["latin"], weight: ["500", "600"], display: "swap" });
const geistMono = Geist_Mono({ variable: "--font-geist-mono", subsets: ["latin"], display: "swap" });

export async function generateMetadata(): Promise<Metadata> {
  const b = await publicGetOr<Branding | null>("branding", null, 60);
  const name = b?.app_name || "Projeto Nexus";
  return {
    title: { default: name, template: `%s — ${name}` },
    description: b?.app_description || "Plataforma corporativa modular (Microkernel) em conformidade com e-MAG, DSGov e LGPD.",
  };
}

export default async function RootLayout({ children }: { children: ReactNode }) {
  const cookieStore = await cookies();
  const theme = cookieStore.get("nexus-theme")?.value;
  const dataTheme = theme === "dark" || theme === "light" ? theme : undefined;

  // Identidade da organização (API, cache de 60s) + preferências e-MAG do
  // visitante (cookie) — aplicadas em <html> ANTES do 1º paint.
  const api = await publicGetOr<Branding | null>("branding", null, 60);
  const branding = mergeServerBranding(api, parseBrandingCookie(cookieStore.get(BRANDING_COOKIE)?.value));
  const tokenCss = tokensToCss(branding.tokens);

  return (
    <html
      lang="pt-BR"
      data-theme={dataTheme}
      data-high-contrast={branding.highContrast ? "true" : undefined}
      data-font-scale={String(branding.fontSizeScale || 100)}
      className={`${inter.variable} ${newsreader.variable} ${geistMono.variable} h-full antialiased`}
    >
      <head>{tokenCss && <style id="nexus-white-label">{tokenCss}</style>}</head>
      <body className="flex min-h-full flex-col">
        <Providers initialBranding={branding}>{children}</Providers>
      </body>
    </html>
  );
}
