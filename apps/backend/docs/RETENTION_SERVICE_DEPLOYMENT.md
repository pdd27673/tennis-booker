# Tennis Court Data Retention Service - Deployment Guide

This document provides comprehensive instructions for deploying, configuring, and monitoring the Tennis Court Data Retention Service.

## Overview

The Tennis Court Data Retention Service is an intelligent data cleanup system that automatically removes old court slots from MongoDB while preserving data that matches active user preferences or has already triggered notifications. The service runs as a Kubernetes CronJob and includes comprehensive monitoring and alerting.

## Architecture

### Components

1. **Retention Service Application** (`cmd/retention-service/`)
   - Go application with cron scheduling capability
   - Configurable retention policies
   - Dry-run mode for testing
   - Comprehensive logging and metrics

2. **Docker Container** (`Dockerfile.retention`)
   - Multi-stage build for optimized image size
   - Non-root user for security
   - Health checks and proper signal handling

3. **Kubernetes Resources**
   - CronJob for scheduled execution
   - ConfigMaps for configuration
   - PersistentVolumeClaim for metrics storage
   - ServiceMonitor and PrometheusRule for monitoring

4. **Monitoring and Alerting**
   - Prometheus metrics collection
   - Grafana dashboard
   - Alert rules for failures and anomalies
   - Structured logging with JSON output

## Quick Start

### Prerequisites

- Kubernetes cluster with kubectl access
- Docker for building images
- Go 1.23+ for local development
- MongoDB instance accessible from the cluster

### Basic Deployment

```bash
# 1. Build and deploy with default settings
make docker-build-retention
make k8s-deploy-retention

# 2. Or use the deployment script
./scripts/deploy-retention-service.sh

# 3. Verify deployment
kubectl get cronjob tennis-court-retention-service -n tennis-booker
```

### Test Deployment

```bash
# Run in dry-run mode first
./scripts/deploy-retention-service.sh --dry-run

# Run test job
make k8s-test-retention

# Check logs
kubectl logs -l app=tennis-court-retention-service -n tennis-booker
```

## Configuration

### Environment Variables

The service is configured through environment variables:

#### Core Retention Settings
- `RETENTION_WINDOW_HOURS`: Hours before slots are eligible for deletion (default: 168 = 7 days)
- `RETENTION_BATCH_SIZE`: Number of slots to process per batch (default: 1000, max: 10000)
- `RETENTION_DRY_RUN`: Enable dry-run mode (default: false)

#### Scheduling
- `RETENTION_CRON_EXPRESSION`: Cron expression for scheduling (default: "0 3 * * *" = daily at 3 AM UTC)
- `RETENTION_RUN_ONCE`: Run once and exit instead of scheduling (default: false)

#### Logging and Monitoring
- `RETENTION_LOG_LEVEL`: Log level - "info" or "debug" (default: info)
- `RETENTION_LOG_FORMAT`: Log format - "text" or "json" (default: json)
- `RETENTION_ENABLE_METRICS`: Enable metrics collection (default: true)
- `RETENTION_METRICS_FILE`: Metrics output file path (default: /var/metrics/retention-metrics.json)

#### Database
- `MONGO_URI`: MongoDB connection string
- `DATABASE_NAME`: Database name (default: tennis_booker)

#### Vault Integration (Optional)
- `VAULT_ADDR`: Vault server address
- `VAULT_TOKEN`: Vault authentication token

### Kubernetes Configuration

Update the CronJob manifest (`k8s/retention-service-cronjob.yaml`) to customize:

```yaml
env:
- name: RETENTION_WINDOW_HOURS
  value: "168"  # 7 days
- name: RETENTION_BATCH_SIZE
  value: "1000"
- name: RETENTION_DRY_RUN
  value: "false"
```

## Deployment Options

### 1. Kubernetes CronJob (Recommended)

The primary deployment method using Kubernetes CronJob:

```bash
# Deploy with default schedule (daily at 3 AM UTC)
kubectl apply -f k8s/retention-service-cronjob.yaml

# Deploy with custom schedule
# Edit the schedule field in the manifest:
# schedule: "0 2 * * *"  # 2 AM UTC
kubectl apply -f k8s/retention-service-cronjob.yaml
```

### 2. Standalone Container

For testing or non-Kubernetes environments:

```bash
# Build image
docker build -f Dockerfile.retention -t tennis-booker/retention-service:latest .

# Run once in dry-run mode
docker run --rm \
  -e RETENTION_RUN_ONCE=true \
  -e RETENTION_DRY_RUN=true \
  -e MONGO_URI="mongodb://localhost:27017" \
  tennis-booker/retention-service:latest

# Run with custom schedule
docker run -d \
  -e RETENTION_CRON_EXPRESSION="0 4 * * *" \
  -e MONGO_URI="mongodb://localhost:27017" \
  tennis-booker/retention-service:latest
```

### 3. Local Development

For development and testing:

```bash
# Run tests
make test

# Run in test mode (dry-run with debug logging)
make run-retention-test

# Run in dry-run mode
make run-retention-dry-run

# Run with custom configuration
RETENTION_WINDOW_HOURS=24 RETENTION_DRY_RUN=true go run ./cmd/retention-service/main.go
```

## Monitoring and Alerting

### Metrics Collection

The service collects comprehensive metrics:

- **Execution Metrics**: Duration, start/end times
- **Data Metrics**: Slots found, checked, deleted
- **Performance Metrics**: Batch processing statistics
- **Error Metrics**: Error counts and types

Metrics are saved to `/var/metrics/retention-metrics.json` and can be collected by monitoring systems.

### Prometheus Integration

Deploy monitoring resources:

```bash
kubectl apply -f k8s/retention-service-monitoring.yaml
```

This creates:
- ServiceMonitor for Prometheus scraping
- PrometheusRule with alert definitions
- Grafana dashboard configuration

### Alert Rules

The service includes alerts for:

1. **Job Failures**: Immediate alert when retention job fails
2. **Long Running Jobs**: Warning when job runs longer than 2 hours
3. **No Recent Success**: Warning when no successful run in 48 hours
4. **Suspended CronJob**: Warning when CronJob is suspended

### Grafana Dashboard

Import the dashboard from the ConfigMap:

```bash
kubectl get configmap retention-service-dashboard -n tennis-booker -o jsonpath='{.data.dashboard\.json}' | jq .
```

The dashboard shows:
- Job success rate over time
- Last job duration
- Job history and status
- Error trends

## Operational Procedures

### Daily Operations

1. **Check Job Status**
   ```bash
   kubectl get cronjob tennis-court-retention-service -n tennis-booker
   kubectl get jobs -l app=tennis-court-retention-service -n tennis-booker
   ```

2. **Review Logs**
   ```bash
   kubectl logs -l app=tennis-court-retention-service -n tennis-booker --tail=100
   ```

3. **Check Metrics**
   ```bash
   kubectl exec -it <pod-name> -n tennis-booker -- cat /var/metrics/retention-metrics.json
   ```

### Troubleshooting

#### Job Failures

1. Check job logs:
   ```bash
   kubectl describe job <job-name> -n tennis-booker
   kubectl logs job/<job-name> -n tennis-booker
   ```

2. Common issues:
   - Database connectivity problems
   - Insufficient permissions
   - Resource limits exceeded
   - Configuration errors

#### Performance Issues

1. Monitor resource usage:
   ```bash
   kubectl top pod -l app=tennis-court-retention-service -n tennis-booker
   ```

2. Adjust batch size if needed:
   ```bash
   kubectl patch cronjob tennis-court-retention-service -n tennis-booker -p '{"spec":{"jobTemplate":{"spec":{"template":{"spec":{"containers":[{"name":"retention-service","env":[{"name":"RETENTION_BATCH_SIZE","value":"500"}]}]}}}}}}'
   ```

#### Data Validation

1. Run in dry-run mode to preview deletions:
   ```bash
   kubectl create job --from=cronjob/tennis-court-retention-service retention-test -n tennis-booker
   kubectl patch job retention-test -n tennis-booker -p '{"spec":{"template":{"spec":{"containers":[{"name":"retention-service","env":[{"name":"RETENTION_DRY_RUN","value":"true"}]}]}}}}'
   ```

### Maintenance

#### Updating Configuration

1. Update ConfigMap:
   ```bash
   kubectl patch configmap retention-service-config -n tennis-booker --patch '{"data":{"retention-window-hours":"240"}}'
   ```

2. Restart CronJob (delete next scheduled job):
   ```bash
   kubectl delete job -l app=tennis-court-retention-service -n tennis-booker
   ```

#### Scaling and Performance Tuning

1. **Batch Size**: Increase for better performance, decrease for lower memory usage
2. **Retention Window**: Adjust based on data growth and business requirements
3. **Schedule**: Consider data volume and system load when scheduling

#### Backup and Recovery

1. **Before Major Changes**: Run dry-run mode to preview impact
2. **Database Backups**: Ensure MongoDB backups are current before retention runs
3. **Rollback**: If issues occur, temporarily suspend CronJob and restore from backup

## Security Considerations

### Container Security

- Runs as non-root user (UID 1001)
- Read-only root filesystem where possible
- Minimal base image (Alpine Linux)
- No unnecessary capabilities

### Network Security

- Uses Kubernetes service accounts
- Network policies can restrict database access
- TLS encryption for database connections

### Secrets Management

- Database credentials stored in Kubernetes secrets
- Vault integration for enhanced secret management
- No hardcoded credentials in images or manifests

## Testing Strategy

### Unit Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage
```

### Integration Tests

```bash
# Test against real MongoDB instance
MONGO_URI="mongodb://test-instance:27017" make test

# Test in Kubernetes
make k8s-test-retention
```

### End-to-End Testing

1. **Staging Environment**: Deploy to staging with test data
2. **Dry-Run Validation**: Always test with dry-run mode first
3. **Monitoring Validation**: Verify alerts and metrics work correctly

## Performance Characteristics

### Resource Requirements

- **CPU**: 100m request, 500m limit
- **Memory**: 128Mi request, 512Mi limit
- **Storage**: 1Gi for metrics (optional)

### Scalability

- Processes up to 10,000 slots per batch
- Typical runtime: 5-30 minutes depending on data volume
- Memory usage scales with batch size

### Database Impact

- Uses efficient MongoDB queries with proper indexing
- Batch processing minimizes lock contention
- Configurable batch sizes prevent overwhelming the database

## Changelog and Versioning

### Version 1.0.0 (Initial Release)

- Basic retention logic with preference matching
- Kubernetes CronJob deployment
- Comprehensive monitoring and alerting
- Dry-run mode for safe testing
- Configurable retention policies

### Future Enhancements

- Advanced retention policies (e.g., venue-specific rules)
- Integration with data archival systems
- Enhanced metrics and reporting
- Multi-region deployment support

## Support and Troubleshooting

### Common Issues

1. **"No MongoDB URI available"**: Check secrets and Vault configuration
2. **"Permission denied"**: Verify service account permissions
3. **"Job timeout"**: Increase activeDeadlineSeconds or reduce batch size
4. **"No slots deleted"**: Check retention window and preference matching logic

### Getting Help

1. Check logs first: `kubectl logs -l app=tennis-court-retention-service`
2. Review metrics: Check Grafana dashboard or metrics file
3. Validate configuration: Run in dry-run mode
4. Check dependencies: Verify database connectivity and user preferences

### Contact Information

- **Development Team**: Backend team
- **Operations**: Platform team
- **Monitoring**: SRE team

---

For additional information, see:
- [API Documentation](./api/)
- [Database Schema](./database/)
- [Monitoring Runbook](./monitoring/) 