#!/bin/bash

echo "=== Stock Panel Email Configuration Setup ==="
echo "This script will help you configure email settings for password reset functionality."
echo ""

# Check if .env file exists, if not create it
if [ ! -f .env ]; then
    echo "Creating .env file..."
    touch .env
fi

echo "Please provide your email configuration:"
echo ""

# SMTP Host
read -p "SMTP Host (default: smtp.gmail.com): " smtp_host
smtp_host=${smtp_host:-smtp.gmail.com}

# SMTP User (email)
read -p "Email Address: " smtp_user
if [ -z "$smtp_user" ]; then
    echo "Error: Email address is required"
    exit 1
fi

# SMTP Password
read -s -p "Email Password/App Password: " smtp_pass
echo ""
if [ -z "$smtp_pass" ]; then
    echo "Error: Password is required"
    exit 1
fi

# From Email (optional)
read -p "From Email (default: same as email address): " from_email
from_email=${from_email:-$smtp_user}

# Write to .env file
echo "Writing configuration to .env file..."
cat > .env << EOF
# Email Configuration
SMTP_HOST=$smtp_host
SMTP_USER=$smtp_user
SMTP_PASS=$smtp_pass
FROM_EMAIL=$from_email
EOF

echo ""
echo "Configuration saved to .env file!"
echo ""
echo "To use this configuration, run the application with:"
echo "source .env && go run cmd/stock-panel/main.go"
echo ""
echo "Or add this to your shell profile (.bashrc, .zshrc, etc.):"
echo "export SMTP_USER=\"$smtp_user\""
echo "export SMTP_PASS=\"$smtp_pass\""
echo "export SMTP_HOST=\"$smtp_host\""
echo "export FROM_EMAIL=\"$from_email\""
echo ""
echo "Note: Make sure to keep your .env file secure and never commit it to version control!"
