# Tennis Booker Production Deployment Guide

This guide covers the complete production deployment of Tennis Booker on Oracle Cloud Infrastructure (OCI) using Docker Compose, Traefik reverse proxy, and Let's Encrypt SSL certificates.

## 🏗️ Architecture Overview

```
Internet → DuckDNS → OCI Instance → Traefik → Services
                                      ├── Frontend (React + Nginx)
                                      ├── Backend API (Go)
                                      ├── MongoDB
                                      ├── Redis
                                      ├── Vault
                                      ├── Notification Service
                                      └── Scraper Service
```

## 📋 Prerequisites

### 1. OCI Infrastructure
- OCI Always Free tier account
- ARM Ampere A1 compute instance (2 OCPUs, 12GB RAM)
- Public IP address assigned
- Ports 80, 443, and 8080 open in security lists

### 2. Domain Setup
- DuckDNS account (free at https://www.duckdns.org/)
- Subdomain configured (e.g., `mytennis.duckdns.org`)
- DuckDNS token obtained

### 3. Email Configuration
- Gmail account with App Password enabled
- App Password generated for SMTP access

## 🚀 Quick Start Deployment

### Step 1: Provision Infrastructure

1. **Deploy OCI Infrastructure with Terraform:**
   ```bash
   cd infrastructure/terraform
   cp terraform.tfvars.example terraform.tfvars
   # Edit terraform.tfvars with your OCI credentials
   terraform init
   terraform plan
   terraform apply
   ```

2. **Connect to your OCI instance:**
   ```bash
   ssh -i ~/.oci/oci_api_key.pem opc@<your-instance-ip>
   ```

### Step 2: Setup Application

1. **Clone the repository:**
   ```bash
   git clone https://github.com/yourusername/tennis-booker.git
   cd tennis-booker
   ```

2. **Configure environment:**
   ```bash
   cp env.prod.example .env.prod
   # Edit .env.prod with your actual values
   ```

3. **Generate Traefik password:**
   ```bash
   ./infrastructure/traefik/generate-password.sh admin your-secure-password
   # Copy the output hash to TRAEFIK_PASSWORD_HASH in .env.prod
   ```

4. **Deploy the application:**
   ```bash
   chmod +x infrastructure/scripts/deploy.sh
   ./infrastructure/scripts/deploy.sh
   ```

## ⚙️ Configuration Details

### Environment Variables (.env.prod)

```bash
# Domain Configuration
DOMAIN_NAME=mytennis.duckdns.org
DUCKDNS_TOKEN=your-duckdns-token

# SSL Configuration
ACME_EMAIL=your-email@example.com

# Traefik Dashboard
TRAEFIK_USER=admin
TRAEFIK_PASSWORD_HASH=$2y$10$generated.hash.here

# Database Configuration
MONGO_ROOT_USERNAME=admin
MONGO_ROOT_PASSWORD=secure-mongo-password
REDIS_PASSWORD=secure-redis-password

# Application Configuration
JWT_SECRET=your-very-long-jwt-secret
USER_EMAIL=admin@yourdomain.com

# Email Configuration
GMAIL_EMAIL=your-gmail@gmail.com
GMAIL_PASSWORD=your-gmail-app-password
FROM_EMAIL=your-gmail@gmail.com

# Scraper Configuration
SCRAPER_INTERVAL=300  # 5 minutes
```

### Service URLs

After deployment, your services will be available at:

- **Frontend:** `https://mytennis.duckdns.org`
- **API:** `https://api.mytennis.duckdns.org`
- **Traefik Dashboard:** `https://traefik.mytennis.duckdns.org`

## 🔧 Management Commands

### Service Management

```bash
# View service status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f [service-name]

# Restart services
docker-compose -f docker-compose.prod.yml restart

# Stop all services
docker-compose -f docker-compose.prod.yml down

# Start all services
docker-compose -f docker-compose.prod.yml up -d
```

### Updates

```bash
# Update to latest code and rebuild
./infrastructure/scripts/update.sh

# Manual update
git pull origin main
docker-compose -f docker-compose.prod.yml build --no-cache
docker-compose -f docker-compose.prod.yml up -d --force-recreate
```

### Backups

```bash
# Create backup
./infrastructure/scripts/backup.sh

# List backups
ls -la backups/

# Restore from backup
./infrastructure/scripts/restore.sh backups/20240101_120000
```

### DuckDNS Management

```bash
# Manual IP update
./infrastructure/scripts/update-duckdns.sh

# Check automatic update service
sudo systemctl status duckdns-update.timer
sudo journalctl -u duckdns-update.service
```

## 🔍 Monitoring & Troubleshooting

### Health Checks

```bash
# Check all services
docker-compose -f docker-compose.prod.yml ps

# Test endpoints
curl -I https://mytennis.duckdns.org
curl -I https://api.mytennis.duckdns.org/health
```

### Log Analysis

```bash
# Traefik logs (SSL certificate issues)
docker-compose -f docker-compose.prod.yml logs -f traefik

# Backend API logs
docker-compose -f docker-compose.prod.yml logs -f backend

# Database logs
docker-compose -f docker-compose.prod.yml logs -f mongodb

# Scraper logs
docker-compose -f docker-compose.prod.yml logs -f scraper-service
```

### Common Issues

#### SSL Certificate Not Generated
```bash
# Check Traefik logs
docker-compose -f docker-compose.prod.yml logs traefik

# Verify domain resolution
nslookup mytennis.duckdns.org

# Update DuckDNS IP
./infrastructure/scripts/update-duckdns.sh
```

#### Service Won't Start
```bash
# Check service logs
docker-compose -f docker-compose.prod.yml logs [service-name]

# Check resource usage
docker stats

# Restart specific service
docker-compose -f docker-compose.prod.yml restart [service-name]
```

#### Database Connection Issues
```bash
# Check MongoDB logs
docker-compose -f docker-compose.prod.yml logs mongodb

# Test MongoDB connection
docker exec -it tennis-mongodb mongosh -u admin -p

# Check environment variables
docker-compose -f docker-compose.prod.yml config
```

## 🔒 Security Considerations

### Firewall Configuration
```bash
# OCI Security Lists (via OCI Console)
- Ingress: 0.0.0.0/0 → Port 80 (HTTP)
- Ingress: 0.0.0.0/0 → Port 443 (HTTPS)
- Ingress: Your-IP/32 → Port 8080 (Traefik Dashboard)
- Ingress: Your-IP/32 → Port 22 (SSH)

# Instance Firewall
sudo firewall-cmd --list-all
```

### SSL/TLS Configuration
- Automatic SSL certificate generation via Let's Encrypt
- HTTP to HTTPS redirection enforced
- Strong cipher suites configured
- HSTS headers enabled

### Access Control
- Traefik dashboard protected with basic authentication
- Database access restricted to internal network
- Non-root containers for all services
- Secrets managed via environment variables

## 📊 Performance Optimization

### Resource Limits
```yaml
# In docker-compose.prod.yml
services:
  backend:
    deploy:
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 256M
```

### Caching Strategy
- Redis for session storage and caching
- Nginx static file caching
- Traefik response compression
- Browser caching headers

### Database Optimization
- MongoDB indexes on frequently queried fields
- Connection pooling
- Regular backup and cleanup

## 🔄 Backup & Recovery

### Automated Backups
```bash
# Setup cron job for daily backups
crontab -e
# Add: 0 2 * * * /opt/tennis-booker/infrastructure/scripts/backup.sh
```

### Backup Contents
- MongoDB database dump
- Redis data snapshot
- Vault secrets
- Traefik SSL certificates
- Configuration files

### Recovery Process
1. Stop all services
2. Restore data from backup
3. Restart services
4. Verify functionality

## 📈 Scaling Considerations

### Horizontal Scaling
- Load balancer configuration with Traefik
- Database replication setup
- Shared storage for uploads

### Vertical Scaling
- Increase OCI instance resources
- Adjust Docker resource limits
- Monitor performance metrics

## 🆘 Support & Maintenance

### Regular Maintenance Tasks
- Weekly security updates
- Monthly backup verification
- Quarterly dependency updates
- SSL certificate renewal (automatic)

### Monitoring Setup
- Docker container health checks
- Disk space monitoring
- SSL certificate expiration alerts
- Service availability monitoring

### Emergency Procedures
1. **Service Down:** Check logs, restart services
2. **SSL Issues:** Verify domain resolution, check Traefik logs
3. **Database Issues:** Check connections, restore from backup
4. **High Load:** Scale resources, check for issues

## 📞 Getting Help

- **Documentation:** Check this guide and inline comments
- **Logs:** Always check service logs first
- **Community:** GitHub Issues for bug reports
- **Emergency:** Use backup and restore procedures

---

**Note:** This deployment guide assumes familiarity with Docker, Linux system administration, and basic networking concepts. Always test changes in a development environment before applying to production. 