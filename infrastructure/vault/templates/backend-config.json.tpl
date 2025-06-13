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
  "jwt": {
    {{- with secret "kv/data/tennisapp/prod/jwt" }}
    "secret": "{{ .Data.data.secret }}"
    {{- end }}
  },
  "email": {
    {{- with secret "kv/data/tennisapp/prod/email" }}
    "gmail_email": "{{ .Data.data.email }}",
    "gmail_password": "{{ .Data.data.password }}",
    "smtp_host": "{{ .Data.data.smtp_host }}",
    "smtp_port": "{{ .Data.data.smtp_port }}"
    {{- end }}
  },
  "vault": {
    "addr": "http://vault:8200"
  },
  "api": {
    "port": "8080"
  }
} 