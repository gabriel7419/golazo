# Reddit OAuth Setup

Enable Reddit OAuth for improved goal replay link retrieval. OAuth provides higher rate limits and eliminates CAPTCHA issues compared to the public API.

## Current Status

**Note**: Reddit has restricted script app creation and OAuth access. The OAuth setup may not work for all users due to Reddit's API changes. Golazo works perfectly with the public API fallback.

## Why Use OAuth (When Available)?

- **600 requests/hour** (vs. 5/minute with public API)
- **No CAPTCHA blocks** for more reliable link retrieval
- **Better success rate** for finding goal replays
- **Automatic fallback** to public API if OAuth is unavailable

## Setup Steps (If Available)

### 1. Account Requirements

Before creating an app, ensure your Reddit account:
- Has email verification completed
- Is at least 30 days old
- Has some account activity/karma

### 2. Create Reddit App

1. Visit https://www.reddit.com/prefs/apps
2. Click "Create App" or "Create Another App"
3. Configure:
   - **Name**: `golazo-app`
   - **App type**: `script`
   - **Description**: `Goal replay link retrieval`
   - **Redirect URI**: `http://localhost:8080`
4. Click "Create app"

**Note**: If app creation fails, Reddit may have disabled script apps for your account or region.

### 3. Get Credentials (If App Created)

After creation, note:
- **Client ID**: String under the app name
- **Client Secret**: The "secret" field

### 4. Set Environment Variables

If you successfully created an app, set these environment variables:

```bash
export REDDIT_CLIENT_ID="your_client_id_here"
export REDDIT_CLIENT_SECRET="your_client_secret_here"
export REDDIT_USERNAME="your_reddit_username"
export REDDIT_PASSWORD="your_reddit_password"
```

**macOS/Linux**: Add to `~/.bashrc`, `~/.zshrc`, or `~/.profile`

**Windows**: Use System Properties → Environment Variables

## Fallback Behavior

Golazo automatically uses the public Reddit API when OAuth is unavailable, with built-in CAPTCHA handling and retry logic. The app functions normally without OAuth credentials, but may run into some limitation due to API rate limits.
