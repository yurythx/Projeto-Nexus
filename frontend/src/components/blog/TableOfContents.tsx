"use client";

import { useEffect, useMemo, useState } from "react";
import { ListCollapse } from "lucide-react";

import { slugifyHeading } from "@/lib/blog/slugify";

interface HeadingItem {
  id: string;
  text: string;
  level: number;
}

interface TableOfContentsProps {
  source: string;
}

export function TableOfContents({ source }: TableOfContentsProps) {
  const [activeId, setActiveId] = useState<string>("");

  const headings = useMemo(() => {
    const lines = source.replace(/\r\n/g, "\n").split("\n");
    const items: HeadingItem[] = [];

    for (const line of lines) {
      const match = /^(#{2,3})\s+(.*)$/.exec(line);
      if (match) {
        const level = match[1]?.length ?? 2;
        const text = match[2]?.trim() ?? "";
        if (text) {
          items.push({
            id: slugifyHeading(text),
            text,
            level,
          });
        }
      }
    }
    return items;
  }, [source]);

  useEffect(() => {
    if (headings.length === 0) return;

    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            setActiveId(entry.target.id);
          }
        }
      },
      { rootMargin: "0px 0px -70% 0px", threshold: 0.1 },
    );

    for (const h of headings) {
      const el = document.getElementById(h.id);
      if (el) observer.observe(el);
    }

    return () => observer.disconnect();
  }, [headings]);

  if (headings.length === 0) return null;

  return (
    <nav
      aria-label="Sumário do artigo"
      className="sticky top-24 flex flex-col gap-3 rounded-xl border border-surface-border bg-surface/60 p-4 backdrop-blur-md"
    >
      <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-muted">
        <ListCollapse size={14} aria-hidden="true" />
        <span>Neste Artigo</span>
      </div>

      <ul className="flex flex-col gap-1.5 text-xs">
        {headings.map((h) => {
          const isActive = activeId === h.id;
          return (
            <li
              key={h.id}
              className={`${h.level === 3 ? "pl-3 text-[11px]" : "font-medium"}`}
            >
              <a
                href={`#${h.id}`}
                onClick={(e) => {
                  e.preventDefault();
                  const target = document.getElementById(h.id);
                  if (target) {
                    target.scrollIntoView({ behavior: "smooth" });
                    setActiveId(h.id);
                  }
                }}
                className={`block py-1 transition-colors hover:text-accent ${
                  isActive
                    ? "font-semibold text-accent underline underline-offset-4"
                    : "text-muted hover:text-foreground"
                }`}
              >
                {h.text}
              </a>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
