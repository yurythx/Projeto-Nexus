"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { ArrowRight, BookOpen, Calendar, Clock, Newspaper, Search, X } from "lucide-react";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/Card";
import type { Post } from "@/lib/nexus/types";

interface PublicBlogListClientProps {
  initialPosts: Post[];
}

export function PublicBlogListClient({ initialPosts }: PublicBlogListClientProps) {
  const [search, setSearch] = useState("");
  const [selectedTag, setSelectedTag] = useState<string>("Todos");

  // Derive tags from posts or content
  const tags = useMemo(() => {
    return [
      "Todos",
      "Infraestrutura",
      "Automação",
      "Segurança",
      "Comunicação",
      "Gestão de TI",
    ];
  }, []);

  const filteredPosts = useMemo(() => {
    const q = search.trim().toLowerCase();
    return initialPosts.filter((post) => {
      if (selectedTag !== "Todos") {
        const textToMatch = `${post.title} ${post.summary} ${post.body}`.toLowerCase();
        if (!textToMatch.includes(selectedTag.toLowerCase())) return false;
      }

      if (!q) return true;

      const title = post.title.toLowerCase();
      const summary = (post.summary || "").toLowerCase();
      const body = (post.body || "").toLowerCase();
      return title.includes(q) || summary.includes(q) || body.includes(q);
    });
  }, [initialPosts, selectedTag, search]);

  const formatDate = (dateStr?: string | null) => {
    if (!dateStr) return "";
    try {
      return new Intl.DateTimeFormat("pt-BR", {
        day: "2-digit",
        month: "short",
        year: "numeric",
      }).format(new Date(dateStr));
    } catch {
      return dateStr;
    }
  };

  const estimateReadTime = (body = "") => {
    const words = body.split(/\s+/).length;
    const minutes = Math.max(1, Math.ceil(words / 200));
    return `${minutes} min de leitura`;
  };

  return (
    <div className="flex flex-col gap-8">
      {/* Search and Filters */}
      <div className="flex flex-col gap-4">
        <div className="relative max-w-md">
          <label htmlFor="blog-search" className="sr-only">
            Buscar artigos
          </label>
          <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-muted">
            <Search size={18} aria-hidden="true" />
          </div>
          <input
            id="blog-search"
            type="search"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Buscar por artigos, tecnologias, empresas..."
            className="w-full rounded-lg border border-surface-border bg-surface-1 py-2 pl-10 pr-9 text-sm text-foreground placeholder:text-muted focus:border-accent focus:outline-none focus:ring-1 focus:ring-accent"
          />
          {search && (
            <button
              type="button"
              onClick={() => setSearch("")}
              className="absolute inset-y-0 right-0 flex items-center pr-3 text-muted hover:text-foreground"
              aria-label="Limpar busca"
            >
              <X size={16} aria-hidden="true" />
            </button>
          )}
        </div>

        <div role="tablist" aria-label="Filtros do blog" className="flex flex-wrap items-center gap-2">
          {tags.map((tag) => {
            const isSelected = selectedTag === tag;
            return (
              <button
                key={tag}
                type="button"
                role="tab"
                aria-selected={isSelected}
                onClick={() => setSelectedTag(tag)}
                className={`inline-flex items-center gap-1.5 rounded-full px-3.5 py-1.5 text-xs font-medium transition-all ${
                  isSelected
                    ? "bg-accent text-accent-foreground font-semibold shadow-sm"
                    : "bg-surface-2 text-muted hover:bg-surface-3 hover:text-foreground"
                }`}
              >
                {tag}
              </button>
            );
          })}
        </div>
      </div>

      {/* Posts Grid */}
      {filteredPosts.length === 0 ? (
        <div className="flex flex-col items-center gap-3 rounded-xl border border-dashed border-surface-border py-16 text-center text-muted">
          <Newspaper size={32} aria-hidden="true" />
          <p className="max-w-md text-sm">
            Nenhuma publicação encontrada para o critério selecionado.
          </p>
          {(search || selectedTag !== "Todos") && (
            <button
              type="button"
              onClick={() => {
                setSearch("");
                setSelectedTag("Todos");
              }}
              className="mt-2 text-xs font-semibold text-accent hover:underline"
            >
              Limpar filtros
            </button>
          )}
        </div>
      ) : (
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {filteredPosts.map((post) => (
            <Link key={post.id} href={`/blog/${post.slug}`} className="group block h-full">
              <Card className="flex h-full flex-col border-surface-border transition-all duration-200 hover:-translate-y-0.5 hover:border-accent/40 hover:shadow-md">
                <CardHeader className="flex flex-col gap-2.5">
                  <div className="flex items-center justify-between text-xs text-muted">
                    <span className="inline-flex items-center gap-1">
                      <Calendar size={13} aria-hidden="true" />
                      {formatDate(post.published_at || post.updated_at)}
                    </span>
                    <span className="inline-flex items-center gap-1">
                      <Clock size={13} aria-hidden="true" />
                      {estimateReadTime(post.body)}
                    </span>
                  </div>
                  <CardTitle as="h2" className="text-lg font-bold leading-snug text-foreground group-hover:text-accent">
                    {post.title}
                  </CardTitle>
                </CardHeader>
                <CardContent className="flex flex-1 flex-col justify-between gap-4">
                  <CardDescription className="line-clamp-3 text-sm leading-relaxed text-muted">
                    {post.summary}
                  </CardDescription>
                  <span className="inline-flex items-center gap-1.5 text-xs font-semibold text-accent transition-transform group-hover:translate-x-1">
                    Ler artigo completo <ArrowRight size={13} aria-hidden="true" />
                  </span>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
