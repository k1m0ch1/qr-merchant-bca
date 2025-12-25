# BCA QR Merchant Service - Architecture & Authentication Flow

## System Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     Frontend Application                         │
│              https://qr.klikbca.com (Angular)                   │
│                                                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  Login Page  │  │ Dashboard    │  │ Transactions │          │
│  │  Component   │  │ Component    │  │ Component    │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│         │                  │                  │                  │
│         └──────────────────┴──────────────────┘                 │
│                           │                                      │
│                    ┌──────▼──────┐                              │
│                    │  Auth       │                              │
│                    │  Service    │                              │
│                    └──────┬──────┘                              │
│                           │                                      │
│                    ┌──────▼──────┐                              │
│                    │  Firebase   │                              │
│                    │  SDK        │                              │
│                    └──────┬──────┘                              │
└───────────────────────────┼──────────────────────────────────────┘
                            │
                ┌───────────┴───────────┐
                │                       │
        ┌───────▼────────┐      ┌──────▼───────┐
        │   Firebase     │      │    Nginx     │
        │   Backend      │      │   Proxy      │
        │                │      │              │
        │   Project:     │      │ qr.klikbca   │
        │ merchant-bca   │      │   .com/api/* │
        └────────────────┘      └──────┬───────┘
                                       │
                        ┌──────────────┴──────────────┐
                        │                             │
                ┌───────▼────────┐          ┌─────────▼────────┐
                │   BMS Backend  │          │  MSSI Backend    │
                │                │          │                  │
                │ bms.ebanksvc   │          │ mssi.ebanksvc    │
                │   .bca.co.id   │          │   .bca.co.id     │
                └────────────────┘          └──────────────────┘
```

## Authentication Flow Analysis

### Expected Flow (Hypothesis)

```
┌──────┐                ┌──────────┐              ┌─────────┐              ┌──────────┐
│Client│                │ Frontend │              │Firebase │              │ Backend  │
└──┬───┘                └────┬─────┘              └────┬────┘              └────┬─────┘
   │                         │                         │                        │
   │  1. Navigate to         │                         │                        │
   │     /login              │                         │                        │
   ├────────────────────────>│                         │                        │
   │                         │                         │                        │
   │  2. Load Angular App    │                         │                        │
   │     + Firebase SDK      │                         │                        │
   │<────────────────────────┤                         │                        │
   │                         │                         │                        │
   │  3. User enters         │                         │                        │
   │     email & password    │                         │                        │
   │     (plain text)        │                         │                        │
   │────────────────────────>│                         │                        │
   │                         │                         │                        │
   │                         │  4. Encrypt password    │                        │
   │                         │     using AES-256-CBC   │                        │
   │                         │     Key: 9C0XAVRJ...    │                        │
   │                         │                         │                        │
   │                         │  5. POST to Backend     │                        │
   │                         │     session/v1.0.0/add  │                        │
   │                         │     {                   │                        │
   │                         │       email: "...",     │                        │
   │                         │       password: "base64"│                        │
   │                         │     }                   │                        │
   │                         │─────────────────────────┼───────────────────────>│
   │                         │                         │                        │
   │                         │  6. Backend validates   │                        │
   │                         │     & creates Firebase  │                        │
   │                         │     custom token        │                        │
   │                         │                         │                        │
   │                         │  7. Return custom token │                        │
   │                         │<────────────────────────┼────────────────────────┤
   │                         │                         │                        │
   │                         │  8. signInWithCustom   │                        │
   │                         │     Token(customToken) │                        │
   │                         │────────────────────────>│                        │
   │                         │                         │                        │
   │                         │  9. Firebase ID token  │                        │
   │                         │     (JWT)              │                        │
   │                         │<────────────────────────┤                        │
   │                         │                         │                        │
   │  10. Store ID token     │                         │                        │
   │      Use as Bearer      │                         │                        │
   │      for API calls      │                         │                        │
   │                         │                         │                        │
   │  11. API calls with     │                         │                        │
   │      Authorization:     │                         │                        │
   │      Bearer {idToken}   │                         │                        │
   │                         │─────────────────────────┼───────────────────────>│
   │                         │                         │                        │
```

### Current Implementation Status

#### ✅ Completed
1. **Password Encryption**: AES-256-CBC encryption implemented in Go
   - Key: `9C0XAVRJ6PQB86TVTAD6SK6XD01PSCIK`
   - Random IV generation (16 bytes)
   - PKCS7 padding
   - Base64 encoding

2. **HTTP Client Setup**:
   - Cookie jar for session management
   - Proper headers (User-Agent, Referer, Origin, etc.)
   - Visit login page first to get cookies

3. **API Endpoint Discovery**:
   - Identified backend URLs
   - Mapped API endpoints
   - Found Firebase configuration

#### ❌ Blocking Issues

1. **405 Not Allowed from Nginx Proxy**
   ```
   Request:  POST https://qr.klikbca.com/api/session/v1.0.0/add
   Response: 405 Not Allowed (nginx/1.18.0)
   ```
   **Possible causes:**
   - CSRF token missing
   - Invalid request signature
   - Session cookie requirement
   - Angular-specific headers missing

2. **401 Unauthorized from Direct Backend**
   ```
   Request:  POST https://bms.ebanksvc.bca.co.id/session/v1.0.0/add
   Response: 401 Unauthorized
   Headers:  www-authenticate: Bearer
   ```
   **Issue:** Backend expects Bearer token, but this is the login endpoint!

3. **Firebase Password Auth Disabled**
   ```
   Request:  POST https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword
   Response: PASSWORD_LOGIN_DISABLED
   ```
   **Issue:** Cannot use standard Firebase email/password auth

## Investigation Results (2025-12-25)

### Test 1: Direct API Call with Encryption
```bash
# Go CLI attempt
./bcaqr.exe transactions -e email -p password

# Result
Error: 405 Not Allowed from nginx
```

### Test 2: Curl with Proper Headers
```bash
curl -X POST "https://qr.klikbca.com/api/session/v1.0.0/add" \
  -H "Content-Type: application/json" \
  -H "Origin: https://qr.klikbca.com" \
  -H "Referer: https://qr.klikbca.com/login" \
  -d '{"email":"test@test.com","password":"test"}'

# Result
405 Not Allowed
```

### Test 3: Direct Backend Call
```bash
curl -X POST "https://bms.ebanksvc.bca.co.id/session/v1.0.0/add" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"test"}'

# Result
401 Unauthorized
www-authenticate: Bearer
```

### Test 4: Firebase Authentication
```bash
curl "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=AIza..." \
  -d '{"email":"some email @gmail.com","password":"some password","returnSecureToken":true}'

# Result
PASSWORD_LOGIN_DISABLED
```

### Test 5: JavaScript Analysis
**Findings:**
- Firebase SDK is loaded
- Uses `signInWithCustomToken` method (not `signInWithEmailAndPassword`)
- Contains encryption function `encryptMessi` using AES-CBC
- References to `grant_type: 'client_credentials'` (OAuth 2.0)
- Session endpoint: `session/v1.0.0/add`

## Hypothesized Actual Flow

Based on the findings, the likely authentication flow is:

1. **User submits email + password** (plain text in browser)
2. **Frontend encrypts password** using AES-256-CBC with static key
3. **Frontend makes request** to custom backend (NOT Firebase directly)
   - Endpoint: Backend API (bms.ebanksvc.bca.co.id?)
   - Payload: `{email: "...", password: "encrypted"}`
   - Problem: This endpoint requires authentication!
4. **Backend validates credentials** and generates Firebase custom token
5. **Frontend uses custom token** to authenticate with Firebase
6. **Firebase returns ID token** (JWT)
7. **ID token used as Bearer** for all subsequent API calls

## Missing Pieces

### 1. Initial Authentication
**Question:** How does the first request to `session/v1.0.0/add` get authorized?

**Possibilities:**
- A. Pre-authenticated session cookie from visiting login page
- B. CSRF token from initial page load
- C. Client credentials OAuth flow (app-level auth)
- D. Request signature using static secret
- E. Different endpoint for initial login (not discovered yet)

### 2. Nginx Proxy Configuration
**Question:** Why is nginx blocking our requests with 405?

**Possibilities:**
- A. Requires specific Angular headers (X-Requested-With, etc.)
- B. Validates request origin/referer strictly
- C. CSRF token validation
- D. Session state requirement
- E. Rate limiting or IP-based blocking

### 3. Request Payload Format
**Question:** Is the payload structure correct?

**Current assumption:**
```json
{
  "email": "someemail@gmail.com",
  "password": "base64_encrypted_password"
}
```

**Might actually be:**
```json
{
  "email": "someemail@gmail.com",
  "password": "base64_encrypted_password",
  "grant_type": "client_credentials",
  "firebase_token": "...",
  // other fields?
}
```

## Next Investigation Steps

### Priority 1: Capture Real Browser Traffic
Use Playwright/browser automation to:
1. Navigate to login page
2. Fill in credentials
3. Capture all network requests (including preflight OPTIONS)
4. Extract:
   - All request headers
   - All cookies
   - Request payload
   - Response data
   - CSRF tokens
   - Session initialization sequence

### Priority 2: Analyze Angular Authentication Service
Search JavaScript for:
- `AuthService` or similar service class
- `login()` function implementation
- HTTP interceptors
- Token management
- Request builders

### Priority 3: Test Alternative Endpoints
Try different endpoint paths:
- `/session/add` (without version)
- `/auth/login`
- `/api/auth/session`
- Direct to mssi backend instead of bms

### Priority 4: Analyze CSP and Security Headers
Review Content-Security-Policy to understand:
- Allowed origins for API calls
- Nonce requirements
- Firebase integration points

## Technical Details

### Encryption Implementation (Go)
```go
// Key from JavaScript
key := "9C0XAVRJ6PQB86TVTAD6SK6XD01PSCIK"

// Generate random IV (16 bytes)
iv := make([]byte, aes.BlockSize)
rand.Read(iv)

// Encrypt with AES-256-CBC + PKCS7
block, _ := aes.NewCipher([]byte(key))
mode := cipher.NewCBCEncrypter(block, iv)
padded := pkcs7Pad(plaintext, aes.BlockSize)
ciphertext := make([]byte, len(padded))
mode.CryptBlocks(ciphertext, padded)

// Prepend IV and encode base64
result := append(iv, ciphertext...)
encoded := base64.StdEncoding.EncodeToString(result)
```

### Request Headers (Current)
```
Content-Type: application/json
Accept: application/json, text/plain, */*
Origin: https://qr.klikbca.com
Referer: https://qr.klikbca.com/login
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36
sec-ch-ua: "Not_A Brand";v="8", "Chromium";v="120"
sec-ch-ua-mobile: ?0
sec-ch-ua-platform: "Windows"
Sec-Fetch-Dest: empty
Sec-Fetch-Mode: cors
Sec-Fetch-Site: same-origin
```

### Response Headers from Nginx
```
HTTP/1.1 405 Not Allowed
cache-control: no-store, no-cache, must-revalidate
content-security-policy: default-src 'self'; script-src 'self' 'nonce-...'
Set-Cookie: cookis=!APjfAFN2OYmm3OMlc4yt+Qek+XD1aZWvzIuTD2RlYL...
Set-Cookie: TS01dc34de=01379333c2bf55ab46897a21455e83cfd50c...
```

## Resources

### URLs
- Frontend: https://qr.klikbca.com
- Login Page: https://qr.klikbca.com/login
- Main JS: https://qr.klikbca.com/main.2aee9935593dc1600215.js
- Backend 1: https://bms.ebanksvc.bca.co.id
- Backend 2: https://mssi.ebanksvc.bca.co.id
- Firebase Project: https://merchant-bca.firebaseio.com

### Firebase Config
```javascript
{
  apiKey: "AIzaSyAnGQ6VXc8JTKCg84mtIWalDGs1hdxgVrY",
  projectId: "merchant-bca",
  databaseURL: "https://merchant-bca.firebaseio.com",
  storageBucket: "merchant-bca.appspot.com"
}
```

### Encryption Key
```
Key: 9C0XAVRJ6PQB86TVTAD6SK6XD01PSCIK
Algorithm: AES-256-CBC
Padding: PKCS7
IV: Random 16 bytes (prepended to ciphertext)
Encoding: Base64
```

## Conclusion

The authentication system uses a hybrid approach combining:
1. Custom backend authentication (encrypted credentials)
2. Firebase custom token authentication
3. JWT bearer tokens for API authorization

The main blocker is understanding the initial authentication step that gets past the nginx 405 error. This requires capturing real browser traffic to see the complete request/response flow including any hidden tokens, signatures, or session state.

**Status:** Investigation ongoing - browser capture needed
**Last Updated:** 2025-12-25
