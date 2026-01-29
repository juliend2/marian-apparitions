# Deployment with Ansible

This directory contains Ansible playbooks for provisioning and deploying the Mariana Apparitions web application to an Ubuntu server.

## Architecture

The deployment creates:
- 3 instances of the app running on ports 8081, 8082, 8083
- nginx reverse proxy with round-robin load balancing
- Let's Encrypt TLS certificates with automatic renewal
- systemd services for each app instance
- UFW firewall (allows SSH, HTTP, HTTPS)

## Prerequisites

**On your local machine:**
- Ansible installed (`pip install ansible` or use your package manager)
- SSH access to the target Ubuntu server
- The app binary built (run `make build` from project root)

**On the target server:**
- Fresh Ubuntu 20.04+ installation
- Root or sudo access
- Domain name pointing to the server's IP address

## Configuration

1. **Edit `vars.yml`** with your settings:
   ```yaml
   domain_name: your-domain.example.com
   letsencrypt_email: your-email@example.com
   ```

2. **Edit `inventory.ini`** with your server details:
   ```ini
   [marianapparitions]
   your-server.example.com ansible_user=ubuntu
   ```

   Or use an IP address:
   ```ini
   [marianapparitions]
   192.168.1.100 ansible_user=ubuntu
   ```

## Usage

### Initial Deployment

1. Build the application binary:
   ```bash
   cd /home/julien/Workbench/marianapparitions
   make build
   ```

2. Run the playbook:
   ```bash
   cd ansible
   ansible-playbook -i inventory.ini provision.yml
   ```

3. If you need to provide SSH key or become password:
   ```bash
   ansible-playbook -i inventory.ini provision.yml --private-key=~/.ssh/your_key --ask-become-pass
   ```

### Updating the Application

To deploy a new version:

1. Build the new binary with `make build`
2. Run the playbook again:
   ```bash
   cd ansible
   ansible-playbook -i inventory.ini provision.yml
   ```

The playbook is idempotent - it will only update the binary and restart services if the binary has changed.

### Managing Services

SSH into the server and use systemctl:

```bash
# Check status of all instances
sudo systemctl status marianapparitions-1
sudo systemctl status marianapparitions-2
sudo systemctl status marianapparitions-3

# View logs
sudo journalctl -u marianapparitions-1 -f
sudo journalctl -u marianapparitions-2 -f
sudo journalctl -u marianapparitions-3 -f

# Restart an instance
sudo systemctl restart marianapparitions-1

# Check nginx status
sudo systemctl status nginx

# View nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

### Certificate Renewal

Let's Encrypt certificates are automatically renewed via a cron job that runs daily at 3:00 AM. To manually renew:

```bash
sudo certbot renew
sudo systemctl reload nginx
```

## Directory Structure

```
ansible/
├── provision.yml              # Main playbook
├── vars.yml                   # Configuration variables
├── inventory.ini              # Server inventory
├── templates/
│   ├── marianapparitions.service.j2  # Systemd service template
│   └── nginx.conf.j2                 # Nginx configuration template
└── README.md                  # This file
```

## Customization

### Changing Number of Instances

Edit `vars.yml`:
```yaml
app_instances: 5  # Change from 3 to 5
app_base_port: 8081  # Will create instances on 8081-8085
```

### Changing Ports

Edit `vars.yml`:
```yaml
app_base_port: 9000  # Instances will run on 9000, 9001, 9002
```

### Adding Static File Serving

If you add static files to your app, update the nginx template to serve them directly:

```nginx
location /static/ {
    root {{ app_dir }};
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

## Troubleshooting

**Playbook fails on certbot:**
- Ensure your domain DNS is properly configured and pointing to the server
- Let's Encrypt requires port 80 to be accessible from the internet
- You can comment out the certbot tasks for initial testing with HTTP only

**App instances won't start:**
- Check logs: `sudo journalctl -u marianapparitions-1 -n 50`
- Verify binary permissions: `ls -la /opt/marianapparitions/app`
- Check if ports are available: `sudo netstat -tlnp | grep 808`

**Nginx returns 502 Bad Gateway:**
- Check if app instances are running: `sudo systemctl status marianapparitions-*`
- Verify ports in nginx config match systemd services
- Check nginx error log: `sudo tail -f /var/log/nginx/error.log`

## Security Notes

- The app runs as a dedicated user with limited privileges
- systemd services include security hardening (NoNewPrivileges, PrivateTmp)
- UFW firewall is enabled with minimal required ports
- Consider adding fail2ban for SSH protection
- Database file is stored in `/var/lib/marianapparitions/` with restricted permissions
