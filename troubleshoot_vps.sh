#!/bin/bash

# ===============================
# VPS Troubleshooting Script
# ===============================
# This script helps diagnose issues on VPS deployment
# Usage: ./troubleshoot_vps.sh

set +e  # Don't exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# ===============================
# Helper Functions
# ===============================

print_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_check() {
    echo -e "${CYAN}[CHECK]${NC} $1"
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

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_fix() {
    echo -e "${YELLOW}[FIX]${NC} $1"
}

# ===============================
# System Checks
# ===============================

check_system_info() {
    print_header "1. SYSTEM INFORMATION"
    
    print_check "Operating System"
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        echo "  OS: $NAME $VERSION"
        print_success "OS detected: $NAME"
    else
        print_warning "Cannot detect OS"
    fi
    
    print_check "Kernel Version"
    uname -r
    
    print_check "System Uptime"
    uptime
    
    print_check "Memory Usage"
    free -h
    
    print_check "Disk Usage"
    df -h | grep -E '^/dev/|Filesystem'
    
    print_check "CPU Information"
    lscpu | grep -E 'Model name|CPU\(s\)|Thread\(s\)'
}

check_network() {
    print_header "2. NETWORK CONNECTIVITY"
    
    print_check "Network Interfaces"
    ip addr show | grep -E 'inet |^[0-9]:'
    
    print_check "Default Gateway"
    ip route | grep default
    
    print_check "DNS Resolution"
    if nslookup google.com > /dev/null 2>&1; then
        print_success "DNS resolution working"
    else
        print_error "DNS resolution failed"
        print_fix "Check /etc/resolv.conf"
    fi
    
    print_check "Internet Connectivity"
    if ping -c 3 8.8.8.8 > /dev/null 2>&1; then
        print_success "Internet connectivity OK"
    else
        print_error "Cannot reach internet"
        print_fix "Check firewall and network settings"
    fi
    
    print_check "Public IP Address"
    PUBLIC_IP=$(curl -s ifconfig.me 2>/dev/null || echo "Unable to detect")
    echo "  Public IP: $PUBLIC_IP"
}

check_ports() {
    print_header "3. PORT STATUS"
    
    print_check "Checking if ports are listening..."
    
    # Check port 9001 (Application)
    if netstat -tuln | grep -q ':9001 '; then
        print_success "Port 9001 (Application) is listening"
        netstat -tuln | grep ':9001 '
    else
        print_error "Port 9001 (Application) is NOT listening"
        print_fix "Start the application: make run"
    fi
    
    # Check port 3306 (MySQL)
    if netstat -tuln | grep -q ':3306 '; then
        print_success "Port 3306 (MySQL) is listening"
    else
        print_error "Port 3306 (MySQL) is NOT listening"
        print_fix "Start MySQL: sudo systemctl start mysql"
    fi
    
    # Check port 80 (HTTP)
    if netstat -tuln | grep -q ':80 '; then
        print_success "Port 80 (HTTP) is listening"
    else
        print_warning "Port 80 (HTTP) is NOT listening"
        print_fix "Setup nginx: sudo systemctl start nginx"
    fi
    
    # Check port 443 (HTTPS)
    if netstat -tuln | grep -q ':443 '; then
        print_success "Port 443 (HTTPS) is listening"
    else
        print_warning "Port 443 (HTTPS) is NOT listening"
        print_fix "Setup nginx with SSL certificate"
    fi
    
    print_check "All listening ports:"
    netstat -tuln | grep LISTEN
}

check_firewall() {
    print_header "4. FIREWALL STATUS"
    
    # Check UFW (Ubuntu/Debian)
    if command -v ufw &> /dev/null; then
        print_check "UFW Status"
        sudo ufw status verbose || true
        
        if sudo ufw status | grep -q "Status: active"; then
            print_info "UFW is active"
            print_check "Checking if required ports are allowed..."
            
            if sudo ufw status | grep -q "80"; then
                print_success "Port 80 (HTTP) is allowed"
            else
                print_warning "Port 80 (HTTP) is NOT allowed"
                print_fix "sudo ufw allow 80/tcp"
            fi
            
            if sudo ufw status | grep -q "443"; then
                print_success "Port 443 (HTTPS) is allowed"
            else
                print_warning "Port 443 (HTTPS) is NOT allowed"
                print_fix "sudo ufw allow 443/tcp"
            fi
            
            if sudo ufw status | grep -q "22"; then
                print_success "Port 22 (SSH) is allowed"
            else
                print_error "Port 22 (SSH) is NOT allowed - DANGEROUS!"
                print_fix "sudo ufw allow 22/tcp"
            fi
        else
            print_info "UFW is not active"
        fi
    fi
    
    # Check iptables
    if command -v iptables &> /dev/null; then
        print_check "iptables Rules"
        sudo iptables -L -n -v | head -20
    fi
    
    # Check firewalld (CentOS/RHEL)
    if command -v firewall-cmd &> /dev/null; then
        print_check "firewalld Status"
        sudo firewall-cmd --state 2>/dev/null || print_info "firewalld not running"
        sudo firewall-cmd --list-all 2>/dev/null || true
    fi
}

check_application() {
    print_header "5. APPLICATION STATUS"
    
    print_check "Checking if application is running..."
    if pgrep -f "cmd/server/main.go" > /dev/null; then
        print_success "Application process is running"
        ps aux | grep -E "cmd/server/main.go" | grep -v grep
    else
        print_error "Application process is NOT running"
        print_fix "Start application: make run"
    fi
    
    print_check "Checking application logs..."
    if [ -f "logs/app.log" ]; then
        print_success "Log file exists"
        print_info "Last 10 log entries:"
        tail -10 logs/app.log
    else
        print_warning "Log file not found at logs/app.log"
    fi
    
    print_check "Testing local application endpoint..."
    if curl -s http://127.0.0.1:9001/health > /dev/null 2>&1; then
        print_success "Application responds on localhost:9001"
        curl -s http://127.0.0.1:9001/health | jq '.' 2>/dev/null || curl -s http://127.0.0.1:9001/health
    else
        print_error "Application does NOT respond on localhost:9001"
        print_fix "Check application logs and restart: make run"
    fi
}

check_database() {
    print_header "6. DATABASE STATUS"
    
    print_check "MySQL Service Status"
    if systemctl is-active --quiet mysql 2>/dev/null || systemctl is-active --quiet mysqld 2>/dev/null; then
        print_success "MySQL service is running"
        systemctl status mysql 2>/dev/null || systemctl status mysqld 2>/dev/null | head -5
    else
        print_error "MySQL service is NOT running"
        print_fix "Start MySQL: sudo systemctl start mysql"
    fi
    
    print_check "MySQL Connection"
    if [ -f "configs/.env" ]; then
        source configs/.env
        if mysql -h"${DB_HOST:-localhost}" -u"${DB_USER}" -p"${DB_PASSWORD}" -e "SELECT 1;" > /dev/null 2>&1; then
            print_success "MySQL connection successful"
            
            print_check "Database exists"
            if mysql -h"${DB_HOST:-localhost}" -u"${DB_USER}" -p"${DB_PASSWORD}" -e "USE ${DB_NAME};" > /dev/null 2>&1; then
                print_success "Database '${DB_NAME}' exists"
                
                print_check "Tables in database"
                mysql -h"${DB_HOST:-localhost}" -u"${DB_USER}" -p"${DB_PASSWORD}" -e "USE ${DB_NAME}; SHOW TABLES;"
            else
                print_error "Database '${DB_NAME}' does NOT exist"
                print_fix "Run migrations: make migrate-up"
            fi
        else
            print_error "Cannot connect to MySQL"
            print_fix "Check credentials in configs/.env"
        fi
    else
        print_warning "configs/.env file not found"
        print_fix "Create configs/.env with database credentials"
    fi
}

check_nginx() {
    print_header "7. NGINX STATUS"
    
    if command -v nginx &> /dev/null; then
        print_check "nginx Installation"
        nginx -v 2>&1
        
        print_check "nginx Service Status"
        if systemctl is-active --quiet nginx; then
            print_success "nginx service is running"
            systemctl status nginx | head -5
        else
            print_warning "nginx service is NOT running"
            print_fix "Start nginx: sudo systemctl start nginx"
        fi
        
        print_check "nginx Configuration Test"
        if sudo nginx -t 2>&1; then
            print_success "nginx configuration is valid"
        else
            print_error "nginx configuration has errors"
            print_fix "Check nginx config: sudo nginx -t"
        fi
        
        print_check "nginx Configuration Files"
        if [ -f "/etc/nginx/sites-available/contact-api" ]; then
            print_success "API config exists: /etc/nginx/sites-available/contact-api"
        else
            print_warning "API config not found"
            print_fix "Create nginx config for the API"
        fi
        
        if [ -L "/etc/nginx/sites-enabled/contact-api" ]; then
            print_success "API config is enabled"
        else
            print_warning "API config is NOT enabled"
            print_fix "sudo ln -s /etc/nginx/sites-available/contact-api /etc/nginx/sites-enabled/"
        fi
        
        print_check "Testing HTTP access"
        if curl -s http://localhost/health > /dev/null 2>&1; then
            print_success "nginx proxy is working"
            curl -s http://localhost/health | jq '.' 2>/dev/null || curl -s http://localhost/health
        else
            print_warning "Cannot access application through nginx"
            print_fix "Check nginx proxy configuration"
        fi
    else
        print_warning "nginx is NOT installed"
        print_fix "Install nginx: sudo apt-get install nginx"
    fi
}

check_ssl() {
    print_header "8. SSL/TLS CERTIFICATE STATUS"
    
    print_check "Checking for SSL certificates..."
    
    if [ -d "/etc/letsencrypt/live" ]; then
        print_info "Let's Encrypt certificates found:"
        sudo ls -la /etc/letsencrypt/live/ 2>/dev/null || print_warning "Cannot access certificate directory"
    else
        print_warning "No Let's Encrypt certificates found"
        print_fix "Setup SSL: sudo certbot --nginx -d yourdomain.com"
    fi
    
    if command -v certbot &> /dev/null; then
        print_check "certbot is installed"
        certbot --version
    else
        print_warning "certbot is NOT installed"
        print_fix "Install certbot: sudo apt-get install certbot python3-certbot-nginx"
    fi
}

check_environment() {
    print_header "9. ENVIRONMENT CONFIGURATION"
    
    print_check "Checking .env file..."
    if [ -f "configs/.env" ]; then
        print_success "configs/.env exists"
        print_info "Environment variables (sensitive values hidden):"
        grep -v '^#' configs/.env | grep -v '^$' | sed 's/=.*/=***/' || true
    else
        print_error "configs/.env NOT found"
        print_fix "Create configs/.env from configs/.env.example"
    fi
    
    print_check "Checking Go installation..."
    if command -v go &> /dev/null; then
        print_success "Go is installed"
        go version
    else
        print_error "Go is NOT installed"
        print_fix "Install Go: https://golang.org/doc/install"
    fi
    
    print_check "Checking Go modules..."
    if [ -f "go.mod" ]; then
        print_success "go.mod exists"
    else
        print_error "go.mod NOT found"
        print_fix "Run: go mod init user-service"
    fi
}

check_logs() {
    print_header "10. RECENT ERRORS IN LOGS"
    
    print_check "Checking application logs for errors..."
    if [ -f "logs/app.log" ]; then
        echo "Last 5 error entries:"
        grep -i "error\|fail\|fatal" logs/app.log | tail -5 || print_info "No errors found in logs"
    else
        print_warning "Application log file not found"
    fi
    
    print_check "Checking nginx error logs..."
    if [ -f "/var/log/nginx/error.log" ]; then
        echo "Last 5 nginx errors:"
        sudo tail -5 /var/log/nginx/error.log || print_info "No recent nginx errors"
    fi
    
    print_check "Checking system logs..."
    if command -v journalctl &> /dev/null; then
        echo "Last 5 system errors:"
        sudo journalctl -p err -n 5 --no-pager || true
    fi
}

generate_recommendations() {
    print_header "11. RECOMMENDATIONS & QUICK FIXES"
    
    echo ""
    print_info "Common Issues and Solutions:"
    echo ""
    
    echo "1. Application won't start:"
    echo "   - Check if port 9001 is already in use: lsof -ti:9001"
    echo "   - Kill existing process: make kill-9001"
    echo "   - Check logs: tail -f logs/app.log"
    echo ""
    
    echo "2. Database connection fails:"
    echo "   - Verify MySQL is running: sudo systemctl status mysql"
    echo "   - Check credentials in configs/.env"
    echo "   - Test connection: mysql -u user -p"
    echo ""
    
    echo "3. Cannot access via public IP:"
    echo "   - Check if nginx is running: sudo systemctl status nginx"
    echo "   - Check firewall: sudo ufw status"
    echo "   - Test locally first: curl http://localhost/health"
    echo ""
    
    echo "4. 502 Bad Gateway error:"
    echo "   - Application is not running on port 9001"
    echo "   - Check nginx upstream config"
    echo "   - Restart application: make run"
    echo ""
    
    echo "5. SSL/HTTPS not working:"
    echo "   - Install certbot: sudo apt-get install certbot python3-certbot-nginx"
    echo "   - Get certificate: sudo certbot --nginx -d yourdomain.com"
    echo "   - Check certificate: sudo certbot certificates"
    echo ""
}

# ===============================
# Main Execution
# ===============================

main() {
    clear
    echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║   VPS Troubleshooting Script           ║${NC}"
    echo -e "${GREEN}║   Contact Management API               ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
    echo ""
    print_info "This script will check common issues on your VPS"
    print_info "Run with sudo for full diagnostics"
    echo ""
    
    check_system_info
    check_network
    check_ports
    check_firewall
    check_application
    check_database
    check_nginx
    check_ssl
    check_environment
    check_logs
    generate_recommendations
    
    print_header "TROUBLESHOOTING COMPLETE"
    echo ""
    print_info "Review the checks above to identify issues"
    print_info "Look for [✗] markers indicating problems"
    print_info "Follow the [FIX] suggestions to resolve issues"
    echo ""
    
    # Save report
    REPORT_FILE="troubleshoot_report_$(date +%Y%m%d_%H%M%S).txt"
    echo "Saving report to: $REPORT_FILE"
    echo "To save a complete report, run:"
    echo "  ./troubleshoot_vps.sh | tee $REPORT_FILE"
    echo ""
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    print_warning "Not running as root. Some checks may be incomplete."
    print_info "For full diagnostics, run: sudo ./troubleshoot_vps.sh"
    echo ""
fi

# Run main function
main "$@"
