import type { Metadata } from "next";
import Link from "next/link";
import {
  ArrowRight,
  BookOpen,
  Boxes,
  CheckCircle2,
  Clock,
  Cloud,
  Cpu,
  FileText,
  HelpCircle,
  KeyRound,
  Layers,
  Lock,
  Mail,
  MessageSquare,
  Network,
  Newspaper,
  Server,
  ShieldCheck,
  Workflow,
  Wrench,
  Zap,
} from "lucide-react";

import { PublicShell } from "@/components/layout/PublicShell";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/Card";
import { getServerBranding } from "@/lib/branding/server";
import { listPublishedServices } from "@/lib/catalog/catalog";
import { listPublicPosts } from "@/lib/blog/publicBlog";
import { ServiceIcon, getCategoryIconStyles } from "@/lib/catalog/icons";
import { APP_URL } from "@/lib/env";

export const dynamic = "force-dynamic";

const description =
  "Plataforma completa de soluções especializadas de TI, infraestrutura como código, automação, mensageria e nuvem privada com novidades e catálogo interativo.";

export const metadata: Metadata = {
  title: "Início — Soluções de TI, Catálogo & Blog & Novidades",
  description,
  alternates: { canonical: `${APP_URL}` },
  openGraph: {
    title: "Início — Soluções de TI, Catálogo & Blog & Novidades",
    description,
    type: "website",
    url: `${APP_URL}`,
  },
};

const HIGHLIGHTED_SERVICES = [
  {
    slug: "whatsapp-api",
    title: "WhatsApp Business API",
    category: "Comunicação",
    icon: "shield",
    summary:
      "Conecte o número da empresa à API oficial da Meta. Zero risco de ban, múltiplos agentes simultâneos e templates homologados.",
  },
  {
    slug: "terraform",
    title: "Terraform",
    category: "Infraestrutura",
    icon: "cloud",
    summary:
      "Infraestrutura como Código (IaC) declarativa para Proxmox e multi-cloud. Versionamento Git, state locking e drift detection.",
  },
  {
    slug: "ansible",
    title: "Ansible",
    category: "Automação",
    icon: "cpu",
    summary:
      "Automação e gerência de configuração agentless via SSH e WinRM. Playbooks idempotentes em YAML e hardening em lote.",
  },
  {
    slug: "proxmox",
    title: "Proxmox VE",
    category: "Infraestrutura",
    icon: "server",
    summary:
      "Plataforma open-source completa de virtualização empresarial, alta disponibilidade, containers LXC e replicação de storages ZFS.",
  },
  {
    slug: "zabbix",
    title: "Zabbix Enterprise",
    category: "Gestão de TI",
    icon: "eye",
    summary:
      "Monitoramento proativo em tempo real de servidores, links de rede, bancos de dados e serviços com dashboards e alertas no Telegram.",
  },
  {
    slug: "glpi",
    title: "GLPI ITSM",
    category: "Gestão de TI",
    icon: "wrench",
    summary:
      "Central de atendimento (Helpdesk) alinhada às melhores práticas ITIL, inventário automatizado de ativos e gestão de contratos com SLA.",
  },
  {
    slug: "nextcloud",
    title: "Nextcloud Hub",
    category: "Infraestrutura",
    icon: "cloud",
    summary:
      "Nuvem privada para armazenamento seguro, compartilhamento criptografado de arquivos corporativos e colaboração em documentos.",
  },
  {
    slug: "keycloak",
    title: "Keycloak IAM & SSO",
    category: "Segurança e Identidade",
    icon: "key-round",
    summary:
      "Gerenciamento centralizado de identidades, Single Sign-On (SSO), federação com Active Directory e autenticação multifator (MFA).",
  },
];

export default async function HomePage() {
  const [allServices, recentPosts, brand] = await Promise.all([
    listPublishedServices().catch(() => []),
    listPublicPosts("", "", 1, 10).catch(() => []),
    getServerBranding().catch(() => null),
  ]);

  const serviceCount = allServices.length > 0 ? allServices.length : 26;
  const displayedServices = allServices.length > 0 ? allServices.slice(0, 8) : HIGHLIGHTED_SERVICES;

  const publishedSlugs = new Set(allServices.map((s) => s.slug));
  const activePosts = recentPosts.filter((post) => {
    if (post.slug.startsWith("artigo-")) {
      const serviceSlug = post.slug.replace(/^artigo-/, "");
      return allServices.length === 0 || publishedSlugs.has(serviceSlug);
    }
    return true;
  });

  return (
    <PublicShell>
      <div className="flex flex-col gap-20 pb-20">
        {/* HERO SECTION */}
        <section className="relative overflow-hidden border-b border-surface-border bg-gradient-to-b from-surface-1 to-surface-2/40 px-6 py-20 sm:py-28">
          <div className="mx-auto flex max-w-5xl flex-col items-center text-center">
            <span className="inline-flex items-center gap-2 rounded-full border border-sky-200 bg-sky-50 px-4 py-1.5 text-xs font-semibold text-sky-800 shadow-xs dark:border-sky-800/60 dark:bg-sky-950/60 dark:text-sky-300">
              <Zap size={14} aria-hidden="true" /> {brand?.appName || brand?.orgName || "Plataforma Digital de Serviços & Infraestrutura"}
            </span>

            <h1 className="mt-6 text-4xl font-extrabold tracking-tight text-foreground sm:text-5xl md:text-6xl">
              Soluções Tecnológicas Especializadas{" "}
              <span className="text-accent underline decoration-accent/40 decoration-wavy underline-offset-8">
                Sob Medida
              </span>
            </h1>

            <p className="mt-6 max-w-3xl text-lg leading-relaxed text-muted sm:text-xl">
              Do atendimento homologado na <strong>WhatsApp Business API</strong> à automação declarativa com{" "}
              <strong>Terraform</strong> e <strong>Ansible</strong>. Conheça nosso catálogo com{" "}
              <strong className="text-foreground">{serviceCount} serviços corporativos</strong> implantados com total
              soberania, segurança e suporte contínuo.
            </p>

            {/* CTAs */}
            <div className="mt-10 flex flex-wrap items-center justify-center gap-4">
              <Link
                href="/servicos"
                className="inline-flex items-center gap-2 rounded-xl bg-accent px-6 py-3.5 text-base font-semibold text-accent-foreground shadow-md transition-all hover:-translate-y-0.5 hover:bg-accent/90"
              >
                Explorar Catálogo de Serviços <ArrowRight size={18} aria-hidden="true" />
              </Link>

              <Link
                href="/blog"
                className="inline-flex items-center gap-2 rounded-xl border border-surface-border bg-surface px-6 py-3.5 text-base font-semibold text-foreground shadow-xs transition-all hover:bg-surface-2"
              >
                <BookOpen size={18} aria-hidden="true" /> Acessar o Blog & Novidades
              </Link>
            </div>

            {/* Key Metric Highlights */}
            <div className="mt-14 grid w-full max-w-4xl grid-cols-2 gap-4 border-t border-surface-border pt-8 sm:grid-cols-4">
              <div className="flex flex-col items-center">
                <span className="text-2xl font-black text-foreground sm:text-3xl">{serviceCount}+</span>
                <span className="text-xs font-medium text-muted">Serviços Especializados</span>
              </div>
              <div className="flex flex-col items-center">
                <span className="text-2xl font-black text-foreground sm:text-3xl">100%</span>
                <span className="text-xs font-medium text-muted">Soberania dos Dados</span>
              </div>
              <div className="flex flex-col items-center">
                <span className="text-2xl font-black text-foreground sm:text-3xl">5 a 10 dias</span>
                <span className="text-xs font-medium text-muted">Prazo de Implantação</span>
              </div>
              <div className="flex flex-col items-center">
                <span className="text-2xl font-black text-foreground sm:text-3xl">Go + Microkernel</span>
                <span className="text-xs font-medium text-muted">Arquitetura de Alta Vazão</span>
              </div>
            </div>
          </div>
        </section>

        {/* SECTION 1: SERVIÇOS EM DESTAQUE */}
        <section className="mx-auto flex w-full max-w-6xl flex-col gap-10 px-6">
          <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
            <div className="flex flex-col gap-2">
              <p className="dateline">Catálogo Especializado</p>
              <h2 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
                Soluções Estratégicas para sua Operação
              </h2>
              <p className="max-w-2xl text-base text-muted">
                Implementação personalizada, treinamento e monitoramento contínuo das principais ferramentas do mercado.
              </p>
            </div>
            <Link
              href="/servicos"
              className="inline-flex items-center gap-1.5 text-sm font-semibold text-accent hover:underline"
            >
              Ver todas as {serviceCount} soluções <ArrowRight size={15} aria-hidden="true" />
            </Link>
          </div>

          <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
            {displayedServices.map((svc) => {
              const iconStyle = getCategoryIconStyles(svc.category);
              return (
                <Link key={svc.slug} href={`/servicos/${svc.slug}`} className="group block h-full">
                  <Card className="flex h-full flex-col border-surface-border transition-all duration-200 hover:-translate-y-1 hover:border-primary/40 hover:shadow-lg">
                    <CardHeader className="flex flex-col gap-3">
                      <div className="flex items-center justify-between">
                        <span className={`inline-flex h-11 w-11 items-center justify-center rounded-xl shadow-xs transition-transform group-hover:scale-105 ${iconStyle.badge}`}>
                          <ServiceIcon name={svc.icon} size={22} aria-hidden="true" />
                        </span>
                        <span className={`rounded-md border px-2 py-0.5 text-[11px] font-semibold ${iconStyle.tag}`}>
                          {svc.category || "Geral"}
                        </span>
                      </div>
                      <CardTitle as="h3" className="text-base font-bold text-foreground group-hover:text-primary">
                        {svc.title}
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-1 flex-col justify-between gap-4">
                      <CardDescription className="line-clamp-3 text-xs leading-relaxed text-muted">
                        {svc.summary}
                      </CardDescription>
                      <span className="inline-flex items-center gap-1 text-xs font-semibold text-primary">
                        Detalhes da solução <ArrowRight size={12} aria-hidden="true" />
                      </span>
                    </CardContent>
                  </Card>
                </Link>
              );
            })}
          </div>

          <div className="rounded-2xl border border-surface-border bg-surface-2/60 p-8 text-center sm:flex sm:items-center sm:justify-between sm:text-left">
            <div className="flex flex-col gap-1">
              <h3 className="text-lg font-bold text-foreground">
                Precisa de uma topologia completa para sua infraestrutura?
              </h3>
              <p className="text-sm text-muted">
                Integramos comunicação, storages, redes corporativas e orquestração de servidores sob medida.
              </p>
            </div>
            <div className="mt-4 flex shrink-0 items-center justify-center gap-3 sm:mt-0">
              <Link
                href="/servicos"
                className="inline-flex items-center gap-2 rounded-lg bg-accent px-5 py-2.5 text-sm font-semibold text-accent-foreground shadow-xs hover:bg-accent/90"
              >
                Catálogo Completo
              </Link>
              <Link
                href="/contato"
                className="inline-flex items-center gap-2 rounded-lg border border-surface-border bg-surface px-5 py-2.5 text-sm font-medium text-foreground hover:bg-surface-3"
              >
                Fale Conosco
              </Link>
            </div>
          </div>
        </section>

        {/* SECTION 2: COMO A PLATAFORMA FOI CONSTRUÍDA (MENCIONANDO) */}
        <section className="border-y border-surface-border bg-surface-1/70 px-6 py-16">
          <div className="mx-auto flex max-w-6xl flex-col gap-12">
            <div className="flex flex-col gap-3">
              <p className="dateline">Arquitetura da Plataforma</p>
              <h2 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
                Como Esta Plataforma Foi Desenvolvida
              </h2>
              <p className="max-w-3xl text-base text-muted">
                Concebida sob princípios rigorosos de isolamento de módulos, alta performance e conformidade com os
                padrões federais de acessibilidade, segurança e governança de dados.
              </p>
            </div>

            <div className="grid gap-6 md:grid-cols-3">
              <div className="flex flex-col gap-3 rounded-xl border border-surface-border bg-surface p-6 shadow-xs">
                <div className="flex h-11 w-11 items-center justify-center rounded-xl border border-blue-200 bg-blue-50 text-blue-700 shadow-xs dark:border-blue-800/60 dark:bg-blue-950/60 dark:text-blue-300">
                  <Workflow size={22} aria-hidden="true" />
                </div>
                <h3 className="text-lg font-bold text-foreground">Core Microkernel em Go</h3>
                <p className="text-xs leading-relaxed text-muted">
                  O backend foi desenvolvido em <strong>Go (Golang)</strong> com arquitetura Microkernel: o Kernel
                  gerencia ciclo de vida, transações atômicas e barramento de eventos (RabbitMQ), enquanto os módulos
                  operam como plug-ins totalmente desacoplados.
                </p>
              </div>

              <div className="flex flex-col gap-3 rounded-xl border border-surface-border bg-surface p-6 shadow-xs">
                <div className="flex h-11 w-11 items-center justify-center rounded-xl border border-indigo-200 bg-indigo-50 text-indigo-700 shadow-xs dark:border-indigo-800/60 dark:bg-indigo-950/60 dark:text-indigo-300">
                  <Boxes size={22} aria-hidden="true" />
                </div>
                <h3 className="text-lg font-bold text-foreground">Next.js 16 & Padrão DSGov</h3>
                <p className="text-xs leading-relaxed text-muted">
                  Interface moderna construída em <strong>Next.js 16 (React 19)</strong> com Turbopack, estrita
                  conformidade com o <strong>Design System de Governo (DSGov)</strong>, zero layout shift e renderização
                  otimizada para velocidade e experiência do usuário.
                </p>
              </div>

              <div className="flex flex-col gap-3 rounded-xl border border-surface-border bg-surface p-6 shadow-xs">
                <div className="flex h-11 w-11 items-center justify-center rounded-xl border border-emerald-200 bg-emerald-50 text-emerald-700 shadow-xs dark:border-emerald-800/60 dark:bg-emerald-950/60 dark:text-emerald-300">
                  <ShieldCheck size={22} aria-hidden="true" />
                </div>
                <h3 className="text-lg font-bold text-foreground">e-MAG, LGPD & AppSec</h3>
                <p className="text-xs leading-relaxed text-muted">
                  Total aderência às diretrizes da SGD/MGI: acessibilidade <strong>e-MAG 2.0 (WCAG 2.1 AA)</strong> com
                  VLibras nativo, mascaramento automático de PII em logs pela LGPD e fundação blindada na matriz{" "}
                  <strong>OWASP Top 10</strong>.
                </p>
              </div>
            </div>

            <div className="flex items-center justify-between border-t border-surface-border pt-4 text-xs text-muted">
              <span>Stack: Go 1.23 · Next.js 16 · PostgreSQL 16 · Redis 7 · RabbitMQ · Proxmox VE · Caddy TLS</span>
              <Link href="/sobre" className="font-semibold text-primary hover:underline">
                Conhecer arquitetura completa →
              </Link>
            </div>
          </div>
        </section>

        {/* SECTION 3: FORTE APEGO AO BLOG */}
        <section className="mx-auto flex w-full max-w-6xl flex-col gap-10 px-6">
          <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
            <div className="flex flex-col gap-2">
              <span className="inline-flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider text-primary">
                <Newspaper size={16} aria-hidden="true" /> Artigos, Origens & Novidades
              </span>
              <h2 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
                Blog & Novidades: Aprofunde-se em Cada Solução
              </h2>
              <p className="max-w-2xl text-base text-muted">
                Publicações completas e abertas sobre como cada serviço surgiu, suas aplicações reais em produção,
                vantagens competitivas e quais gigantes globais confiam nele.
              </p>
            </div>
            <Link
              href="/blog"
              className="inline-flex items-center gap-1.5 text-sm font-semibold text-primary hover:underline"
            >
              Explorar todas as novidades <ArrowRight size={15} aria-hidden="true" />
            </Link>
          </div>

          {/* Posts Highlights */}
          {activePosts.length === 0 ? (
            <div className="flex flex-col items-center justify-center rounded-2xl border border-dashed border-surface-border py-16 text-center text-muted">
              <BookOpen size={32} aria-hidden="true" />
              <p className="mt-3 text-sm">Carregando novidades do Blog...</p>
            </div>
          ) : (
            <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {activePosts.slice(0, 6).map((post) => (
                <Link key={post.id} href={`/blog/${post.slug}`} className="group block h-full">
                  <Card className="flex h-full flex-col border-surface-border transition-all duration-200 hover:-translate-y-1 hover:border-primary/40 hover:shadow-lg">
                    <CardHeader className="flex flex-col gap-2">
                      <div className="flex items-center justify-between text-xs text-muted">
                        <span className="inline-flex items-center gap-1 font-semibold text-primary">
                          Blog & Novidades
                        </span>
                        <span className="inline-flex items-center gap-1">
                          <Clock size={12} aria-hidden="true" />
                          5 min de leitura
                        </span>
                      </div>
                      <CardTitle as="h3" className="text-base font-bold leading-snug text-foreground group-hover:text-primary">
                        {post.title}
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-1 flex-col justify-between gap-4">
                      <CardDescription className="line-clamp-3 text-xs leading-relaxed text-muted">
                        {post.summary}
                      </CardDescription>
                      <span className="inline-flex items-center gap-1 text-xs font-semibold text-accent transition-transform group-hover:translate-x-1">
                        Ler novidade completa <ArrowRight size={13} aria-hidden="true" />
                      </span>
                    </CardContent>
                  </Card>
                </Link>
              ))}
            </div>
          )}

          <div className="text-center">
            <Link
              href="/blog"
              className="inline-flex items-center gap-2 rounded-xl bg-accent px-6 py-3 text-sm font-semibold text-accent-foreground shadow-sm transition-colors hover:bg-accent/90"
            >
              Ver Todas as Novidades no Blog <ArrowRight size={16} aria-hidden="true" />
            </Link>
          </div>
        </section>

        {/* SECTION 4: PILARES PÚBLICOS (SOBRE, CONTATO, FAQ) */}
        <section className="mx-auto flex w-full max-w-6xl flex-col gap-6 px-6">
          <div className="grid gap-6 sm:grid-cols-3">
            <Link href="/sobre" className="group block h-full">
              <div className="flex h-full flex-col gap-3 rounded-2xl border border-surface-border bg-surface p-6 transition-all hover:border-primary/40 hover:shadow-md">
                <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
                  <Layers size={20} aria-hidden="true" />
                </span>
                <h3 className="text-base font-bold text-foreground group-hover:text-primary">
                  Sobre a Plataforma
                </h3>
                <p className="text-xs leading-relaxed text-muted">
                  Conheça nossa arquitetura tecnológica, conformidade e-MAG / DSGov, governança e diretrizes de
                  segurança.
                </p>
                <span className="mt-auto inline-flex items-center gap-1 text-xs font-semibold text-primary">
                  Saber mais <ArrowRight size={12} aria-hidden="true" />
                </span>
              </div>
            </Link>

            <Link href="/faq" className="group block h-full">
              <div className="flex h-full flex-col gap-3 rounded-2xl border border-surface-border bg-surface p-6 transition-all hover:border-accent/40 hover:shadow-md">
                <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg bg-accent/10 text-accent">
                  <HelpCircle size={20} aria-hidden="true" />
                </span>
                <h3 className="text-base font-bold text-foreground group-hover:text-accent">
                  Perguntas Frequentes (FAQ)
                </h3>
                <p className="text-xs leading-relaxed text-muted">
                  Tire dúvidas sobre prazos de implantação, infraestrutura dedicada, SLAs e modelos de contratação.
                </p>
                <span className="mt-auto inline-flex items-center gap-1 text-xs font-semibold text-accent">
                  Consultar FAQ <ArrowRight size={12} aria-hidden="true" />
                </span>
              </div>
            </Link>

            <Link href="/contato" className="group block h-full">
              <div className="flex h-full flex-col gap-3 rounded-2xl border border-surface-border bg-surface p-6 transition-all hover:border-success/40 hover:shadow-md">
                <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg bg-success/10 text-success">
                  <MessageSquare size={20} aria-hidden="true" />
                </span>
                <h3 className="text-base font-bold text-foreground group-hover:text-success">
                  Fale Conosco
                </h3>
                <p className="text-xs leading-relaxed text-muted">
                  Envie suas demandas ou tire dúvidas diretamente com nossa equipe técnica de atendimento.
                </p>
                <span className="mt-auto inline-flex items-center gap-1 text-xs font-semibold text-success">
                  Abrir contato <ArrowRight size={12} aria-hidden="true" />
                </span>
              </div>
            </Link>
          </div>
        </section>
      </div>
    </PublicShell>
  );
}
