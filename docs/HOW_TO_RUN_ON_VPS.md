# 🚀 VPS Deployment Guide - Hackathon Contact API

Complete guide untuk deploy aplikasi Go ke VPS dengan berbagai metode.

---

## 📋 Table of Contents

1. [Quick Start](#quick-start)
2. [Prerequisites](#prerequisites)
3. [Method 1: Automated Deployment (Recommended)](#method-1-automated-deployment)
4. [Method 2: Manual Deployment](#method-2-manual-deployment)
5. [Method 3: Docker Deployment](#method-3-docker-deployment)
6. [Post-Deployment Setup](#post-deployment-setup)
7. [Troubleshooting](#troubleshooting)

---

## 🎯 Quick Start

**Fastest way to deploy (5 minutes):**

```bash
# 1. Build and deploy
./deploy_to_vps.sh ubuntu@13.229.87.19

# 2. SSH to VPS
ssh ubuntu@13.229.87.19

# 3. Configure database
nano ~/hackathon_1/configs/.env

# 4. Run application
cd ~/hackathon_1
./bin/server &

# 5. Setup nginx
sudo ./setup_nginx.sh

# 6. Test
curl http://13.229.87.19/api/v1/ping
```

✅ Done! API running on http://13.229.87.19

---

## 📦 Prerequisites

### VPS Requirements:
- OS: Ubuntu 20.04+ / Debian 11+
- RAM: 512MB minimum (1GB recommended)
- Storage: 10GB minimum
- Go: 1.23+ installed
- MySQL: 8.0+ installed
- Nginx: Latest stable

### Check VPS Setup:

```bash
# SSH to VPS
ssh user@YOUR_VPS_IP

# Check Go version
go version  # Should be 1.23+

# If not installed
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Check MySQL
sudo systemctl status mysql

# If not installed
sudo apt update
sudo apt install -y mysql-server
sudo mysql_secure_installation

# Check Nginx (will install later if needed)
nginx -v
```

---

## 🤖 Method 1: Automated Deployment (Recommended)

### Step 1: Deploy from Local Machine

```bash
# From your local project directory
cd /Users/aleph/Sites/project/go/src/hackathon/hackathon_1

# Deploy (ganti 'ubuntu' dengan username VPS Anda)
./deploy_to_vps.sh ubuntu@13.229.87.19
```

Script akan otomatis:
- ✅ Build aplikasi locally
- ✅ Create deployment package
- ✅ Upload ke VPS
- ✅ Extract dan setup permissions

### Step 2: Setup Database di VPS

```bash
# SSH to VPS
ssh ubuntu@13.229.87.19

# Create database
mysql -u root -p << EOF
CREATE DATABASE IF NOT EXISTS hackathon_getcontact;
CREATE USER IF NOT EXISTS 'yudo'@'localhost' IDENTIFIED BY 'yudo123';
GRANT ALL PRIVILEGES ON hackathon_getcontact.* TO 'yudo'@'localhost';
FLUSH PRIVILEGES;
EOF

# Update .env file
cd ~/hackathon_1
nano configs/.env
```

Edit `.env`:
```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=yudo
DB_PASSWORD=yudo123
DB_NAME=hackathon_getcontact

SERVER_PORT=9001

JWT_SECRET=HackthonII-2025
JWT_EXPIRATION=168h
```

### Step 3: Run Database Migration

```bash
cd ~/hackathon_1

# Run migration manually
go run ./cmd/migrate/main.go -command=up

# Or let the app auto-migrate on startup
```

### Step 4: Run Application

**Option A: Foreground (for testing)**
```bash
cd ~/hackathon_1
./bin/server
```

**Option B: Background with nohup**
```bash
cd ~/hackathon_1
nohup ./bin/server > logs/app.log 2>&1 &

# Check if running
ps aux | grep server

# View logs
tail -f logs/app.log
```

**Option C: Using systemd (recommended for production)**
```bash
# Create systemd service
sudo tee /etc/systemd/system/contact-api.service << EOF
[Unit]
Description=Contact API Service
After=network.target mysql.service

[Service]
Type=simple
User=$USER
WorkingDirectory=$HOME/hackathon_1
ExecStart=$HOME/hackathon_1/bin/server
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable contact-api
sudo systemctl start contact-api

# Check status
sudo systemctl status contact-api

# View logs
sudo journalctl -u contact-api -f
```

### Step 5: Setup Nginx Reverse Proxy

```bash
cd ~/hackathon_1

# For HTTP (simple, recommended for testing)
sudo ./setup_nginx.sh

# For HTTPS with self-signed cert (IP address)
sudo ./setup_https_with_ip.sh 13.229.87.19

# For HTTPS with domain (recommended for production)
# sudo ./setup_nginx.sh api.yourdomain.com
# sudo apt install -y certbot python3-certbot-nginx
# sudo certbot --nginx -d api.yourdomain.com
```

### Step 6: Test API

```bash
# Test from VPS
curl http://localhost:9001/api/v1/ping
curl http://localhost:9001/api/v1/health

# Test from external (via nginx)
curl http://13.229.87.19/api/v1/ping

# Test registration
curl --location 'http://13.229.87.19/api/v1/auth/register' \
--header 'Content-Type: application/json' \
--data-raw '{
    "full_name": "Test User",
    "email": "test@example.com",
    "password": "password123"
}'
```

---

## 🔧 Method 2: Manual Deployment

### Step 1: Build Locally

```bash
cd /Users/aleph/Sites/project/go/src/hackathon/hackathon_1
make build
```

### Step 2: Upload to VPS

```bash
# Create directory on VPS
ssh ubuntu@13.229.87.19 "mkdir -p ~/hackathon_1/{bin,configs,logs}"

# Upload binary
scp bin/server ubuntu@13.229.87.19:~/hackathon_1/bin/

# Upload configs (excluding sensitive .env)
scp -r configs/* ubuntu@13.229.87.19:~/hackathon_1/configs/

# Upload scripts
scp setup_nginx.sh ubuntu@13.229.87.19:~/hackathon_1/
scp troubleshoot_vps.sh ubuntu@13.229.87.19:~/hackathon_1/
```

### Step 3: Setup and Run

Follow steps 2-6 from Method 1 above.

---

## 🐳 Method 3: Docker Deployment

### Step 1: Build Docker Image

```bash
cd /Users/aleph/Sites/project/go/src/hackathon/hackathon_1

# Build image
docker build -t contact-api:latest .

# Save image to tar
docker save contact-api:latest | gzip > contact-api.tar.gz
```

### Step 2: Upload to VPS

```bash
# Upload image
scp contact-api.tar.gz ubuntu@13.229.87.19:~/

# Upload docker-compose
scp docker-compose.yml ubuntu@13.229.87.19:~/
```

### Step 3: Run with Docker on VPS

```bash
# SSH to VPS
ssh ubuntu@13.229.87.19

# Load image
docker load < contact-api.tar.gz

# Create .env file
cat > .env << EOF
DB_HOST=mysql
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=hackathon_getcontact
SERVER_PORT=9001
JWT_SECRET=HackthonII-2025
JWT_EXPIRATION=168h
EOF

# Run with docker-compose
docker-compose up -d

# Check logs
docker-compose logs -f
```

---

## ⚙️ Post-Deployment Setup

### 1. Configure Firewall

```bash
# Allow required ports
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS (optional)
sudo ufw allow 9001/tcp  # Go app (for debugging)

# Enable firewall
sudo ufw --force enable

# Check status
sudo ufw status
```

### 2. Setup Auto-restart on Reboot

Already configured if using systemd service (Method 1, Option C).

For nohup method, add to crontab:
```bash
crontab -e

# Add this line
@reboot cd /home/ubuntu/hackathon_1 && nohup ./bin/server > logs/app.log 2>&1 &
```

### 3. Setup Log Rotation

```bash
sudo tee /etc/logrotate.d/contact-api << EOF
/home/$USER/hackathon_1/logs/*.log {
    daily
    missingok
    rotate 14
    compress
    delaycompress
    notifempty
    create 0644 $USER $USER
}
EOF

# Test
sudo logrotate -d /etc/logrotate.d/contact-api
```

### 4. Setup Monitoring (Optional)

```bash
# Install htop for monitoring
sudo apt install -y htop

# Monitor resources
htop

# Monitor application
watch -n 2 'ps aux | grep server'

# Monitor logs
tail -f ~/hackathon_1/logs/app.log
```

---

## 🔍 Troubleshooting

### Run Automated Troubleshoot

```bash
cd ~/hackathon_1
./troubleshoot_vps.sh

# Save report
./troubleshoot_vps.sh > troubleshoot_report.txt
```

### Common Issues

#### Issue 1: Application not starting

```bash
# Check logs
tail -f ~/hackathon_1/logs/app.log

# Check if port is already in use
sudo lsof -i :9001

# Kill existing process
sudo kill -9 $(lsof -ti:9001)

# Restart application
cd ~/hackathon_1
./bin/server &
```

#### Issue 2: Database connection error

```bash
# Check MySQL running
sudo systemctl status mysql

# Test connection
mysql -u yudo -pyudo123 -e "USE hackathon_getcontact; SHOW TABLES;"

# Check .env file
cat ~/hackathon_1/configs/.env

# Reset database user
mysql -u root -p << EOF
DROP USER IF EXISTS 'yudo'@'localhost';
CREATE USER 'yudo'@'localhost' IDENTIFIED BY 'yudo123';
GRANT ALL PRIVILEGES ON hackathon_getcontact.* TO 'yudo'@'localhost';
FLUSH PRIVILEGES;
EOF
```

#### Issue 3: Nginx not working

```bash
# Check nginx status
sudo systemctl status nginx

# Check nginx config
sudo nginx -t

# Restart nginx
sudo systemctl restart nginx

# Check error logs
sudo tail -f /var/log/nginx/error.log
```

#### Issue 4: Cannot access from external

```bash
# Check if app is running
curl http://localhost:9001/api/v1/ping

# Check nginx is running
sudo lsof -i :80

# Check firewall
sudo ufw status

# Allow port 80
sudo ufw allow 80/tcp
sudo ufw reload

# Check from VPS
curl http://localhost/api/v1/ping

# If works locally but not externally, check cloud provider firewall/security groups
```

#### Issue 5: High memory usage

```bash
# Check memory
free -h

# Check processes
ps aux --sort=-%mem | head -n 10

# Restart application
sudo systemctl restart contact-api

# Or if using nohup
pkill -f "bin/server"
cd ~/hackathon_1
nohup ./bin/server > logs/app.log 2>&1 &
```

---

## 📊 Useful Commands

### Application Management

```bash
# Check if running
ps aux | grep server

# Start application
cd ~/hackathon_1 && ./bin/server &

# Stop application
pkill -f "bin/server"

# Restart application
pkill -f "bin/server" && cd ~/hackathon_1 && ./bin/server &

# View logs
tail -f ~/hackathon_1/logs/app.log

# With systemd
sudo systemctl start contact-api
sudo systemctl stop contact-api
sudo systemctl restart contact-api
sudo systemctl status contact-api
sudo journalctl -u contact-api -f
```

### Database Management

```bash
# Connect to database
mysql -u yudo -pyudo123 hackathon_getcontact

# Check tables
mysql -u yudo -pyudo123 -e "USE hackathon_getcontact; SHOW TABLES;"

# Check users
mysql -u yudo -pyudo123 -e "USE hackathon_getcontact; SELECT id, email, full_name FROM users LIMIT 10;"

# Backup database
mysqldump -u yudo -pyudo123 hackathon_getcontact > backup_$(date +%Y%m%d).sql

# Restore database
mysql -u yudo -pyudo123 hackathon_getcontact < backup_20250116.sql
```

### Nginx Management

```bash
# Check config
sudo nginx -t

# Reload config
sudo nginx -s reload

# Restart nginx
sudo systemctl restart nginx

# View access logs
sudo tail -f /var/log/nginx/contact-api-access.log

# View error logs
sudo tail -f /var/log/nginx/error.log
```

### Network Debugging

```bash
# Check open ports
sudo netstat -tlnp

# Check specific port
sudo lsof -i :9001
sudo lsof -i :80

# Test from VPS
curl http://localhost:9001/api/v1/ping
curl http://localhost/api/v1/ping

# Check firewall
sudo ufw status verbose
```

---

## 🎯 Quick Reference

### One-liner Deploy & Run

```bash
# Full deployment in one command
./deploy_to_vps.sh ubuntu@13.229.87.19 && \
ssh ubuntu@13.229.87.19 'cd ~/hackathon_1 && \
./bin/server > logs/app.log 2>&1 & \
sudo ./setup_nginx.sh'
```

### Health Check Script

Create `health_check.sh` on VPS:

```bash
#!/bin/bash

APP_URL="http://localhost:9001/api/v1/ping"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $APP_URL)

if [ "$RESPONSE" = "200" ]; then
    echo "✅ Application is healthy"
    exit 0
else
    echo "❌ Application is down (HTTP $RESPONSE)"
    echo "Restarting application..."
    pkill -f "bin/server"
    cd ~/hackathon_1
    nohup ./bin/server > logs/app.log 2>&1 &
    exit 1
fi
```

Add to crontab:
```bash
crontab -e
*/5 * * * * /home/ubuntu/hackathon_1/health_check.sh
```

---

## 📚 Additional Resources

- [VPS_DEPLOYMENT.md](./VPS_DEPLOYMENT.md) - Detailed deployment guide
- [SCRIPTS_REFERENCE.md](./SCRIPTS_REFERENCE.md) - Scripts documentation
- [VPS_QUICK_FIX.md](./VPS_QUICK_FIX.md) - Quick fixes for common issues

---

## 💡 Best Practices

1. ✅ **Always use systemd** for production (auto-restart, logging)
2. ✅ **Setup log rotation** to prevent disk space issues
3. ✅ **Use HTTPS with domain** for production (not self-signed)
4. ✅ **Setup monitoring** and health checks
5. ✅ **Regular database backups**
6. ✅ **Keep Go version updated**
7. ✅ **Use environment variables** for sensitive data
8. ✅ **Setup firewall** properly
9. ✅ **Monitor resource usage** (CPU, memory, disk)
10. ✅ **Document your configuration**

---

**Happy Deploying! 🚀**

For issues, run: `./troubleshoot_vps.sh`
