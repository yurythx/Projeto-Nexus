# -*- coding: utf-8 -*-
import subprocess
import os

src_dir = '/root/Projetos/Projeto-Aurora-v3'
dst_dir = '/root/Projetos/assistencia-social'

# 1. Stop existing projeto-aurora containers (preserve volumes!)
print("Stopping old projeto-aurora containers...")
subprocess.run(["docker", "compose", "-p", "projeto-aurora", "down"], cwd=src_dir)

# 2. Rename directory
print(f"Renaming directory {src_dir} -> {dst_dir}...")
if os.path.exists(src_dir) and not os.path.exists(dst_dir):
    os.rename(src_dir, dst_dir)
elif os.path.exists(src_dir) and os.path.exists(dst_dir):
    # If dst exists, move contents
    pass

# 3. Update docker-compose.yml in dst_dir
compose_path = os.path.join(dst_dir, 'docker-compose.yml')
with open(compose_path, 'r', encoding='utf-8') as f:
    c = f.read()

# Replace project name and network
c = c.replace('name: projeto-aurora', 'name: assistencia-social')
c = c.replace('aurora_internal', 'assistencia_internal')

# Update volumes to point to existing data
old_vols = '''volumes:
  postgres_data: null
  rabbitmq_data: null
  minio_data: null
  redis_data: null'''

new_vols = '''volumes:
  postgres_data:
    external: true
    name: projeto-aurora_postgres_data
  rabbitmq_data:
    external: true
    name: projeto-aurora_rabbitmq_data
  minio_data:
    external: true
    name: projeto-aurora_minio_data
  redis_data:
    external: true
    name: projeto-aurora_redis_data'''

c = c.replace(old_vols, new_vols)

with open(compose_path, 'w', encoding='utf-8') as f:
    f.write(c)
print("docker-compose.yml updated with name: assistencia-social and external volumes")

# 4. Update symlink /opt/assistencia-social
subprocess.run(["ln", "-sfn", dst_dir, "/opt/assistencia-social"])
print("Symlink /opt/assistencia-social updated")

# 5. Start new stack with docker compose up -d
print("Starting new assistencia-social stack...")
subprocess.run(["docker", "compose", "up", "-d"], cwd=dst_dir)
print("Migration to assistencia-social stack complete!")
