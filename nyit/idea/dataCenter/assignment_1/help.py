def bcube_hosts(level: int, cube: int) -> (int, int):
    """
    Calculate the two host numbers for a given BCube level and cube index.
    
    level : BCube level (1-based)
    cube  : cube index at that level
    """
    # Calculate the first host
    first_digit = (cube // 2) * (10 ** (level - 1)) * 2
    second_digit = cube % 2
    host1 = first_digit + second_digit
    
    # Calculate second host
    level_offset = 10 ** (level - 1) * 2
    host2 = host1 + level_offset
    
    return host1, host2

for cube in range(8):
    print(f"Level 2, cube {cube}: {bcube_hosts(2, cube)}")
