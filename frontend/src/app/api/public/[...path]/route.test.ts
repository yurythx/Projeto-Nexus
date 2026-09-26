// @vitest-environment node
import { NextRequest } from "next/server";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/lib/api/backendUrl", () => ({ BACKEND_INTERNAL_URL: "http://api:8080" }));

import { GET, POST } from "./route";

const params = (...path: string[]) => ({ params: Promise.resolve({ path }) });

describe("proxy BFF público (allowlist)", () => {
  let upstream: ReturnType<typeof vi.fn>;
  beforeEach(() => {
    upstream = vi.fn(async () => new Response('{"data":1,"error":null}', { status: 200, headers: { "content-type": "application/json" } }));
    vi.stubGlobal("fetch", upstream);
  });
  afterEach(() => vi.unstubAllGlobals());

  it("encaminha rotas da allowlist, sem token, com o IP do visitante", async () => {
    const res = await GET(new NextRequest("http://app/api/public/v1/catalog/services?q=x", { headers: { "x-forwarded-for": "203.0.113.9" } }), params("v1", "catalog", "services"));
    expect(res.status).toBe(200);
    const [url, init] = upstream.mock.calls[0]!;
    expect(String(url)).toBe("http://api:8080/api/v1/catalog/services?q=x");
    expect(init.headers.Authorization).toBeUndefined();
    expect(init.headers["X-Forwarded-For"]).toBe("203.0.113.9");
    expect(init.body).toBeUndefined();

    await POST(new NextRequest("http://app/x", { method: "POST", body: '{"a":1}', headers: { "x-request-id": "r1" } }), params("v1", "contact", "messages"));
    const post = upstream.mock.calls[1]![1];
    expect(new TextDecoder().decode(post.body)).toBe('{"a":1}');
    expect(post.headers["X-Request-ID"]).toBe("r1");
    expect(post.headers["X-Forwarded-For"]).toBeUndefined();
  });

  it("fora da allowlist (rota, método ou '..') responde 404", async () => {
    const cases: [string, string[]][] = [
      ["GET", ["v1", "me"]],
      ["POST", ["v1", "catalog", "services"]],
      ["GET", ["v1", "catalog", "..", "admin"]],
    ];
    for (const [method, path] of cases) {
      const req = new NextRequest("http://app/x", { method, body: method === "POST" ? "{}" : undefined });
      const res = await (method === "GET" ? GET : POST)(req, params(...path));
      expect(res.status).toBe(404);
    }
    expect(upstream).not.toHaveBeenCalled();
  });

  it("backend fora do ar vira 503; content-type ausente vira JSON", async () => {
    upstream.mockRejectedValueOnce(new Error("timeout"));
    expect((await GET(new NextRequest("http://app/x"), params("v1", "branding"))).status).toBe(503);
    upstream.mockResolvedValueOnce(new Response("{}", { status: 200, headers: {} }));
    const res = await GET(new NextRequest("http://app/x"), params("v1", "branding"));
    expect(res.headers.get("content-type")).toMatch(/json|text/);
  });
});
