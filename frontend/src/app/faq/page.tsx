import type { Metadata } from "next";
import Link from "next/link";

import { PublicShell } from "@/components/layout/PublicShell";
import { Card, CardContent } from "@/components/ui/Card";
import { APP_URL } from "@/lib/env";

const description = "Perguntas frequentes sobre os serviços socioassistenciais da SEMPRAS.";

export const metadata: Metadata = {
  title: "Perguntas frequentes",
  description,
  alternates: { canonical: `${APP_URL}/faq` },
  openGraph: { title: "FAQ — Assistência Social", description, type: "website", url: `${APP_URL}/faq` },
};

// Conteúdo estático — mesma decisão de /sobre. FAQ por serviço (mais
// específica) vive na própria página do serviço (/servicos/[slug]), ver
// catalog/domain.FAQItem; esta é só a FAQ institucional/geral.
const FAQS = [
  {
    question: "O que é a Rede de Assistência Social (SEMPRAS / SUAS)?",
    answer:
      "É o sistema integrado de gestão e atendimento dos serviços socioassistenciais da Prefeitura de Rondonópolis, conectando os 9 CRAS, CREAS, Centro POP e programas de transferência de renda em conformidade com o SUAS.",
  },
  {
    question: "Como funciona a central de serviços ao cidadão?",
    answer:
      "Cada serviço socioassistencial listado em /servicos tem sua própria página informativa, detalhando documentação necessária, público-alvo e formas de atendimento nos polos municipais.",
  },
  {
    question: "Como os servidores e técnicos acessam a plataforma interna?",
    answer:
      "O acesso à área autenticada é integrado à credencial de rede corporativa (Active Directory / Keycloak), com permissões e visibilidade de unidades e setores (CRAS, Centro POP, CREAS) atribuídas de acordo com a lotação funcional do servidor.",
  },
  {
    question: "Como entrar em contato com a equipe socioassistencial?",
    answer: "Pelo formulário oficial em /contato ou diretamente nas unidades físicas dos CRAS e CREAS do município.",
  },
];

const faqJsonLd = {
  "@context": "https://schema.org",
  "@type": "FAQPage",
  mainEntity: FAQS.map((f) => ({
    "@type": "Question",
    name: f.question,
    acceptedAnswer: { "@type": "Answer", text: f.answer },
  })),
};

export default function FaqPage() {
  return (
    <PublicShell>
      {/* JSON-LD estático a partir da constante FAQS acima, sem entrada de usuário. */}
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(faqJsonLd) }} />

      <div className="mx-auto flex w-full max-w-3xl flex-col gap-8 px-6 py-14">
        <div className="flex flex-col gap-3">
          <p className="dateline">FAQ</p>
          <h1 className="text-3xl font-bold text-foreground sm:text-4xl">Perguntas frequentes</h1>
          <p className="text-muted">{description}</p>
        </div>

        <div className="flex flex-col gap-4">
          {FAQS.map((faq) => (
            <Card key={faq.question}>
              <CardContent className="flex flex-col gap-2 pt-5">
                <h2 className="font-semibold text-foreground">{faq.question}</h2>
                <p className="text-sm text-muted">{faq.answer}</p>
              </CardContent>
            </Card>
          ))}
        </div>

        <p className="text-center text-sm text-muted">
          Não encontrou o que procurava?{" "}
          <Link href="/contato" className="font-semibold text-accent hover:underline">
            Fale com a gente
          </Link>
          .
        </p>
      </div>
    </PublicShell>
  );
}
