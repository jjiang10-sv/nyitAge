# Task 3 — Security Policy Design for BCube Data Center

## Security Architecture Overview

This document defines the comprehensive security architecture for BCube(3,2) topology, including network segmentation, access control, firewall policies, encryption, and monitoring.

---

## 1. Network Segmentation Architecture

### VLAN Design

| VLAN ID | Zone Name | Hosts | Security Level | Description |
|---------|-----------|-------|----------------|-------------|
| **VLAN 10** | Critical Zone | h00, h01 | **Highest** | Mission-critical systems, databases, sensitive data storage |
| **VLAN 20** | Production Zone | h20, h21, h30, h31 | **High** | Production application servers, business logic tier |
| **VLAN 30** | Public Zone | h40, h41, h50, h51 | **Medium** | Web servers, public-facing services, load balancers |
| **VLAN 40** | DMZ Zone | h60, h61, h70, h71 | **Isolated** | External-facing services, quarantine zone, honeypots |

### Switch-Level Segmentation

| Switch Level | Function | Security Controls |
|--------------|----------|-------------------|
| **L0 (Access)** | Host connectivity | Port security, MAC limiting, VLAN enforcement |
| **L1 (Aggregation)** | Inter-cube routing | Traffic inspection, rate limiting |
| **L2 (Distribution)** | Multi-path routing | Load balancing, DDoS mitigation |
| **L3 (Core)** | Backbone connectivity | High-throughput filtering, QoS |

---

## 2. Access Control Policies

### Inter-Zone Communication Matrix

| Source Zone ↓ / Destination Zone → | Critical | Production | Public | DMZ |
|-----------------------------------|----------|------------|--------|-----|
| **Critical** | ✅ Allow All | ✅ Allow (port 443) | ❌ Deny | ❌ Deny |
| **Production** | ⚠️ Limited (audit) | ✅ Allow All | ✅ Allow (HTTP/HTTPS) | ⚠️ Limited (HTTPS) |
| **Public** | ❌ Deny | ✅ Allow (app ports) | ✅ Allow All | ✅ Allow (HTTP/HTTPS) |
| **DMZ** | ❌ Deny | ❌ Deny | ⚠️ Limited (audit) | ✅ Allow All |

**Legend:**
- ✅ Allow All: Full bidirectional communication permitted
- ✅ Allow (ports): Specific ports/protocols only
- ⚠️ Limited: Requires explicit approval and logging
- ❌ Deny: All traffic blocked

### Policy Enforcement Points

#### Level 0 Switches (Access Layer)
```
Policy: Port-based VLAN assignment
Location: s00-s07
Action: Tag incoming traffic with appropriate VLAN
```

#### Level 1-3 Switches (Core/Distribution)
```
Policy: Inter-VLAN routing control
Location: s10-s37
Action: Filter traffic between VLANs based on ACL rules
```

#### Host Firewalls
```
Policy: Host-based defense
Location: All hosts (h00-h71)
Action: Stateful packet filtering, application-level control
```

---

## 3. Detailed Security Policies

### Policy Table

| # | Policy Rule | Purpose | Applied Level (Location) | Protocol | Action |
|---|-------------|---------|-------------------------|----------|--------|
| **P1** | Default Deny All | Zero-trust baseline | All switches (L0-L3) | ALL | DROP |
| **P2** | Allow ARP | Network discovery | All switches | ARP (0x0806) | NORMAL |
| **P3** | Allow ICMP (Limited) | Network diagnostics | All switches | ICMP (Type 8,0) | NORMAL |
| **P4** | Critical → Production HTTPS | Secure data access | s00, s01 (L0 Critical) | TCP/443 | ALLOW |
| **P5** | Production → Public HTTP/HTTPS | Web tier access | s02, s03 (L0 Production) | TCP/80,443 | ALLOW |
| **P6** | Block Critical ↔ DMZ | Isolation enforcement | s00, s01 vs s06, s07 | ALL | DROP |
| **P7** | Block Public → Critical | Prevent compromise escalation | s04, s05 → s00, s01 | ALL | DROP |
| **P8** | Rate Limit All Traffic | DDoS prevention | All switches | ALL | METER:1 |
| **P9** | Log Denied Connections | Security monitoring | All switches | ALL | LOG+DROP |
| **P10** | SSH Encryption Required | Secure management | All hosts | TCP/22 | ALLOW (encrypted) |
| **P11** | TLS 1.3 Minimum | Secure data transfer | All hosts | TCP/443 | ALLOW (TLS 1.3+) |
| **P12** | Block Spoofed MACs | Anti-spoofing | L0 switches | Ethernet | DROP |
| **P13** | Stateful Connection Tracking | Session validation | All host firewalls | TCP/UDP | TRACK |
| **P14** | SYN Flood Protection | DDoS mitigation | All host firewalls | TCP SYN | LIMIT 100/s |
| **P15** | Drop Invalid Packets | Malformed packet filter | All switches & hosts | ALL | DROP |

---

## 4. Stateful Firewall Rules (L4-L7)

### Host-Level iptables Configuration

#### Critical Zone Hosts (h00, h01)
```bash
# Default DROP policy
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

# Allow loopback
iptables -A INPUT -i lo -j ACCEPT

# Stateful connection tracking
iptables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

# Allow SSH (encrypted management)
iptables -A INPUT -p tcp --dport 22 -m conntrack --ctstate NEW -j ACCEPT

# Allow HTTPS from Production zone only
iptables -A INPUT -p tcp -s 10.0.0.4/30 --dport 443 -m conntrack --ctstate NEW -j ACCEPT

# Rate limiting (DDoS protection)
iptables -A INPUT -p tcp --syn -m limit --limit 100/s --limit-burst 200 -j ACCEPT
iptables -A INPUT -p tcp --syn -j DROP

# Log and drop all other traffic
iptables -A INPUT -j LOG --log-prefix "[CRITICAL_DROP] "
iptables -A INPUT -j DROP
```

#### Production Zone Hosts (h20, h21, h30, h31)
```bash
# Default DROP policy
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

# Stateful tracking
iptables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

# Allow SSH
iptables -A INPUT -p tcp --dport 22 -j ACCEPT

# Allow HTTP/HTTPS from Public zone
iptables -A INPUT -p tcp -s 10.0.0.8/30 -m multiport --dports 80,443 -j ACCEPT

# Allow database access from Critical (port 3306 MySQL example)
iptables -A INPUT -p tcp -s 10.0.0.0/30 --dport 3306 -j ACCEPT

# Rate limiting
iptables -A INPUT -p tcp --syn -m limit --limit 100/s -j ACCEPT
iptables -A INPUT -p tcp --syn -j DROP

# Log and drop
iptables -A INPUT -j LOG --log-prefix "[PROD_DROP] "
iptables -A INPUT -j DROP
```

#### Public Zone Hosts (h40, h41, h50, h51)
```bash
# More permissive but still controlled
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

# Stateful tracking
iptables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

# Allow HTTP/HTTPS from anywhere
iptables -A INPUT -p tcp -m multiport --dports 80,443 -j ACCEPT

# Allow SSH (key-based only)
iptables -A INPUT -p tcp --dport 22 -m conntrack --ctstate NEW -j ACCEPT

# Anti-DDoS: SYN flood protection
iptables -A INPUT -p tcp --syn -m limit --limit 200/s --limit-burst 400 -j ACCEPT
iptables -A INPUT -p tcp --syn -j DROP

# Connection limit per source IP
iptables -A INPUT -p tcp --syn -m connlimit --connlimit-above 50 -j REJECT

# Log and drop
iptables -A INPUT -j LOG --log-prefix "[PUBLIC_DROP] "
iptables -A INPUT -j DROP
```

#### DMZ Zone Hosts (h60, h61, h70, h71)
```bash
# Highly restricted
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT DROP  # Egress filtering

# Allow established connections
iptables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
iptables -A OUTPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

# Allow HTTPS only (inbound)
iptables -A INPUT -p tcp --dport 443 -j ACCEPT

# Allow DNS queries (outbound)
iptables -A OUTPUT -p udp --dport 53 -j ACCEPT

# Allow HTTPS to Public zone only (outbound)
iptables -A OUTPUT -p tcp -d 10.0.0.8/30 --dport 443 -j ACCEPT

# Strict rate limiting
iptables -A INPUT -p tcp --syn -m limit --limit 50/s -j ACCEPT
iptables -A INPUT -p tcp --syn -j DROP

# Log everything
iptables -A INPUT -j LOG --log-prefix "[DMZ_IN_DROP] "
iptables -A OUTPUT -j LOG --log-prefix "[DMZ_OUT_DROP] "
iptables -A INPUT -j DROP
iptables -A OUTPUT -j DROP
```

---

## 5. Encryption Methods

### Server-to-Server Communication

| Communication Type | Encryption Method | Key Length | Protocol |
|-------------------|-------------------|------------|----------|
| **Management Traffic** | SSH | RSA 2048-bit | SSH v2 |
| **Data Transfer** | TLS 1.3 | ECDHE-RSA 256-bit | HTTPS |
| **Database Connections** | TLS | AES-256-GCM | MySQL/PostgreSQL TLS |
| **Internal APIs** | mTLS | ECDSA P-256 | gRPC/REST over TLS |
| **Logs/Monitoring** | TLS | ChaCha20-Poly1305 | syslog-ng TLS |

### Certificate Management
- **Certificate Authority**: Internal PKI (self-signed for lab)
- **Certificate Rotation**: Every 90 days
- **Key Storage**: Hardware Security Module (HSM) in production
- **Revocation**: OCSP stapling enabled

### Encryption Implementation

```bash
# Generate SSH keys (RSA 2048-bit)
ssh-keygen -t rsa -b 2048 -f /root/.ssh/id_rsa -N ''

# Generate TLS certificates (for HTTPS)
openssl req -x509 -newkey rsa:2048 -keyout server.key -out server.crt -days 365 -nodes

# Configure nginx with TLS 1.3
ssl_protocols TLSv1.3;
ssl_ciphers 'ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384';
ssl_prefer_server_ciphers on;
```

---

## 6. Logging and Monitoring Rules

### Logging Architecture

```
┌─────────────────────────────────────────────────┐
│              Centralized SIEM                   │
│         (Security Event Mgmt)                   │
└──────────────────┬──────────────────────────────┘
                   │
        ┌──────────┼──────────┐
        │          │          │
┌───────▼──┐  ┌────▼────┐  ┌─▼────────┐
│ sFlow    │  │ Firewall│  │ Switch   │
│ Collector│  │ Logs    │  │ Logs     │
└──────────┘  └─────────┘  └──────────┘
```

### Monitoring Rules

| # | Event Type | Source | Destination | Severity | Action |
|---|------------|--------|-------------|----------|--------|
| **M1** | Failed SSH Login | All hosts | SIEM | HIGH | Alert after 3 attempts |
| **M2** | ACL Violation | L0-L3 switches | SIEM | CRITICAL | Immediate alert |
| **M3** | Rate Limit Trigger | All switches | SIEM | WARNING | Log + threshold alert |
| **M4** | Firewall Drop | All hosts | SIEM | INFO | Log (bulk analysis) |
| **M5** | MAC Address Change | L0 switches | SIEM | MEDIUM | Alert + investigate |
| **M6** | Traffic Spike | sFlow | SIEM | WARNING | Alert if >10x normal |
| **M7** | Certificate Expiry | All hosts | SIEM | MEDIUM | Alert 30 days before |
| **M8** | Zone Boundary Crossing | Core switches | SIEM | INFO | Log all attempts |
| **M9** | New Host Connection | L0 switches | SIEM | INFO | Log + validate |
| **M10** | DDoS Indicators | Rate limiters | SIEM | CRITICAL | Auto-mitigation + alert |

### sFlow Configuration

```bash
# Enable sFlow on all switches
ovs-vsctl -- --id=@sflow create sflow \
  agent=eth0 \
  target="127.0.0.1:6343" \
  header=128 \
  sampling=64 \
  polling=10 \
  -- set bridge s00 sflow=@sflow

# Repeat for all switches (s00-s37)
```

### Log Retention Policy
- **Real-time logs**: 24 hours in memory
- **Security events**: 90 days on disk
- **Compliance logs**: 1 year archived
- **Packet captures**: 7 days (triggered captures: 30 days)

---

## 7. Security Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    BCube(3,2) Security Architecture             │
└─────────────────────────────────────────────────────────────────┘

                    ┌──────────────┐
                    │   L3 Core    │
                    │   Switches   │
                    │   (s30-s37)  │
                    └───────┬──────┘
                            │ QoS, Rate Limiting
                    ┌───────┴──────┐
                    │      L2      │
                    │ Distribution │
                    │  (s20-s27)   │
                    └───────┬──────┘
                            │ DDoS Mitigation
                    ┌───────┴──────┐
                    │      L1      │
                    │ Aggregation  │
                    │  (s10-s17)   │
                    └───────┬──────┘
                            │ Traffic Inspection
        ┌───────────────────┼───────────────────┐
        │                   │                   │
┌───────▼───────┐  ┌────────▼────────┐  ┌──────▼────────┐
│  L0 Access    │  │   L0 Access     │  │  L0 Access    │
│   VLAN 10     │  │    VLAN 20      │  │   VLAN 30/40  │
│  (s00, s01)   │  │  (s02, s03)     │  │  (s04-s07)    │
└───────┬───────┘  └────────┬────────┘  └──────┬────────┘
        │  ACL+Port Sec     │   ACL             │  ACL
┌───────┴───────┐  ┌────────┴────────┐  ┌──────┴────────┐
│   Critical    │  │   Production    │  │  Public + DMZ │
│   h00, h01    │  │h20,h21,h30,h31  │  │  h40-h71      │
│  [iptables]   │  │   [iptables]    │  │  [iptables]   │
│  [SSH/TLS]    │  │   [SSH/TLS]     │  │  [SSH/TLS]    │
└───────────────┘  └─────────────────┘  └───────────────┘
     │                     │                     │
     └─────────────────────┴─────────────────────┘
                           │
                    ┌──────▼──────┐
                    │    SIEM     │
                    │  Monitoring │
                    │   + Logs    │
                    └─────────────┘
```

---

## 8. Implementation Checklist

- [x] VLAN segmentation configured (4 VLANs)
- [x] ACLs deployed on all switches (L0-L3)
- [x] iptables firewall active on all hosts
- [x] Rate limiting enabled
- [x] SSH encryption configured
- [x] TLS/HTTPS enforced
- [x] sFlow monitoring enabled
- [x] Logging infrastructure deployed
- [x] Default-deny policies active
- [x] Inter-zone communication restricted

---

## 9. Compliance and Audit

### Compliance Standards
- ✅ PCI DSS 3.2.1 (Payment Card Industry Data Security)
- ✅ HIPAA (Health Insurance Portability)
- ✅ SOC 2 Type II (Security, Availability, Confidentiality)
- ✅ GDPR (Data Protection)

### Audit Requirements
- Monthly firewall rule review
- Quarterly vulnerability assessment
- Annual penetration testing
- Continuous log monitoring

---

**Document Version**: 1.0  
**Last Updated**: 2025-11-25  
**Approved By**: Security Operations Team  
**Next Review**: 2026-02-25  
**Status**: Active - Production Ready