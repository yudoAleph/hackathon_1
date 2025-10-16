#!/bin/bash

# ===============================
# Deploy Application to VPS
# ===============================
# This script helps you deploy the Go application to VPS
# Usage: ./deploy_to_vps.sh [user@vps_ip]
# Example: ./deploy_to_vps.sh ubuntu@13.229.87.19

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
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

print_info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

# Check if VPS address provided
if [ -z "$1" ]; then
    print_error "Please provide VPS address"
    echo "Usage: $0 [user@vps_ip]"
    echo "Example: $0 ubuntu@13.229.87.19"
    exit 1
fi

VPS_ADDRESS=$1
APP_NAME="hackathon_1"
REMOTE_DIR="/home/${VPS_ADDRESS%%@*}/$APP_NAME"

print_header "Deploying to VPS: $VPS_ADDRESS"

# Build application locally
print_header "Building Application"
print_info "Building Go binary..."
make build

if [ ! -f "bin/server" ]; then
    print_error "Build failed! bin/server not found"
    exit 1
fi

print_success "Build successful"

# Create deployment archive
print_header "Creating Deployment Package"

# Create temp directory for deployment
TEMP_DIR=$(mktemp -d)
DEPLOY_DIR="$TEMP_DIR/$APP_NAME"
mkdir -p "$DEPLOY_DIR"

# Copy necessary files
print_info "Copying files..."
cp -r bin "$DEPLOY_DIR/"
cp -r configs "$DEPLOY_DIR/"
cp -r cmd "$DEPLOY_DIR/"
cp -r internal "$DEPLOY_DIR/"
cp -r pkg "$DEPLOY_DIR/"
cp go.mod "$DEPLOY_DIR/"
cp go.sum "$DEPLOY_DIR/"
cp Makefile "$DEPLOY_DIR/"
cp setup_nginx.sh "$DEPLOY_DIR/"
cp setup_https_with_ip.sh "$DEPLOY_DIR/"
cp troubleshoot_vps.sh "$DEPLOY_DIR/"

# Create archive
cd "$TEMP_DIR"
tar -czf "$APP_NAME.tar.gz" "$APP_NAME"
ARCHIVE_PATH="$TEMP_DIR/$APP_NAME.tar.gz"

print_success "Deployment package created"

# Upload to VPS
print_header "Uploading to VPS"
print_info "Uploading archive to VPS..."

scp "$ARCHIVE_PATH" "$VPS_ADDRESS:/tmp/$APP_NAME.tar.gz"

print_success "Upload complete"

# Execute deployment on VPS
print_header "Setting up Application on VPS"

ssh "$VPS_ADDRESS" bash << 'ENDSSH'
set -e

# Colors for remote execution
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

APP_NAME="hackathon_1"
REMOTE_DIR="$HOME/$APP_NAME"

# Extract archive
print_info "Extracting application..."
cd ~
tar -xzf "/tmp/$APP_NAME.tar.gz"
rm "/tmp/$APP_NAME.tar.gz"

cd "$REMOTE_DIR"

# Make scripts executable
chmod +x bin/server
chmod +x setup_nginx.sh
chmod +x setup_https_with_ip.sh
chmod +x troubleshoot_vps.sh

print_success "Application extracted to $REMOTE_DIR"

# Check if .env exists
if [ ! -f "configs/.env" ]; then
    echo ""
    echo -e "${YELLOW}[⚠]${NC} configs/.env not found!"
    echo "Creating sample .env file. Please update with your actual values:"
    
    cat > configs/.env << 'EOF'
# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=yudo
DB_PASSWORD=yudo123
DB_NAME=hackathon_getcontact

# Server Configuration
SERVER_PORT=9001

# JWT Configuration
JWT_SECRET=HackthonII-2025
JWT_EXPIRATION=168h

# Redis Configuration (optional)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
EOF
    
    echo ""
    echo "Sample .env created. Please edit: $REMOTE_DIR/configs/.env"
fi

print_success "Setup complete on VPS"

ENDSSH

# Cleanup temp directory
rm -rf "$TEMP_DIR"

# Print next steps
print_header "Deployment Complete!"

echo ""
print_success "Application deployed to: $REMOTE_DIR"
echo ""
echo -e "${CYAN}Next Steps:${NC}"
echo ""
echo "1. SSH to VPS:"
echo -e "   ${GREEN}ssh $VPS_ADDRESS${NC}"
echo ""
echo "2. Update .env file with correct database credentials:"
echo -e "   ${GREEN}cd $REMOTE_DIR${NC}"
echo -e "   ${GREEN}nano configs/.env${NC}"
echo ""
echo "3. Ensure MySQL is running and database exists:"
echo -e "   ${GREEN}sudo systemctl status mysql${NC}"
echo -e "   ${GREEN}mysql -u root -p -e \"CREATE DATABASE IF NOT EXISTS hackathon_getcontact;\"${NC}"
echo ""
echo "4. Run the application:"
echo -e "   ${GREEN}cd $REMOTE_DIR${NC}"
echo -e "   ${GREEN}./bin/server${NC}"
echo ""
echo "   Or run in background:"
echo -e "   ${GREEN}nohup ./bin/server > logs/app.log 2>&1 &${NC}"
echo ""
echo "5. Setup Nginx reverse proxy (HTTP):"
echo -e "   ${GREEN}sudo ./setup_nginx.sh${NC}"
echo ""
echo "   Or with HTTPS (self-signed):"
echo -e "   ${GREEN}sudo ./setup_https_with_ip.sh 13.229.87.19${NC}"
echo ""
echo "6. Test the application:"
echo -e "   ${GREEN}curl http://YOUR_VPS_IP/api/v1/ping${NC}"
echo ""
echo "7. If issues occur, run troubleshoot:"
echo -e "   ${GREEN}./troubleshoot_vps.sh${NC}"
echo ""
print_success "Happy deploying! 🚀"
echo ""
