import bcrypt
import subprocess

password = "Admin@2026"
password_hash = bcrypt.hashpw(password.encode('utf-8'), bcrypt.gensalt()).decode('utf-8')

sql = f"""
INSERT INTO users (id, keycloak_subject, username, email, display_name, active, password_hash, roles, created_at, updated_at)
VALUES (gen_random_uuid(), NULL, 'admin', 'admin@assistenciasocial.local', 'Administrador SEMPRAS', true, '{password_hash}', ARRAY['aurora-admin', 'aurora-user'], now(), now())
ON CONFLICT (username) DO UPDATE SET
  password_hash = '{password_hash}',
  active = true,
  failed_login_attempts = 0,
  locked_until = NULL,
  roles = ARRAY['aurora-admin', 'aurora-user'],
  display_name = 'Administrador SEMPRAS',
  email = 'admin@assistenciasocial.local',
  updated_at = now();
"""

proc = subprocess.run(
    ["docker", "exec", "-i", "projeto-aurora-postgres-1", "psql", "-U", "aurora", "-d", "aurora"],
    input=sql,
    text=True,
    capture_output=True
)

print("Return code:", proc.returncode)
print("Stdout:", proc.stdout)
print("Stderr:", proc.stderr)
print(f"User 'admin' configured with password: {password}")
