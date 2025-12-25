# CRITICAL AUTHENTICATION DISCOVERY

## The Missing Piece Found!

### Request Details

**URL**: `https://bms.ebanksvc.bca.co.id/session/v1.0.0/add` (DIRECT, not through proxy!)  
**Method**: POST  
**Status**: 201 Created ✅

### Critical Headers (NEW!)

```
X-OS: AAABm1ZQY5oMGykxTTQ0-lEOEJXjJelCsSQSei1RfGkzA0vUpM9luL8MI_vwUlWnhA1JmJK7hg
x-app-version: AAABm1ZQY5wMGykxTTQ0-lEOEJXjJelCHp7PJiVQcbDOLyyOER1_FfzFGR6AZkKiq6RgJ8iGHq4
x-encrypt-mcb: true
Content-Type: application/json
Origin: https://qr.klikbca.com
Referer: https://qr.klikbca.com/
```

### THE KEY DISCOVERY: Encrypted Payload

**The ENTIRE request body is encrypted** - Not JSON with email/password fields!

```
Payload (100 bytes): 
AAABm1ZQY5wMGykxTTQ0-lEOEJXjJelCiPGRvbI3Ud7aGZgVQ3M3fFVOEfIyv8cAIJFyVMSbqRCgi-RxtKNcx4jqoFPYHz5DtSnI
```

This is NOT:
```json
{
  "email": "someemail@gmail.com",
  "password": "encrypted_password"
}
```

This IS:
```
Single encrypted blob containing everything
```

## What This Means

1. **No 405 Error Anymore**: Request goes DIRECTLY to `bms.ebanksvc.bca.co.id`, not through nginx proxy
2. **Special Headers Required**: `X-OS` and `x-app-version` are mandatory (encrypted/encoded)
3. **Full Payload Encryption**: The entire JSON payload is encrypted as one block
4. **Different Encryption**: This might be the "MCB" encryption (`x-encrypt-mcb: true`)

## Next Steps

1. Find how `X-OS` and `x-app-version` headers are generated
2. Find the "MCB" encryption function (not just "Messi" encryption)
3. Understand what data goes into the encrypted payload
4. Decrypt these header values to understand their format
5. Implement the complete encryption flow in Go

## Header Analysis

Both `X-OS` and `x-app-version` start with `AAABm1ZQY5`:
- Could be timestamp/nonce prefix
- Followed by encrypted data
- Base64url encoded (uses `-` and `_`)

Length:
- X-OS: 70 chars
- x-app-version: 75 chars  
- Payload: 100 chars

All use same encoding scheme!
