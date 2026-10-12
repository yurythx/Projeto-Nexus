import type { Metadata } from "next";

import { PublicShell } from "@/components/layout/PublicShell";
import { listPublicPosts } from "@/lib/blog/publicBlog";
import { listPublishedServices } from "@/lib/catalog/catalog";
import { APP_URL } from "@/lib/env";
import { PublicBlogListClient } from "./PublicBlogListClient";

export const dynamic = "force-dynamic";

const description =
  "Artigos aprofundados, novidades e análises técnicas sobre as principais soluções de infraestrutura, automação, comunicação e segurança.";

export const metadata: Metadata = {
  title: "Blog & Novidades",
  description,
  alternates: { canonical: `${APP_URL}/blog` },
  openGraph: {
    title: "Blog & Novidades",
    description,
    type: "website",
    url: `${APP_URL}/blog`,
  },
};

export default async function BlogPublicPage() {
  const [posts, services] = await Promise.all([
    listPublicPosts("", "", 1, 100).catch(() => []),
    listPublishedServices().catch(() => []),
  ]);

  const publishedSlugs = new Set(services.map((s) => s.slug));
  const filteredPosts = posts.filter((post) => {
    if (post.slug.startsWith("artigo-")) {
      const serviceSlug = post.slug.replace(/^artigo-/, "");
      return services.length === 0 || publishedSlugs.has(serviceSlug);
    }
    return true;
  });

  return (
    <PublicShell>
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-10 px-6 py-12">
        <div className="flex flex-col gap-3">
          <p className="dateline">Blog & Novidades</p>
          <h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
            Blog & Novidades
          </h1>
          <p className="max-w-3xl text-base text-muted">{description}</p>
        </div>

        <PublicBlogListClient initialPosts={filteredPosts} />
      </div>
    </PublicShell>
  );
}
