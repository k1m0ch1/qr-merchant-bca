# Encryption Details for qr.klikbca.com

## Discovered Configuration

From analyzing the JavaScript (`main.2aee9935593dc1600215.js`):

```javascript
// Encryption Constants
KEY_LENGTH: 256 (0x100)
SALT_LENGTH: 16 (0x10)
ITERATION: 12 (0xc)
GCM_TAG_LENGTH: 16 (0x10)
CIPHER_ALGORITHM: 'CBC' (AES-CBC)

// Encryption Keys
key: {
  'messi': '9C0XAVRJ6PQB86TVTAD6SK6XD01PSCIK',
  'mcb': (another key),
  'oldQrms': 'kmud$b823087VAJSD0919bva^FG95JNzx{}'
}

// IV & Salt
rawIv: randomly generated WordArray (8 bytes)
rawSalt: randomly generated WordArray (8 bytes)
parsedIv: Utf8.parse(rawIv)
parsedSalt: Utf8.parse(rawSalt)
base64Iv: Base64.stringify(parsedIv)
base64Salt: Base64.stringify(parsedSalt)
```

## Encryption Functions

- `encryptMessi(plaintext)` - Encrypts data using AES-CBC
- `decryptMessi(ciphertext)` - Decrypts data using AES-CBC

## API Endpoints

```
Frontend: https://qr.klikbca.com
Backend APIs (via CSP):
  - https://mssi.ebanksvc.bca.co.id/
  - https://bms.ebanksvc.bca.co.id/

API Routes (proxied through qr.klikbca.com/api/*):
  - Login: /api/session/v1.0.0/add
  - Transactions: /api/transaction-v2/v2.0.0/list
  - Outlet List: /api/outlet/v1.0.0/list
  - User Member: /api/user/v1.0.0/member
```

## Implementation Notes

### Password Encryption
The password needs to be encrypted using:
1. AES-256-CBC encryption
2. Key: '9C0XAVRJ6PQB86TVTAD6SK6XD01PSCIK'
3. Random IV (initialization vector) - 16 bytes
4. PKCS7 padding
5. Prepend IV to ciphertext
6. Base64 encode the result (IV + ciphertext)

The encrypted password is then sent in the JSON payload to the login API.

### Implementation Status ✅

**Completed:**
- ✅ AES-256-CBC encryption implemented in Go (`internal/client/crypto.go`)
- ✅ Password encryption with random IV and PKCS7 padding
- ✅ Base64 encoding of encrypted data
- ✅ CLI tool with `login` and `transactions` commands
- ✅ Session persistence to file
- ✅ Cookie jar for session management

**Blocking Issue:**
- ⚠️ API endpoint returns `405 Not Allowed` from nginx
- The server appears to have additional protection beyond standard CORS
- May require CSRF tokens, specific session cookies, or other authentication
- Need to capture actual browser requests to understand complete flow

**Investigation Results (2025-12-25):**

1. **Firebase Configuration Discovered:**
   - API Key: `AIzaSyAnGQ6VXc8JTKCg84mtIWalDGs1hdxgVrY`
   - Project ID: `merchant-bca`
   - Password authentication is DISABLED on Firebase
   - Uses custom token authentication flow

2. **Backend Discovery:**
   - Direct backend: `https://bms.ebanksvc.bca.co.id`
   - Alternative: `https://mssi.ebanksvc.bca.co.id`
   - Both require Bearer token authentication
   - Login endpoint paradox: login requires auth!

3. **Test Results:**
   - ✅ Encryption implementation works correctly
   - ❌ Nginx proxy blocks all requests with 405
   - ❌ Direct backend requires Bearer token
   - ❌ Firebase standard auth is disabled

**Authentication Flow Hypothesis:**
1. Frontend encrypts password using AES-256-CBC
2. Backend validates and creates Firebase custom token
3. Frontend uses custom token to get Firebase ID token
4. ID token used as Bearer for all API calls

**Critical Gap:** Unknown how initial auth request gets authorized
- Possible CSRF token from page load
- Possible session cookie requirement
- Possible OAuth client credentials flow
- Possible request signature mechanism

**Next Steps:**
1. **PRIORITY:** Capture real browser network traffic with Playwright
2. Analyze complete request/response cycle including:
   - All headers (including Angular-specific)
   - All cookies and their sequence
   - CSRF tokens or nonces
   - Request signatures
3. Reverse engineer Angular AuthService
4. Update Go client to replicate exact browser behavior

**Documentation:**
- See `arch-flow.md` for complete authentication flow analysis
- See `CLAUDE.md` for credentials and configuration
