import type { Metadata } from "next";
import Link from "next/link";
import { ArrowRight, HelpCircle, MessageSquare } from "lucide-react";

import { PublicShell } from "@/components/layout/PublicShell";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";
import { APP_URL } from "@/lib/env";

const description =
  "Perguntas frequentes sobre a plataforma tecnológica, catálogo de serviços de TI, implantação, infraestrutura e segurança.";

export const metadata: Metadata = {
  title: "Perguntas Frequentes (FAQ)",
  description,
  alternates: { canonical: `${APP_URL}/faq` },
  openGraph: {
    title: "Perguntas Frequentes (FAQ)",
    description,
    type: "website",
    url: `${APP_URL}/faq`,
  },
};

const FAQ_SECTIONS = [
  {
    category: "Plataforma e Arquitetura",
    items: [
      {
        q: "O que é o Projeto Nexus e como foi desenvolvido?",
        a: "O Projeto Nexus é uma plataforma digital institucional de alto desempenho construída sobre uma arquitetura Microkernel em Go (Golang) com desacoplamento modular completo. O frontend utiliza Next.js 16 com React 19, persistência em PostgreSQL 16 com isolamento de transações, mensageria RabbitMQ e orquestração Docker sobre Proxmox VE.",
      },
      {
        q: "Quais normas de conformidade pública e governamental a plataforma atende?",
        a: "A plataforma foi desenhada em estrita conformidade com as diretrizes da Secretaria de Governo Digital (SGD/MGI), respeitando o e-MAG 2.0 (Acessibilidade Digital WCAG 2.1 AA com VLibras), o Padrão Digital de Governo (DSGov), a LGPD (Lei 13.709/2018 com mascaramento PII nativo) e a matriz de segurança cibernética OWASP Top 10 Enterprise.",
      },
      {
        q: "Nossa organização mantém total soberania sobre os dados e infraestrutura?",
        a: "Sim. Toda a stack é desenhada para operar de forma dedicada (On-Premises ou VPS privada), sem dependência de nuvens proprietárias opacas. Seus dados, bancos de dados, backups e chaves criptográficas permanecem sob sua total custódia e governança.",
      },
    ],
  },
  {
    category: "Serviços Especializados de TI",
    items: [
      {
        q: "Como funciona o processo de contratação e implantação de um serviço?",
        a: "O processo é dividido em três fases claras: 1) Levantamento técnico e arquitetura personalizada; 2) Configuração, hardening de segurança e testes em homologação; 3) Entrega em produção, documentação técnica viva e treinamento prático para a sua equipe.",
      },
      {
        q: "Qual a diferença entre a WhatsApp Business API Oficial e soluções não-oficiais?",
        a: "A API Oficial é conectada diretamente aos servidores da Meta com número verificado. Não há risco de ban de número, permite múltiplos atendentes simultâneos sem desconexões, suporta mensagens ativas em massa aprovadas (templates HSM) e integra-se perfeitamente a CRMs e helpdesks como Chatwoot.",
      },
      {
        q: "Como o Terraform e o Ansible se complementam na prática?",
        a: "O Terraform cuida da Infraestrutura como Código (IaC): cria as máquinas virtuais no Proxmox ou nuvem, define redes, IPs e storages. O Ansible entra em seguida para configurar o sistema operacional, aplicar patches de segurança, criar usuários e instalar os softwares de forma automatizada e idempotente (agentless).",
      },
      {
        q: "Por que migrar para o Proxmox VE e storages TrueNAS?",
        a: "Proxmox VE e TrueNAS entregam virtualização e armazenamento de nível enterprise com sistemas de arquivos resilientes (ZFS), replicação de dados e alta disponibilidade sem custos astronômicos de licenciamento por socket ou core, permitindo economia imediata de até 70% nos custos de TI.",
      },
    ],
  },
  {
    category: "Suporte, Prazos e Segurança",
    items: [
      {
        q: "Qual é o prazo médio de entrega e implementação dos serviços?",
        a: "A maioria das soluções padronizadas (como Chatwoot, Zabbix, GLPI, Nextcloud, TrueNAS e VPNs Tailscale) é entregue em até 5 a 10 dias úteis, com acompanhamento assistido e garantia de operação contínua.",
      },
      {
        q: "Existe suporte contínuo após a entrega?",
        a: "Sim. Oferecemos pacotes de serviços gerenciados, rotinas de backup corporativo automatizado, monitoramento proativo de métricas e suporte técnico especializado via canais dedicados.",
      },
      {
        q: "Como entrar em contato para tirar dúvidas específicas?",
        a: "Você pode navegar pelo catálogo em /servicos, consultar artigos técnicos no nosso Blog (/blog) ou enviar uma mensagem pelo formulário público em /contato.",
      },
    ],
  },
];

export default function FAQPage() {
  return (
    <PublicShell>
      <div className="mx-auto flex w-full max-w-5xl flex-col gap-12 px-6 py-12">
        <div className="flex flex-col gap-3">
          <p className="dateline">Central de Ajuda & Dúvidas</p>
          <h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
            Perguntas Frequentes (FAQ)
          </h1>
          <p className="max-w-2xl text-base text-muted">{description}</p>
        </div>

        <div className="flex flex-col gap-10">
          {FAQ_SECTIONS.map((sec) => (
            <section key={sec.category} className="flex flex-col gap-4">
              <h2 className="text-xl font-bold text-foreground border-b border-surface-border pb-2">
                {sec.category}
              </h2>
              <div className="grid gap-4">
                {sec.items.map((item, idx) => (
                  <Card key={idx} className="border-surface-border">
                    <CardHeader className="flex flex-row items-start gap-3 pb-2">
                      <HelpCircle size={20} className="mt-0.5 text-accent shrink-0" aria-hidden="true" />
                      <CardTitle as="h3" className="text-base font-semibold text-foreground">
                        {item.q}
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="pl-4 sm:pl-11 text-sm leading-relaxed text-muted">
                      {item.a}
                    </CardContent>
                  </Card>
                ))}
              </div>
            </section>
          ))}
        </div>

        {/* CTA */}
        <div className="rounded-xl border border-surface-border bg-surface-2 p-6 sm:p-8 text-center sm:text-left sm:flex sm:items-center sm:justify-between">
          <div className="flex flex-col gap-1">
            <h3 className="text-lg font-bold text-foreground">Ainda tem dúvidas?</h3>
            <p className="text-sm text-muted">
              Nossa equipe está pronta para avaliar seu cenário e desenhar a solução ideal.
            </p>
          </div>
          <div className="mt-4 sm:mt-0 flex flex-wrap shrink-0 items-center justify-center sm:justify-start gap-3">
            <Link
              href="/contato"
              className="inline-flex items-center gap-2 rounded-lg bg-accent px-5 py-2.5 text-sm font-semibold text-accent-foreground shadow-xs hover:bg-accent/90"
            >
              <MessageSquare size={16} aria-hidden="true" /> Fale Conosco
            </Link>
            <Link
              href="/servicos"
              className="inline-flex items-center gap-1 text-sm font-medium text-accent hover:underline"
            >
              Ver Serviços <ArrowRight size={14} aria-hidden="true" />
            </Link>
          </div>
        </div>
      </div>
    </PublicShell>
  );
}
