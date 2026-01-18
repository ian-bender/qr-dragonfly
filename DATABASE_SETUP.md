# Database Setup Guide

## PostgreSQL Configuration

This application uses PostgreSQL for persistent storage of QR codes and click analytics.

### Local Development

#### 1. Install PostgreSQL

**macOS (Homebrew):**

```bash
brew install postgresql@16
brew services start postgresql@16
```

**Ubuntu/Debian:**

```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
```

#### 2. Create Database

```bash
# Connect to PostgreSQL
psql postgres

# Create database and user
CREATE DATABASE qrdragonfly;
CREATE USER qruser WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE qrdragonfly TO qruser;
\q
```

#### 3. Set Environment Variables

Add to your `.env` file or export in shell:

```bash
# QR Service
export DATABASE_URL="postgresql://qruser:your_secure_password@localhost:5432/qrdragonfly"

# Click Service (can use same database)
export DATABASE_URL="postgresql://qruser:your_secure_password@localhost:5432/qrdragonfly"
```

### Production Deployment

#### Heroku Postgres

```bash
# Add Heroku Postgres addon
heroku addons:create heroku-postgresql:mini -a your-app-name

# DATABASE_URL is automatically set
# Verify with:
heroku config:get DATABASE_URL -a your-app-name
```

#### AWS RDS

1. Create RDS PostgreSQL instance
2. Configure security group to allow inbound on port 5432
3. Set connection string:

```bash
DATABASE_URL="postgresql://username:password@your-rds-instance.region.rds.amazonaws.com:5432/database_name"
```

#### Railway

```bash
# Add PostgreSQL plugin
railway add postgresql

# Railway automatically sets $DATABASE_URL
```

### Connection Pool Settings

The application uses conservative connection pool settings suitable for small instances:

- **Max Open Connections:** 10
- **Max Idle Connections:** 5
- **Connection Max Lifetime:** 30 minutes

To adjust for higher traffic, edit `cmd/server/main.go`:

```go
sqlDB.SetMaxOpenConns(25)      // Increase for more concurrency
sqlDB.SetMaxIdleConns(10)      // Keep ~40% of max open
sqlDB.SetConnMaxLifetime(30 * time.Minute)
```

### Schema Management

Schemas are automatically created/migrated on startup using GORM AutoMigrate.

#### QR Service Schema

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE qr_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    label VARCHAR NOT NULL,
    url VARCHAR NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX qr_codes_active_idx ON qr_codes(active);
CREATE INDEX qr_codes_created_at_idx ON qr_codes(created_at DESC);
```

#### Click Service Schema

```sql
CREATE TABLE click_daily_stats (
    qr_code_id VARCHAR NOT NULL,
    day DATE NOT NULL,
    total INTEGER NOT NULL DEFAULT 0,
    region_counts JSONB,
    hour00 INTEGER NOT NULL DEFAULT 0,
    hour01 INTEGER NOT NULL DEFAULT 0,
    -- ... (hour02-hour22)
    hour23 INTEGER NOT NULL DEFAULT 0,
    last_at TIMESTAMP NOT NULL,
    last_country VARCHAR NOT NULL DEFAULT '',
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    PRIMARY KEY (qr_code_id, day)
);
```

### Backup Strategy

#### Automated Backups (Heroku)

```bash
# Enable automated backups
heroku pg:backups:schedule DATABASE_URL --at '02:00 America/New_York' -a your-app

# Manual backup
heroku pg:backups:capture -a your-app

# Download backup
heroku pg:backups:download -a your-app
```

#### Manual Backup (PostgreSQL)

```bash
# Dump database
pg_dump -U qruser -h localhost qrdragonfly > backup_$(date +%Y%m%d).sql

# Restore from backup
psql -U qruser -h localhost qrdragonfly < backup_20260117.sql
```

#### Automated Backup Script

Create `scripts/backup-db.sh`:

```bash
#!/bin/bash
BACKUP_DIR="$HOME/backups/qrdragonfly"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

# Dump database
pg_dump $DATABASE_URL > "$BACKUP_DIR/backup_$DATE.sql"

# Compress
gzip "$BACKUP_DIR/backup_$DATE.sql"

# Keep only last 30 days
find $BACKUP_DIR -name "*.sql.gz" -mtime +30 -delete

echo "Backup completed: backup_$DATE.sql.gz"
```

Schedule with cron:

```bash
crontab -e
# Add: 0 2 * * * /path/to/scripts/backup-db.sh
```

### Performance Optimization

#### Add Indexes for Common Queries

```sql
-- For user-specific QR code filtering (future enhancement)
CREATE INDEX qr_codes_user_id_idx ON qr_codes(user_id) WHERE user_id IS NOT NULL;

-- For click analytics date range queries
CREATE INDEX click_daily_stats_day_idx ON click_daily_stats(day DESC);
```

#### Analyze Query Performance

```sql
EXPLAIN ANALYZE SELECT * FROM qr_codes WHERE active = true ORDER BY created_at DESC;
```

### Monitoring

#### Check Connection Count

```sql
SELECT count(*) FROM pg_stat_activity WHERE datname = 'qrdragonfly';
```

#### Check Database Size

```sql
SELECT pg_size_pretty(pg_database_size('qrdragonfly'));
```

#### Check Table Sizes

```sql
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

### Troubleshooting

#### Connection Refused

- Verify PostgreSQL is running: `pg_isready`
- Check DATABASE_URL format
- Verify firewall allows port 5432

#### Too Many Connections

- Increase max_connections in postgresql.conf
- Reduce connection pool size in application
- Check for connection leaks

#### Slow Queries

- Run VACUUM: `VACUUM ANALYZE;`
- Add missing indexes
- Check pg_stat_statements for slow queries

### Migration from In-Memory

If you're migrating from in-memory storage:

1. **Export existing data** (if running in-memory)
   - Data is lost on restart, so export before deploying
2. **Set DATABASE_URL** environment variable

3. **Restart services** - schema will auto-create

4. **Verify connection:**
   ```bash
   # Check logs for:
   # "qr-service using postgres storage"
   # "click-service using postgres storage"
   ```

### Security Best Practices

- ✅ Use strong passwords (20+ characters)
- ✅ Enable SSL/TLS connections in production
- ✅ Restrict network access (VPC, security groups)
- ✅ Rotate credentials regularly
- ✅ Use read-only credentials for analytics
- ✅ Enable audit logging
- ✅ Regular backups to separate storage
- ✅ Test disaster recovery procedures
