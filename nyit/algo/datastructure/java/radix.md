The `countingSortByDigit` method is a crucial part of the Radix Sort algorithm, which sorts integers by processing individual digits. Here's a breakdown of how this method works and why it is effective:

### Overview of Counting Sort by Digit

1. **Purpose**: The method sorts the array `arr` based on the digit represented by `exp` (exponent), which indicates the current digit's place value (units, tens, hundreds, etc.).

2. **Input Parameters**:
   - `int[] arr`: The array of integers to be sorted.
   - `int exp`: The current digit's place value (1 for units, 10 for tens, etc.).

### Steps in the Method

1. **Initialization**:
   - `int n = arr.length;`: Get the length of the input array.
   - `int[] output = new int[n];`: Create an output array to hold the sorted numbers.
   - `int[] count = new int[10];`: Create a count array to store the frequency of each digit (0-9).

2. **Count Occurrences**:
   - The first loop iterates through each number in `arr`, calculates the digit at the current place value using `(num / exp) % 10`, and increments the corresponding index in the `count` array. This effectively counts how many times each digit appears in the current place.

3. **Cumulative Count**:
   - The second loop modifies the `count` array to store the cumulative count of digits. This means that each index in `count` will now represent the position of that digit in the output array. For example, if `count[2]` is 5, it means that the digit '2' will occupy the first 5 positions in the output array.

4. **Build Output Array**:
   - The third loop constructs the output array in reverse order. This is done to maintain the stability of the sort (i.e., if two numbers have the same digit, their relative order from the original array is preserved). It uses the cumulative counts to place each number in the correct position in the `output` array.

5. **Copy Back to Original Array**:
   - Finally, the method copies the sorted numbers from the `output` array back to the original `arr`, effectively sorting the array based on the current digit.

### Why It Works

- **Stability**: The method is stable, meaning that it preserves the relative order of equal elements. This is crucial for Radix Sort, which processes digits from the least significant to the most significant.
  
- **Efficiency**: Counting Sort operates in linear time \( O(n + k) \), where \( n \) is the number of elements in the array and \( k \) is the range of the input (in this case, 10 for digits 0-9). This makes it very efficient for sorting integers.

- **Digit-by-Digit Sorting**: By sorting the numbers digit by digit, Radix Sort can handle larger numbers effectively, as it breaks down the sorting process into manageable parts.

In summary, the `countingSortByDigit` method is a well-structured implementation of the Counting Sort algorithm tailored for Radix Sort, allowing for efficient and stable sorting of integers based on their individual digits.


The provided Java code implements the **Bucket Sort** algorithm, which is a distribution-based sorting algorithm. Here's a breakdown of how the code works:

### Overview of Bucket Sort
Bucket Sort works by distributing the elements of an array into a number of buckets. Each bucket is then sorted individually, either using a different sorting algorithm or recursively applying the bucket sort. Finally, the sorted buckets are concatenated to form the final sorted array.

### Code Explanation

1. **Method Declaration**:
   ```java
   public static void bucketSort(int[] arr) {
   ```
   This declares a static method `bucketSort` that takes an array of integers as input.

2. **Check for Empty Array**:
   ```java
   if (arr.length == 0) return;
   ```
   If the input array is empty, the method returns immediately, as there is nothing to sort.

3. **Finding Minimum and Maximum Values**:
   ```java
   int min = arr[0];
   int max = arr[0];
   for (int num : arr) {
       if (num < min) min = num;
       if (num > max) max = num;
   }
   ```
   The code initializes `min` and `max` to the first element of the array and then iterates through the array to find the actual minimum and maximum values. This is important for determining the range of values to distribute into buckets.

4. **Creating Buckets**:
   ```java
   int numBuckets = arr.length;
   List<List<Integer>> buckets = new ArrayList<>(numBuckets);
   for (int i = 0; i < numBuckets; i++) {
       buckets.add(new ArrayList<>());
   }
   ```
   The number of buckets is set to the length of the array. A list of lists (buckets) is created, where each inner list will hold the elements that fall into that bucket.

5. **Distributing Elements into Buckets**:
   ```java
   for (int num : arr) {
       int bucketIndex;
       if (max == min) {
           bucketIndex = 0; // All elements are the same
       } else {
           bucketIndex = (int) ((num - min) * (numBuckets - 1L) / (max - min));
       }
       buckets.get(bucketIndex).add(num);
   }
   ```
   Each element in the array is placed into a bucket based on its value. If all elements are the same (i.e., `max` equals `min`), they all go into the first bucket. Otherwise, the bucket index is calculated using a formula that maps the value of the element to the appropriate bucket based on the range of values.

6. **Sorting Each Bucket and Merging**:
   ```java
   int index = 0;
   for (List<Integer> bucket : buckets) {
       Collections.sort(bucket); // Use built-in sort (TimSort)
       for (int num : bucket) {
           arr[index++] = num;
       }
   }
   ```
   Each bucket is sorted using the built-in `Collections.sort()` method (which uses TimSort). After sorting, the elements from each bucket are copied back into the original array in order.

### Conclusion
The `bucketSort` method effectively sorts an array of integers by distributing them into buckets, sorting each bucket, and then merging the results. This algorithm is particularly efficient when the input is uniformly distributed over a range, as it can achieve linear time complexity under ideal conditions.


**Bucket Sort Algorithm Explanation:**

Bucket Sort is a distribution-based sorting algorithm that divides elements into multiple "buckets," sorts each bucket individually, and then concatenates the sorted buckets. It works best when the input is **uniformly distributed**. The steps are:

1. **Determine Range and Create Buckets:**
   - Find the minimum and maximum values in the array.
   - Create a fixed number of empty buckets (e.g., equal to the array size).

2. **Scatter Elements into Buckets:**
   - Distribute elements into buckets based on their value. For example, elements in a specific range go to a corresponding bucket.

3. **Sort Individual Buckets:**
   - Sort each bucket using a simple algorithm like Insertion Sort or a built-in sort.

4. **Concatenate Buckets:**
   - Merge all sorted buckets back into the original array.

---

**Java Code:**

```java
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

public class BucketSort {

    public static void bucketSort(int[] arr) {
        if (arr.length == 0) return;

        // Find min and max values
        int min = arr[0];
        int max = arr[0];
        for (int num : arr) {
            if (num < min) min = num;
            if (num > max) max = num;
        }

        // Determine number of buckets (using array length)
        int numBuckets = arr.length;
        List<List<Integer>> buckets = new ArrayList<>(numBuckets);
        for (int i = 0; i < numBuckets; i++) {
            buckets.add(new ArrayList<>());
        }

        // Distribute elements into buckets
        for (int num : arr) {
            int bucketIndex;
            if (max == min) {
                bucketIndex = 0; // All elements are the same
            } else {
                bucketIndex = (int) ((num - min) * (numBuckets - 1L) / (max - min));
            }
            buckets.get(bucketIndex).add(num);
        }

        // Sort each bucket and merge
        int index = 0;
        for (List<Integer> bucket : buckets) {
            Collections.sort(bucket); // Use built-in sort (TimSort)
            for (int num : bucket) {
                arr[index++] = num;
            }
        }
    }

    public static void main(String[] args) {
        int[] arr = {30, 40, 10, 80, 5, 12, 35};
        bucketSort(arr);
        System.out.println("Sorted array:");
        for (int num : arr) {
            System.out.print(num + " ");
        }
    }
}
```

---

**Time Complexity:**

- **Best Case:** O(n)  
  Achieved when elements are uniformly distributed, and each bucket contains roughly the same number of elements. Sorting each bucket takes constant time (if buckets are small).

- **Average Case:** O(n + k)  
  Where `k` is the number of buckets. This assumes a uniform distribution and efficient sorting of buckets (e.g., O(1) per bucket).

- **Worst Case:** O(n²)  
  Occurs when all elements are placed into a single bucket, reducing Bucket Sort to the complexity of the sorting algorithm used for buckets (e.g., Insertion Sort).

---

**Comparison with Quadratic Sorts:**

| Algorithm       | Average Time Complexity | Why Bucket Sort is Faster                                                                                                                                 |
|-----------------|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Bucket Sort** | O(n + k)                | Breaks the problem into smaller subproblems (buckets). For uniform data, each bucket is small, and sorting them takes linear or near-linear time overall. |
| Insertion Sort  | O(n²)                   | Performs pairwise swaps for each element. Inefficient for large datasets.                                                                                |
| Bubble Sort     | O(n²)                   | Repeatedly swaps adjacent elements, leading to excessive comparisons.                                                                                    |
| Selection Sort  | O(n²)                   | Scans the unsorted portion for the minimum element, requiring O(n²) comparisons.                                                                          |

**Key Advantages of Bucket Sort:**
- Exploits **data distribution** to reduce the problem size.
- Achieves **linear time** for uniformly distributed data.
- Avoids costly pairwise comparisons used in quadratic algorithms.

**Limitations:**
- Requires prior knowledge of data range.
- Performance degrades for non-uniformly distributed data.