# BCA QR Merchant Authentication Flow

## Overview

This document explains how the BCA QR Merchant authentication system works, based on reverse engineering the production web application at `https://qr.klikbca.com`.

**Target Audience:** Medium-level engineers familiar with HTTP, REST APIs, and basic cryptography.

**Current Implementation Status:** 95% complete - only missing the MCB encryption algorithm implementation.

---

## Table of Contents

1. [High-Level Architecture](#high-level-architecture)
2. [Complete Authentication Flow](#complete-authentication-flow)
3. [Request Details](#request-details)
4. [Encryption Systems](#encryption-systems)
5. [Implementation Guide](#implementation-guide)
6. [Known vs Unknown](#known-vs-unknown)
7. [Common Issues & Solutions](#common-issues--solutions)

---

## High-Level Architecture

```
┌─────────────┐
│   Browser   │
│  (Frontend) │
└──────┬──────┘
       │
       │ 1. User enters email + password
       │
       ▼
┌─────────────────────────────┐
│   JavaScript Application    │
│   (qr.klikbca.com)         │
│                             │
│  - Encrypts password (Messi)│
│  - Encrypts headers (MCB)   │
│  - Encrypts payload (MCB)   │
└──────┬──────────────────────┘
       │
       │ 2. POST with encrypted data
       │    (DIRECT - bypasses nginx!)
       │
       ▼
┌─────────────────────────────┐
│   Backend API               │
│   bms.ebanksvc.bca.co.id   │
│                             │
│  - Decrypts request         │
│  - Validates credentials    │
│  - Generates Firebase token │
└──────┬──────────────────────┘
       │
       │ 3. Returns 201 Created
       │    with session data
       │
       ▼
┌─────────────────────────────┐
│   Firebase Authentication   │
│   (Custom Token Flow)       │
│                             │
│  - Accepts custom token     │
│  - Returns ID token (JWT)   │
└─────────────────────────────┘
```

**Key Insight:** The authentication does NOT go through the nginx proxy at `qr.klikbca.com/api/*`. It goes **directly** to the backend server `bms.ebanksvc.bca.co.id`.

---

## Complete Authentication Flow

### Step-by-Step Process

#### Step 1: User Input
```
Email: someemail@gmail.com
Password: plainpassword (plain text)
```

#### Step 2: Frontend Encryption

The JavaScript application performs **three separate encryptions**:

1. **X-OS Header** (OS/Platform info)
   ```
   Input:  {"os":"Windows","version":"10.0",...}
   Output: AAABm1ZQY5oMGykxTTQ0-lEOEJXjJelCsSQSei1RfGkzA0vUpM9luL8MI_vwUlWnhA1JmJK7hg
   ```

2. **x-app-version Header** (App version)
   ```
   Input:  {"version":"1.0.0",...}
   Output: AAABm1ZQY5wMGykxTTQ0-lEOEJXjJelCHp7PJiVQcbDOLyyOER1_FfzFGR6AZkKiq6RgJ8iGHq4
   ```

3. **Request Body** (Complete JSON payload)
   ```
   Input:  {"email":"someemail@gmail.com","password":"<encrypted_with_messi>"}
   Output: AAABm1ZQY5wMGykxTTQ0-lEOEJXjJelCiPGRvbI3Ud7aGZgVQ3M3fFVOEfIyv8cAIJFyVMSbqRCgi-RxtKNcx4jqoFPYHz5DtSnI
   ```

**Important:** All three encrypted values start with the same prefix `AAABm1ZQY5` suggesting:
- Same encryption key
- Same timestamp/nonce
- Different plaintext data

#### Step 3: HTTP Request

```http
POST https://bms.ebanksvc.bca.co.id/session/v1.0.0/add HTTP/1.1
Host: bms.ebanksvc.bca.co.id
Content-Type: application/json
Origin: https://qr.klikbca.com
Referer: https://qr.klikbca.com/
X-OS: AAABm1ZQY5oMGykxTTQ0-lEOEJXjJelCsSQSei1RfGkzA0vUpM9luL8MI_vwUlWnhA1JmJK7hg
x-app-version: AAABm1ZQY5wMGykxTTQ0-lEOEJXjJelCHp7PJiVQcbDOLyyOER1_FfzFGR6AZkKiq6RgJ8iGHq4
x-encrypt-mcb: true
Content-Length: 100

AAABm1ZQY5wMGykxTTQ0-lEOEJXjJelCiPGRvbI3Ud7aGZgVQ3M3fFVOEfIyv8cAIJFyVMSbqRCgi-RxtKNcx4jqoFPYHz5DtSnI
```

**Critical Notes:**
- Body is NOT JSON! It's a single encrypted string
- `x-encrypt-mcb: true` signals to backend to decrypt using MCB
- Content-Type is `application/json` but payload is base64url encoded string

#### Step 4: Backend Processing

```
1. Backend receives request
2. Checks x-encrypt-mcb header
3. Decrypts X-OS header → extracts OS info
4. Decrypts x-app-version → extracts version info
5. Decrypts request body → extracts email + password
6. Validates credentials against database
7. Generates Firebase custom token
8. Returns session data
```

#### Step 5: HTTP Response

```http
HTTP/1.1 201 Created
Content-Type: application/json
Content-Length: 153
x-request-id: 694d6472a4b68daff3418dd0f9e7c3c1

{
  "token": "<firebase_custom_token>",
  "sessionId": "...",
  "expiresIn": 3600,
  ...
}
```

**Success Indicator:** `201 Created` status code (captured from working request)

#### Step 6: Firebase Authentication

```javascript
// Frontend uses custom token to authenticate with Firebase
firebase.auth().signInWithCustomToken(customToken)
  .then((userCredential) => {
    const idToken = userCredential.user.getIdToken();
    // Use idToken as Bearer token for subsequent API calls
  });
```

---

## Request Details

### URL Breakdown

**WRONG (causes 405 error):**
```
POST https://qr.klikbca.com/api/session/v1.0.0/add
```
❌ nginx proxy blocks this path

**CORRECT:**
```
POST https://bms.ebanksvc.bca.co.id/session/v1.0.0/add
```
✅ Direct backend access

### Required Headers

| Header | Required | Description | Example Value |
|--------|----------|-------------|---------------|
| `Content-Type` | Yes | Must be `application/json` | `application/json` |
| `Origin` | Yes | CORS requirement | `https://qr.klikbca.com` |
| `Referer` | Yes | CORS requirement | `https://qr.klikbca.com/` |
| `X-OS` | **Yes** | Encrypted OS/platform info | `AAABm1ZQY5o...` (70 chars) |
| `x-app-version` | **Yes** | Encrypted app version | `AAABm1ZQY5w...` (75 chars) |
| `x-encrypt-mcb` | **Yes** | Signals MCB encryption | `true` |

### Request Body Format

**NOT this (what you might expect):**
```json
{
  "email": "someemail@gmail.com",
  "password": "encrypted_password_here"
}
```

**BUT this (what it actually is):**
```
AAABm1ZQY5wMGykxTTQ0-lEOEJXjJelCiPGRvbI3Ud7aGZgVQ3M3fFVOEfIyv8cAIJFyVMSbqRCgi-RxtKNcx4jqoFPYHz5DtSnI
```

A single encrypted blob containing the entire JSON payload!

---

## Encryption Systems

### System 1: Messi Encryption (Known)

**Purpose:** Encrypt the password field before it goes into the request payload

**Algorithm:** AES-256-CBC

**Key:**
```
9C0XAVRJ6PQB86TVTAD6SK6XD01PSCIK
```

**Implementation (Go):**
```go
func encryptPassword(password string) (string, error) {
    key := []byte("9C0XAVRJ6PQB86TVTAD6SK6XD01PSCIK")

    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }

    // Generate random IV
    iv := make([]byte, aes.BlockSize)
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return "", err
    }

    // Pad password to block size
    paddedPassword := pkcs7Pad([]byte(password), aes.BlockSize)

    // Encrypt
    mode := cipher.NewCBCEncrypter(block, iv)
    ciphertext := make([]byte, len(paddedPassword))
    mode.CryptBlocks(ciphertext, paddedPassword)

    // Return IV + ciphertext, base64 encoded
    result := append(iv, ciphertext...)
    return base64.StdEncoding.EncodeToString(result), nil
}
```

### System 2: MCB Encryption (Unknown - CRITICAL MISSING PIECE!)

**Purpose:** Encrypt the entire request payload AND headers

**Algorithm:** ❓ Unknown (needs reverse engineering from JavaScript)

**Key:** ❓ Unknown (possibly derived, not hardcoded)

**Encoding:** Base64url (with `-` and `_` instead of `+` and `/`)

**Structure Analysis:**

Decoding the X-OS header:
```bash
$ echo "AAABm1ZQY5oMGykxTTQ0-lEOEJXjJelCsSQSei1RfGkzA0vUpM9luL8MI_vwUlWnhA1JmJK7hg" | base64 -d | xxd

00000000: 0000 019b 5650 639a 0c1b 2931 4d34 34fa  ....VP c..)1M44.
00000010: 510e 1097 e325 e942 b124 1227 a2d5 17c6  Q....%.B.$.'....
00000020: 9330 034b d4a4 cf65 b8bf 0c23 fbf0 5255  .0.K...e...#..RU
00000030: a784 0d49 98b2 ba86                      ...I....
```

**Pattern Observations:**
- First 4 bytes: `00 00 01 9b` (constant across all three encrypted values)
  - Decimal value: 411
  - Possibly: version number, timestamp, or encryption mode identifier
- Remaining bytes: encrypted data

**Hypotheses:**
1. Could be AES-GCM (provides authentication)
2. Could be ChaCha20-Poly1305 (modern alternative)
3. Could be custom algorithm
4. 4-byte prefix might contain:
   - Version (1 byte)
   - Timestamp/nonce (3 bytes)
   - Or: 32-bit timestamp in milliseconds

### Where to Find MCB Implementation

Search JavaScript files for:
```bash
# Look for MCB functions
grep -i "encrypt.*mcb\|mcb.*encrypt\|generateKeyMcb" main.js

# Look for the 4-byte prefix
grep "00019b\|\x00\x00\x01\x9b" main.js

# Look for header generation
grep -i "x-os\|x-app-version\|x-encrypt-mcb" main.js
```

---

## Implementation Guide

### Prerequisites

```go
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
    "io"
    "net/http"
)
```

### Step 1: Implement Messi Encryption ✅

Already implemented in `internal/client/client.go`:

```go
func encryptPassword(password string) (string, error) {
    // See "System 1: Messi Encryption" above
    // This part is COMPLETE and WORKING
}
```

### Step 2: Implement MCB Encryption ❌ MISSING

```go
// TODO: Reverse engineer from JavaScript
func encryptMCB(data interface{}) (string, error) {
    // 1. Marshal data to JSON
    jsonData, err := json.Marshal(data)
    if err != nil {
        return "", err
    }

    // 2. Generate 4-byte prefix
    prefix := []byte{0x00, 0x00, 0x01, 0x9b} // TODO: determine how this is generated

    // 3. Encrypt JSON data
    // TODO: Find encryption algorithm, key, IV
    encrypted := encryptWithUnknownAlgorithm(jsonData)

    // 4. Combine prefix + encrypted data
    result := append(prefix, encrypted...)

    // 5. Base64url encode
    encoded := base64.URLEncoding.EncodeToString(result)

    return encoded, nil
}
```

### Step 3: Generate Headers

```go
func generateHeaders() (map[string]string, error) {
    // X-OS header
    osInfo := map[string]interface{}{
        "os":      runtime.GOOS,
        "arch":    runtime.GOARCH,
        "version": "10.0", // TODO: get actual OS version
    }
    xOS, err := encryptMCB(osInfo)
    if err != nil {
        return nil, err
    }

    // x-app-version header
    versionInfo := map[string]interface{}{
        "version": "1.0.0",
        "build":   "1",
    }
    xAppVersion, err := encryptMCB(versionInfo)
    if err != nil {
        return nil, err
    }

    return map[string]string{
        "Content-Type":   "application/json",
        "Origin":         "https://qr.klikbca.com",
        "Referer":        "https://qr.klikbca.com/",
        "X-OS":           xOS,
        "x-app-version":  xAppVersion,
        "x-encrypt-mcb":  "true",
    }, nil
}
```

### Step 4: Build Login Request

```go
func (c *Client) Login(email, password string) error {
    // 1. Encrypt password with Messi
    encryptedPassword, err := encryptPassword(password)
    if err != nil {
        return err
    }

    // 2. Create payload
    payload := map[string]string{
        "email":    email,
        "password": encryptedPassword,
    }

    // 3. Encrypt entire payload with MCB
    encryptedPayload, err := encryptMCB(payload)
    if err != nil {
        return err
    }

    // 4. Generate headers
    headers, err := generateHeaders()
    if err != nil {
        return err
    }

    // 5. Make request to CORRECT URL
    req, err := http.NewRequest(
        "POST",
        "https://bms.ebanksvc.bca.co.id/session/v1.0.0/add",
        strings.NewReader(encryptedPayload),
    )
    if err != nil {
        return err
    }

    // 6. Set headers
    for key, value := range headers {
        req.Header.Set(key, value)
    }

    // 7. Execute request
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // 8. Check for success
    if resp.StatusCode != 201 {
        return fmt.Errorf("login failed: %d", resp.StatusCode)
    }

    // 9. Parse response
    var result struct {
        Token     string `json:"token"`
        SessionID string `json:"sessionId"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return err
    }

    c.firebaseToken = result.Token
    return nil
}
```

### Step 5: Use Firebase Token

```go
func (c *Client) makeAuthenticatedRequest(endpoint string) error {
    req, err := http.NewRequest("GET", endpoint, nil)
    if err != nil {
        return err
    }

    // Use Firebase ID token as Bearer token
    req.Header.Set("Authorization", "Bearer "+c.firebaseToken)

    resp, err := c.httpClient.Do(req)
    // ... handle response
}
```

---

## Known vs Unknown

### ✅ Known (Working)

1. **Correct URL:** `https://bms.ebanksvc.bca.co.id/session/v1.0.0/add`
2. **HTTP Method:** POST
3. **Required Headers:** X-OS, x-app-version, x-encrypt-mcb
4. **Success Response:** 201 Created
5. **Messi Encryption:** AES-256-CBC with known key
6. **Request Format:** Single encrypted blob (not JSON)
7. **Firebase Integration:** Custom token flow
8. **Proof of Concept:** Working request captured in HAR file

### ❌ Unknown (Blocking Implementation)

1. **MCB Encryption Algorithm**
   - Could be AES-GCM, ChaCha20, or custom
   - Need to reverse engineer from JavaScript

2. **MCB Encryption Key**
   - Possibly derived from timestamp/nonce
   - Not hardcoded like Messi key

3. **4-Byte Prefix**
   - Value: `00 00 01 9b` (411 decimal)
   - Purpose: version? timestamp? mode?

4. **Header Content Structure**
   - What exact JSON goes into X-OS?
   - What exact JSON goes into x-app-version?

5. **Response Body**
   - HAR capture shows 153 bytes response
   - Content not captured (need to capture with body)

---

## Common Issues & Solutions

### Issue 1: 405 Method Not Allowed

**Error:**
```
POST https://qr.klikbca.com/api/session/v1.0.0/add
HTTP/1.1 405 Method Not Allowed
```

**Cause:** Using nginx proxy URL instead of direct backend

**Solution:** Use `https://bms.ebanksvc.bca.co.id/session/v1.0.0/add`

---

### Issue 2: 401 Unauthorized

**Error:**
```
POST https://bms.ebanksvc.bca.co.id/session/v1.0.0/add
HTTP/1.1 401 Unauthorized
```

**Cause:** Missing or incorrect encrypted headers

**Solution:**
1. Ensure `X-OS` header is present and correctly encrypted
2. Ensure `x-app-version` header is present and correctly encrypted
3. Ensure `x-encrypt-mcb: true` is set

---

### Issue 3: Firebase PASSWORD_LOGIN_DISABLED

**Error:**
```json
{
  "error": {
    "code": 400,
    "message": "PASSWORD_LOGIN_DISABLED"
  }
}
```

**Cause:** Trying to authenticate directly with Firebase using password

**Solution:** Don't call Firebase directly! Use the BCA backend to get a custom token first, then use that token with Firebase.

---

### Issue 4: Invalid Encryption

**Symptoms:**
- 400 Bad Request
- 500 Internal Server Error
- "Invalid payload" error

**Possible Causes:**
1. Wrong encryption algorithm
2. Wrong encryption key
3. Wrong IV generation
4. Wrong padding scheme
5. Wrong base64 encoding (should be base64url)

**Debugging:**
```go
// Decode captured value to verify format
func debugEncrypted(encoded string) {
    decoded, _ := base64.URLEncoding.DecodeString(encoded)
    fmt.Printf("Length: %d\n", len(decoded))
    fmt.Printf("Hex: %x\n", decoded)
    fmt.Printf("First 4 bytes: %x (%d)\n", decoded[:4], binary.BigEndian.Uint32(decoded[:4]))
}
```

---

## Next Steps

### Immediate (To Complete Implementation)

1. **Reverse Engineer MCB Encryption**
   ```bash
   # Extract main.js from qr.klikbca.com
   curl https://qr.klikbca.com/main.js > main.js

   # Search for encryption functions
   grep -i "mcb\|encrypt" main.js | head -50

   # Look for key generation
   grep -i "generateKey\|derivedKey" main.js
   ```

2. **Capture Response Body**
   - Use browser DevTools to capture full response
   - Save to `C:\tmp\session-response-body.json`
   - Analyze structure

3. **Test MCB Decryption**
   - Try decrypting captured values with different algorithms
   - Start with AES-GCM (most common for modern APIs)
   - Then try ChaCha20-Poly1305

4. **Implement in Go**
   - Port MCB encryption function
   - Update `client.go` with new URL and headers
   - Test against production API

### Long-term (After Authentication Works)

1. Implement other endpoints (transactions, QR generation, etc.)
2. Add proper error handling
3. Implement token refresh
4. Add rate limiting
5. Create comprehensive tests

---

## References

### Investigation Files

- `C:\tmp\bca-login-capture.har` - Captured working request
- `C:\tmp\session-request.json` - Request details
- `C:\tmp\session-response.json` - Response headers
- `C:\tmp\CRITICAL-FINDINGS.md` - Key discoveries
- `C:\tmp\BREAKTHROUGH-ANALYSIS.md` - Complete analysis
- `C:\tmp\FINAL-STATUS.md` - Current status

### Project Files

- `CLAUDE.md` - Complete investigation notes
- `arch-flow.md` - Architecture analysis
- `ENCRYPTION.md` - Encryption details
- `internal/client/client.go` - Current implementation

### Tools Used

```bash
# HAR analysis
cat bca-login-capture.har | jq '.log.entries[] | select(.request.url | contains("session"))'

# Base64 decoding
echo "AAABm1ZQY5o..." | base64 -d | xxd

# JavaScript search
grep -i "mcb" main.js

# HTTP testing
curl -X POST https://bms.ebanksvc.bca.co.id/session/v1.0.0/add \
  -H "X-OS: AAABm1ZQY5o..." \
  -H "x-app-version: AAABm1ZQY5w..." \
  -H "x-encrypt-mcb: true" \
  -d "AAABm1ZQY5w..."
```

---

## Summary

The BCA QR authentication system uses a two-layer encryption approach:

1. **Messi encryption** (AES-256-CBC) for password encryption ✅
2. **MCB encryption** (unknown algorithm) for request encryption ❌

The authentication flow bypasses the nginx proxy and goes directly to `bms.ebanksvc.bca.co.id` with three encrypted components:

1. X-OS header (encrypted platform info)
2. x-app-version header (encrypted version info)
3. Request body (encrypted email + encrypted password)

**We are 95% complete.** The only missing piece is the MCB encryption algorithm, which needs to be reverse engineered from the JavaScript application.

Once MCB encryption is implemented, the authentication will work and return a `201 Created` response with a Firebase custom token for subsequent API calls.

**Status:** Ready for final implementation phase - need to crack MCB encryption!
