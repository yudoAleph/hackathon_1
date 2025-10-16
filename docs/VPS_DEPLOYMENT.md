# VPS Deployment Guide

Complete guide for deploying the Contact Management API on a VPS with nginx reverse proxy.

## Prerequisites

- Ubuntu 20.04+ or Debian 10+ VPS
- Root or sudo access
- Domain name (optional, for HTTPS)
- MySQL 8.0+
- Go 1.20+

## Quick Start

```bash
# 1. Clone repository
git clone <repository-url>
cd hackathon_1

# 2. Setup nginx reverse proxy
sudo ./setup_nginx.sh

# 3. Configure environment
cp configs/.env.example configs/.env
nano configs/.env

# 4. Run application
make run
```

## Detailed Setup

### 1. System Preparation

#### Update System

```bash
sudo apt-get update
sudo apt-get upgrade -y
```

#### Install Dependencies

```bash
# Install required packages
sudo apt-get install -y git curl wget build-essential nginx mysql-server

# Install Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

### 2. Database Setup

#### Secure MySQL Installation

```bash
sudo mysql_secure_installation
```

#### Create Database and User

```bash
sudo mysql -u root -p

# In MySQL console:
CREATE DATABASE hackathon_getcontact CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'apiuser'@'localhost' IDENTIFIED BY 'SecurePassword123!';
GRANT ALL PRIVILEGES ON hackathon_getcontact.* TO 'apiuser'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

#### Test Connection

```bash
mysql -u apiuser -p hackathon_getcontact
```

### 3. Application Setup

#### Clone Repository

```bash
cd /opt
sudo git clone <repository-url> contact-api
cd contact-api
sudo chown -R $USER:$USER /opt/contact-api
```

#### Configure Environment

```bash
cp configs/.env.example configs/.env
nano configs/.env
```

Edit the following variables:

```env
# Server
PORT=9001
ENVIRONMENT=production

# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=apiuser
DB_PASSWORD=SecurePassword123!
DB_NAME=hackathon_getcontact

# JWT
JWT_SECRET=your-super-secret-jwt-key-min-32-chars

# Redis (optional)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
```

#### Install Dependencies

```bash
go mod download
go mod tidy
```

#### Run Database Migrations

```bash
make migrate-up
```

#### Build Application

```bash
make build
```

### 4. Nginx Reverse Proxy Setup

#### Option A: IP-Based Access (HTTP Only)

For testing or internal use without domain:

```bash
sudo ./setup_nginx.sh
```

This will configure nginx to proxy requests from port 80 to your application on port 9001.

**Access:** http://YOUR_VPS_IP/health

#### Option B: Domain-Based Access (HTTP + HTTPS)

For production with SSL certificate:

```bash
sudo ./setup_nginx.sh yourdomain.com
```

This will:
- Configure nginx with your domain
- Setup HTTP to HTTPS redirect
- Use temporary self-signed certificate
- Prepare for Let's Encrypt SSL

**Access:** http://yourdomain.com/health

#### Setup SSL with Let's Encrypt

After DNS is properly configured:

```bash
# Install certbot
sudo apt-get install -y certbot python3-certbot-nginx

# Get SSL certificate
sudo certbot --nginx -d yourdomain.com

# Test auto-renewal
sudo certbot renew --dry-run
```

### 5. Firewall Configuration

```bash
# Enable firewall
sudo ufw enable

# Allow SSH (IMPORTANT!)
sudo ufw allow 22/tcp

# Allow HTTP and HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Check status
sudo ufw status
```

### 6. Run Application

#### Option A: Foreground (for testing)

```bash
make run
```

Press `Ctrl+C` to stop.

#### Option B: Background with nohup

```bash
nohup make run > app.log 2>&1 &
```

#### Option C: Systemd Service (recommended)

Create systemd service file:

```bash
sudo nano /etc/systemd/system/contact-api.service
```

Add the following content:

```ini
[Unit]
Description=Contact Management API
After=network.target mysql.service

[Service]
Type=simple
User=your-username
WorkingDirectory=/opt/contact-api
ExecStart=/usr/local/go/bin/go run /opt/contact-api/cmd/server/main.go
Restart=always
RestartSec=10
StandardOutput=append:/opt/contact-api/logs/app.log
StandardError=append:/opt/contact-api/logs/app.log

# Environment
Environment="PATH=/usr/local/go/bin:/usr/bin:/bin"

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service to start on boot
sudo systemctl enable contact-api

# Start service
sudo systemctl start contact-api

# Check status
sudo systemctl status contact-api

# View logs
sudo journalctl -u contact-api -f
```

### 7. Verify Deployment

#### Test Local Access

```bash
curl http://localhost:9001/health
curl http://localhost/health
```

#### Test Public Access

```bash
curl http://YOUR_VPS_IP/health
# or
curl http://yourdomain.com/health
```

#### Run Full API Tests

```bash
./test_api.sh http://YOUR_VPS_IP
# or
./test_api.sh http://yourdomain.com
```

## Troubleshooting

### Run Diagnostic Script

```bash
sudo ./troubleshoot_vps.sh
```

This will check:
- System information
- Network connectivity
- Port status
- Firewall configuration
- Application status
- Database connectivity
- Nginx configuration
- SSL certificates
- Logs and errors

### Common Issues

#### 1. Application Not Starting

```bash
# Check if port is in use
lsof -ti:9001

# Kill existing process
make kill-9001

# Check logs
tail -f logs/app.log

# Restart application
make run
```

#### 2. Database Connection Failed

```bash
# Check MySQL status
sudo systemctl status mysql

# Test connection
mysql -u apiuser -p hackathon_getcontact

# Check credentials in configs/.env
cat configs/.env | grep DB_
```

#### 3. 502 Bad Gateway (Nginx)

```bash
# Check if application is running
curl http://localhost:9001/health

# Check nginx error logs
sudo tail -f /var/log/nginx/contact-api-error.log

# Restart nginx
sudo systemctl restart nginx
```

#### 4. Cannot Access from Public IP

```bash
# Check firewall
sudo ufw status

# Allow HTTP/HTTPS if not already allowed
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Check nginx is listening on public IP
sudo netstat -tuln | grep :80
```

#### 5. SSL Certificate Issues

```bash
# Check certificate status
sudo certbot certificates

# Renew certificate manually
sudo certbot renew

# Check nginx SSL config
sudo nginx -t
```

## Monitoring & Maintenance

### View Application Logs

```bash
# Real-time application logs
tail -f logs/app.log

# With systemd
sudo journalctl -u contact-api -f

# Nginx access logs
sudo tail -f /var/log/nginx/contact-api-access.log

# Nginx error logs
sudo tail -f /var/log/nginx/contact-api-error.log
```

### Application Management

```bash
# Check status
sudo systemctl status contact-api

# Start
sudo systemctl start contact-api

# Stop
sudo systemctl stop contact-api

# Restart
sudo systemctl restart contact-api

# View logs
sudo journalctl -u contact-api -n 100
```

### Database Maintenance

```bash
# Backup database
mysqldump -u apiuser -p hackathon_getcontact > backup_$(date +%Y%m%d).sql

# Restore database
mysql -u apiuser -p hackathon_getcontact < backup_20250101.sql

# Check database size
mysql -u apiuser -p -e "
SELECT 
    table_schema AS 'Database',
    ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) AS 'Size (MB)'
FROM information_schema.tables
WHERE table_schema = 'hackathon_getcontact'
GROUP BY table_schema;
"
```

### Log Rotation

Create logrotate configuration:

```bash
sudo nano /etc/logrotate.d/contact-api
```

Add:

```
/opt/contact-api/logs/app.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 your-username your-username
    sharedscripts
    postrotate
        systemctl reload contact-api > /dev/null 2>&1 || true
    endscript
}
```

### SSL Certificate Renewal

Certbot automatically renews certificates. To test:

```bash
# Dry run
sudo certbot renew --dry-run

# Force renewal
sudo certbot renew --force-renewal
```

### Update Application

```bash
# Pull latest changes
cd /opt/contact-api
git pull

# Install dependencies
go mod tidy

# Run migrations
make migrate-up

# Restart application
sudo systemctl restart contact-api

# Check status
sudo systemctl status contact-api
```

## Performance Tuning

### Nginx Optimization

Edit `/etc/nginx/nginx.conf`:

```nginx
worker_processes auto;
worker_connections 2048;

# Enable gzip compression
gzip on;
gzip_vary on;
gzip_proxied any;
gzip_comp_level 6;
gzip_types text/plain text/css text/xml text/javascript 
           application/json application/javascript application/xml+rss;
```

### MySQL Optimization

Edit `/etc/mysql/mysql.conf.d/mysqld.cnf`:

```ini
[mysqld]
# Connection settings
max_connections = 200
connect_timeout = 10

# Buffer settings
innodb_buffer_pool_size = 1G
innodb_log_file_size = 256M

# Query cache
query_cache_type = 1
query_cache_size = 32M
```

### Application Optimization

- Enable caching (Redis)
- Use connection pooling
- Optimize database queries with indexes
- Implement rate limiting
- Use CDN for static assets

## Security Checklist

- ✅ Use strong database passwords
- ✅ Configure firewall (UFW)
- ✅ Enable HTTPS with valid SSL certificate
- ✅ Keep system and packages updated
- ✅ Use non-root user for application
- ✅ Disable root SSH login
- ✅ Use SSH keys instead of passwords
- ✅ Enable fail2ban for brute force protection
- ✅ Regular backups
- ✅ Monitor logs for suspicious activity
- ✅ Keep JWT secret secure and complex
- ✅ Implement rate limiting
- ✅ Use environment variables for secrets

## Monitoring Setup

### Install Monitoring Tools

```bash
# htop for process monitoring
sudo apt-get install htop

# netdata for real-time monitoring
bash <(curl -Ss https://my-netdata.io/kickstart.sh)
```

### Setup Health Check Monitoring

Use external monitoring services:
- UptimeRobot (https://uptimerobot.com)
- Pingdom (https://www.pingdom.com)
- StatusCake (https://www.statuscake.com)

Monitor endpoint: `http://yourdomain.com/health`

## Backup Strategy

### Automated Backup Script

Create `/opt/contact-api/backup.sh`:

```bash
#!/bin/bash
BACKUP_DIR="/opt/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# Database backup
mysqldump -u apiuser -p'password' hackathon_getcontact > "$BACKUP_DIR/db_$DATE.sql"

# Compress
gzip "$BACKUP_DIR/db_$DATE.sql"

# Keep only last 30 days
find "$BACKUP_DIR" -name "db_*.sql.gz" -mtime +30 -delete
```

Add to crontab:

```bash
crontab -e

# Daily backup at 2 AM
0 2 * * * /opt/contact-api/backup.sh
```

## Additional Resources

- [Nginx Documentation](https://nginx.org/en/docs/)
- [MySQL Documentation](https://dev.mysql.com/doc/)
- [Let's Encrypt](https://letsencrypt.org/)
- [Go Documentation](https://go.dev/doc/)
- [Systemd Service Management](https://www.freedesktop.org/software/systemd/man/systemd.service.html)

## Support

For issues and questions:
1. Check logs: `./troubleshoot_vps.sh`
2. Review this documentation
3. Check application logs: `tail -f logs/app.log`
4. Test endpoints: `./test_api.sh`
