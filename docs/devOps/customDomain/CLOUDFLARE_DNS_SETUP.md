# CloudFront with Cloudflare DNS Setup Guide

## 🎯 Overview

This guide shows how to use **Cloudflare DNS** with **AWS CloudFront** and **ACM certificates**, combining:
- ✅ AWS ACM for FREE SSL certificates
- ✅ Cloudflare for FREE DNS hosting + DDoS protection
- ✅ CloudFront for global CDN

**Cost:** $0/month for DNS (saves $0.50/month vs Route53!)

---

## 📋 Prerequisites

1. **Domain name** owned by you
2. **Cloudflare account** (free tier is fine)
3. **Domain added to Cloudflare** with active nameservers
4. **AWS account** with Pulumi configured

---

## 🚀 Step-by-Step Setup

### Step 1: Update Pulumi Code

Add ACM certificate to `pulumi/index.ts`:

```typescript
// After the tags section (around line 22)

// Create us-east-1 provider for ACM certificate (required for CloudFront)
const useast1Provider = new aws.Provider("useast1", { region: "us-east-1" });

// Create ACM certificate for custom domain
const certificate = domainName ? new aws.acm.Certificate(
    `${projectName}-${stackName}-cert`,
    {
        domainName: domainName,
        validationMethod: "DNS",
        tags,
    },
    { provider: useast1Provider }
) : undefined;
```

Update CloudFront viewer certificate (around line 125):

```typescript
viewerCertificate: domainName && certificate ? {
    acmCertificateArn: certificate.arn,
    sslSupportMethod: "sni-only",
    minimumProtocolVersion: "TLSv1.2_2021",
} : {
    cloudfrontDefaultCertificate: true,
},
```

Add certificate export (around line 220):

```typescript
export const certificateArn = certificate?.arn;
export const cloudFrontUrl = domainName 
    ? `https://${domainName}` 
    : pulumi.interpolate`https://${cloudFrontDistribution.domainName}`;
```

### Step 2: Deploy Initial Infrastructure

```bash
# Set your domain
pulumi config set domain_name yourdomain.com

# Deploy to create ACM certificate
pulumi up
```

**Output:**
```
+ aws:acm:Certificate sunlink-dev-cert creating
~ aws:cloudfront:Distribution sunlink-dev-cdn updating
```

### Step 3: Get ACM Validation Records

**Option A: Using AWS CLI**
```bash
# Get certificate ARN
CERT_ARN=$(pulumi stack output certificateArn -s dev)

# Get validation records
aws acm describe-certificate \
  --certificate-arn "$CERT_ARN" \
  --region us-east-1 \
  --query 'Certificate.DomainValidationOptions[0].ResourceRecord'
```

**Output:**
```json
{
    "Name": "_abc123def456.yourdomain.com.",
    "Type": "CNAME",
    "Value": "_xyz789ghi012.acm-validations.aws."
}
```

**Option B: Using AWS Console**
1. Go to **AWS Console** → **Certificate Manager** → **us-east-1**
2. Click on your certificate
3. Copy the CNAME record details

### Step 4: Add Validation Record to Cloudflare

1. **Log in to Cloudflare** → Select your domain
2. **Go to DNS** → **Records**
3. **Click "Add record":**
   ```
   Type: CNAME
   Name: _abc123def456
   Target: _xyz789ghi012.acm-validations.aws.
   Proxy status: DNS only (gray cloud) ⚠️ IMPORTANT
   TTL: Auto
   ```
4. **Click "Save"**

**⚠️ Critical:** Make sure proxy is **OFF** (gray cloud, not orange)

### Step 5: Wait for Certificate Validation

```bash
# Check validation status
aws acm describe-certificate \
  --certificate-arn "$CERT_ARN" \
  --region us-east-1 \
  --query 'Certificate.Status'
```

**Expected:** `"ISSUED"` (takes 5-30 minutes)

**Monitor in real-time:**
```bash
# Watch until status is ISSUED
watch -n 10 'aws acm describe-certificate --certificate-arn $(pulumi stack output certificateArn) --region us-east-1 --query Certificate.Status'
```

### Step 6: Get CloudFront Domain

```bash
# Get CloudFront distribution domain
pulumi stack output cloudFrontDomain
```

**Output:** `d1234abcd5678.cloudfront.net`

### Step 7: Add CNAME to Cloudflare

Back in **Cloudflare DNS**:

1. **Click "Add record":**
   ```
   Type: CNAME
   Name: @ (for apex) or www (for subdomain)
   Target: d1234abcd5678.cloudfront.net
   Proxy status: DNS only (gray cloud) ⚠️ IMPORTANT
   TTL: Auto
   ```
2. **Click "Save"**

**⚠️ Critical:** Proxy must be **OFF** for CloudFront to work

### Step 8: Wait for CloudFront Deployment

```bash
# Check CloudFront status
aws cloudfront get-distribution \
  --id $(pulumi stack output cloudFrontId) \
  --query 'Distribution.Status'
```

**Expected:** `"Deployed"` (takes 15-20 minutes)

### Step 9: Test Your Domain

```bash
# Test HTTPS
curl -I https://yourdomain.com

# Expected response:
HTTP/2 200
server: CloudFront
x-cache: Hit from cloudfront
```

**Or open in browser:** `https://yourdomain.com`

✅ Should show valid SSL certificate!

---

## 🔧 Important Cloudflare Settings

### SSL/TLS Mode

**Cloudflare Dashboard** → **SSL/TLS** → **Overview**

Set to: **Full (strict)** ✅

| Mode | Description | CloudFront Compatible? |
|------|-------------|----------------------|
| Off | No encryption | ❌ Not recommended |
| Flexible | CF↔User encrypted only | ❌ Won't work |
| Full | CF↔CF, CF↔Origin encrypted | ⚠️ Works but insecure |
| **Full (strict)** | Valid cert required | ✅ **Recommended** |

### Proxy Status

For both DNS records, ensure:
- ☁️ **Orange cloud (proxied):** ❌ **Will break CloudFront**
- ☁️ **Gray cloud (DNS only):** ✅ **Correct**

**Why?** CloudFront needs direct connection, not proxied through Cloudflare.

### Page Rules (Optional)

Add these for better performance:

**Cloudflare Dashboard** → **Rules** → **Page Rules**

1. **Always Use HTTPS**
   ```
   URL: http://*yourdomain.com/*
   Setting: Always Use HTTPS
   ```

2. **Cache Everything** (optional for static sites)
   ```
   URL: yourdomain.com/*
   Settings:
     - Cache Level: Cache Everything
     - Edge Cache TTL: 1 month
   ```

---

## 💰 Cost Comparison

### With Cloudflare DNS (This Setup)

| Service | Cost | Details |
|---------|------|---------|
| **S3 Storage** | ~$0.02/month | 50 MB static files |
| **CloudFront** | FREE | First 1 TB (12 months) |
| **ACM Certificate** | FREE | SSL certificate |
| **Cloudflare DNS** | FREE | Unlimited queries |
| **Cloudflare DDoS** | FREE | Unlimited protection |

**Total: ~$0.02/month** (essentially FREE!)

### With Route53 DNS

| Service | Cost | Details |
|---------|------|---------|
| **S3 Storage** | ~$0.02/month | 50 MB static files |
| **CloudFront** | FREE | First 1 TB (12 months) |
| **ACM Certificate** | FREE | SSL certificate |
| **Route53 Zone** | $0.50/month | Hosted zone |
| **Route53 Queries** | FREE | First 1B queries |

**Total: ~$0.52/month**

**Savings with Cloudflare: $0.50/month = $6/year** 💰

---

## 🚨 Troubleshooting

### Issue 1: Certificate Stuck in "Pending Validation"

**Symptom:** Certificate status remains `PENDING_VALIDATION` for >30 minutes

**Solutions:**

1. **Check DNS propagation:**
   ```bash
   dig _abc123def456.yourdomain.com CNAME
   
   # Should return the acm-validations.aws record
   ```

2. **Verify Cloudflare record:**
   - Proxy status is **OFF** (gray cloud)
   - Record name is exact (including subdomain parts)
   - No trailing dots in Cloudflare

3. **Check Cloudflare Universal SSL:**
   - **Cloudflare** → **SSL/TLS** → **Edge Certificates**
   - Ensure "Universal SSL" is enabled

### Issue 2: "This site can't be reached"

**Symptom:** Domain doesn't resolve

**Solutions:**

1. **Check CNAME exists:**
   ```bash
   dig yourdomain.com CNAME
   
   # Should return CloudFront domain
   ```

2. **Verify proxy is OFF:**
   - Cloudflare DNS record must be gray cloud
   - Click record → Toggle off proxy

3. **Check DNS propagation:**
   ```bash
   # Use Google DNS
   dig @8.8.8.8 yourdomain.com
   ```

### Issue 3: SSL Certificate Error in Browser

**Symptom:** "Your connection is not private" or "NET::ERR_CERT_AUTHORITY_INVALID"

**Solutions:**

1. **Check ACM certificate status:**
   ```bash
   aws acm describe-certificate \
     --certificate-arn $(pulumi stack output certificateArn) \
     --region us-east-1 \
     --query 'Certificate.Status'
   ```
   Should be: `"ISSUED"`

2. **Verify CloudFront is using certificate:**
   ```bash
   aws cloudfront get-distribution \
     --id $(pulumi stack output cloudFrontId) \
     --query 'Distribution.DistributionConfig.ViewerCertificate'
   ```

3. **Check domain matches:**
   - Certificate domain: `yourdomain.com`
   - Cloudflare CNAME: `yourdomain.com`
   - Must match exactly!

### Issue 4: "Too Many Redirects"

**Symptom:** Browser shows redirect loop error

**Solution:**

Set Cloudflare SSL mode to **Full (strict)**:
- **Cloudflare** → **SSL/TLS** → **Overview**
- Select "Full (strict)"

### Issue 5: CloudFront Distribution Not Updating

**Symptom:** Old content still showing after deployment

**Solutions:**

1. **Invalidate CloudFront cache:**
   ```bash
   aws cloudfront create-invalidation \
     --distribution-id $(pulumi stack output cloudFrontId) \
     --paths "/*"
   ```

2. **Check Cloudflare cache:**
   - **Cloudflare** → **Caching** → **Configuration**
   - Click "Purge Everything"

---

## ✅ Verification Checklist

After setup, verify:

- [ ] ACM certificate status is `ISSUED`
- [ ] CloudFront distribution status is `Deployed`
- [ ] Cloudflare CNAME record exists (gray cloud)
- [ ] Cloudflare SSL mode is "Full (strict)"
- [ ] `curl https://yourdomain.com` returns 200 OK
- [ ] Browser shows valid SSL certificate (green padlock)
- [ ] No SSL warnings or errors
- [ ] Page loads correctly

---

## 🔄 Updating/Removing Domain

### Change Domain

```bash
# Update domain
pulumi config set domain_name newdomain.com

# Redeploy
pulumi up

# Repeat steps 3-9 for new domain
```

### Remove Custom Domain

```bash
# Remove domain config
pulumi config rm domain_name

# Redeploy
pulumi up
```

CloudFront will revert to default domain (`xxxxx.cloudfront.net`).

---

## 🌟 Cloudflare Bonus Features (FREE)

With Cloudflare, you also get:

1. **Analytics Dashboard**
   - Traffic stats
   - Threat analytics
   - Performance insights

2. **DDoS Protection**
   - Unlimited mitigation
   - Automatic protection
   - No extra cost

3. **Web Application Firewall (WAF)**
   - Free tier available
   - Block malicious traffic
   - Custom rules

4. **Email Routing**
   - Forward emails to Gmail
   - No email hosting needed

5. **Page Rules**
   - URL redirects
   - Custom caching
   - Security headers

---

## 📊 Performance Comparison

### Cloudflare Proxy OFF (This Setup)

✅ **Best for:** Static sites with CloudFront
- Origin: S3 via CloudFront edge
- SSL: ACM (AWS-managed)
- DDoS: CloudFront (limited free tier)
- Cost: ~$0.02/month

### Cloudflare Proxy ON (Alternative)

✅ **Best for:** Dynamic sites, better free DDoS
- Origin: S3 via Cloudflare edge → CloudFront
- SSL: Cloudflare-managed
- DDoS: Cloudflare (unlimited free)
- Cost: FREE

**Note:** Proxy ON requires different CloudFront setup (origin certificate).

---

## 🎉 Summary

**You now have:**
- ✅ Custom domain with FREE SSL
- ✅ FREE Cloudflare DNS (saves $6/year)
- ✅ FREE DDoS protection
- ✅ Global CDN via CloudFront
- ✅ Professional infrastructure

**Total cost: ~$0.02/month (essentially FREE!)**

**Next steps:**
- Configure Cloudflare analytics
- Set up email routing
- Enable Cloudflare security features
- Monitor with CloudWatch
