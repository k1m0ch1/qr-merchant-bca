# Manual Network Traffic Capture Instructions

## Quick Method (Recommended)

### Using Browser DevTools

1. Open Chrome/Edge browser
2. Press `F12` to open Developer Tools
3. Go to **Network** tab
4. Check "Preserve log" checkbox
5. Navigate to https://qr.klikbca.com/login
6. Enter credentials:
   - Email: someemail@gmail.com
   - Password: somepassword
7. Click "Masuk" (Login)
8. Wait for response
9. Right-click in Network tab → **"Save all as HAR with content"**
10. Save as `C:\tmp\bca-login.har`

### What to Look For

After clicking login, look for these requests in the Network tab:

1. **session/v1.0.0/add** or similar authentication endpoint
2. Check:
   - Request Method (POST/GET)
   - Request Headers (all of them!)
   - Request Payload (the JSON body)
   - Response Status
   - Response Headers
   - Response Body

### Key Information to Capture

**Request Headers:**
```
Content-Type: ?
Authorization: ?
X-CSRF-Token: ?
Cookie: ?
Origin: ?
Referer: ?
User-Agent: ?
... (capture ALL headers)
```

**Request Payload:**
```json
{
  "email": "?",
  "password": "?",
  ... (any other fields?)
}
```

**Response:**
```
Status Code: ?
Headers: ?
Body: ?
```

## Alternative: Using Fiddler/Burp Suite

1. Install Fiddler or Burp Suite
2. Configure browser proxy to 127.0.0.1:8888 (Fiddler) or 8080 (Burp)
3. Navigate and login
4. Export captured traffic

## Alternative: Using curl with cookies

If login succeeds in browser:

1. Login in browser with DevTools open
2. Find the successful request
3. Right-click → Copy → Copy as cURL
4. Save to file

## Files to Create

After capture, create these files in `C:\tmp\`:

1. `request-headers.txt` - All request headers
2. `request-payload.json` - The JSON sent
3. `response-headers.txt` - All response headers
4. `response-body.json` - The JSON received
5. `cookies.txt` - All cookies before/after request
6. `bca-login.har` - Full HAR export

## Expected Endpoints

Look for requests to:
- `https://qr.klikbca.com/api/session/v1.0.0/add`
- `https://bms.ebanksvc.bca.co.id/session/v1.0.0/add`
- `https://mssi.ebanksvc.bca.co.id/session/v1.0.0/add`
- Any Firebase authentication URLs
- Any token endpoints

## What We Need to Find

1. **The correct endpoint URL** (which backend?)
2. **The exact request format** (JSON structure)
3. **Required headers** (CSRF tokens, etc.)
4. **Cookie requirements** (session cookies?)
5. **The encryption format** (is password encrypted in payload?)
6. **Response tokens** (Firebase custom token? JWT?)

## Troubleshooting

If login fails in browser too:
- Check console for JavaScript errors
- Verify credentials are correct
- Check if account is locked/disabled
- Try incognito mode
