import urllib.request
import json

url = "http://127.0.0.1:8005/api/v1/auth/login"
payload = {"username": "admin", "password": "Admin@2026"}
data = json.dumps(payload).encode("utf-8")
req = urllib.request.Request(url, data=data, headers={"Content-Type": "application/json"})

try:
    with urllib.request.urlopen(req) as resp:
        body = json.loads(resp.read().decode("utf-8"))
        print("LOGIN HTTP STATUS:", resp.status)
        print("FULL BODY KEYS:", list(body.get("data", {}).keys()))
        print("FULL BODY:", json.dumps(body.get("data", {}), indent=2)[:300])
except urllib.error.HTTPError as e:
    print("HTTP ERROR:", e.code, e.read().decode("utf-8"))
except Exception as e:
    print("OTHER ERROR:", e)
