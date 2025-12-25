# BREAKTHROUGH: Complete Authentication Flow Discovered!

## Date: 2025-12-25

## The Complete Picture

### What Actually Happens

The login does NOT go through the nginx proxy! It goes **directly** to the backend:

```
POST https://bms.ebanksvc.bca.co.id/session/v1.0.0/add
```

### Required Headers (CRITICAL!)

```http
X-OS: AAABm1ZQY5oMGykxTTQ0-lEOEJXjJelCsSQSei1RfGkzA0vUpM9luL8MI_vwUlWnhA1JmJK7hg
x-app-version: AAABm1ZQY5wMGykxTTQ0-lEOEJXjJelCHp7PJiVQcbDOLyyOER1_FfzFGR6AZkKiq6RgJ8iGHq4
x-encrypt-mcb: true
Content-Type: application/json
Origin: https://qr.klikbca.com
Referer: https://qr.klikbca.com/
```

### Encrypted Payload (ENTIRE REQUEST)

**NOT**  this (what we thought):
```json
{
  "email": "someemail@gmail.com",
  "password": "base64_encrypted_password_here"
}
```

**BUT** this (what it actually is):
```
AAABm1ZQY5wMGykxTTQ0-lEOEJXjJelCiPGRvbI3Ud7aGZgVQ3M3fFVOEfIyv8cAIJFyVMSbqRCgi-RxtKNcx4jqoFPYHz5DtSnI
```

A single encrypted blob!

## Binary Analysis

Decoding the `X-OS` header:
```
Base64: AAABm1ZQY5oMGykxTTQ0-lEOEJXjJelCsSQSei1RfGkzA0vUpM9luL8MI_vwUlWnhA1JmJK7hg
Hex:    00 00 01 9b 56 50 63 9a 0c 1b 29 31 4d 34 34 ...
```

**Structure Hypothesis:**
- Bytes 0-3: Header/version/timestamp (00 00 01 9b = 411 decimal)
- Bytes 4+: Encrypted data

**All three use same prefix:** `AAABm1ZQY5` → `00 00 01 9b 56 50 63 9a`

This suggests:
1. Same encryption key for all three
2. Same timestamp/nonce
3. Different plaintext data

## The MCB Encryption

From JavaScript references and header `x-encrypt-mcb: true`, we know there's a separate MCB encryption (not the "Messi" encryption we found earlier).

**Two encryption systems:**
1. **Messi encryption**: AES-256-CBC with key `9C0XAVRJ6PQB86TVTAD6SK6XD01PSCIK`
2. **MCB encryption**: Unknown (this is what's used for the actual request!)

## What Needs to be Encrypted

### For `X-OS` Header:
Likely: OS information (Windows, version, etc.)

### For `x-app-version` Header:
Likely: App version number

### For Request Body:
The JSON:
```json
{
  "email": "saomeemail@gmail.com",
  "password": "somepassword"  // or already encrypted with Messi?
}
```

## Pattern Analysis

All three encrypted values:
- Use base64url encoding (with `-` and `_`)
- Start with same 4-byte header: `00 00 01 9b`
- Followed by encrypted data
- No obvious IV prepended (unlike Messi encryption)

## Next Steps to Implement

1. **Find MCB encryption key**
   - Search JavaScript for "mcb" key
   - Look for `generateKeyMcb` or similar
   - Check if it's related to the "mcb" key mentioned in ENCRYPTION.md

2. **Understand encryption parameters**
   - Algorithm (likely AES, but which mode?)
   - IV generation (static? derived?)
   - Padding scheme

3. **Identify what goes into each encrypted value**
   - X-OS: What data?
   - x-app-version: What data?
   - Payload: Is password pre-encrypted with Messi first?

4. **Implement in Go**
   - MCB encryption function
   - Header generation
   - Complete request builder

## Key Files to Search in JS

Look for:
- `encrypt.*mcb` (case insensitive)
- `generateKeyMcb` or `generateKeyMcbV2`
- Header value: `00 00 01 9b` (might be in hex form)
- `x-encrypt-mcb`
- `X-OS` header generation

## Why We Got 405 Before

We were calling:
```
POST https://qr.klikbca.com/api/session/v1.0.0/add
```

But should call:
```
POST https://bms.ebanksvc.bca.co.id/session/v1.0.0/add
```

**With the special encrypted headers!**

## Success Indicator

The captured request returned:
- Status: **201 Created** ✅
- This means the authentication worked!

## Summary

We now know:
1. ✅ Correct URL (direct backend, not proxy)
2. ✅ Required headers (X-OS, x-app-version, x-encrypt-mcb)
3. ✅ Payload format (fully encrypted, not JSON)
4. ❌ MCB encryption algorithm (NEED TO FIND)
5. ❌ MCB encryption key (NEED TO FIND)
6. ❌ Data structure before encryption (NEED TO FIND)

## Confidence Level

**90% of the way there!**

The only missing piece is the MCB encryption implementation. Once we find that in the JavaScript, we can replicate the exact request and authenticate successfully.
