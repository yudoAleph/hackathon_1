# Scripts Quick Reference

This document provides a quick reference for all utility scripts in the project.

## Available Scripts

### 1. test_api.sh - API Testing Script

**Purpose:** Comprehensive testing of all API endpoints

**Usage:**
```bash
# Test local server
./test_api.sh

# Test remote server
./test_api.sh http://your-server-ip
./test_api.sh https://yourdomain.com
```

**What it tests:**
- ✅ Health check endpoint
- ✅ Ping endpoint
- ✅ User registration (success & duplicate)
- ✅ User login (success & invalid credentials)
- ✅ Get user profile
- ✅ Update user profile
- ✅ Create contact
- ✅ List contacts with pagination
- ✅ Search contacts
- ✅ Get contact details
- ✅ Update contact
- ✅ Delete contact
- ✅ Unauthorized access handling

**Output:**
- Color-coded test results (green = pass, red = fail)
- JSON responses for each test
- Test summary with pass/fail counts

**Requirements:**
- `jq` command-line JSON processor
- Server must be running and accessible

**Exit codes:**
- `0` - All tests passed
- `1` - Some tests failed

---

### 2. troubleshoot_vps.sh - VPS Diagnostic Script

**Purpose:** Diagnose common issues on VPS deployment

**Usage:**
```bash
# Basic check
./troubleshoot_vps.sh

# Full diagnostics (recommended)
sudo ./troubleshoot_vps.sh
```

**What it checks:**

1. **System Information**
   - OS version and kernel
   - System uptime
   - Memory and disk usage
   - CPU information

2. **Network Connectivity**
   - Network interfaces and IP addresses
   - DNS resolution
   - Internet connectivity
   - Public IP detection

3. **Port Status**
   - Port 9001 (Application)
   - Port 3306 (MySQL)
   - Port 80 (HTTP)
   - Port 443 (HTTPS)

4. **Firewall Configuration**
   - UFW status and rules
   - iptables rules
   - firewalld status (CentOS)

5. **Application Status**
   - Process running check
   - Application logs review
   - Local endpoint test

6. **Database Status**
   - MySQL service status
   - Database connection test
   - Tables verification

7. **Nginx Status**
   - Installation check
   - Service status
   - Configuration validation
   - Proxy functionality test

8. **SSL/TLS Certificates**
   - Let's Encrypt certificates
   - Certbot installation

9. **Environment Configuration**
   - .env file validation
   - Go installation
   - Dependencies check

10. **Recent Errors**
    - Application error logs
    - Nginx error logs
    - System error logs

**Output:**
- Color-coded status indicators
- Detailed check results
- Fix suggestions for problems
- Recommendations section

**Save report:**
```bash
./troubleshoot_vps.sh | tee troubleshoot_report.txt
```

---

### 3. setup_nginx.sh - Nginx Reverse Proxy Setup

**Purpose:** Automated nginx configuration for reverse proxy

**Usage:**
```bash
# IP-based access (HTTP only)
sudo ./setup_nginx.sh

# Domain-based access (HTTP + HTTPS)
sudo ./setup_nginx.sh yourdomain.com
```

**What it does:**

1. **Installation**
   - Checks and installs nginx if needed
   - Backs up existing configuration

2. **Configuration**
   - Creates reverse proxy config
   - Sets up upstream to localhost:9001
   - Configures security headers
   - Sets up CORS headers
   - Configures timeouts and limits

3. **Domain Setup (if provided)**
   - HTTP to HTTPS redirect
   - SSL configuration placeholder
   - Prepares for Let's Encrypt

4. **IP-Based Setup (no domain)**
   - HTTP-only configuration
   - Accepts requests on any IP

5. **Finalization**
   - Validates nginx configuration
   - Enables and starts nginx
   - Configures firewall (UFW)

**Post-Setup Steps:**

For SSL certificate (domain-based):
```bash
sudo apt-get install certbot python3-certbot-nginx
sudo certbot --nginx -d yourdomain.com
```

**Configuration files created:**
- `/etc/nginx/sites-available/contact-api`
- `/etc/nginx/sites-enabled/contact-api`

**Logs:**
- `/var/log/nginx/contact-api-access.log`
- `/var/log/nginx/contact-api-error.log`

**Access URLs:**

*IP-based:*
- Local: `http://localhost/health`
- Public: `http://YOUR_IP/health`

*Domain-based:*
- HTTP: `http://yourdomain.com/health`
- HTTPS: `https://yourdomain.com/health` (after SSL)

---

## Common Workflows

### Initial Deployment

```bash
# 1. Setup environment
cp configs/.env.example configs/.env
nano configs/.env

# 2. Setup nginx
sudo ./setup_nginx.sh yourdomain.com

# 3. Start application
make run

# 4. Test deployment
./test_api.sh http://yourdomain.com
```

### Troubleshooting Issues

```bash
# 1. Run diagnostic
sudo ./troubleshoot_vps.sh

# 2. Check specific issues based on output
# Example: If application not running
make kill-9001
make run

# 3. Verify fix
curl http://localhost:9001/health
./test_api.sh
```

### Regular Testing

```bash
# Test local development
make run  # Terminal 1
./test_api.sh  # Terminal 2

# Test production
./test_api.sh https://yourdomain.com
```

### SSL Certificate Setup

```bash
# 1. Ensure nginx is configured with domain
sudo ./setup_nginx.sh yourdomain.com

# 2. Install certbot
sudo apt-get install certbot python3-certbot-nginx

# 3. Get certificate
sudo certbot --nginx -d yourdomain.com

# 4. Test auto-renewal
sudo certbot renew --dry-run

# 5. Verify HTTPS
curl https://yourdomain.com/health
```

## Script Dependencies

### test_api.sh
- **Required:** `jq`, `curl`
- **Install:** `brew install jq` (macOS) or `apt-get install jq` (Ubuntu)

### troubleshoot_vps.sh
- **Required:** `curl`, `netstat`, `systemctl`, `mysql`
- **Optional:** `jq` (for formatted output)
- **Best with:** `sudo` access

### setup_nginx.sh
- **Required:** `sudo` access
- **Installs:** `nginx` (if not present)
- **Optional:** `certbot` (for SSL)

## Environment Variables

All scripts respect these environment variables:

```bash
# Base URL for API testing
export BASE_URL="http://localhost:9001"
./test_api.sh

# Custom application port
export APP_PORT="9001"
sudo ./setup_nginx.sh
```

## Troubleshooting Scripts

### test_api.sh fails with "jq not found"

```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# CentOS/RHEL
sudo yum install jq
```

### troubleshoot_vps.sh shows permission errors

```bash
# Run with sudo for full diagnostics
sudo ./troubleshoot_vps.sh
```

### setup_nginx.sh fails

```bash
# Check if running as root
sudo ./setup_nginx.sh

# Check nginx installation
nginx -v

# Check configuration syntax
sudo nginx -t
```

## Tips & Best Practices

### Testing
- Always test locally before deploying to production
- Use `test_api.sh` in CI/CD pipelines
- Run tests after each deployment

### Troubleshooting
- Run `troubleshoot_vps.sh` before asking for help
- Save troubleshoot reports for reference
- Check logs directory for detailed errors

### Nginx Setup
- Use domain names for production (enables HTTPS)
- Test configuration before restarting: `sudo nginx -t`
- Keep backups of working configurations
- Monitor nginx logs regularly

### Security
- Never commit `.env` files
- Use strong passwords for database
- Enable HTTPS in production
- Keep firewall enabled with minimal open ports
- Regular security updates: `sudo apt-get update && sudo apt-get upgrade`

## Additional Resources

- [VPS Deployment Guide](VPS_DEPLOYMENT.md) - Complete deployment instructions
- [Nginx Documentation](https://nginx.org/en/docs/) - Official nginx docs
- [Let's Encrypt](https://letsencrypt.org/) - Free SSL certificates
- [API Documentation](../README.md) - Main project README

## Support

For issues with scripts:

1. Check this quick reference
2. Review error messages carefully
3. Run diagnostic script: `./troubleshoot_vps.sh`
4. Check logs: `tail -f logs/app.log`
5. Verify environment: `cat configs/.env`

## Script Updates

All scripts are version controlled. To update:

```bash
git pull origin main
chmod +x *.sh
```

## Contributing

When modifying scripts:
- Maintain backward compatibility
- Update this documentation
- Test on clean VPS before committing
- Add error handling for edge cases
- Keep output user-friendly with colors
