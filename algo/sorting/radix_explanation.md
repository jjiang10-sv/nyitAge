# Radix Sort Example: [-2, 1, -1, 2]

## The Magic of Sign Bit Flipping

The key insight is that `maxbit = -1 << 31 = 0x80000000` (sign bit mask).

When we XOR with this mask:
- **Negative numbers**: Sign bit 1 → 0 (become smaller values)
- **Positive numbers**: Sign bit 0 → 1 (become larger values)

## Step-by-Step Trace

### Original Data: [-2, 1, -1, 2]

### Step 1: XOR with Sign Bit Mask

| Original | Binary (32-bit) | After XOR | New Value |
|----------|----------------|-----------|-----------|
| -2 | `11111111111111111111111111111110` | `01111111111111111111111111111110` | 2147483646 |
| 1  | `00000000000000000000000000000001` | `10000000000000000000000000000001` | -2147483647 |
| -1 | `11111111111111111111111111111111` | `01111111111111111111111111111111` | 2147483647 |
| 2  | `00000000000000000000000000000010` | `10000000000000000000000000000010` | -2147483646 |

**Key Insight**: After XOR, all original negatives now have smaller values than original positives!

### Step 2: Convert to Little-Endian Bytes

| XOR Value | Little-Endian Bytes |
|-----------|-------------------|
| 2147483646 | `[254, 255, 255, 127]` |
| -2147483647 | `[1, 0, 0, 128]` |
| 2147483647 | `[255, 255, 255, 127]` |
| -2147483646 | `[2, 0, 0, 128]` |

### Step 3: Radix Sort by Each Byte Position

**Byte Position 0 (Least Significant):**
- Bucket 1: `[1,0,0,128]` (from 1)
- Bucket 2: `[2,0,0,128]` (from 2)  
- Bucket 254: `[254,255,255,127]` (from -2)
- Bucket 255: `[255,255,255,127]` (from -1)

**Result**: `[1,0,0,128], [2,0,0,128], [254,255,255,127], [255,255,255,127]`

**Byte Positions 1 & 2**: No change (same byte values within groups)

**Byte Position 3 (Most Significant - Sign Bit):**
- Bucket 127: `[254,255,255,127], [255,255,255,127]` (original negatives)
- Bucket 128: `[1,0,0,128], [2,0,0,128]` (original positives)

**Final byte order**: `[254,255,255,127], [255,255,255,127], [1,0,0,128], [2,0,0,128]`

### Step 4: Convert Back and Restore Sign

| Bytes | Intermediate | After XOR Back | Original |
|-------|-------------|----------------|----------|
| `[254,255,255,127]` | 2147483646 | -2 | ✓ |
| `[255,255,255,127]` | 2147483647 | -1 | ✓ |
| `[1,0,0,128]` | -2147483647 | 1 | ✓ |
| `[2,0,0,128]` | -2147483646 | 2 | ✓ |

## Final Result: [-2, -1, 1, 2] ✅

## Why This Works

1. **Sign bit flip** ensures negatives sort before positives
2. **Two's complement** ordering is preserved within each group
3. **Little-endian** byte order means we naturally do LSD radix sort
4. **Stable sorting** at each byte position maintains relative order
5. **Final XOR** restores original values

The algorithm cleverly transforms the problem so that a simple byte-wise radix sort produces the correct final ordering for signed integers!