# Docker Daemon Security Configuration - Installation Guide

## Overview

This file contains Docker daemon security settings that improve CIS Docker Benchmark compliance (Section 2).

## Configuration Details

```json
{
  "icc": false, // Disable inter-container communication on default bridge
  "log-level": "info", // Standard logging level
  "userns-remap": "default", // Enable user namespace remapping
  "disable-legacy-registry": true, // Disable insecure registries
  "live-restore": true, // Keep containers alive during daemon downtime
  "userland-proxy": false, // Use kernel port forwarding (better performance)
  "no-new-privileges": true // Prevent container privilege escalation
}
```

## Installation Instructions

### For Development (Local Machine)

1. **Backup existing configuration** (if exists):

   ```bash
   sudo cp /etc/docker/daemon.json /etc/docker/daemon.json.backup
   ```

2. **Copy new configuration**:

   ```bash
   sudo cp daemon.json /etc/docker/daemon.json
   ```

3. **Restart Docker daemon**:

   ```bash
   sudo systemctl restart docker
   ```

4. **Verify configuration**:
   ```bash
   docker info | grep -i "userns\|live restore"
   ```

### For Production Server

1. **Review configuration** before applying:

   - `userns-remap` creates new user namespace (may affect existing containers)
   - `icc: false` isolates containers on default bridge (use custom networks for inter-container communication)
   - `live-restore: true` requires systemd or similar init system

2. **Test in staging environment first**:

   ```bash
   # Apply configuration
   sudo cp daemon.json /etc/docker/daemon.json
   sudo systemctl restart docker

   # Test application
   docker-compose up -d
   docker ps  # Verify all containers are running
   ```

3. **Monitor for issues**:
   ```bash
   sudo journalctl -u docker -f
   ```

## Important Notes

### User Namespace Remapping (`userns-remap: default`)

⚠️ **This creates a new storage location for images and containers!**

- Old location: `/var/lib/docker/`
- New location: `/var/lib/docker/<uid>.<gid>/`

**Migration steps**:

1. Stop all containers: `docker-compose down`
2. Apply daemon.json and restart Docker
3. Rebuild images: `docker-compose build`
4. Restart containers: `docker-compose up -d`

### Inter-Container Communication (`icc: false`)

When `icc: false` is set:

- Containers on **default bridge** cannot communicate
- Our application uses **custom networks** (already configured in docker-compose.yml)
- **No changes needed** - custom networks are unaffected

### Live Restore (`live-restore: true`)

Benefits:

- Containers keep running during Docker daemon updates
- Reduces downtime during maintenance

Requirements:

- systemd or similar init system
- Not compatible with Docker Swarm mode

## Troubleshooting

### Issue: Containers fail to start after applying config

**Check 1**: User namespace permissions

```bash
# Verify subuid/subgid are configured
grep dockremap /etc/subuid /etc/subgid

# If missing, add them:
echo "dockremap:100000:65536" | sudo tee -a /etc/subuid
echo "dockremap:100000:65536" | sudo tee -a /etc/subgid
sudo systemctl restart docker
```

**Check 2**: Volume permissions

```bash
# Volumes may need permission adjustment
docker-compose down
docker volume ls
# Recreate volumes with correct permissions
docker-compose up -d
```

### Issue: Cannot access containers from host

This is expected with `userns-remap`. Use:

```bash
docker exec -it container_name sh
```

### Issue: High CPU usage after enabling userland-proxy: false

This should actually **reduce** CPU usage. If experiencing issues:

```bash
# Check iptables rules
sudo iptables -t nat -L -n -v

# Verify kernel forwarding is enabled
sysctl net.ipv4.ip_forward
```

## Reverting Changes

If you need to revert:

1. **Restore backup**:

   ```bash
   sudo cp /etc/docker/daemon.json.backup /etc/docker/daemon.json
   ```

2. **Or remove user namespace remapping only**:

   ```bash
   sudo nano /etc/docker/daemon.json
   # Remove the "userns-remap" line
   ```

3. **Restart Docker**:
   ```bash
   sudo systemctl restart docker
   ```

## Verification

After installation, verify with CIS Docker Benchmark:

```bash
# Clone benchmark tool
cd ~
git clone https://github.com/docker/docker-bench-security.git
cd docker-bench-security

# Run scan
sudo sh docker-bench-security.sh

# Check Section 2 results
sudo sh docker-bench-security.sh | grep -A 20 "Section 2"
```

Expected improvements:

- ✅ 2.2: Network traffic restricted
- ✅ 2.9: User namespace support enabled
- ✅ 2.14: Containers restricted from acquiring new privileges
- ✅ 2.15: Live restore enabled
- ✅ 2.16: Userland proxy disabled

## References

- [Docker daemon configuration](https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-configuration-file)
- [User namespace remapping](https://docs.docker.com/engine/security/userns-remap/)
- [CIS Docker Benchmark Section 2](https://www.cisecurity.org/benchmark/docker)

---

**Note**: This configuration is already applied in `docker-compose.yml` through service-level security options. The `daemon.json` provides **daemon-wide** security controls.
