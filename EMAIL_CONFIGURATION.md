# Email Configuration & Customization

## Overview

AWS Cognito handles authentication emails (verification, password reset). This guide covers customizing these templates and configuring email delivery.

## Email Types

1. **Verification Email** - Sent when users register
2. **Password Reset Email** - Sent when users request password reset
3. **Change Email Verification** - Sent when users change their email (if enabled)
4. **Welcome Email** - Custom email after successful signup (optional)

## AWS Cognito Email Configuration

### Option 1: Amazon SES (Production)

For production, use Amazon SES for better deliverability and custom domains.

#### Setup Steps

1. **Verify Your Domain in SES**

```bash
# AWS Console → SES → Verified Identities → Create Identity
# Choose "Domain" and enter your domain (e.g., qrdragonfly.com)
```

2. **Configure DNS Records**

Add the TXT, CNAME, and MX records provided by SES to your DNS provider.

3. **Update Cognito User Pool**

```bash
aws cognito-idp update-user-pool \
  --user-pool-id YOUR_USER_POOL_ID \
  --email-configuration \
    SourceArn=arn:aws:ses:REGION:ACCOUNT_ID:identity/your-domain.com \
    ReplyToEmailAddress=support@your-domain.com \
    EmailSendingAccount=DEVELOPER \
    From=QR Dragonfly <noreply@your-domain.com>
```

**Benefits:**

- Custom "From" address (noreply@yourdomain.com)
- Better deliverability
- Higher sending limits (50,000 emails/day vs 50/day)
- Custom reply-to address

### Option 2: Cognito Default (Development)

Free tier includes 50 emails/day. Suitable for development only.

**Limitations:**

- Generic "From" address (no-reply@verificationemail.com)
- 50 emails per day limit
- Lower deliverability
- Cannot customize sender name

## Customizing Email Templates

### Via AWS Console

1. Go to **Cognito Console** → Your User Pool → **Messaging** → **Email**
2. Choose **Verification type**: Code or Link
3. Customize templates

### Via AWS CLI

```bash
# Update verification email template
aws cognito-idp update-user-pool \
  --user-pool-id YOUR_USER_POOL_ID \
  --verification-message-template \
    'DefaultEmailOption=CONFIRM_WITH_CODE,
     EmailMessage="Welcome to QR Dragonfly! Your verification code is {####}. This code expires in 24 hours.",
     EmailSubject="Verify your QR Dragonfly account"'
```

### Email Template Variables

Cognito provides these placeholders:

| Variable             | Description                  |
| -------------------- | ---------------------------- |
| `{####}`             | Verification code (6 digits) |
| `{username}`         | Username                     |
| `{##Verify Email##}` | Link for email verification  |

## Custom Email Templates

### Verification Email Template

**Subject:** Verify your QR Dragonfly account

**HTML Body:**

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Verify your account</title>
    <style>
      body {
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
          "Helvetica Neue", Arial, sans-serif;
        line-height: 1.6;
        color: #333;
        max-width: 600px;
        margin: 0 auto;
        padding: 20px;
      }
      .header {
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        color: white;
        padding: 30px;
        text-align: center;
        border-radius: 8px 8px 0 0;
      }
      .content {
        background: #f9fafb;
        padding: 30px;
        border-radius: 0 0 8px 8px;
      }
      .code-box {
        background: white;
        border: 2px solid #667eea;
        border-radius: 8px;
        padding: 20px;
        text-align: center;
        margin: 20px 0;
      }
      .code {
        font-size: 32px;
        font-weight: bold;
        letter-spacing: 8px;
        color: #667eea;
      }
      .button {
        display: inline-block;
        background: #667eea;
        color: white;
        padding: 12px 30px;
        text-decoration: none;
        border-radius: 6px;
        margin: 20px 0;
      }
      .footer {
        text-align: center;
        margin-top: 30px;
        font-size: 12px;
        color: #666;
      }
    </style>
  </head>
  <body>
    <div class="header">
      <h1>Welcome to QR Dragonfly! 🚀</h1>
    </div>
    <div class="content">
      <p>Hi there,</p>
      <p>
        Thanks for signing up! To complete your registration, please verify your
        email address by entering this code:
      </p>

      <div class="code-box">
        <div class="code">{####}</div>
      </div>

      <p><strong>This code expires in 24 hours.</strong></p>

      <p>
        If you didn't create an account with QR Dragonfly, you can safely ignore
        this email.
      </p>

      <p>Questions? Reply to this email and we'll help you out!</p>

      <p>Best,<br />The QR Dragonfly Team</p>
    </div>
    <div class="footer">
      <p>QR Dragonfly - Create and track QR codes with analytics</p>
      <p>
        <a href="https://your-domain.com/privacy">Privacy Policy</a> |
        <a href="https://your-domain.com/terms">Terms of Service</a>
      </p>
    </div>
  </body>
</html>
```

**Text Version:**

```
Welcome to QR Dragonfly!

Thanks for signing up! To complete your registration, verify your email with this code:

{####}

This code expires in 24 hours.

If you didn't create an account, you can safely ignore this email.

Best,
The QR Dragonfly Team

---
QR Dragonfly - Create and track QR codes with analytics
https://your-domain.com
```

### Password Reset Email Template

**Subject:** Reset your QR Dragonfly password

**HTML Body:**

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Reset your password</title>
    <style>
      /* Same styles as verification email */
    </style>
  </head>
  <body>
    <div class="header">
      <h1>Password Reset Request</h1>
    </div>
    <div class="content">
      <p>Hi there,</p>
      <p>
        We received a request to reset your password. Use this code to reset
        your password:
      </p>

      <div class="code-box">
        <div class="code">{####}</div>
      </div>

      <p><strong>This code expires in 1 hour.</strong></p>

      <p>
        If you didn't request a password reset, you can safely ignore this
        email. Your password will not be changed.
      </p>

      <p>For security, never share this code with anyone.</p>

      <p>Best,<br />The QR Dragonfly Team</p>
    </div>
    <div class="footer">
      <p>QR Dragonfly - Create and track QR codes with analytics</p>
      <p>
        <a href="https://your-domain.com/privacy">Privacy Policy</a> |
        <a href="https://your-domain.com/terms">Terms of Service</a>
      </p>
    </div>
  </body>
</html>
```

## Applying Templates via Terraform

Create `terraform/cognito_email_templates.tf`:

```hcl
resource "aws_cognito_user_pool" "main" {
  name = "qr-dragonfly-users"

  email_configuration {
    email_sending_account = "DEVELOPER" # or "COGNITO_DEFAULT"
    source_arn           = aws_ses_email_identity.main.arn
    reply_to_email       = "support@your-domain.com"
    from_email_address   = "QR Dragonfly <noreply@your-domain.com>"
  }

  verification_message_template {
    default_email_option = "CONFIRM_WITH_CODE"
    email_subject        = "Verify your QR Dragonfly account"
    email_message        = file("${path.module}/templates/verification_email.html")
  }

  # Password reset template
  admin_create_user_config {
    invite_message_template {
      email_subject = "Reset your QR Dragonfly password"
      email_message = file("${path.module}/templates/password_reset_email.html")
    }
  }
}

resource "aws_ses_email_identity" "main" {
  email = "your-domain.com"
}
```

## Custom Welcome Email (Post-Signup)

Create a Lambda function triggered after Cognito signup to send custom welcome emails.

### Lambda Function

Create `backend/user-service/lambda/welcome-email/main.go`:

```go
package main

import (
    "context"
    "fmt"
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/ses"
)

func handlePostConfirmation(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
    email := event.Request.UserAttributes["email"]
    name := event.Request.UserAttributes["name"]

    if email == "" {
        return event, nil
    }

    if err := sendWelcomeEmail(email, name); err != nil {
        fmt.Printf("Failed to send welcome email: %v\n", err)
        // Don't fail the signup, just log
    }

    return event, nil
}

func sendWelcomeEmail(email, name string) error {
    sess := session.Must(session.NewSession())
    svc := ses.New(sess)

    subject := "Welcome to QR Dragonfly! 🎉"
    htmlBody := fmt.Sprintf(`
        <h1>Welcome aboard, %s!</h1>
        <p>We're excited to have you on QR Dragonfly.</p>
        <p>Here's what you can do to get started:</p>
        <ul>
            <li>Create your first QR code</li>
            <li>Track clicks and analytics</li>
            <li>Upgrade to unlock more features</li>
        </ul>
        <a href="https://your-domain.com" style="display:inline-block;background:#667eea;color:white;padding:12px 30px;text-decoration:none;border-radius:6px;">Get Started</a>
    `, name)

    input := &ses.SendEmailInput{
        Source: aws.String("QR Dragonfly <noreply@your-domain.com>"),
        Destination: &ses.Destination{
            ToAddresses: []*string{aws.String(email)},
        },
        Message: &ses.Message{
            Subject: &ses.Content{Data: aws.String(subject)},
            Body: &ses.Body{
                Html: &ses.Content{Data: aws.String(htmlBody)},
            },
        },
    }

    _, err := svc.SendEmail(input)
    return err
}

func main() {
    lambda.Start(handlePostConfirmation)
}
```

### Deploy Lambda

```bash
cd backend/user-service/lambda/welcome-email
GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
zip function.zip bootstrap
aws lambda create-function \
  --function-name qr-dragonfly-welcome-email \
  --runtime provided.al2 \
  --role arn:aws:iam::ACCOUNT_ID:role/lambda-execution-role \
  --handler bootstrap \
  --zip-file fileb://function.zip
```

### Connect to Cognito

```bash
aws cognito-idp update-user-pool \
  --user-pool-id YOUR_USER_POOL_ID \
  --lambda-config PostConfirmation=arn:aws:lambda:REGION:ACCOUNT_ID:function:qr-dragonfly-welcome-email
```

## Email Deliverability Best Practices

### 1. SPF Records

Add to DNS:

```
v=spf1 include:amazonses.com ~all
```

### 2. DKIM

AWS SES provides DKIM records - add all 3 CNAME records to DNS.

### 3. DMARC

Add to DNS:

```
_dmarc.your-domain.com TXT "v=DMARC1; p=quarantine; rua=mailto:dmarc@your-domain.com"
```

### 4. Warm Up IP

If using dedicated IP, gradually increase sending volume over 2-4 weeks.

### 5. Monitor Bounces

Set up SNS topics for bounce notifications:

```bash
aws ses set-identity-notification-topic \
  --identity your-domain.com \
  --notification-type Bounce \
  --sns-topic arn:aws:sns:REGION:ACCOUNT_ID:ses-bounces
```

## Testing Emails

### Test in Development

```bash
# Use Mailhog for local email testing
docker run -d -p 1025:1025 -p 8025:8025 mailhog/mailhog

# Configure SES to use Mailhog (development only)
# Or use AWS SES Sandbox mode
```

### View Test Emails

Access Mailhog UI: http://localhost:8025

### Send Test Email via SES

```bash
aws ses send-email \
  --from "noreply@your-domain.com" \
  --to "test@example.com" \
  --subject "Test Email" \
  --text "This is a test email"
```

## Monitoring Email Delivery

### CloudWatch Metrics

Monitor in AWS Console → SES → Sending Statistics:

- Send rate
- Bounce rate
- Complaint rate

### Set Up Alerts

```bash
aws cloudwatch put-metric-alarm \
  --alarm-name high-bounce-rate \
  --alarm-description "Alert when bounce rate > 5%" \
  --metric-name Reputation.BounceRate \
  --namespace AWS/SES \
  --statistic Average \
  --period 3600 \
  --evaluation-periods 1 \
  --threshold 0.05 \
  --comparison-operator GreaterThanThreshold
```

## Troubleshooting

### Emails Not Delivered

1. Check SES sending status (Sandbox vs Production)
2. Verify domain/email in SES
3. Check bounce/complaint rates
4. Review CloudWatch logs
5. Verify DNS records (SPF, DKIM, DMARC)

### Emails Going to Spam

1. Authenticate with SPF, DKIM, DMARC
2. Avoid spam trigger words
3. Include unsubscribe link
4. Maintain low bounce/complaint rates
5. Warm up IP address gradually

### Rate Limits

- **Cognito Default**: 50 emails/day
- **SES Sandbox**: 200 emails/day, can only send to verified addresses
- **SES Production**: Request limit increase (start at 50,000/day)

Request limit increase:

```
AWS Console → SES → Account Dashboard → Request Production Access
```

## Cost Estimates

### SES Pricing (us-east-1)

- $0.10 per 1,000 emails
- Free tier: 62,000 emails/month (if hosted on EC2/Lambda)

### Example Costs

| Users   | Emails/Month | Cost      |
| ------- | ------------ | --------- |
| 100     | 600          | $0.06     |
| 1,000   | 6,000        | $0.60     |
| 10,000  | 60,000       | Free tier |
| 100,000 | 600,000      | $54.00    |

## Next Steps

1. ✅ Verify domain in SES
2. ✅ Configure DNS records
3. ✅ Update Cognito email config
4. ✅ Customize email templates
5. ✅ Test email delivery
6. ✅ Set up monitoring
7. ✅ Request production access (if needed)
8. ✅ Implement welcome email Lambda (optional)
