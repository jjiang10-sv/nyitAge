# Custom Domain Cost Analysis

## 💰 AWS Costs Breakdown

### Current Setup (CloudFront Default Domain)

| Service | Cost | Details |
|---------|------|---------|
| **S3 Storage** | ~$0.023/GB/month | Standard storage for static files |
| **S3 Requests** | ~$0.005/1000 PUT | File uploads during deployment |
| **CloudFront** | First 1 TB FREE/month | Data transfer out (Free Tier first year) |
| **CloudFront** | $0.085/GB after 1TB | Additional data transfer |
| **CloudFront Requests** | $0.0075/10,000 | HTTP/HTTPS requests |
| **Route53 (optional)** | $0 | Not used with default domain |
| **ACM Certificate** | $0 | Not used with default domain |

**Estimated Monthly Cost (Low Traffic):** ~$1-5/month

---

### With Custom Domain Setup

| Service | Cost | Details |
|---------|------|---------|
| **S3 Storage** | ~$0.023/GB/month | Standard storage for static files |
| **S3 Requests** | ~$0.005/1000 PUT | File uploads during deployment |
| **CloudFront** | First 1 TB FREE/month | Data transfer out (Free Tier first year) |
| **CloudFront** | $0.085/GB after 1TB | Additional data transfer |
| **CloudFront Requests** | $0.0075/10,000 | HTTP/HTTPS requests |
| **Route53 Hosted Zone** | **$0.50/month** | ⚠️ NEW COST |
| **Route53 Queries** | $0.40/million | Standard queries (first 1B free) |
| **ACM Certificate** | **FREE** ✅ | Public SSL certificates are free! |

**Estimated Monthly Cost (Low Traffic):** ~$1.50-5.50/month

**Additional Cost:** **$0.50/month** (Route53 hosted zone only)

---

## 📊 Cost Scenarios

### Scenario 1: Personal Website / Low Traffic
```
Traffic: 10,000 page views/month
Data Transfer: ~5 GB/month
Files: 50 MB

S3 Storage:        $0.05 × 0.05 GB = $0.00
S3 Requests:       Negligible
CloudFront Data:   FREE (under 1 TB)
CloudFront Req:    $0.0075 × 1 = $0.01
Route53 Zone:      $0.50
Route53 Queries:   FREE

Total: ~$0.51/month
Additional with custom domain: +$0.50/month
```

### Scenario 2: Small Business / Medium Traffic
```
Traffic: 100,000 page views/month
Data Transfer: ~50 GB/month
Files: 200 MB

S3 Storage:        $0.023 × 0.2 GB = $0.00
S3 Requests:       Negligible
CloudFront Data:   FREE (under 1 TB)
CloudFront Req:    $0.0075 × 10 = $0.08
Route53 Zone:      $0.50
Route53 Queries:   FREE

Total: ~$0.58/month
Additional with custom domain: +$0.50/month
```

### Scenario 3: Growing App / High Traffic
```
Traffic: 1,000,000 page views/month
Data Transfer: ~500 GB/month
Files: 500 MB

S3 Storage:        $0.023 × 0.5 GB = $0.01
S3 Requests:       Negligible
CloudFront Data:   FREE (under 1 TB)
CloudFront Req:    $0.0075 × 100 = $0.75
Route53 Zone:      $0.50
Route53 Queries:   FREE

Total: ~$1.26/month
Additional with custom domain: +$0.50/month
```

### Scenario 4: Enterprise / Very High Traffic
```
Traffic: 10,000,000 page views/month
Data Transfer: ~2 TB/month
Files: 1 GB

S3 Storage:        $0.023 × 1 GB = $0.02
S3 Requests:       $0.01
CloudFront Data:   1 TB FREE + 1 TB × $0.085 = $85.00
CloudFront Req:    $0.0075 × 1000 = $7.50
Route53 Zone:      $0.50
Route53 Queries:   FREE

Total: ~$93.03/month
Additional with custom domain: +$0.50/month
```

---

## 🎯 Key Insights

### What's FREE
- ✅ **ACM SSL Certificates** - Public certificates are completely free
- ✅ **First 1 TB CloudFront data transfer** - Free tier for the first 12 months
- ✅ **Route53 queries** - First 1 billion queries/month

### What Costs Money
- 💵 **Route53 Hosted Zone** - $0.50/month (the main additional cost)
- 💵 **CloudFront data over 1 TB** - $0.085/GB (only for high traffic)
- 💵 **S3 storage** - ~$0.023/GB (minimal for static sites)

### Cost Comparison: Custom Domain vs Default

| Traffic Level | Default Domain | Custom Domain | Difference |
|--------------|----------------|---------------|------------|
| **Low** (10K views) | $0.01/month | $0.51/month | +$0.50 |
| **Medium** (100K views) | $0.08/month | $0.58/month | +$0.50 |
| **High** (1M views) | $0.76/month | $1.26/month | +$0.50 |
| **Enterprise** (10M views) | $92.53/month | $93.03/month | +$0.50 |

**Conclusion:** Custom domain adds a **flat $0.50/month** regardless of traffic level.

---

## 💡 Cost Optimization Tips

### 1. Use CloudFront Regional Edge Caches
Already configured with `PriceClass_100` (North America + Europe)
- Saves ~40% vs `PriceClass_All` (global)

### 2. Enable Compression
Already enabled: `compress: true`
- Reduces bandwidth by 60-80% for text files

### 3. Optimize Cache TTL
Current: 1 hour default, 24 hour max
- Reduces origin requests
- Lowers CloudFront request costs

### 4. Use S3 Intelligent-Tiering (for larger apps)
```typescript
storageClass: "INTELLIGENT_TIERING"
```
- Automatically moves infrequently accessed files to cheaper tiers
- Saves 40-70% on storage costs

### 5. Minimize File Count
- Bundle JavaScript/CSS files
- Use image sprites where possible
- Reduces S3 request costs

---

## 📅 Annual Cost Comparison

### Personal Website (Low Traffic)
- **Without custom domain:** $0.12/year
- **With custom domain:** $6.12/year
- **Additional:** $6/year

### Small Business (Medium Traffic)
- **Without custom domain:** $0.96/year  
- **With custom domain:** $6.96/year
- **Additional:** $6/year

### Enterprise (High Traffic)
- **Without custom domain:** $1,110/year
- **With custom domain:** $1,116/year
- **Additional:** $6/year

---

## ✅ Recommendation

**For most users, the custom domain is worth it:**
- Professional appearance
- Brand recognition
- SEO benefits
- Only **$0.50/month** (~$6/year) additional cost
- FREE SSL certificate included

**Cost is so low that it's a no-brainer for:**
- ✅ Business websites
- ✅ Portfolio sites
- ✅ SaaS applications
- ✅ Any production application

**Consider default domain only if:**
- ❌ Testing/development environment
- ❌ Internal tools
- ❌ Temporary projects

---

## 🔧 Setup Steps to Enable Custom Domain

1. **Set domain in Pulumi config:**
   ```bash
   pulumi config set domain_name yourdomain.com
   ```

2. **Ensure Route53 hosted zone exists:**
   - Either create manually in AWS Console
   - Or let Pulumi create it (add to index.ts)

3. **Deploy infrastructure:**
   ```bash
   pulumi up
   ```

4. **Verify certificate:**
   - ACM will create DNS validation records
   - Add CNAME records to your DNS (automatic with Route53)
   - Wait ~5-10 minutes for validation

5. **Test custom domain:**
   ```bash
   curl https://yourdomain.com
   ```

---

## 📊 Cost Monitoring

### CloudWatch Metrics (Free)
Monitor your actual costs with CloudWatch:
- CloudFront requests
- Data transfer
- Cache hit ratio

### AWS Cost Explorer
Set up budget alerts:
```bash
# Get notified if costs exceed $10/month
aws budgets create-budget --budget file://budget.json
```

### Pulumi Stack Tagging
All resources tagged with:
- `Project: sunlink-bright-dashboard`
- `Stack: dev` or `production`  
- `ManagedBy: Pulumi`

Use tags in Cost Explorer to track costs per environment.

---

## 🎉 Summary

**Custom domain setup adds:**
- ✅ FREE ACM SSL certificate
- ✅ Route53 DNS management
- ✅ Professional domain name
- 💰 **Only $0.50/month additional cost**

**Total infrastructure cost for typical usage:**
- Development: ~$1-2/month
- Production (low-medium traffic): ~$2-10/month
- Production (high traffic): ~$50-200/month

**The custom domain is negligible compared to total costs!**
