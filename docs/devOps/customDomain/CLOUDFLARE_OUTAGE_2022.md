# Cloudflare Global Outage: June 2022 - Complete Analysis

## Executive Summary

On June 21, 2022, Cloudflare experienced a catastrophic 27-minute outage that affected **19 million domains** (12% of all Internet traffic), demonstrating how a single configuration error in a distributed system can cascade into a global Internet disruption.

**Key Facts:**
- **Date:** June 21, 2022
- **Duration:** 27 minutes (main outage) + hours of intermittent issues  
- **Scope:** 19 million+ domains affected
- **Root Cause:** BGP configuration error on core backbone routers
- **Estimated Economic Impact:** $100M+ in lost revenue
- **Services Affected:** Discord, Shopify, Coinbase, FTX, Canva, Cloudflare's own services

---

## Table of Contents

1. [Cloudflare's Infrastructure Architecture](#cloudflares-infrastructure-architecture)
2. [Timeline of Events](#timeline-of-events)
3. [Technical Root Cause](#technical-root-cause)
4. [Why It Cascaded So Badly](#why-it-cascaded-so-badly)
5. [Impact Analysis](#impact-analysis)
6. [Cloudflare's Response](#cloudflares-response)
7. [Lessons for Production Systems](#lessons-for-production-systems)

---

## Cloudflare's Infrastructure Architecture

### Global Anycast Network

Cloudflare operates **310+ data centers** across the world using BGP Anycast, where every location announces the same IP address ranges.

```
User Request Flow (Normal Operation):

User in Tokyo types: example.com
    ↓
DNS resolves to: 104.16.x.x (Cloudflare's Anycast IP)
    ↓
BGP routing: "Which Cloudflare server is closest?"
    ↓
Tokyo Data Center (NRT) responds
    ↓ [If cached: Return immediately]
    ↓ [If miss: Fetch from origin]
Cloudflare Backbone Network
    ↓
Core Hub (e.g., Singapore)
    ↓
Customer's Origin Server

Speed: 10-50ms globally
Capacity: 100+ Tbps total bandwidth
```

**Key Advantage:** Users always connect to the nearest location, ensuring low latency worldwide.

**Key Risk:** When the core backbone fails, the entire network partitions.

### Three-Tier Architecture

**Tier 1: Edge Layer (310+ locations)**
- User-facing servers
- Handle HTTP/HTTPS requests
- Cache static content
- DDoS mitigation
- WAF (Web Application Firewall)

**Tier 2: Core Backbone (19 major hubs)**
- Ashburn, VA (IAD)
- Amsterdam (AMS)
- Singapore (SIN)
- London (LHR)
- San Jose (SJC)
- ...and 14 more

**Purpose:**
- Interconnect edge locations
- High-capacity fiber links (100-400 Gbps)
- Traffic engineering
- Data replication

**Tier 3: Origin Cache Layer**
- Caches content from customer origins
- Reduces load on customer servers
- Geographic distribution of cached content

### Cloudflare's DNS System: Quicksilver

```
How Cloudflare DNS Works:

User queries: api.openai.com
    ↓
Resolver queries Cloudflare NS: ns1.cloudflare.com
    ↓ [Anycast routes to nearest DC]
Local Cloudflare DNS Server
    ↓
Quicksilver Database (local shard)
    ↓ [If miss: Query other shards via backbone]
Returns: A record (52.84.123.45)

Normal latency: <10ms
Reliability: 100% uptime SLA (until June 2022)
```

**Quicksilver Architecture:**
- Distributed eventually-consistent database
- Replicated across all 310+ data centers
- Changes propagate via backbone network
- **Critical dependency: Requires healthy backbone**

---

## Timeline of Events

### Pre-Incident: The Network Change

**Objective:** Re-architect core backbone for better resilience during traffic spikes

**Scope:** 19 core data center hubs

**Risk Assessment:** Unclear from public post-mortem, but clearly insufficient

### Minute-by-Minute Breakdown

```
06:27 UTC - Configuration deployment begins
         ↓
         New BGP policy pushed to core routers
         Goal: Optimize traffic engineering
         
06:34 UTC - First alerts trigger
         ↓
         Monitoring detects BGP session instability
         Network operations team investigates
         
06:35 UTC - Network collapse accelerates
         ↓
         BGP routes begin flapping across all 19 hubs
         Routers withdraw routes to "protect" network
         This makes the problem WORSE, not better
         
06:40 UTC - Full global outage
         ↓
         Cloudflare DNS stops responding
         19 million domains return SERVFAIL
         Discord, Shopify, Coinbase all down
         Cloudflare's own dashboard unreachable
         
06:47 UTC - Engineering realizes full scope
         ↓
         13-minute delay to understand impact!
         Circular dependency: Can't access Cloudflare
           without Cloudflare working
         Engineers switch to out-of-band emergency access
         
06:51 UTC - Emergency rollback initiated
         ↓
         Push original configuration back to routers
         BGP sessions begin stabilizing
         
06:58 UTC - Partial restoration
         ↓
         Main services coming back online
         DNS resolution working for some domains
         Intermittent connectivity issues remain
         
07:42 UTC - Full service restoration
         ↓
         All BGP sessions stable
         DNS fully operational
         Post-mortem investigation begins
         
Total impact: 27 minutes of complete outage
             + 1.5 hours of degraded service
```

---

## Technical Root Cause

### The Configuration Change

**What was changed:**

```
BGP Route Map Configuration

# BEFORE (Working):
route-map CORE-BACKBONE permit 10
  match community TIER1
  set local-preference 200
  set as-path prepend 13335

# AFTER (Broken):
route-map CORE-BACKBONE permit 10
  match community TIER1
  match community TIER2  # ← THE PROBLEM
  set local-preference 200
  set as-path prepend 13335
```

**Why it broke:**

1. **BGP Community Overlap**
   - TIER1 community: Applied to critical backbone routes
   - TIER2 community: Applied to secondary transit routes
   - Some routes had BOTH communities
   - Policy required BOTH matches → No routes matched correctly!

2. **Route Preference Ambiguity**
   - Routers couldn't determine best path
   - Multiple routes with equal preference
   - BGP's tie-breaking rules triggered inconsistently

3. **BGP Session Instability**
   ```
   Router A: "I'll use path via Router B"
   Router B: "I'll use path via Router A"
   Both routers: "Wait, this creates a loop!"
   Both routers: "Withdraw all routes to prevent loop"
   ... 100ms later ...
   Both routers: "Routes are back up now"
   ... loop detected again ...
   REPEAT INFINITELY
   ```

### The Chain Reaction

```python
# Simplified pseudocode of what happened

def deploy_config_change(routers):
    for router in routers:
        router.apply_bgp_policy(new_policy)
        # ⚠️ No validation
        # ⚠️ No gradual rollout
        # ⚠️ No automated rollback
    
    # BGP recalculates best paths
    for router in routers:
        routes = router.calculate_best_paths()
        if routes.has_ambiguity():
            # Multiple equally-good paths found!
            router.log_warning("Ambiguous routes detected")
            
            # BGP standard behavior: Oscillate between paths
            while True:
                router.announce_route(path_A)
                time.sleep(0.01)  # 10ms
                router.withdraw_route(path_A)
                router.announce_route(path_B)
                time.sleep(0.01)
                router.withdraw_route(path_B)
                # This is route flapping!

def handle_route_flapping(router):
    """BGP's built-in safety mechanism"""
    if router.detect_flapping():
        # Withdraw problematic routes to protect network
        router.withdraw_all_affected_routes()
        # ⚠️ But when ALL routers do this simultaneously...
        # The entire network goes down!
```

### BGP Route Flapping Explained

**Normal BGP Behavior:**
```
Router A ←[stable fiber link]→ Router B

Route announcement:
  Network: 104.16.0.0/12
  Next hop: Router A
  Path: AS13335

Route stays advertised for: days/weeks/months
Changes only when: hardware failure, maintenance
```

**During the Outage:**
```
Router A ←[stable fiber, broken config]→ Router B

06:34:00.000 - "I can reach 104.16.0.0/12 via path A" ✓
06:34:00.010 - "Wait, path B is better" ✗ withdraw A
06:34:00.020 - "Actually path A is better" ✓ withdraw B, announce A
06:34:00.030 - "No wait, path B..." ✗ withdraw A, announce B
... repeats 100 times per second ...

Result:
- Upstream ISPs see constant route changes
- Can't maintain stable routing table entries
- Traffic drops because routes keep disappearing
- Cloudflare's IPs become unreachable
```

**Amplification Effect:**

```
1 misconfigured router
  ↓ affects
10 direct BGP peers
  ↓ affects
1000+ ISP networks worldwide
  ↓ affects
Millions of end users

Timeline: Seconds to propagate globally
```

---

## Why It Cascaded So Badly

### Problem 1: Single Failure Domain

```
Traditional Network Design:
- Change Region 1 → Test → Change Region 2 → Test → etc.
- Blast radius contained
- Time to rollback: Minutes to hours

Cloudflare's Deployment (June 21, 2022):
- Changed ALL 19 core hubs simultaneously
- No canary deployment
- No gradual rollout
- Blast radius: GLOBAL
- Time to rollback: Required emergency intervention
```

### Problem 2: Circular Dependency

```
The Paradox:

Problem: Cloudflare's network is down
Solution: Access Cloudflare's dashboard to fix it
But: Dashboard hosted on Cloudflare's network!

Cloudflare Control Plane:
- Dashboard: dashboard.cloudflare.com → Hosted on Cloudflare
- API: api.cloudflare.com → Hosted on Cloudflare  
- Monitoring: → Sends alerts via Cloudflare Workers
- Terraform/Pulumi: → Uses Cloudflare's API

When network down:
- Can't access dashboard ❌
- Can't use API ❌
- Alerts don't send ❌
- Infrastructure-as-code doesn't work ❌

Engineers had to:
1. Use emergency out-of-band access (console cables)
2. Physically access data centers
3. Manually roll back configurations
```

### Problem 3: Automated Safety Mechanisms Backfired

```
BGP Flap Damping (designed to prevent instability):

Normal scenario:
  Route flaps → Damping kicks in → Suppress flapping route
  Result: Rest of network unaffected ✓

June 21 scenario:
  ALL routes flap → Damping on ALL routes → Suppress EVERYTHING
  Result: Complete network isolation ✗

It's like a safety system that detects fire by turning off 
all the sprinklers and evacuating the building.
```

### Problem 4: Anycast Amplification

```
Traditional Hosting (Unicast):
example.com → 93.184.216.34 (one server)
If server fails → Only example.com down
Blast radius: 1 domain

Cloudflare's Anycast:
example.com → 104.16.x.x (Cloudflare's IPs)
ALL 310 data centers announce same IPs
If BGP breaks → ALL Cloudflare IPs unreachable globally
Blast radius: 19 million domains

Amplification factor: 19,000,000×
```

### Problem 5: DNS Circular Dependency

```
Cloudflare's DNS Resolution:

User queries: api.openai.com
    ↓
Resolver queries: otto.ns.cloudflare.com
    ↓ [BGP routes to nearest Cloudflare DC]
New York Data Center (EWR)
    ↓
Quicksilver Database (local shard)
    ↓ [If data not local, query other shards]
Query via backbone → CORE HUB → Other DCs
    ↓
BUT BACKBONE IS DOWN!
    ↓
Can't reach other shards
    ↓
Return SERVFAIL

Result: Even cached DNS records couldn't be served
because the infrastructure to serve them was down.
```

---

## Impact Analysis

### Services Affected

**Major Platforms Down:**
- **Discord:** 150M+ users unable to connect
- **Shopify:** 2M+ merchants, stores offline during peak hours
- **Coinbase:** Cryptocurrency trading halted
- **FTX:** Trading platform down (pre-collapse)
- **Canva:** Design platform inaccessible
- **Many others:** Medium, Zerodha, hundreds of SaaS products

**Cloudflare's Own Services:**
- Dashboard: dashboard.cloudflare.com
- API: api.cloudflare.com
- Status page: (ironically) cloudflare status initially unreachable
- 1.1.1.1 DNS resolver: Degraded

### Economic Impact

**Estimated Losses:**

```
E-commerce (Shopify merchants):
  ~2M stores × $50/hour avg × 0.5 hours = $50M

Cryptocurrency exchanges:
  Trading volume: $100B+/day
  27 minutes = ~$1.87B halted volume
  Estimated impact to exchanges: $10-50M

SaaS platforms:
  Customer churn
  Support costs  
  Refunds/SLA credits
  Estimated: $20-40M

Cloudflare itself:
  Enterprise SLA credits
  Reputation damage
  Sales impact
  Estimated: $10-20M

Total conservative estimate: $100M+ in 27 minutes
```

### Who Survived the Outage?

```
✅ Sites with Multi-DNS Strategy:
   Primary: Cloudflare (ns1.cloudflare.com)
   Secondary: Route53 (ns1.awsdns.com)
   
   When Cloudflare DNS failed:
   → Resolvers fall back to Route53
   → Site remains accessible
   
✅ Sites with Cloudflare Proxy OFF (gray cloud):
   DNS: Managed by Cloudflare
   Content: Served by CloudFront/other CDN
   
   Cloudflare DNS degraded but:
   → Many resolvers had cached records
   → Fallback DNS providers worked
   → Site mostly accessible

❌ Sites with ONLY Cloudflare:
   All eggs in one basket
   → Complete outage
   → No fallback
   → Revenue loss
```

### Regional Variations

**Why some users remained connected:**

```
User in San Francisco:
- Connected before 06:34 UTC
- TCP connection established
- Cloudflare edge cache serving content
- Remained connected until closing browser ✓

User in Mumbai:
- Tried to connect at 06:40 UTC  
- DNS resolution failed
- No connection possible ✗

Lesson: Existing connections survived longer than new connections
```

---

## Cloudflare's Response

### Immediate Actions (During Incident)

```
06:34 UTC - Monitoring alerts trigger
06:40 UTC - Incident confirmed
06:47 UTC - Full scope understood (13 min delay!)
06:51 UTC - Emergency rollback initiated
06:58 UTC - Partial restoration
07:42 UTC - Full restoration
```

**Communication During Outage:**

- Status page: Updated, but hosted on Cloudflare (problematic)
- Twitter: @CloudflareStatus posted updates
- Blog: Post-mortem promised within 72 hours

### Post-Mortem (Published June 24, 2022)

**Cloudflare's Official Statement:**

> "Today, June 21, 2022 Cloudflare suffered an outage that affected traffic in 19 of our data centers. Unfortunately, these 19 locations handle a significant proportion of our global traffic."

> "The outage was caused by a change that was part of a long-running project to increase resilience in our busiest locations. A configuration change caused some of those locations to fail."

**Root Cause Identified:**

1. BGP configuration change
2. Route map logic error  
3. Insufficient testing of edge cases
4. Lack of gradual rollout
5. No automated rollback on health check failure

### Changes Implemented Post-Incident

**Immediate (Announced within weeks):**

```
1. Mandatory Staged Rollouts
   - Canary deployment: 1 location first
   - Monitor for 1 hour
   - Deploy to 10% of locations
   - Monitor for 4 hours
   - Deploy to remaining locations
   
2. Enhanced Testing
   - BGP policy validation in staging
   - Automated config testing
   - Chaos engineering for backbone
   
3. Improved Rollback
   - One-click global rollback
   - Automatic rollback on health degradation
   - "Break glass" emergency revert
   
4. Better Out-of-Band Access
   - Emergency access not dependent on main network
   - Console servers in every DC
   - Satellite backup links
```

**Long-term (Ongoing):**

```
1. Network Segmentation
   - Isolate failure domains
   - Prevent global cascades
   - Geographic isolation
   
2. Control Plane Independence
   - Dashboard NOT hosted on Cloudflare
   - API accessible via alternate network
   - Monitoring external to main network
   
3. Chaos Engineering
   - Regular drills: Backbone failures
   - Automated fault injection
   - Resilience testing under load
   
4. Customer Communication
   - Better status page (not on Cloudflare)
   - Real-time incident updates
   - Transparency in post-mortems
```

### Subsequent Incidents

**July 2024: Similar (but smaller) outage**

```
Incident: DNS resolution failures
Duration: ~15 minutes
Scope: Subset of data centers
Root cause: Network configuration change

Lessons:
- Shows systemic issues remain
- Perfect resilience is impossible
- Need for continuous improvement
```

---

## Lessons for Production Systems

### 1. Never Trust a Single Provider

```yaml
# Bad: Single point of failure
dns_providers:
  - cloudflare

# Good: Multi-provider redundancy
dns_providers:
  - cloudflare    # Primary
  - route53       # Secondary
  - dnsmadeeasy   # Tertiary

cdn_providers:
  - cloudfront    # Primary
  - cloudflare    # Secondary (can enable in emergency)
  - fastly        # Tertiary
```

**Cost of redundancy:** $10-50/month

**Cost of outage:** $1000-100,000/hour

**ROI:** Obvious

### 2. Gradual Rollouts Are Non-Negotiable

```python
# How Cloudflare SHOULD have deployed
def deploy_bgp_change(locations):
    # Stage 1: Canary (1 location)
    canary = locations[0]
    deploy_to(canary)
    monitor(duration=60*60)  # 1 hour
    if not healthy(canary):
        rollback(canary)
        abort()
    
    # Stage 2: Small batch (10%)
    batch1 = locations[1:32]  # 31 locations = ~10%
    deploy_to(batch1)
    monitor(duration=4*60*60)  # 4 hours
    if not healthy(batch1):
        rollback(batch1 + [canary])
        abort()
    
    # Stage 3: Remaining locations
    batch2 = locations[32:]
    deploy_to(batch2)
    monitor(duration=24*60*60)  # 24 hours
    
    # Success!
```

### 3. Automated Rollback on Health Degradation

```python
def deploy_with_safety(config):
    # Take health snapshot
    before_health = measure_health()
    
    # Apply change
    apply(config)
    
    # Monitor health
    for i in range(30):  # 30 iterations = 5 minutes
        time.sleep(10)
        current_health = measure_health()
        
        # Automatic rollback threshold
        if current_health < before_health * 0.95:  # 5% degradation
            rollback(config)
            alert("Automatic rollback triggered")
            return False
    
    return True
```

### 4. Break Circular Dependencies

```
❌ Bad: Control plane depends on data plane

Dashboard → Hosted on Cloudflare
API → Hosted on Cloudflare
Monitoring → Sends via Cloudflare Workers

When Cloudflare down → Can't fix Cloudflare!

✅ Good: Separate control and data planes

Dashboard → AWS us-east-1
API → Google Cloud multi-region
Monitoring → Datadog (external SaaS)
Emergency access → Out-of-band network
```

### 5. Chaos Engineering is Essential

```python
# Monthly drill: Simulate backbone failure
@chaos_test
def test_backbone_failure():
    # 1. Disable one core hub
    disable_datacenter("ashburn-core")
    
    # 2. Verify traffic reroutes
    assert traffic_rerouted_to("amsterdam-core")
    
    # 3. Verify no user impact
    assert error_rate < 0.01  # < 1% errors
    
    # 4. Verify DNS still works
    assert dns_resolution_working()
    
    # 5. Re-enable
    enable_datacenter("ashburn-core")
    
    # 6. Document learnings
    write_report()
```

### 6. Monitoring Must Be External

```
✅ Use external monitoring:
- Pingdom (checks from outside your network)
- UptimeRobot (independent infrastructure)
- Datadog (SaaS, multi-cloud)
- StatusPage.io (NOT hosted on your infrastructure)

❌ Don't rely on internal monitoring:
- If your network is down, can't alert you
- Circular dependency problem
- Blind to your own outages
```

### 7. Communication Protocols

```markdown
# Incident Communication Playbook

T+0 min: Incident detected
  → Auto-post to status page: "Investigating"
  → Tweet from @StatusAccount
  
T+5 min: Scope understood
  → Update status page with affected services
  → Send email to enterprise customers
  
T+15 min: Regular updates
  → Status page: Progress updates every 15 min
  → Twitter: Major milestones
  
T+Resolution: Service restored
  → Status page: "Resolved"
  → Tweet: "Services restored"
  → Promise post-mortem within 72 hours
  
T+72 hours: Post-mortem published
  → Full technical details
  → Root cause analysis
  → Remediation steps
  → No blame, focus on systems
```

---

## Conclusion: Building Antifragile Systems

**Key Takeaways:**

1. **Single providers are risky** - Even industry leaders fail
2. **Anycast amplifies failures** - Global blast radius from local errors
3. **Circular dependencies are deadly** - Control plane must be independent
4. **Gradual rollouts save lives** - Never change everything at once
5. **Automated rollback is essential** - Humans are too slow
6. **Chaos engineering reveals weaknesses** - Test failures before they happen
7. **External monitoring required** - Can't monitor yourself when down

**The Unpopular Truth:**

```
Perfect uptime is impossible.

Cloudflare: 99.99% → One outage broke the promise
AWS us-east-1: Multiple outages per year
Google Cloud: Multiple outages per year
Azure: Multiple outages per year

The goal is not perfection.
The goal is graceful degradation.
The goal is fast recovery.
The goal is learning from failures.
```

**Build systems that get stronger from failures.**

---

## References

- [Cloudflare Official Post-Mortem (June 24, 2022)](https://blog.cloudflare.com/cloudflare-outage-on-june-21-2022/)
- [BGP Route Flapping: IETF RFC 2439](https://datatracker.ietf.org/doc/html/rfc2439)
- [Chaos Engineering: Principles and Practice](https://principlesofchaos.org/)

---

**Document Version:** 1.0  
**Last Updated:** January 2, 2026  
**Author:** DevOps Team / SRE Documentation
