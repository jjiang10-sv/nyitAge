

// public class RadixSort {

//     public static void radixSort(int[] arr) {
//         if (arr.length == 0) return;
//         // found the max value of the array
//         int max = getMax(arr);
//         // loop through the digits in the max val
//         for (int exp = 1; max / exp > 0; exp *= 10) {
//             countingSortByDigit(arr, exp);
//         }
//     }

//     private static int getMax(int[] arr) {
//         int max = arr[0];
//         for (int num : arr) {
//             if (num > max) max = num;
//         }
//         return max;
//     }

//     private static void countingSortByDigit(int[] arr, int exp) {
//         int n = arr.length;
//         //Create an output array to hold the sorted numbers.
//         int[] output = new int[n];
//         //Create a count array to store the frequency of each digit (0-9).
//         int[] count = new int[10];

//         //The first loop iterates through each number in arr, 
//         //calculates the digit at the current place value 
//         //using (num / exp) % 10, and increments the corresponding index 
//         //in the count array. This effectively counts how many times 
//         //each digit appears in the current place.
//         for (int num : arr) {
//             int digit = (num / exp) % 10;
//             count[digit]++;
//         }

//         //The second loop modifies the count array to store the cumulative
//         // count of digits. This means that each index in count will 
//         //now represent the position of that digit in the output array.
//         // For example, if count[2] is 5, it means that the digit '2' will
//         // occupy the first 5 positions in the output array.
//         for (int i = 1; i < 10; i++) {
//             count[i] += count[i - 1];
//         }

//         // Build output array in reverse for stability

//         //The third loop constructs the output array in reverse order. 
//         //This is done to maintain the stability of the sort 
//         //(i.e., if two numbers have the same digit, their relative order 
//         //from the original array is preserved). It uses the cumulative 
//         //counts to place each number in the correct position 
//         //in the `output` array.
//         for (int i = n - 1; i >= 0; i--) {
//             int digit = (arr[i] / exp) % 10;
//             output[count[digit] - 1] = arr[i];
//             count[digit]--;
//         }

//         // Copy output back to original array
//         System.arraycopy(output, 0, arr, 0, n);
//     }

//     public static void main(String[] args) {
//         int[] arr = {170, 45, 75, 90, 802, 24, 2, 66};
//         radixSort(arr);
//         System.out.println("Sorted array:");
//         for (int num : arr) {
//             System.out.print(num + " ");
//         }
//     }
// }

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
        //The number of buckets is set to the length of the array. 
        //A list of lists (buckets) is created, where each inner list will 
        //hold the elements that fall into that bucket.
        int numBuckets = arr.length;
        List<List<Integer>> buckets = new ArrayList<>(numBuckets);
        for (int i = 0; i < numBuckets; i++) {
            buckets.add(new ArrayList<>());
        }

        // Distribute elements into buckets
        //Each element in the array is placed into a bucket based 
        //on its value. If all elements are the same (i.e., max equals min), 
        //they all go into the first bucket. Otherwise, 
        //the bucket index is calculated using a formula that 
        //maps the value of the element to the appropriate bucket 
        //based on the range of values.
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
        //Each bucket is sorted using the built-in Collections.sort() 
        //method (which uses TimSort). After sorting, 
        //the elements from each bucket are copied back into 
        //the original array in order.
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
        System.out.println("Bucket Sorted array:");
        for (int num : arr) {
            System.out.print(num + " ");
        }
    }
}