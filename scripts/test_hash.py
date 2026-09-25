try:
    import bcrypt
    print("bcrypt: available")
    h = bcrypt.hashpw(b"Admin@2026", bcrypt.gensalt()).decode()
    print("hash:", h)
except Exception as e:
    print("bcrypt error:", e)
