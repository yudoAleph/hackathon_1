#!/bin/bash

# ===============================
# Setup HTTPS with IP Address using Self-Signed Certificate
# ===============================
# ⚠️ WARNING: Self-signed certificates will show browser warnings!
# Only use this for development/testing, NOT for production!
#
# Usage: sudo ./setup_https_with_ip.sh 13.229.87.19

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[⚠]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    print_error "Please run as root: sudo $0"
    exit 1
fi

# Get IP address
IP_ADDRESS=${1:-$(curl -s ifconfig.me)}

print_header "Setup HTTPS with Self-Signed Certificate for IP: $IP_ADDRESS"

print_warning "⚠️  Self-signed certificates will show security warnings in browsers!"
print_warning "⚠️  Clients will need to trust the certificate or disable SSL verification"
echo ""

# Install nginx if not installed
if ! command -v nginx &> /dev/null; then
    echo "Installing nginx..."
    apt update
    apt install -y nginx
    print_success "Nginx installed"
fi

# Create SSL directory
SSL_DIR="/etc/nginx/ssl"
mkdir -p $SSL_DIR

print_header "Generating Self-Signed Certificate"

# Generate self-signed certificate valid for 365 days
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout $SSL_DIR/selfsigned.key \
    -out $SSL_DIR/selfsigned.crt \
    -subj "/C=ID/ST=Jakarta/L=Jakarta/O=Development/CN=$IP_ADDRESS" \
    -addext "subjectAltName=IP:$IP_ADDRESS"

chmod 600 $SSL_DIR/selfsigned.key
chmod 644 $SSL_DIR/selfsigned.crt

print_success "Self-signed certificate generated"
echo "  Certificate: $SSL_DIR/selfsigned.crt"
echo "  Private Key: $SSL_DIR/selfsigned.key"

# Generate strong DH parameters (this may take a while)
print_header "Generating DH Parameters (this may take a few minutes...)"
if [ ! -f $SSL_DIR/dhparam.pem ]; then
    openssl dhparam -out $SSL_DIR/dhparam.pem 2048
    print_success "DH parameters generated"
else
    print_success "DH parameters already exist"
fi

# Create nginx configuration
print_header "Configuring Nginx"

NGINX_CONF="/etc/nginx/sites-available/contact-api-https"

cat > $NGINX_CONF << EOF
# HTTP - Redirect to HTTPS
server {
    listen 80;
    server_name $IP_ADDRESS;
    
    # Redirect all HTTP traffic to HTTPS
    return 301 https://\$server_name\$request_uri;
}

# HTTPS
server {
    listen 443 ssl http2;
    server_name $IP_ADDRESS;

    # SSL Certificate (Self-Signed)
    ssl_certificate $SSL_DIR/selfsigned.crt;
    ssl_certificate_key $SSL_DIR/selfsigned.key;
    ssl_dhparam $SSL_DIR/dhparam.pem;

    # SSL Configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384';
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Proxy to application
    location / {
        proxy_pass http://127.0.0.1:9001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Logging
    access_log /var/log/nginx/contact-api-https-access.log;
    error_log /var/log/nginx/contact-api-https-error.log;
}
EOF

# Enable site
ln -sf $NGINX_CONF /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default

# Test nginx configuration
print_header "Testing Nginx Configuration"
if nginx -t; then
    print_success "Nginx configuration is valid"
else
    print_error "Nginx configuration is invalid!"
    exit 1
fi

# Configure firewall
print_header "Configuring Firewall"

if command -v ufw &> /dev/null; then
    ufw allow 80/tcp comment 'HTTP'
    ufw allow 443/tcp comment 'HTTPS'
    ufw allow 9001/tcp comment 'Go App'
    ufw --force enable
    print_success "Firewall configured"
    ufw status
fi

# Restart nginx
print_header "Starting Nginx"
systemctl restart nginx
systemctl enable nginx

if systemctl is-active --quiet nginx; then
    print_success "Nginx is running"
else
    print_error "Failed to start nginx"
    systemctl status nginx
    exit 1
fi

# Print completion message
print_header "Setup Complete!"

echo ""
print_success "HTTPS is now enabled with self-signed certificate"
echo ""
print_warning "⚠️  IMPORTANT: Browser/Client Configuration Required!"
echo ""
echo "Because this uses a self-signed certificate, clients will show warnings."
echo ""
echo "To test with curl (disable SSL verification):"
echo ""
echo -e "${GREEN}curl -k --location 'https://$IP_ADDRESS/api/v1/auth/register' \\${NC}"
echo -e "${GREEN}--header 'Content-Type: application/json' \\${NC}"
echo -e "${GREEN}--data-raw '{${NC}"
echo -e "${GREEN}    \"full_name\": \"Jaka Tarub\",${NC}"
echo -e "${GREEN}    \"email\": \"jaka.tarub@example.com\",${NC}"
echo -e "${GREEN}    \"phone\": \"\",${NC}"
echo -e "${GREEN}    \"password\": \"password123\"${NC}"
echo -e "${GREEN}}'${NC}"
echo ""
echo "Note: -k flag disables SSL certificate verification"
echo ""
echo "For Postman:"
echo "  Settings → SSL Certificate Verification → Turn OFF"
echo ""
echo "For mobile apps:"
echo "  - iOS: Add certificate to trust store"
echo "  - Android: Add certificate as trusted CA"
echo "  - Or: Disable SSL pinning in dev build (NOT recommended for production)"
echo ""
echo "Certificate location: $SSL_DIR/selfsigned.crt"
echo ""
print_warning "⚠️  For production, use a domain name with Let's Encrypt certificate!"
echo ""
