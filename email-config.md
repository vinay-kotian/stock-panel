# Email Configuration for Password Reset

To enable email sending for password reset functionality, you need to configure SMTP settings using environment variables.

## Environment Variables

Set the following environment variables before running the application:

### Required Variables
- `SMTP_USER`: Your email address (e.g., "your-email@gmail.com")
- `SMTP_PASS`: Your email password or app password

### Optional Variables
- `SMTP_HOST`: SMTP server host (default: "smtp.gmail.com")
- `FROM_EMAIL`: From email address (default: same as SMTP_USER)

## Gmail Setup (Recommended)

1. **Enable 2-Factor Authentication** on your Google account
2. **Generate an App Password**:
   - Go to Google Account settings
   - Security → 2-Step Verification → App passwords
   - Generate a new app password for "Mail"
3. **Set Environment Variables**:
   ```bash
   export SMTP_USER="your-email@gmail.com"
   export SMTP_PASS="your-app-password"
   ```

## Other Email Providers

### Outlook/Hotmail
```bash
export SMTP_HOST="smtp-mail.outlook.com"
export SMTP_USER="your-email@outlook.com"
export SMTP_PASS="your-password"
```

### Yahoo
```bash
export SMTP_HOST="smtp.mail.yahoo.com"
export SMTP_USER="your-email@yahoo.com"
export SMTP_PASS="your-app-password"
```

### Custom SMTP Server
```bash
export SMTP_HOST="your-smtp-server.com"
export SMTP_PORT="587"
export SMTP_USER="your-username"
export SMTP_PASS="your-password"
```

## Testing

1. Set the environment variables
2. Run the application
3. Go to the forgot password page
4. Enter an email address
5. Check if the email is sent successfully

## Fallback Mode

If email is not configured, the application will:
- Log the reset link to the console
- Still return a success response to the user
- Display a message about email configuration

## Security Notes

- Never commit email credentials to version control
- Use app passwords instead of regular passwords when possible
- Consider using a dedicated email service for production (SendGrid, Mailgun, etc.)
