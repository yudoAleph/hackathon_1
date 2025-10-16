#!/bin/bash

# ===============================
# Nginx Setup Script
# ===============================
# This script sets up nginx as a reverse proxy for the Contact Management API
# Usage: sudo ./setup_nginx.sh [domain]
# Example: sudo ./setup_nginx.sh api.example.com

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

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[⚠]${NC} $1"
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    print_error "This script must be run as root"
    echo "Usage: sudo ./setup_nginx.sh [domain]"
    exit 1
fi

DOMAIN=${1:-""}
APP_PORT="9001"
PUBLIC_IP=$(curl -s ifconfig.me 2>/dev/null || echo "")

print_header "NGINX REVERSE PROXY SETUP"
echo ""
print_info "Application Port: $APP_PORT"
print_info "Public IP: $PUBLIC_IP"
if [ -n "$DOMAIN" ]; then
    print_info "Domain: $DOMAIN"
else
    print_warning "No domain specified, will setup for IP-based access"
fi
echo ""

# Install nginx if not already installed
print_header "1. CHECKING NGINX INSTALLATION"
if ! command -v nginx &> /dev/null; then
    print_warning "nginx is not installed"
    print_info "Installing nginx..."
    apt-get update
    apt-get install -y nginx
    print_success "nginx installed"
else
    print_success "nginx is already installed"
    nginx -v
fi

# Stop nginx to make changes
print_info "Stopping nginx..."
systemctl stop nginx || true

# Backup existing default config
print_header "2. BACKING UP EXISTING CONFIGURATION"
if [ -f "/etc/nginx/sites-available/default" ]; then
    BACKUP_FILE="/etc/nginx/sites-available/default.backup.$(date +%Y%m%d_%H%M%S)"
    cp /etc/nginx/sites-available/default "$BACKUP_FILE"
    print_success "Backed up default config to $BACKUP_FILE"
fi

# Create nginx configuration
print_header "3. CREATING NGINX CONFIGURATION"

NGINX_CONFIG="/etc/nginx/sites-available/contact-api"

if [ -n "$DOMAIN" ]; then
    # Configuration with domain
    cat > "$NGINX_CONFIG" <<EOF
# Contact Management API - Reverse Proxy Configuration
# Domain: $DOMAIN

upstream contact_api {
    server 127.0.0.1:$APP_PORT;
    keepalive 32;
}

# HTTP Server - Redirect to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name $DOMAIN;

    # Redirect all HTTP to HTTPS
    return 301 https://\$host\$request_uri;
}

# HTTPS Server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name $DOMAIN;

    # SSL Configuration (will be configured by certbot)
    ssl_certificate /etc/ssl/certs/ssl-cert-snakeoil.pem;
    ssl_certificate_key /etc/ssl/private/ssl-cert-snakeoil.key;
    
    # SSL Security Settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security Headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;

    # Client body size limit (for file uploads)
    client_max_body_size 10M;

    # Logging
    access_log /var/log/nginx/contact-api-access.log;
    error_log /var/log/nginx/contact-api-error.log;

    # Root location - Health check
    location /health {
        proxy_pass http://contact_api/health;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        
        # Timeouts
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    # API endpoints
    location / {
        proxy_pass http://contact_api;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        
        # Timeouts
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
        
        # CORS headers (if needed)
        add_header Access-Control-Allow-Origin * always;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
        add_header Access-Control-Allow-Headers "Authorization, Content-Type" always;
        
        # Handle preflight requests
        if (\$request_method = 'OPTIONS') {
            return 204;
        }
    }
}
EOF
else
    # Configuration for IP-based access (HTTP only)
    cat > "$NGINX_CONFIG" <<EOF
# Contact Management API - Reverse Proxy Configuration
# IP-based access

upstream contact_api {
    server 127.0.0.1:$APP_PORT;
    keepalive 32;
}

# HTTP Server
server {
    listen 80 default_server;
    listen [::]:80 default_server;
    server_name _;

    # Security Headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Client body size limit
    client_max_body_size 10M;

    # Logging
    access_log /var/log/nginx/contact-api-access.log;
    error_log /var/log/nginx/contact-api-error.log;

    # Root location - Health check
    location /health {
        proxy_pass http://contact_api/health;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        
        # Timeouts
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    # API endpoints
    location / {
        proxy_pass http://contact_api;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        
        # Timeouts
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
        
        # CORS headers (if needed)
        add_header Access-Control-Allow-Origin * always;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
        add_header Access-Control-Allow-Headers "Authorization, Content-Type" always;
        
        # Handle preflight requests
        if (\$request_method = 'OPTIONS') {
            return 204;
        }
    }
}
EOF
fi

print_success "Created nginx configuration: $NGINX_CONFIG"

# Enable the site
print_header "4. ENABLING SITE"
rm -f /etc/nginx/sites-enabled/default
rm -f /etc/nginx/sites-enabled/contact-api
ln -s "$NGINX_CONFIG" /etc/nginx/sites-enabled/contact-api
print_success "Site enabled"

# Test nginx configuration
print_header "5. TESTING NGINX CONFIGURATION"
if nginx -t; then
    print_success "Nginx configuration is valid"
else
    print_error "Nginx configuration test failed"
    exit 1
fi

# Start nginx
print_header "6. STARTING NGINX"
systemctl start nginx
systemctl enable nginx
print_success "Nginx started and enabled"

# Configure firewall
print_header "7. CONFIGURING FIREWALL"
if command -v ufw &> /dev/null; then
    print_info "Configuring UFW firewall..."
    ufw allow 'Nginx Full' || ufw allow 80/tcp && ufw allow 443/tcp
    ufw allow 22/tcp  # Ensure SSH is allowed
    print_success "Firewall configured"
else
    print_warning "UFW not found, skipping firewall configuration"
fi

# Print status
print_header "8. SETUP COMPLETE"
echo ""
print_success "Nginx reverse proxy has been configured successfully!"
echo ""
print_info "Configuration details:"
echo "  - Upstream: 127.0.0.1:$APP_PORT"
echo "  - Config file: $NGINX_CONFIG"
echo "  - Logs: /var/log/nginx/contact-api-*.log"
echo ""

if [ -n "$DOMAIN" ]; then
    print_info "Access your API at:"
    echo "  - HTTP:  http://$DOMAIN"
    echo "  - HTTPS: https://$DOMAIN (after SSL setup)"
    echo ""
    print_warning "HTTPS is using a self-signed certificate"
    print_info "To setup proper SSL with Let's Encrypt:"
    echo "  1. Ensure your domain points to this server: $PUBLIC_IP"
    echo "  2. Install certbot: apt-get install certbot python3-certbot-nginx"
    echo "  3. Run: certbot --nginx -d $DOMAIN"
else
    print_info "Access your API at:"
    echo "  - Local:  http://localhost"
    echo "  - Public: http://$PUBLIC_IP"
    echo ""
    print_warning "HTTPS is not configured for IP-based access"
    print_info "To enable HTTPS, rerun this script with a domain name:"
    echo "  sudo ./setup_nginx.sh yourdomain.com"
fi

echo ""
print_info "Test the setup:"
echo "  curl http://localhost/health"
if [ -n "$PUBLIC_IP" ]; then
    echo "  curl http://$PUBLIC_IP/health"
fi
echo ""

print_info "Nginx status:"
systemctl status nginx --no-pager -l

echo ""
print_success "Setup completed successfully!"
