import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { ArrowLeft, Calendar, Clock, Layers, Share2 } from "lucide-react";

import { PublicShell } from "@/components/layout/PublicShell";
import { Markdown } from "@/components/nexus/Markdown";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";
import { getPublicPost } from "@/lib/blog/publicBlog";
import { APP_URL } from "@/lib/env";

type Props = { params: Promise<{ ref: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const post = await getPublicPost((await params).ref);
  if (!post) return { title: "Artigo não encontrado" };
  return {
    title: `${post.title} — Blog`,
    description: post.summary,
    alternates: { canonical: `${APP_URL}/blog/${post.slug}` },
    openGraph: {
      title: post.title,
      description: post.summary,
      type: "article",
      url: `${APP_URL}/blog/${post.slug}`,
    },
  };
}

export default async function PublicPostPage({ params }: Props) {
  const { ref } = await params;
  const post = await getPublicPost(ref);
  if (!post) notFound();

  const formatDate = (dateStr?: string | null) => {
    if (!dateStr) return "";
    try {
      return new Intl.DateTimeFormat("pt-BR", {
        day: "2-digit",
        month: "long",
        year: "numeric",
      }).format(new Date(dateStr));
    } catch {
      return dateStr;
    }
  };

  const words = (post.body || "").split(/\s+/).length;
  const readTime = `${Math.max(1, Math.ceil(words / 200))} min de leitura`;

  // Detect related service slug if post title or slug relates to one of our services
  const possibleServiceSlug = post.slug.replace(/^artigo-/, "");

  return (
    <PublicShell>
      <article className="mx-auto flex w-full max-w-4xl flex-col gap-8 px-6 py-12">
        <Link
          href="/blog"
          className="inline-flex items-center gap-1.5 text-sm font-medium text-muted transition-colors hover:text-foreground"
        >
          <ArrowLeft size={15} aria-hidden="true" /> Voltar para o Blog
        </Link>

        <header className="flex flex-col gap-4 border-b border-surface-border pb-8">
          <div className="flex flex-wrap items-center gap-3 text-xs text-muted">
            <span className="inline-flex items-center gap-1 font-medium">
              <Calendar size={13} aria-hidden="true" />
              {formatDate(post.published_at || post.updated_at)}
            </span>
            <span>•</span>
            <span className="inline-flex items-center gap-1">
              <Clock size={13} aria-hidden="true" />
              {readTime}
            </span>
            {post.author_name && (
              <>
                <span>•</span>
                <span>Por {post.author_name}</span>
              </>
            )}
          </div>

          <h1 className="text-3xl font-extrabold tracking-tight text-foreground sm:text-4xl md:text-5xl">
            {post.title}
          </h1>

          {post.summary && (
            <p className="text-lg leading-relaxed text-muted">
              {post.summary}
            </p>
          )}
        </header>

        {post.body && (
          <div className="prose prose-neutral dark:prose-invert max-w-none">
            <Markdown source={post.body} />
          </div>
        )}

        {/* CTA to Catalog Service */}
        <div className="mt-8 rounded-xl border border-accent/20 bg-accent/5 p-6 sm:p-8">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex flex-col gap-1">
              <span className="text-xs font-bold uppercase tracking-wider text-accent">
                Implementação & Serviços Gerenciados
              </span>
              <h3 className="text-lg font-bold text-foreground">
                Deseja implantar esta solução na sua empresa ou órgão público?
              </h3>
              <p className="text-sm text-muted">
                Nossa equipe especializada realiza o planejamento, provisionamento, configuração e suporte contínuo.
              </p>
            </div>
            <div className="flex shrink-0 items-center gap-3 pt-2 sm:pt-0">
              <Link
                href={`/servicos/${possibleServiceSlug}`}
                className="inline-flex items-center justify-center rounded-lg bg-accent px-4 py-2.5 text-sm font-semibold text-accent-foreground shadow-xs transition-colors hover:bg-accent/90"
              >
                Ver no Catálogo
              </Link>
              <Link
                href="/contato"
                className="inline-flex items-center justify-center rounded-lg border border-surface-border bg-surface px-4 py-2.5 text-sm font-medium text-foreground transition-colors hover:bg-surface-2"
              >
                Fale Conosco
              </Link>
            </div>
          </div>
        </div>
      </article>
    </PublicShell>
  );
}
