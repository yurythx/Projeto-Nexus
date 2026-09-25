import { NextRequest, NextResponse } from "next/server";

import { BACKEND_INTERNAL_URL } from "@/lib/api/backendUrl";

// Proxy BFF ANÔNIMO — só para as rotas públicas do backend, numa
// allowlist explícita (método + prefixo). Qualquer outra rota responde
// 404 aqui: este proxy nunca vira um atalho para a API autenticada.
// O IP real do visitante segue no X-Forwarded-For para o rate limit por
// IP do backend (que só confia nele vindo de TRUSTED_PROXIES).
const ALLOW: { method: string; prefix: string }[] = [
  { method: "GET", prefix: "v1/branding" },
  { method: "GET", prefix: "v1/system/public-modules" },
  { method: "GET", prefix: "v1/catalog/" },
  { method: "GET", prefix: "v1/directory/public/" },
  { method: "GET", prefix: "v1/calendar/public/" },
  { method: "GET", prefix: "v1/signum/verify/" },
  { method: "GET", prefix: "v1/transparencia/" },
  { method: "POST", prefix: "v1/contact/messages" },
  { method: "POST", prefix: "v1/lgpd/accept-anon" },
];

async function forward(req: NextRequest, path: string[]): Promise<NextResponse> {
  const joined = path.join("/");
  if (joined.includes("..") || !ALLOW.some((a) => a.method === req.method && joined.startsWith(a.prefix))) {
    return NextResponse.json({ data: null, error: { code: "NOT_FOUND", message: "rota pública inexistente" } }, { status: 404 });
  }
  const target = new URL(`/api/${joined}`, BACKEND_INTERNAL_URL);
  target.search = req.nextUrl.search;
  const headers: Record<string, string> = {
    "Content-Type": req.headers.get("content-type") ?? "application/json",
    "X-Request-ID": req.headers.get("x-request-id") ?? crypto.randomUUID(),
  };
  const xff = req.headers.get("x-forwarded-for");
  if (xff) headers["X-Forwarded-For"] = xff;
  try {
    const res = await fetch(target, {
      method: req.method,
      headers,
      body: req.method === "GET" ? undefined : await req.arrayBuffer(),
      signal: AbortSignal.timeout(10000),
    });
    return new NextResponse(await res.text(), {
      status: res.status,
      headers: { "Content-Type": res.headers.get("content-type") ?? "application/json" },
    });
  } catch {
    return NextResponse.json(
      { data: null, error: { code: "DEPENDENCY_UNAVAILABLE", message: "serviço indisponível no momento" } },
      { status: 503 },
    );
  }
}

export async function GET(req: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  return forward(req, (await params).path);
}

export async function POST(req: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  return forward(req, (await params).path);
}
