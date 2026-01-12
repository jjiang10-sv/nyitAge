def quick_sort(nums,start,end):
    
    if start < end:
        pivot_idx = partition(nums=nums, start=start, end=end)
        #print("after the swaping, nums, i+1", nums, i+1)
        quick_sort(nums=nums,start=start,end=pivot_idx-1)
        quick_sort(nums=nums,start=pivot_idx+1, end=end)

def partition(nums, start, end):
    # [9,2,3,54,32,3,9,10,8]
    # [3,2,3,54,32,9,9,10,8]
    pivot = nums[end]
    i = start -1
    for j in range(start,end):
        if nums[j] <= pivot:
            i +=1
            nums[i], nums[j] = nums[j], nums[i]
    nums[i+1], nums[end] = nums[end], nums[i+1]
    return i+1
        
    
        
    
def test_quick_sort():
    # Helper to call quick_sort and return the sorted list
    def sort_and_return(nums):
        arr = nums[:]
        quick_sort(arr,0,len(arr)-1)
        return arr

    # Test 1: Empty list
    assert sort_and_return([]) == []

    # Test 2: Single element
    assert sort_and_return([1]) == [1]

    # Test 3: Already sorted
    assert sort_and_return([1, 2, 3, 4, 5]) == [1, 2, 3, 4, 5]

    # Test 4: Reverse sorted
    result =  sort_and_return([5, 4, 3, 2, 1]) 
    print(result)
    assert result == [1, 2, 3, 4, 5]

    # Test 5: All elements the same
    assert sort_and_return([7, 7, 7, 7]) == [7, 7, 7, 7]

    # Test 6: Random order
    assert sort_and_return([3, 1, 4, 1, 5, 9, 2, 6, 5]) == sorted([3, 1, 4, 1, 5, 9, 2, 6, 5])

    # Test 7: Negative numbers
    assert sort_and_return([-3, -1, -4, -1, -5, -9, -2, -6, -5]) == sorted([-3, -1, -4, -1, -5, -9, -2, -6, -5])

    # Test 8: Mixed positive and negative
    assert sort_and_return([0, -1, 3, -2, 2, 1]) == sorted([0, -1, 3, -2, 2, 1])

    # Test 9: Large numbers
    assert sort_and_return([1000000, 999999, 1234567, 0, -1000000]) == sorted([1000000, 999999, 1234567, 0, -1000000])

    print("All quick_sort test cases passed.")

# Run the tests
test_quick_sort()
            
