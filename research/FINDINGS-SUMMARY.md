# BCA QR Login Investigation - Summary & Next Steps

## Investigation Date
2025-12-25

## What We Discovered

### ✅ Successfully Found:

1. **Firebase Configuration**
   - API Key: `AIzaSyAnGQ6VXc8JTKCg84mtIWalDGs1hdxgVrY`
   - Project ID: `merchant-bca`
   - Database URL: `https://merchant-bca.firebaseio.com`
   - Storage Bucket: `merchant-bca.appspot.com`
   - **Password Auth**: DISABLED (uses custom tokens)

2. **Encryption Details** (Already Implemented)
   - Algorithm: AES-256-CBC
   - Key: `9C0XAVRJ6PQB86TVTAD6SK6XD01PSCIK`
   - Mode: CBC with PKCS7 padding
   - IV: Random 16 bytes (prepended)
   - Output: Base64 encoded

3. **Backend URLs**
   - Main: `https://bms.ebanksvc.bca.co.id`
   - Alternative: `https://mssi.ebanksvc.bca.co.id`
   - Proxy: `https://qr.klikbca.com/api/*`

4. **API Endpoints**
   - `session/v1.0.0/add` - Login/session creation
   - `user/v1.0.0/member` - User data
   - `outlet/v1.0.0/list` - Outlet listings
   - `transaction-v2/v2.0.0/list` - Transaction history

### ❌ Current Blockers:

1. **405 Not Allowed from Nginx**
   ```
   POST https://qr.klikbca.com/api/session/v1.0.0/add
   Result: 405 Not Allowed
   ```

2. **401 Unauthorized from Direct Backend**
   ```
   POST https://bms.ebanksvc.bca.co.id/session/v1.0.0/add
   Result: 401 Unauthorized (requires Bearer token!)
   ```
   **Paradox**: Login endpoint requires authentication!

3. **Firebase Direct Auth Disabled**
   ```
   POST https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword
   Result: PASSWORD_LOGIN_DISABLED
   ```

## Hypothesized Authentication Flow

```
1. User enters credentials → Frontend
2. Frontend encrypts password (AES-CBC)
3. Frontend calls ??? → Gets some initial token/session
4. Use token to call backend session API
5. Backend returns Firebase custom token
6. Frontend uses custom token → Firebase auth
7. Firebase returns ID token (JWT)
8. Use ID token as Bearer for all API calls
```

**Missing Piece**: Step 3 - How to get initial authorization!

## Test Results

| Test | Method | URL | Result |
|------|--------|-----|--------|
| CLI Tool | POST | /api/session/v1.0.0/add | 405 |
| cURL | POST | /api/session/v1.0.0/add | 405 |
| cURL Direct | POST | bms.ebanksvc.bca.co.id/session/v1.0.0/add | 401 |
| Firebase | POST | identitytoolkit.googleapis.com/... | PASSWORD_LOGIN_DISABLED |

## What We Need (Critical)

**MUST CAPTURE REAL BROWSER TRAFFIC**

We need to see:
1. Exact request URL (which backend?)
2. All request headers (especially auth-related)
3. Request payload structure
4. Cookies before and after
5. Any CSRF tokens or nonces
6. Response with tokens

## How to Capture (DO THIS MANUALLY)

### Method 1: Browser DevTools (Easiest)

1. Open Chrome
2. Press F12 → Network tab
3. ✅ Check "Preserve log"
4. Go to https://qr.klikbca.com/login
5. Login with:
   - Email: some-email@gmail.com
   - Password: mantapjiwa
6. Click "Masuk"
7. In Network tab, find the POST request
8. Right-click → "Save all as HAR with content"
9. Save to: `C:\tmp\bca-login-capture.har`

### Method 2: What to Look For

In the Network tab, look for:
- Any POST request after clicking "Masuk"
- Requests to `/session` or `/add` or `/login`
- Requests to Firebase URLs
- Requests with JSON payloads

### Method 3: Export Specific Request

1. Find the login POST request
2. Right-click → Copy → Copy as cURL
3. Save to: `C:\tmp\login-curl.txt`

## Files Created for Reference

- `C:\tmp\CAPTURE-INSTRUCTIONS.md` - Detailed capture instructions
- `C:\tmp\network-capture.html` - HTML tool (won't work due to CORS)
- `C:\tmp\capture-bca-login.go` - Go script (needs chromedp dependency)
- `C:\tmp\FINDINGS-SUMMARY.md` - This file

## Next Actions

### Immediate (User Must Do):
1. **CAPTURE NETWORK TRAFFIC** using browser DevTools (see method above)
2. Share the HAR file or copy the request details
3. Or use Copy as cURL for the specific request

### After Getting Traffic Data:
1. Analyze the captured request/response
2. Identify missing headers/tokens/cookies
3. Update Go client to replicate exact request
4. Test the updated implementation
5. Document the working authentication flow

## Documentation Updated

All findings have been documented in:
- `CLAUDE.md` - Credentials and configuration
- `arch-flow.md` - Architecture and authentication flow analysis
- `ENCRYPTION.md` - Encryption implementation details

## Why Playwright MCP Failed

The Playwright MCP browser kept:
- Navigating away to extension page
- Timing out on element selectors
- Closing the browser context

Manual capture with DevTools is more reliable for this investigation.

## Summary

**We have everything EXCEPT the initial authentication mechanism.**

The encryption works, the endpoints are known, but we're missing how to authorize the first request. This can only be discovered by capturing real browser traffic during a successful login.

**Priority**: Get HAR export or cURL of working login request from browser.
