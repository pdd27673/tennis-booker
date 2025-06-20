# Scraper Environment Variables - Generated by Vault Agent
{{- with secret "kv/data/tennisapp/prod/database" }}
MONGO_URI="mongodb://{{ .Data.data.username }}:{{ .Data.data.password }}@mongodb:27017/{{ .Data.data.database }}?authSource=admin"
MONGO_DB_NAME="{{ .Data.data.database }}"
{{- end }}

{{- with secret "kv/data/tennisapp/prod/redis" }}
REDIS_ADDR="redis:6379"
REDIS_PASSWORD="{{ .Data.data.password }}"
REDIS_DB=0
{{- end }}

# Vault Configuration
VAULT_ADDR=http://vault:8200

# Scraper Configuration
SCRAPER_INTERVAL=300
SCRAPER_TIMEOUT=30
LOG_LEVEL=INFO 