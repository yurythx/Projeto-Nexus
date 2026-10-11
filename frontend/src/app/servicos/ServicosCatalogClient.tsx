"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { ArrowRight, LayoutGrid, Search, X } from "lucide-react";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/Card";
import { ServiceIcon } from "@/lib/catalog/icons";
import type { CatalogService } from "@/lib/nexus/types";

interface ServicosCatalogClientProps {
  initialServices: CatalogService[];
}

export function ServicosCatalogClient({ initialServices }: ServicosCatalogClientProps) {
  const [search, setSearch] = useState("");
  const [selectedCategory, setSelectedCategory] = useState<string>("Todas");

  // Extract unique categories and counts
  const categories = useMemo(() => {
    const counts = new Map<string, number>();
    for (const svc of initialServices) {
      const cat = svc.category || "Geral";
      counts.set(cat, (counts.get(cat) || 0) + 1);
    }

    const list = Array.from(counts.entries()).map(([name, count]) => ({
      name,
      count,
    }));

    // Sort alphabetically
    list.sort((a, b) => a.name.localeCompare(b.name, "pt-BR"));

    return [{ name: "Todas", count: initialServices.length }, ...list];
  }, [initialServices]);

  // Filter services by search and category
  const filteredServices = useMemo(() => {
    const q = search.trim().toLowerCase();
    return initialServices.filter((svc) => {
      const matchesCategory =
        selectedCategory === "Todas" || (svc.category || "Geral") === selectedCategory;

      if (!matchesCategory) return false;

      if (!q) return true;

      const title = svc.title.toLowerCase();
      const summary = (svc.summary || "").toLowerCase();
      const cat = (svc.category || "").toLowerCase();
      const slug = svc.slug.toLowerCase();

      return (
        title.includes(q) ||
        summary.includes(q) ||
        cat.includes(q) ||
        slug.includes(q)
      );
    });
  }, [initialServices, selectedCategory, search]);

  return (
    <div className="flex flex-col gap-8">
      {/* Search and Filters */}
      <div className="flex flex-col gap-4">
        {/* Search Input */}
        <div className="relative max-w-md">
          <label htmlFor="service-search" className="sr-only">
            Buscar serviços
          </label>
          <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-muted">
            <Search size={18} aria-hidden="true" />
          </div>
          <input
            id="service-search"
            type="search"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Buscar por nome, categoria ou palavra-chave..."
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

        {/* Category Pills */}
        <div
          role="tablist"
          aria-label="Categorias de serviços"
          className="flex flex-wrap items-center gap-2"
        >
          {categories.map((cat) => {
            const isSelected = selectedCategory === cat.name;
            return (
              <button
                key={cat.name}
                type="button"
                role="tab"
                aria-selected={isSelected}
                onClick={() => setSelectedCategory(cat.name)}
                className={`inline-flex items-center gap-1.5 rounded-full px-3.5 py-1.5 text-xs font-medium transition-all ${
                  isSelected
                    ? "bg-accent text-accent-foreground shadow-sm font-semibold"
                    : "bg-surface-2 text-muted hover:bg-surface-3 hover:text-foreground"
                }`}
              >
                <span>{cat.name}</span>
                <span
                  className={`rounded-full px-1.5 py-0.2 text-[10px] ${
                    isSelected
                      ? "bg-black/20 text-white"
                      : "bg-surface-3 text-muted"
                  }`}
                >
                  {cat.count}
                </span>
              </button>
            );
          })}
        </div>
      </div>

      {/* Services Grid or Empty State */}
      {filteredServices.length === 0 ? (
        <div className="flex flex-col items-center gap-3 rounded-xl border border-dashed border-surface-border py-16 text-center text-muted">
          <LayoutGrid size={32} aria-hidden="true" />
          <p className="max-w-md text-sm">
            Nenhum serviço encontrado para o filtro aplicado.
          </p>
          {(search || selectedCategory !== "Todas") && (
            <button
              type="button"
              onClick={() => {
                setSearch("");
                setSelectedCategory("Todas");
              }}
              className="mt-2 text-xs font-semibold text-accent hover:underline"
            >
              Limpar filtros de busca
            </button>
          )}
        </div>
      ) : (
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {filteredServices.map((svc) => (
            <Link key={svc.id} href={`/servicos/${svc.slug}`} className="group block h-full">
              <Card className="flex h-full flex-col border-surface-border transition-all duration-200 hover:-translate-y-0.5 hover:border-accent/40 hover:shadow-md">
                <CardHeader className="flex flex-col gap-3">
                  <div className="flex items-center justify-between gap-2">
                    <span className="inline-flex h-11 w-11 items-center justify-center rounded-xl bg-accent/10 text-accent transition-colors group-hover:bg-accent group-hover:text-accent-foreground">
                      <ServiceIcon name={svc.icon} size={22} aria-hidden="true" />
                    </span>
                    {svc.category && (
                      <span className="rounded-md bg-surface-2 px-2.5 py-1 text-[11px] font-medium text-muted">
                        {svc.category}
                      </span>
                    )}
                  </div>
                  <CardTitle as="h2" className="text-lg font-bold text-foreground">
                    {svc.title}
                  </CardTitle>
                </CardHeader>
                <CardContent className="flex flex-1 flex-col justify-between gap-4">
                  <CardDescription className="line-clamp-3 text-sm leading-relaxed text-muted">
                    {svc.summary}
                  </CardDescription>
                  <span className="inline-flex items-center gap-1.5 text-xs font-semibold text-accent transition-transform group-hover:translate-x-1">
                    Conhecer solução <ArrowRight size={13} aria-hidden="true" />
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
