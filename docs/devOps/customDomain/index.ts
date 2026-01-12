import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import * as fs from "fs";
import * as path from "path";
import * as mime from "mime-types";

// Get configuration
const config = new pulumi.Config();
const projectName = pulumi.getProject();
const stackName = pulumi.getStack();

// Configuration options
const domainName = config.get("domain_name") || "";
const enableCloudFrontLogging = config.getBoolean("enable_cloudfront_logging") || false;
const priceClass = config.get("price_class") || "PriceClass_100";

// Tags for all resources
const tags = {
    Project: projectName,
    Stack: stackName,
    ManagedBy: "Pulumi",
};

// Create us-east-1 provider for ACM certificate (required for CloudFront)
const useast1Provider = new aws.Provider("useast1", { region: "us-east-1" });

// Create ACM certificate for custom domain (if configured)
const certificate = domainName ? new aws.acm.Certificate(
    `${projectName}-${stackName}-cert`,
    {
        domainName: domainName,
        validationMethod: "DNS",
        tags,
    },
    { provider: useast1Provider }
) : undefined;

// Get Route53 hosted zone (if custom domain configured)
const hostedZone = domainName ? aws.route53.getZoneOutput({
    name: domainName,
    privateZone: false,
}) : undefined;

// Create S3 bucket for website hosting
const bucket = new aws.s3.BucketV2(`${projectName}-${stackName}-bucket`, {
    tags,
});

// Configure bucket ownership controls
const ownershipControls = new aws.s3.BucketOwnershipControls(
    `${projectName}-${stackName}-ownership`,
    {
        bucket: bucket.id,
        rule: {
            objectOwnership: "BucketOwnerPreferred",
        },
    }
);

// Block public ACLs but allow CloudFront access via bucket policy
const publicAccessBlock = new aws.s3.BucketPublicAccessBlock(
    `${projectName}-${stackName}-public-access-block`,
    {
        bucket: bucket.id,
        blockPublicAcls: true,
        blockPublicPolicy: false, // Allow bucket policy for CloudFront
        ignorePublicAcls: true,
        restrictPublicBuckets: false,
    }
);

// Create CloudFront Origin Access Control (OAC)
const originAccessControl = new aws.cloudfront.OriginAccessControl(
    `${projectName}-${stackName}-oac`,
    {
        description: `OAC for ${projectName} ${stackName}`,
        originAccessControlOriginType: "s3",
        signingBehavior: "always",
        signingProtocol: "sigv4",
    }
);

// Create CloudFront distribution
const cloudFrontDistribution = new aws.cloudfront.Distribution(
    `${projectName}-${stackName}-cdn`,
    {
        enabled: true,
        isIpv6Enabled: true,
        comment: `${projectName} ${stackName} distribution`,
        defaultRootObject: "index.html",
        priceClass,

        // Origin configuration
        origins: [
            {
                domainName: bucket.bucketRegionalDomainName,
                originId: bucket.id,
                originAccessControlId: originAccessControl.id,
            },
        ],

        // Default cache behavior
        defaultCacheBehavior: {
            allowedMethods: ["GET", "HEAD", "OPTIONS"],
            cachedMethods: ["GET", "HEAD"],
            targetOriginId: bucket.id,
            viewerProtocolPolicy: "redirect-to-https",
            compress: true,

            forwardedValues: {
                queryString: false,
                cookies: {
                    forward: "none",
                },
            },

            minTtl: 0,
            defaultTtl: 3600, // 1 hour
            maxTtl: 86400, // 24 hours
        },

        // Custom error responses for SPA routing
        customErrorResponses: [
            {
                errorCode: 404,
                responseCode: 200,
                responsePagePath: "/index.html",
                errorCachingMinTtl: 10,
            },
            {
                errorCode: 403,
                responseCode: 200,
                responsePagePath: "/index.html",
                errorCachingMinTtl: 10,
            },
        ],

        // Restrictions
        restrictions: {
            geoRestriction: {
                restrictionType: "none",
            },
        },

        // SSL/TLS configuration
        viewerCertificate: domainName && certificate
            ? {
                acmCertificateArn: certificate.arn,
                sslSupportMethod: "sni-only",
                minimumProtocolVersion: "TLSv1.2_2021",
            }
            : {
                cloudfrontDefaultCertificate: true,
            },

        // Optional: Aliases for custom domain
        aliases: domainName ? [domainName] : [],

        tags,
    }
);

// S3 bucket policy to allow CloudFront OAC access
const bucketPolicy = new aws.s3.BucketPolicy(
    `${projectName}-${stackName}-bucket-policy`,
    {
        bucket: bucket.id,
        policy: pulumi
            .all([bucket.arn, cloudFrontDistribution.arn])
            .apply(([bucketArn, distributionArn]) =>
                JSON.stringify({
                    Version: "2012-10-17",
                    Statement: [
                        {
                            Sid: "AllowCloudFrontServicePrincipal",
                            Effect: "Allow",
                            Principal: {
                                Service: "cloudfront.amazonaws.com",
                            },
                            Action: "s3:GetObject",
                            Resource: `${bucketArn}/*`,
                            Condition: {
                                StringEquals: {
                                    "AWS:SourceArn": distributionArn,
                                },
                            },
                        },
                    ],
                })
            ),
    },
    { dependsOn: [publicAccessBlock] }
);

// Upload website files from dist folder
const distDir = path.join(__dirname, "..", "dist");

function uploadDirectory(dirPath: string, prefix: string = ""): void {
    if (!fs.existsSync(dirPath)) {
        console.warn(`Dist directory not found at ${dirPath}. Run 'npm run build' first.`);
        return;
    }

    const items = fs.readdirSync(dirPath);

    for (const item of items) {
        const itemPath = path.join(dirPath, item);
        const relativePath = prefix ? `${prefix}/${item}` : item;
        const stat = fs.statSync(itemPath);

        if (stat.isFile()) {
            const contentType = mime.lookup(itemPath) || "application/octet-stream";

            new aws.s3.BucketObject(
                relativePath.replace(/\//g, "-"),
                {
                    bucket: bucket.id,
                    key: relativePath,
                    source: new pulumi.asset.FileAsset(itemPath),
                    contentType,
                },
                { dependsOn: [bucketPolicy] }
            );
        } else if (stat.isDirectory()) {
            uploadDirectory(itemPath, relativePath);
        }
    }
}

// Upload files
uploadDirectory(distDir);

// Create Route53 DNS record for custom domain (if configured)
const dnsRecord = domainName && hostedZone ? new aws.route53.Record(
    `${projectName}-${stackName}-dns`,
    {
        zoneId: hostedZone.zoneId,
        name: domainName,
        type: "A",
        aliases: [{
            name: cloudFrontDistribution.domainName,
            zoneId: cloudFrontDistribution.hostedZoneId,
            evaluateTargetHealth: false,
        }],
    }
) : undefined;

// Exports
export const bucketName = bucket.id;
export const bucketArn = bucket.arn;
export const cloudFrontId = cloudFrontDistribution.id;
export const cloudFrontDomain = cloudFrontDistribution.domainName;
export const cloudFrontUrl = domainName
    ? `https://${domainName}`
    : pulumi.interpolate`https://${cloudFrontDistribution.domainName}`;
export const customDomainUrl = domainName ? `https://${domainName}` : undefined;
export const certificateArn = certificate?.arn;
