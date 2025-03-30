public class sort {
    static void bubleSort(int arr[], int n) {
        int i,j,temp;
        boolean swapped;
        for (i = 0; i< n-1;i++) {
            swapped = false;
            for (j=0;j<n-i-1;j++){
                if (arr[j] > arr[j+1]) {
                    temp = arr[j];
                    arr[j] = arr[j+1];
                    arr[j+1] = temp;
                    swapped = true;
                }
            };
            // all has been sorted 
            if (swapped) {
                break;
            }
        }
    }

    static void selectSort(int arr[]) {
        int n = arr.length;
        for (int i = 0; i < n; i++) {
            int min_idx = i;
            for (int j = i+1; j < n; j++) {
                if (arr[j] < arr[min_idx]) {
                    // update min_idx
                    min_idx = j;
                }
            }
            //swap mind_idx and i
            int temp = arr[i];
            arr[i] = arr[min_idx];
            arr[min_idx] = temp;
        }
    }

    static void insertionSort(int arr[]) {
        int n = arr.length;
        for (int i = 1; i<n;i++){
            int key = arr[i];
            int j = i-1;
            while (j>=0&&arr[j]>key){
                // move the j item to right 
                arr[j+1] = arr[j];
                j -=1;
            }
            arr[j+1] = key;
        }
    }
    // low, middle and high
    static void merge(int arr[], int l, int m, int r) {
        int n1 = m-l+1;
        int n2 = r-m;
        
        int L[] = new int[n1];
        int R[] = new int[n2];

        for (int i = 0; i < n1; i++) {
            L[i] = arr[l+i];
        }
        for (int j = 0; j < n2; j++) {
            R[j] = arr[m+j];
        }
        int i= 0,j=0;
        int k = l;
        while (i < n1 && j < n2) {
            if (L[i] <= R[j]) {
                arr[k] = L[i];
                i++;
            }else{
                arr[k] = R[j];
                j++;
            }
            k++;
        }
    }

    static void mergeSort(int arr[], int l, int r) {
        if (l < r) {
            int m = (l+r)/2;
            //int m = l + (r-1)/2;
            mergeSort(arr, l, m);
            mergeSort(arr, m+1, r);
            merge(arr, l, m, r);
        }
    }
    // The `countingSortByDigit` method implements a digit-based counting sort, which is a stable sorting algorithm used as a subroutine in the Radix Sort algorithm. Here's a breakdown of how it works and why it is effective:

    // ### Explanation of the Code
    
    // 1. **Input Parameters**:
    //    - `int[] arr`: The array of integers to be sorted.
    //    - `int exp`: The exponent representing the current digit's place value (1 for units, 10 for tens, 100 for hundreds, etc.).
    
    // 2. **Initialization**:
    //    - `int n = arr.length;`: Get the length of the input array.
    //    - `int[] output = new int[n];`: Create an output array to hold the sorted numbers.
    //    - `int[] count = new int[10];`: Create a count array to store the frequency of each digit (0-9).
    
    // 3. **Counting Occurrences**:
    //    - The first loop iterates through each number in the input array and calculates the digit at the current place value (determined by `exp`). It increments the corresponding index in the `count` array.
    //    ```java
    //    int digit = (num / exp) % 10;
    //    count[digit]++;
    //    ```
    
    // 4. **Cumulative Count**:
    //    - The second loop modifies the `count` array to store the cumulative count of digits. This means that each index in the `count` array will now represent the position of that digit in the output array.
    //    ```java
    //    for (int i = 1; i < 10; i++) {
    //        count[i] += count[i - 1];
    //    }
    //    ```
    
    // 5. **Building the Output Array**:
    //    - The third loop builds the output array in reverse order to maintain stability (i.e., the relative order of equal elements is preserved). It uses the cumulative counts to place each number in its correct position in the output array.
    //    ```java
    //    for (int i = n - 1; i >= 0; i--) {
    //        int digit = (arr[i] / exp) % 10;
    //        output[count[digit] - 1] = arr[i];
    //        count[digit]--;
    //    }
    //    ```
    
    // 6. **Copying Back to Original Array**:
    //    - Finally, the output array is copied back to the original array, effectively sorting the numbers based on the current digit.
    //    ```java
    //    System.arraycopy(output, 0, arr, 0, n);
    //    ```
    
    // ### Why It Works
    
    // - **Stable Sorting**: Counting sort is stable, meaning that it preserves the order of equal elements. This is crucial for Radix Sort, which processes digits from least significant to most significant.
      
    // - **Efficiency**: The time complexity of counting sort is \(O(n + k)\), where \(n\) is the number of elements in the input array and \(k\) is the range of the input (in this case, 10 for digits 0-9). This makes it very efficient for sorting integers with a limited range.
    
    // - **Digit-by-Digit Sorting**: By sorting the numbers digit by digit, starting from the least significant digit, Radix Sort can sort numbers of arbitrary length efficiently. Each pass through the digits uses counting sort, which is efficient for the limited range of digits.
    
    // In summary, the `countingSortByDigit` method is a key component of the Radix Sort algorithm, allowing it to sort numbers based on individual digits while maintaining stability and efficiency.
    
    public static void radixSort(int[] arr) {
        if (arr.length == 0) return;

        int max = getMax(arr);
        // 1 10 100 1000
        for (int exp = 1; max / exp > 0; exp *= 10) {
            countingSortByDigit(arr, exp);
        }
    }

    private static int getMax(int[] arr) {
        int max = arr[0];
        for (int num : arr) {
            if (num > max) max = num;
        }
        return max;
    }

    private static void countingSortByDigit(int[] arr, int exp) {
        int n = arr.length;
        int[] output = new int[n];
        int[] count = new int[10];

        // Count occurrences of each digit
        for (int num : arr) {
            int digit = (num / exp) % 10;
            count[digit]++;
        }

        // Compute cumulative counts
        for (int i = 1; i < 10; i++) {
            count[i] += count[i - 1];
        }

        // Build output array in reverse for stability
        for (int i = n - 1; i >= 0; i--) {
            int digit = (arr[i] / exp) % 10;
            output[count[digit] - 1] = arr[i];
            count[digit]--;
        }

        // Copy output back to original array
        System.arraycopy(output, 0, arr, 0, n);
    }

    static void printArray(int arr[], int size) {
        int i;
        for (i=0; i< size;i++){
            System.out.print(arr[i] + " ");
        }
        System.out.println("Done");
    }

    public static void main(String args[]){
        int arr[] = {343,43,21,43,65,76,32,1,23,87};
        int n = arr.length;
        // bubleSort(arr, n);
        // System.out.println("sorted array ");
        printArray(arr, n);
        //selectSort(arr);
        //insertionSort(arr);
        mergeSort(arr, 0, n-1);
        printArray(arr, n);


    }
}

// Time Complexity:

// Radix Sort: O(d*(n + k)), where:

// d: Number of digits in the maximum element.

// n: Number of elements.

// k: Range of digits (10 for base 10).

// Simplified to O(d*n) since k is constant. For fixed-size integers (e.g., 32-bit), d is constant, leading to linear time O(n).

// Comparison with Quadratic Sorts:

// Insertion/Bubble/Selection Sort: Average and worst-case time complexity O(n²) due to pairwise comparisons.

// Radix Sort Advantage: Avoids O(n²) comparisons by exploiting digit structure. Processes each digit in linear time, resulting in O(n) for large datasets with fixed digit counts. This makes Radix Sort significantly faster for large n despite higher constant factors and space usage.

// Conclusion:

// Radix Sort excels in sorting large integer datasets by leveraging digit-wise processing, achieving linear time complexity under fixed digit constraints. It outperforms quadratic comparison-based sorts (insertion, bubble, selection) for large n, albeit with increased space requirements and data-type restrictions.


// **Bucket Sort Algorithm Explanation:**

// Bucket Sort is a distribution-based sorting algorithm that divides elements into multiple "buckets," sorts each bucket individually, and then concatenates the sorted buckets. It works best when the input is **uniformly distributed**. The steps are:

// 1. **Determine Range and Create Buckets:**
//    - Find the minimum and maximum values in the array.
//    - Create a fixed number of empty buckets (e.g., equal to the array size).

// 2. **Scatter Elements into Buckets:**
//    - Distribute elements into buckets based on their value. For example, elements in a specific range go to a corresponding bucket.

// 3. **Sort Individual Buckets:**
//    - Sort each bucket using a simple algorithm like Insertion Sort or a built-in sort.

// 4. **Concatenate Buckets:**
//    - Merge all sorted buckets back into the original array.

// ---

// **Java Code:**

// ```java
// import java.util.ArrayList;
// import java.util.Collections;
// import java.util.List;

// public class BucketSort {

//     public static void bucketSort(int[] arr) {
//         if (arr.length == 0) return;

//         // Find min and max values
//         int min = arr[0];
//         int max = arr[0];
//         for (int num : arr) {
//             if (num < min) min = num;
//             if (num > max) max = num;
//         }

//         // Determine number of buckets (using array length)
//         int numBuckets = arr.length;
//         List<List<Integer>> buckets = new ArrayList<>(numBuckets);
//         for (int i = 0; i < numBuckets; i++) {
//             buckets.add(new ArrayList<>());
//         }

//         // Distribute elements into buckets
//         for (int num : arr) {
//             int bucketIndex;
//             if (max == min) {
//                 bucketIndex = 0; // All elements are the same
//             } else {
//                 bucketIndex = (int) ((num - min) * (numBuckets - 1L) / (max - min));
//             }
//             buckets.get(bucketIndex).add(num);
//         }

//         // Sort each bucket and merge
//         int index = 0;
//         for (List<Integer> bucket : buckets) {
//             Collections.sort(bucket); // Use built-in sort (TimSort)
//             for (int num : bucket) {
//                 arr[index++] = num;
//             }
//         }
//     }

//     public static void main(String[] args) {
//         int[] arr = {30, 40, 10, 80, 5, 12, 35};
//         bucketSort(arr);
//         System.out.println("Sorted array:");
//         for (int num : arr) {
//             System.out.print(num + " ");
//         }
//     }
// }
// ```

// ---

// **Time Complexity:**

// - **Best Case:** O(n)  
//   Achieved when elements are uniformly distributed, and each bucket contains roughly the same number of elements. Sorting each bucket takes constant time (if buckets are small).

// - **Average Case:** O(n + k)  
//   Where `k` is the number of buckets. This assumes a uniform distribution and efficient sorting of buckets (e.g., O(1) per bucket).

// - **Worst Case:** O(n²)  
//   Occurs when all elements are placed into a single bucket, reducing Bucket Sort to the complexity of the sorting algorithm used for buckets (e.g., Insertion Sort).

// ---

// **Comparison with Quadratic Sorts:**

// | Algorithm       | Average Time Complexity | Why Bucket Sort is Faster                                                                                                                                 |
// |-----------------|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
// | **Bucket Sort** | O(n + k)                | Breaks the problem into smaller subproblems (buckets). For uniform data, each bucket is small, and sorting them takes linear or near-linear time overall. |
// | Insertion Sort  | O(n²)                   | Performs pairwise swaps for each element. Inefficient for large datasets.                                                                                |
// | Bubble Sort     | O(n²)                   | Repeatedly swaps adjacent elements, leading to excessive comparisons.                                                                                    |
// | Selection Sort  | O(n²)                   | Scans the unsorted portion for the minimum element, requiring O(n²) comparisons.                                                                          |

// **Key Advantages of Bucket Sort:**
// - Exploits **data distribution** to reduce the problem size.
// - Achieves **linear time** for uniformly distributed data.
// - Avoids costly pairwise comparisons used in quadratic algorithms.

// **Limitations:**
// - Requires prior knowledge of data range.
// - Performance degrades for non-uniformly distributed data.