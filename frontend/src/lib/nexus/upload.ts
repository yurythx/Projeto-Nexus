import type { UploadTicket } from "@/lib/nexus/types";

/** Envia o arquivo direto ao object storage (MinIO) pela URL pré-assinada
 * emitida pelo backend — o binário nunca passa pela API nem pelo BFF. */
export async function putToTicket(ticket: UploadTicket, file: Blob): Promise<void> {
  const res = await fetch(ticket.upload_url, {
    method: ticket.method || "PUT",
    headers: ticket.headers,
    body: file,
  });
  if (!res.ok) throw new Error(`upload recusado pelo armazenamento (HTTP ${res.status})`);
}

/** Tipo MIME com fallback genérico (alguns SOs não informam). */
export function mimeOf(file: File): string {
  return file.type || "application/octet-stream";
}

/** SHA-256 hexadecimal de um arquivo (Web Crypto) — usado para conferir
 * a integridade de documentos assinados pelo Signum. */
export async function sha256Hex(file: Blob): Promise<string> {
  const digest = await crypto.subtle.digest("SHA-256", await file.arrayBuffer());
  return [...new Uint8Array(digest)].map((b) => b.toString(16).padStart(2, "0")).join("");
}
