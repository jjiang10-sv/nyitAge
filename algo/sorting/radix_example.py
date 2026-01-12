import struct
from typing import List

DIGIT = 4
MAXBIT = -1 << 31  # 0x80000000

def main():
    # Simple example with mixed positive/negative numbers
    data = [-2, 1, -1, 2]
    print(f"Original data: {data}\n")
    
    radix_sort_with_trace(data)
    
    print(f"\nFinal sorted result: {data}")

def radix_sort_with_trace(data: List[int]):
    print("=== STEP 1: Convert to bytes with sign bit flip ===")
    
    ds = []  # List of byte arrays
    
    for i, e in enumerate(data):
        print(f"Original: {e} (binary: {e & 0xFFFFFFFF:032b})")
        flipped = e ^ MAXBIT
        print(f"After XOR:  {flipped} (binary: {flipped & 0xFFFFFFFF:032b})")
        
        # Convert to little-endian bytes
        b = list(struct.pack('<i', flipped))  # '<i' = little-endian signed int
        ds.append(b)
        
        print(f"As bytes: {b} (little-endian)")
        print()
    
    print("=== STEP 2: Radix sort by each byte position ===")
    
    for byte_pos in range(DIGIT):
        print(f"\n--- Processing byte position {byte_pos} ---")
        
        # Show current state
        print("Current order: ", end="")
        for i, b in enumerate(ds):
            if i > 0:
                print(", ", end="")
            print(f"{b}", end="")
        print()
        
        # Create counting sort buckets (256 buckets for each possible byte value)
        counting_sort = [[] for _ in range(256)]
        
        # Distribute into buckets
        print(f"Sorting by byte[{byte_pos}]:")
        for i, b in enumerate(ds):
            # Convert signed byte to unsigned (0-255 range)
            bucket_index = b[byte_pos] if b[byte_pos] >= 0 else b[byte_pos] + 256
            counting_sort[bucket_index].append(b)
            print(f"  Item {i}: byte[{byte_pos}]={bucket_index} -> bucket {bucket_index}")
        
        # Show non-empty buckets
        print("Non-empty buckets:")
        for k in range(256):
            if len(counting_sort[k]) > 0:
                print(f"  Bucket {k}: ", end="")
                for i, b in enumerate(counting_sort[k]):
                    if i > 0:
                        print(", ", end="")
                    print(f"{b}", end="")
                print()
        
        # Collect back in order (stable sort)
        ds.clear()
        for bucket in counting_sort:
            ds.extend(bucket)
        
        # Show result after this pass
        print("After sorting: ", end="")
        for i, b in enumerate(ds):
            if i > 0:
                print(", ", end="")
            print(f"{b}", end="")
        print()
    
    print("\n=== STEP 3: Convert back to integers and restore sign ===")
    
    for i, b in enumerate(ds):
        # Convert bytes back to integer
        w = struct.unpack('<i', bytes(b))[0]  # '<i' = little-endian signed int
        restored = w ^ MAXBIT
        
        print(f"Bytes {b} -> {w} -> {restored} (after XOR back)")
        
        data[i] = restored

if __name__ == "__main__":
    main()