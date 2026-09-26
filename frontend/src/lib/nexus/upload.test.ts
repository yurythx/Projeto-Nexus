import { afterEach, describe, expect, it, vi } from "vitest";

import { mimeOf, putToTicket, sha256Hex } from "./upload";

describe("upload direto ao MinIO", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("SHA-256 calculado localmente (Signum)", async () => {
    const hash = await sha256Hex(new Blob(["abc"]));
    expect(hash).toBe("ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad");
  });

  it("PUT na URL pré-assinada com os headers do ticket", async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200 });
    vi.stubGlobal("fetch", fetchMock);
    const blob = new Blob(["x"]);
    await putToTicket(
      { object_key: "k", upload_url: "https://minio.test/put", method: "PUT", headers: { "Content-Type": "application/pdf" }, expires_at: "" },
      blob,
    );
    expect(fetchMock).toHaveBeenCalledWith("https://minio.test/put", { method: "PUT", headers: { "Content-Type": "application/pdf" }, body: blob });
  });

  it("falha do storage vira erro", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false, status: 403 }));
    await expect(
      putToTicket({ object_key: "k", upload_url: "u", method: "PUT", headers: {}, expires_at: "" }, new Blob([])),
    ).rejects.toThrow(/403/);
  });

  it("tipo MIME com fallback", () => {
    expect(mimeOf(new File([""], "a.bin"))).toBe("application/octet-stream");
    expect(mimeOf(new File([""], "a.pdf", { type: "application/pdf" }))).toBe("application/pdf");
  });
});
