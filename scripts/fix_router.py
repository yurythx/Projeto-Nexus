with open('/root/Projetos/Projeto-Aurora-v3/backend/internal/app/router.go', 'r', encoding='utf-8') as f:
    c = f.read()

bad = 'AllowedOrigins: []string{deps.Config.FrontendURL,  http://localhost:3005, http://127.0.0.1:3005, http://192.168.1.42:3005},'
good = 'AllowedOrigins: []string{deps.Config.FrontendURL, "http://localhost:3005", "http://127.0.0.1:3005", "http://192.168.1.42:3005"},'

if bad in c:
    c = c.replace(bad, good)
    with open('/root/Projetos/Projeto-Aurora-v3/backend/internal/app/router.go', 'w', encoding='utf-8') as f:
        f.write(c)
    print("router.go fixed with proper quotes")
else:
    print("bad pattern not found, checking...")
