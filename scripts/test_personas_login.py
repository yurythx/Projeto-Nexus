import json
import urllib.request

creds = [
    ("carlos.diretoria", "Nexus@Rondonopolis2026!"),
    ("mariana.mkt", "Nexus@Rondonopolis2026!"),
    ("lucas.tech", "Nexus@Rondonopolis2026!"),
    ("marcos.ops", "Nexus@Rondonopolis2026!"),
    ("ana.protocolo", "Nexus@Rondonopolis2026!"),
    ("rodrigo.agro", "Nexus@Rondonopolis2026!"),
]

for username, password in creds:
    body = json.dumps({"username": username, "password": password}).encode()
    req = urllib.request.Request("http://127.0.0.1:8010/api/v1/auth/login", data=body, headers={"Content-Type": "application/json"})
    try:
        res = urllib.request.urlopen(req)
        data = json.loads(res.read().decode())
        token = data["data"]["access_token"]
        req_me = urllib.request.Request("http://127.0.0.1:8010/api/v1/me", headers={"Authorization": "Bearer " + token})
        res_me = urllib.request.urlopen(req_me)
        me = json.loads(res_me.read().decode())["data"]
        perms_count = len(me.get("permissions", []))
        scopes_count = len(me.get("scopes", []))

        req_proc = urllib.request.Request("http://127.0.0.1:8010/api/v1/tramite/processos", headers={"Authorization": "Bearer " + token})
        res_proc = urllib.request.urlopen(req_proc)
        procs = json.loads(res_proc.read().decode())["data"]
        procs_count = len(procs) if isinstance(procs, list) else 0
        print(f"[OK] {username}: {me['name']} | escopos={scopes_count} | permissoes={perms_count} | processos={procs_count}")
    except Exception as e:
        print(f"[FAIL] {username}: {e}")

