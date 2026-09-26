"""Blog, Wiki, Catálogo, Contato (formulário público) e Egress."""
import time
import urllib.request

from nxlib import *  # noqa: F403

RUN = int(time.time())
pickr = lambda g: (lambda l: l[RUN % len(l)])(users_in(g))  # noqa: E731
editor = "teste.conteudo"
servidor = pickr("SEMED-ESCOLA-SILVESTRE-ATD")
admin = "teste.admin"
print(f"editor={editor} servidor={servidor}")
FE = open(os.path.join(ROOT, ".env")).read().split("FRONTEND_URL=")[1].split("\n")[0]


def public(method, path, body=None):
    """Chamada anônima pelo proxy público do site (/api/public)."""
    req = urllib.request.Request(f"{FE}/api/public/v1/{path}", method=method,
                                 data=json.dumps(body).encode() if body is not None else None,
                                 headers={"Content-Type": "application/json"})
    try:
        with urllib.request.urlopen(req, timeout=15, context=SSL_CTX) as r:
            return r.status, json.loads(r.read() or b"null")
    except urllib.error.HTTPError as e:
        return e.code, json.loads(e.read() or b"null")


print("\n== Blog")
code, _ = call(servidor, "POST", "blog/posts", {"title": "Post indevido"})
expect("servidor não publica no blog", code, 403)
code, d = call(editor, "POST", "blog/posts", {"title": f"Campanha de vacinação {RUN}", "summary": "Resumo", "body": "Texto do comunicado.", "kind": "comunicado"})
expect("editor cria rascunho", code, (200, 201), d)
if code in (200, 201):
    pid, slug = d["data"]["id"], d["data"].get("slug")
    code, d = call(servidor, "GET", f"blog/posts/{pid}")
    expect("rascunho invisível ao servidor", code, (403, 404))
    code, d = call(editor, "POST", f"blog/posts/{pid}/publish")
    expect("editor publica", code, (200, 204), d)
    code, d = call(servidor, "GET", f"blog/posts/{slug or pid}")
    expect("publicado visível ao servidor", code, 200, d)

print("\n== Wiki")
code, d = call(servidor, "POST", "wiki/pages", {"title": f"Rotina do atendimento {RUN}", "body": "v1", "summary": "criação"})
expect("servidor cria página", code, (200, 201), d)
if code in (200, 201):
    wid, ver = d["data"]["id"], d["data"]["version"]
    edit = {"title": f"Rotina do atendimento {RUN}", "body": "v2", "summary": "revisão"}
    code, d = call(servidor, "PUT", f"wiki/pages/{wid}", edit)
    expect("edição sem a versão aberta é recusada", code, 422, d)
    code, d = call(servidor, "PUT", f"wiki/pages/{wid}", {**edit, "version": ver})
    expect("edita (nova revisão)", code, 200, d)
    code, d = call(servidor, "PUT", f"wiki/pages/{wid}", {**edit, "body": "v2-conflito", "version": ver})
    expect("edição sobre versão antiga = conflito", code, 409, d)
    code, d = call(servidor, "GET", f"wiki/pages/{wid}/revisions")
    expect("histórico de revisões", code, 200, d)
    if code == 200:
        check("2 revisões", len(d["data"]) >= 2, len(d["data"]))
    code, d = call(servidor, "POST", f"wiki/pages/{wid}/revisions/1/restore")
    expect("restaura a v1", code, (200, 201), d)
    code, d = call(servidor, "GET", f"wiki/pages/{wid}")
    if code == 200:
        check("conteúdo voltou para v1", d["data"].get("body") == "v1", d["data"].get("body"))
    code, _ = call(servidor, "DELETE", f"wiki/pages/{wid}")
    expect("servidor não apaga página (só moderador)", code, (403, 404))
    code, _ = call(editor, "DELETE", f"wiki/pages/{wid}")
    expect("gestor de conteúdo apaga", code, (200, 204))

print("\n== Catálogo de serviços")
_, me = call(admin, "GET", "iam/unidades")
upa = next((u["id"] for u in me["data"] if u.get("nome") == "UPA 24h"), None)
code, _ = call(servidor, "POST", "catalog/admin/services", {"title": "Serviço indevido"})
expect("servidor não gerencia catálogo", code, 403)
code, d = call(editor, "POST", "catalog/admin/services", {
    "title": f"Atendimento de urgência {RUN}", "summary": "Pronto atendimento 24h", "category": "Saúde",
    "requirements": ["Documento com foto", "Cartão SUS"], "steps": ["Classificação de risco", "Atendimento"],
    "sla": "Imediato", "cost": "Gratuito", "responsible_unidade_id": upa})
expect("editor cadastra serviço", code, (200, 201), d)
if code in (200, 201):
    sid, sslug = d["data"]["id"], d["data"]["slug"]
    code, _ = public("GET", f"catalog/services/{sslug}")
    expect("rascunho fora do site público", code, 404)
    code, d = call(editor, "POST", f"catalog/admin/services/{sid}/publish")
    expect("sem canal de atendimento não publica (Lei 13.460)", code, 422, d)
    _, cur = call(editor, "GET", f"catalog/admin/services/{sid}")
    svc = {k: cur["data"].get(k) for k in ("title", "slug", "summary", "description", "category", "audience", "requirements", "steps", "sla", "cost", "icon", "responsible_unidade_id", "position")}
    svc["channels"] = [{"type": "presencial", "label": "UPA 24h", "value": "Atendimento por ordem de classificação de risco"}]
    code, d = call(editor, "PUT", f"catalog/admin/services/{sid}", svc)
    expect("adiciona canal de atendimento", code, 200, d)
    code, d = call(editor, "POST", f"catalog/admin/services/{sid}/publish")
    expect("publica", code, (200, 204), d)
    code, d = public("GET", f"catalog/services/{sslug}")
    expect("aparece no site público (anônimo)", code, 200, d)

print("\n== Contato (formulário público anônimo)")
msg = {"name": "Cidadão Fictício", "email": f"cidadao{RUN}@exemplo.test", "subject": "Dúvida de teste",
       "category": "duvida", "message": "Mensagem de teste do formulário público.", "consent": True, "website": ""}
code, d = public("POST", "contact/messages", {**msg, "consent": False})
expect("sem consentimento LGPD é recusado", code, (400, 422), d)
code, d = public("POST", "contact/messages", msg)
expect("visitante envia mensagem", code, (200, 201, 202), d)
code, _ = call(servidor, "GET", "contact/messages")
expect("servidor não lê o contato", code, 403)
code, d = call(editor, "GET", "contact/messages")
expect("gestor lê as mensagens", code, 200, d)
if code == 200:
    mine = [m for m in d["data"] if m.get("email") == msg["email"] or msg["email"] in json.dumps(m)]
    check("mensagem do visitante na caixa", len(mine) == 1, len(mine))
    if mine:
        code, d = call(editor, "PATCH", f"contact/messages/{mine[0]['id']}", {"status": "answered", "notes": "Respondido por e-mail."})
        expect("marca como respondida", code, 200, d)

print("\n== Egress (webhooks) e anti-SSRF")
code, _ = call(servidor, "GET", "egress/targets")
expect("servidor não vê integrações", code, 403)
for url in ("http://192.168.1.42:8010/health", "http://127.0.0.1/", "http://169.254.169.254/latest/meta-data/", "http://postgres:5432/"):
    code, d = call(admin, "POST", "egress/targets", {"name": f"ssrf {RUN}", "kind": "webhook", "url": url, "event_patterns": ["*"]})
    expect(f"destino interno recusado ({url.split('/')[2]})", code, (400, 422), d)
code, d = call(admin, "POST", "egress/targets", {"name": f"Webhook público {RUN}", "kind": "webhook", "url": "https://example.com/nexus-hook", "event_patterns": ["tramite.*"]})
expect("destino público aceito", code, (200, 201), d)
if code in (200, 201):
    code, _ = call(admin, "DELETE", f"egress/targets/{d['data']['id']}")
    expect("remove destino", code, (200, 204))
done()
