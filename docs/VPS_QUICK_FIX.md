# Quick VPS Setup Guide - Fix HTTPS Connection

## ❌ Problem
```
curl --location 'https://13.229.87.19/api/v1/auth/register'
Error: connect ECONNREFUSED 13.229.87.19:443
```

**Root Cause:**
- Port 443 (HTTPS) tidak setup/terbuka
- Nginx belum dikonfigurasi
- SSL certificate belum diinstall

---

## ✅ Solution 1: Setup HTTP (Recommended untuk testing)

### Step 1: Upload script ke VPS
Dari komputer local, jalankan:
```bash
# Pastikan di folder project
cd /Users/aleph/Sites/project/go/src/hackathon/hackathon_1

# Upload script ke VPS (ganti 'user' dengan username VPS Anda)
scp setup_nginx.sh user@13.229.87.19:~/
```

### Step 2: Login ke VPS dan jalankan script
```bash
# Login ke VPS
ssh user@13.229.87.19

# Jalankan script
chmod +x setup_nginx.sh
sudo ./setup_nginx.sh
```

Script akan otomatis:
- ✅ Install nginx
- ✅ Setup reverse proxy dari port 80 → 9001
- ✅ Konfigurasi firewall
- ✅ Start nginx service

### Step 3: Test dengan HTTP (bukan HTTPS)
```bash
# Ganti HTTPS menjadi HTTP
curl --location 'http://13.229.87.19/api/v1/auth/register' \
--header 'Content-Type: application/json' \
--data-raw '{
    "full_name": "Jaka Tarub",
    "email": "jaka.tarub@example.com",
    "phone": "",
    "password": "password123"
  }'
```

✅ **Expected Response:**
```json
{
  "status": 1,
  "status_code": 201,
  "message": "Registration success",
  "data": {
    "id": 1,
    "full_name": "Jaka Tarub",
    "email": "jaka.tarub@example.com",
    "token": {
      "access_token": "eyJ..."
    }
  }
}
```

---

## 🔒 Solution 2: Setup HTTPS (Butuh Domain)

**⚠️ PENTING:** HTTPS membutuhkan domain name, tidak bisa pakai IP address!

### Prerequisites:
1. Punya domain (contoh: `api.example.com`)
2. Point A record domain ke IP VPS: `13.229.87.19`

### Step 1: Setup domain di DNS
```
Type: A Record
Name: api (atau @)
Value: 13.229.87.19
TTL: 3600
```

### Step 2: Upload dan jalankan script dengan domain
```bash
# Upload script ke VPS
scp setup_nginx.sh user@13.229.87.19:~/

# Login ke VPS
ssh user@13.229.87.19

# Jalankan script dengan domain
chmod +x setup_nginx.sh
sudo ./setup_nginx.sh api.example.com
```

### Step 3: Install SSL certificate
```bash
# Install certbot
sudo apt update
sudo apt install -y certbot python3-certbot-nginx

# Get SSL certificate (ganti dengan domain Anda)
sudo certbot --nginx -d api.example.com

# Certbot akan otomatis:
# - Generate SSL certificate dari Let's Encrypt
# - Update nginx config untuk HTTPS
# - Setup auto-renewal
```

### Step 4: Test dengan HTTPS
```bash
curl --location 'https://api.example.com/api/v1/auth/register' \
--header 'Content-Type: application/json' \
--data-raw '{
    "full_name": "Jaka Tarub",
    "email": "jaka.tarub@example.com",
    "phone": "",
    "password": "password123"
  }'
```

---

## 🔧 Solution 3: Manual Setup (Jika script gagal)

### Install Nginx
```bash
sudo apt update
sudo apt install -y nginx
```

### Cek aplikasi berjalan
```bash
# Cek port 9001
sudo lsof -i :9001

# Jika tidak ada, start aplikasi
cd ~/hackathon_1
./bin/server &
```

### Konfigurasi Nginx
```bash
sudo tee /etc/nginx/sites-available/contact-api << 'EOF'
server {
    listen 80;
    server_name 13.229.87.19;

    # Security headers
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;

    location / {
        proxy_pass http://127.0.0.1:9001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    access_log /var/log/nginx/contact-api-access.log;
    error_log /var/log/nginx/contact-api-error.log;
}
EOF
```

### Enable site
```bash
# Enable site
sudo ln -sf /etc/nginx/sites-available/contact-api /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default

# Test config
sudo nginx -t

# Restart nginx
sudo systemctl restart nginx
```

### Konfigurasi Firewall
```bash
# Allow HTTP
sudo ufw allow 80/tcp

# Allow port 9001 (optional, untuk debugging)
sudo ufw allow 9001/tcp

# Enable firewall
sudo ufw --force enable

# Cek status
sudo ufw status
```

---

## 🔍 Troubleshooting

### Cek apakah aplikasi berjalan
```bash
ps aux | grep server
```

### Cek ports
```bash
sudo netstat -tlnp | grep -E ':(80|443|9001)'
```

### Cek nginx status
```bash
sudo systemctl status nginx
```

### Cek nginx error logs
```bash
sudo tail -f /var/log/nginx/error.log
```

### Cek application logs
```bash
tail -f ~/hackathon_1/logs/app.log
```

### Test locally di VPS
```bash
# Test ping
curl http://localhost:9001/api/v1/ping

# Test health
curl http://localhost:9001/api/v1/health
```

### Restart aplikasi
```bash
# Kill existing process
pkill -f "bin/server"

# Start new process
cd ~/hackathon_1
./bin/server &
```

### Restart nginx
```bash
sudo systemctl restart nginx
```

---

## 📝 Summary

### Untuk Testing (Cepat):
1. ✅ Gunakan HTTP: `http://13.229.87.19`
2. ✅ Setup dengan `sudo ./setup_nginx.sh`
3. ✅ Cukup untuk development/testing

### Untuk Production (Secure):
1. 🔒 Butuh domain (tidak bisa pakai IP)
2. 🔒 Setup dengan `sudo ./setup_nginx.sh yourdomain.com`
3. 🔒 Install SSL dengan certbot
4. 🔒 Gunakan HTTPS: `https://yourdomain.com`

---

## 🚀 Quick Commands Reference

```bash
# Upload script ke VPS
scp setup_nginx.sh user@13.229.87.19:~/

# Login ke VPS
ssh user@13.229.87.19

# Setup HTTP only
sudo ./setup_nginx.sh

# Setup HTTPS (dengan domain)
sudo ./setup_nginx.sh api.example.com

# Test HTTP
curl http://13.229.87.19/api/v1/ping

# Test HTTPS (dengan domain)
curl https://api.example.com/api/v1/ping

# Cek status
sudo systemctl status nginx
sudo lsof -i :80
sudo lsof -i :9001
```

---

## ⚠️ Common Issues

### Issue: Port 9001 tidak ada proses
**Fix:**
```bash
cd ~/hackathon_1
./bin/server &
```

### Issue: Nginx tidak start
**Fix:**
```bash
sudo nginx -t  # Check config
sudo systemctl restart nginx
sudo tail -f /var/log/nginx/error.log
```

### Issue: Firewall block
**Fix:**
```bash
sudo ufw allow 80/tcp
sudo ufw reload
```

### Issue: Cannot connect to database
**Fix:**
```bash
# Cek MySQL running
sudo systemctl status mysql

# Cek .env file
cat ~/hackathon_1/configs/.env

# Test connection
mysql -h localhost -u yudo -p hackathon_getcontact
```
