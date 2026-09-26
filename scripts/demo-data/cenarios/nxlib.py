# Utilitários dos cenários de teste (ver docs/DEPLOY.md, "Cenários").
import csv, json, os, subprocess, sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", ".."))
CSV = os.path.join(ROOT, "deploy", "demo", "usuarios.csv")
NX = os.path.join(os.path.dirname(os.path.abspath(__file__)), "nx")
USERS = list(csv.DictReader(open(CSV, encoding="utf-8-sig"), delimiter=";"))
FAILS = []

def call(user, method, path, body=None):
    args = [NX, user, method, path] + ([json.dumps(body)] if body is not None else [])
    out = subprocess.run(args, capture_output=True, text=True).stdout
    first, _, rest = out.partition("\n")
    code = int(first.split()[1]) if first.startswith("HTTP") else 0
    try:
        data = json.loads(rest) if rest.strip() else None
    except json.JSONDecodeError:
        data = rest
    return code, data

def users_in(group):
    return [u["usuario"] for u in USERS if group in u["grupos"].split()]

def check(label, cond, detail=""):
    print(("  OK   " if cond else "  FALHA") + f" {label}" + ("" if cond else f"  -> {str(detail)[:400]}"))
    if not cond:
        FAILS.append(label)

def expect(label, got, want, data=None):
    check(f"{label} [{got}]", got == want if isinstance(want, int) else got in want, data)

def done():
    print(f"\n{len(FAILS)} falha(s)" + (": " + "; ".join(FAILS) if FAILS else ""))
    sys.exit(1 if FAILS else 0)
