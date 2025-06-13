{
  "database": {
    {{- with secret "kv/data/tennisapp/prod/database" }}
    "uri": "mongodb://{{ .Data.data.username }}:{{ .Data.data.password }}@mongodb:27017/{{ .Data.data.database }}?authSource=admin",
    "name": "{{ .Data.data.database }}"
    {{- end }}
  },
  "redis": {
    {{- with secret "kv/data/tennisapp/prod/redis" }}
    "addr": "redis:6379",
    "password": "{{ .Data.data.password }}",
    "db": 0
    {{- end }}
  },
  "vault": {
    "addr": "http://vault:8200"
  },
  "scraper": {
    "interval": 300,
    "timeout": 30,
    "log_level": "INFO",
    "user_agents": [
      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
    ]
  }
} 