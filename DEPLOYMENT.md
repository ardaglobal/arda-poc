# Arda Blockchain Deployment Guide

This guide outlines how to deploy the Arda blockchain application to a Digital Ocean droplet (or any Linux server).

## Prerequisites

- Ubuntu 20.04+ or Debian 11+ server
- Root or sudo access
- Domain name pointed to your server's IP address
- At least 4GB RAM and 25GB disk space recommended

## System Setup

### 1. Update System Packages

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl wget git build-essential git ca-certificates build-base
```

### 2. Install Go 1.24.5 or Greater

Air (the hot-reload tool used by the sidecar) requires Go 1.24.5 or greater.

```bash
# Remove any existing Go installation
sudo rm -rf /usr/local/go

# Download and install Go 1.24.5 (or latest version)
curl -sL "https://go.dev/dl/go1.24.5.linux-amd64.tar.gz" -o /tmp/go.tar.gz
sudo tar -C /usr/local -xzf /tmp/go.tar.gz
rm /tmp/go.tar.gz

# Add Go to PATH
echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.bashrc
export PATH=/usr/local/go/bin:$PATH

# Verify installation
go version
```

### 3. Install Docker (Required by setup script)

```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
rm get-docker.sh

# Log out and back in for Docker group changes to take effect
```

### 4. Clone and Setup Application

```bash
# Clone the repository
git clone <repository-url> /opt/arda-poc
cd /opt/arda-poc

# Run the development environment setup
make setup-dev

# Build the application
make install
```

### 5. Initialize Blockchain Data

```bash
# Run development setup once to initialize blockchain data
# This will create the ~/.arda-poc directory and configuration
make dev
# Press Ctrl+C after the blockchain starts successfully
```

## Nginx Configuration

### 1. Install Nginx

```bash
sudo apt install -y nginx
```

### 2. Configure Nginx

Create the Nginx configuration file:

```bash
sudo tee /etc/nginx/sites-available/arda > /dev/null <<EOF
server {
    listen 80;
    server_name your-domain.com;

    # Redirect all HTTP traffic to HTTPS
    return 301 https://\$server_name\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    # SSL certificate paths (configure after obtaining certificates)
    # ssl_certificate /path/to/your/certificate.crt;
    # ssl_certificate_key /path/to/your/private.key;

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # Proxy to the sidecar API on port 8080
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;

        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";

        # Timeout settings
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Optional: Serve blockchain RPC directly (if needed)
    location /rpc/ {
        proxy_pass http://localhost:26657/;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    # Optional: Serve blockchain REST API directly (if needed)
    location /api/ {
        proxy_pass http://localhost:1317/;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF
```

### 3. Enable the Site

```bash
# Enable the site
sudo ln -s /etc/nginx/sites-available/arda /etc/nginx/sites-enabled/

# Remove default site if it exists
sudo rm -f /etc/nginx/sites-enabled/default

# Test Nginx configuration
sudo nginx -t

# Start and enable Nginx
sudo systemctl start nginx
sudo systemctl enable nginx
```

### 4. SSL Certificate Setup (Recommended)

Install and configure Let's Encrypt SSL certificate:

```bash
# Install Certbot
sudo apt install -y certbot python3-certbot-nginx

# Obtain SSL certificate
sudo certbot --nginx -d your-domain.com

# Test automatic renewal
sudo certbot renew --dry-run
```

## Service Configuration


## Environment Variables

Create a `.env` file in the application directory for configuration:

```bash
sudo tee /opt/arda-poc/.env > /dev/null <<EOF
# Admin authentication key
ADMIN_KEY=your-secure-admin-key-here

# CORS configuration
ALLOWED_ORIGINS=https://your-domain.com
ALLOW_CREDENTIALS=true

# Blockchain configuration
BLOCKCHAIN_REST_API_URL=http://localhost:1317
EOF
```

## Monitoring and Maintenance

### 1. View Logs

```bash
# View sidecar logs
tail -f nohup.out

# View Nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

#

## Security Considerations

1. **Firewall Configuration**: Only expose necessary ports (80, 443, 22)
2. **SSH Security**: Use key-based authentication, disable password login
3. **Regular Updates**: Keep system packages and Go version updated
4. **Backup**: Regularly backup the `~/.arda-poc` directory
5. **Admin Key**: Use a strong, unique admin key and keep it secure
6. **SSL**: Always use HTTPS in production

## Ports Used

- **26656**: P2P networking (internal)
- **26657**: RPC server (internal)
- **1317**: REST API (internal)
- **9090**: gRPC API (internal)
- **8080**: Transaction sidecar API (proxied through Nginx)
- **80/443**: HTTP/HTTPS (public, served by Nginx)

## Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure no other services are using the required ports
2. **Permission issues**: Ensure proper file ownership and permissions
3. **Go version**: Verify Go 1.24.5+ is installed correctly
4. **Network issues**: Check firewall rules and DNS configuration

### Health Checks

```bash
# Check if blockchain is running
curl http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info

# Check if sidecar is running
curl http://localhost:8080/user/list

# Check if Nginx is proxying correctly
curl https://your-domain.com/user/list
```

Remember to replace `your-domain.com` and `your-secure-admin-key-here` with your actual domain and a secure admin key throughout this guide.