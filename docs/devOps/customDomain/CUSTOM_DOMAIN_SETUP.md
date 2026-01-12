# Custom Domain Setup Guide

## 🎯 Quick Start

Enable custom domain in 3 steps:

```bash
# 1. Set your domain name
pulumi config set domain_name yourdomain.com

# 2. Deploy infrastructure
pulumi up

# 3. Wait for certificate validation (~5-10 minutes)
# Then test: curl https://yourdomain.com
```

---

## 📋 Prerequisites

### 1. Domain Name
- You own a domain (e.g., `sunlink.ai`)
- Domain can be from any registrar (GoDaddy, Namecheap, Google Domains, etc.)

### 2. Route53 Hosted Zone
- **Option A:** Already have Route53 hosted zone → Nothing to do ✅
- **Option B:** Need to create hosted zone:
  ```bash
  aws route53 create-hosted-zone --name yourdomain.com --caller-reference $(date +%s)
  ```
- **Option C:** Domain NOT in Route53 → Update nameservers at your registrar

### 3. AWS Permissions
Ensure your AWS credentials have permissions for:
- ACM (Certificate Manager)
- CloudFront
- Route53
- S3

---

## 🚀 Step-by-Step Setup

### Step 1: Configure Domain in Pulumi

```bash
cd pulumi
pulumi config set domain_name yourdomain.com
```

**Verify configuration:**
```bash
pulumi config
```

Expected output:
```
KEY                  VALUE
domain_name          yourdomain.com
enable_cloudfront... false
price_class         PriceClass_100
```

### Step 2: Deploy Infrastructure

```bash
pulumi up
```

**What happens:**
1. Creates ACM certificate in us-east-1
2. Requests DNS validation
3. Creates CloudFront distribution with custom domain
4. Creates Route53 A record pointing to CloudFront
5. Uploads static files to S3

**Expected output:**
```
Updating (dev)

     Type                        Name                                Status
 +   ├─ aws:acm:Certificate     sunlink-dev-cert                    created
 +   ├─ aws:route53:Record      sunlink-dev-dns                     created
 ~   └─ aws:cloudfront:Dist...  sunlink-dev-cdn                     updated

Outputs:
    cloudFrontUrl      : "https://yourdomain.com"
    certificateArn     : "arn:aws:acm:us-east-1:..."
    customDomainUrl    : "https://yourdomain.com"

Resources:
    + 2 created
    ~ 1 updated
    17 unchanged

Duration: 45s
```

### Step 3: Validate ACM Certificate

**If using Route53 (Automatic):**
- Certificate validation happens automatically
- Wait ~5-10 minutes for validation to complete
- No action needed! ✅

**If NOT using Route53 (Manual):**
1. Go to AWS Console → Certificate Manager → us-east-1
2. Click on your certificate
3. Copy the CNAME record name and value
4. Add CNAME record to your DNS provider:
   ```
   Name:  _abc123.yourdomain.com
   Type:  CNAME
   Value: _xyz456.acm-validations.aws.
   ```
5. Wait for validation (5-30 minutes)

**Check validation status:**
```bash
aws acm describe-certificate \
  --certificate-arn $(pulumi stack output certificateArn) \
  --region us-east-1 \
  --query 'Certificate.Status'
```

Expected: `"ISSUED"`

### Step 4: Wait for CloudFront Distribution

CloudFront takes ~15-20 minutes to deploy globally.

**Check distribution status:**
```bash
aws cloudfront get-distribution \
  --id $(pulumi stack output cloudFrontId) \
  --query 'Distribution.Status'
```

Expected: `"Deployed"`

### Step 5: Test Custom Domain

```bash
# Test HTTPS connection
curl -I https://yourdomain.com

# Expected response:
HTTP/2 200
server: CloudFront
content-type: text/html
...
```

**Or open in browser:**
```
https://yourdomain.com
```

---

## 🔧 Troubleshooting

### Issue 1: Certificate Stuck in "Pending Validation"

**Cause:** DNS validation records not found

**Solution:**
```bash
# Check DNS propagation
dig _abc123.yourdomain.com CNAME

# If using Route53, verify record exists
aws route53 list-resource-record-sets \
  --hosted-zone-id $(pulumi stack output hostedZoneId)
```

### Issue 2: "The request could not be satisfied" Error

**Cause:** CloudFront distribution still deploying

**Solution:** Wait 15-20 minutes and try again

### Issue 3: DNS Resolution Fails

**Cause:** Nameservers not updated at registrar

**Solution:**
```bash
# Get Route53 nameservers
aws route53 get-hosted-zone --id $(pulumi stack output hostedZoneId) \
  --query 'DelegationSet.NameServers'

# Update nameservers at your domain registrar:
# ns-123.awsdns-12.com
# ns-456.awsdns-23.net
# ns-789.awsdns-34.org
# ns-012.awsdns-45.co.uk
```

### Issue 4: certificate.arn is undefined

**Cause:** Domain name not configured

**Solution:**
```bash
pulumi config set domain_name yourdomain.com
pulumi up
```

---

## 🔄 Reverting to Default CloudFront Domain

To remove custom domain and use CloudFront default:

```bash
# Remove domain configuration
pulumi config rm domain_name

# Redeploy
pulumi up
```

**What changes:**
- Certificate deleted
- Route53 record deleted
- CloudFront uses default domain (`xxxxx.cloudfront.net`)
- **Saves $0.50/month** (no Route53 hosted zone cost)

---

## 📊 Verification Checklist

After deployment, verify:

- [ ] ACM certificate status is `ISSUED`
- [ ] CloudFront distribution status is `Deployed`  
- [ ] Route53 A record exists and points to CloudFront
- [ ] `curl https://yourdomain.com` returns 200 OK
- [ ] Browser shows valid SSL certificate
- [ ] No SSL warnings in browser
- [ ] `pulumi stack output cloudFrontUrl` shows custom domain

---

## 💰 Cost Reminder

**Custom domain adds:**
- ✅ FREE ACM SSL certificate
- 💰 $0.50/month Route53 hosted zone
- ✅ No additional CloudFront costs

**Total additional cost: ~$0.50/month**

See [COST_ANALYSIS.md](./COST_ANALYSIS.md) for detailed breakdown.

---

## 📚 Additional Configuration

### Add Subdomain

To use `www.yourdomain.com` or `app.yourdomain.com`:

```bash
# Update domain configuration
pulumi config set domain_name app.yourdomain.com

# Redeploy
pulumi up
```

### Multiple Subdomains

```typescript
// In index.ts, update aliases:
aliases: domainName ? [domainName, `www.${domainName}`] : [],
```

### Add WWW Redirect

Create a separate CloudFront distribution that redirects `www` to apex domain, or use Route53 alias records.

---

## 🎉 Success!

Your custom domain is now live with:
- ✅ HTTPS enabled (SSL/TLS)
- ✅ Global CDN (CloudFront)
- ✅ Fast performance
- ✅ Professional domain

**Next steps:**
- Update DNS records to point to CloudFront
- Configure custom error pages
- Set up monitoring with CloudWatch
- Configure CloudFront Functions for advanced routing
