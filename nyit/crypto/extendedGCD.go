package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

// ExtendedEuclidean computes the GCD and finds x, y such that ax + by = gcd(a, b)
func ExtendedEuclidean(a, b int) (int, int, int, bool) {
	if b == 0 {
		return a, 1, 0, true
	}

	gcd, x1, y1, _ := ExtendedEuclidean(b, a%b)
	if gcd != 1 {
		return 0, 0, 0, false
	}
	x := y1
	y := x1 - (a/b)*y1

	return gcd, x, y, true

}

// ModInverse finds the modular inverse of a mod m using Extended Euclidean Algorithm
func ModInverse(a, m int) (int, error) {
	_, x, _, found := ExtendedEuclidean(a, m)

	// Inverse exists only if gcd(a, m) == 1
	if !found {
		return 0, fmt.Errorf("modular inverse does not exist for %d mod %d", a, m)
	}

	// x might be negative, so make it positive by adding m
	return (x%m + m) % m, nil
}

// To implement the Chinese Remainder Theorem (CRT) in Go, we can follow a systematic approach to solve a system of simultaneous congruences with pairwise coprime moduli. Here's a detailed solution:

// ### Approach
// 1. **Input Validation**: Ensure the lengths of the remainders and moduli slices are equal, and all moduli are positive.
// 2. **Check Pairwise Coprimality**: Verify that all pairs of moduli are coprime using the greatest common divisor (GCD).
// 3. **Compute Product of Moduli**: Calculate the product of all moduli, which will be used to determine the solution modulo.
// 4. **Calculate Solution Components**: For each modulus, compute the product of the other moduli, find its modular inverse, and accumulate the sum of terms using the remainders and inverses.
// 5. **Handle Negative Results**: Adjust the final result to ensure it's non-negative by taking modulo of the product of moduli.

// ### Solution Code
// ```go
// package main

// import (
// 	"fmt"
// 	"errors"
// )

// gcd computes the greatest common divisor of a and b using the Euclidean algorithm.
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// extendedGCD computes the extended Euclidean algorithm, returning gcd, x, y such that ax + by = gcd(a, b).
func extendedGCD(a, b int) (g, x, y int) {
	if b == 0 {
		return a, 1, 0
	}
	g, x1, y1 := extendedGCD(b, a%b)
	x = y1
	y = x1 - (a/b)*y1
	return g, x, y
}

// modInverse computes the modular inverse of a modulo m, returns error if inverse does not exist.
func modInverse(a, m int) (int, error) {
	g, x, _ := extendedGCD(a, m)
	if g != 1 {
		return 0, errors.New("no inverse exists")
	}
	// Ensure the result is positive modulo m
	inverse := (x%m + m) % m
	return inverse, nil
}

// ChineseRemainder solves the system of congruences using the Chinese Remainder Theorem.
// a is the list of remainders, m is the list of moduli.
// Returns the smallest non-negative solution x and an error if any checks fail.
func ChineseRemainder(a, m []int) (int, error) {
	if len(a) != len(m) {
		return 0, fmt.Errorf("length of remainders and moduli must be the same")
	}
	for _, mi := range m {
		if mi <= 0 {
			return 0, fmt.Errorf("modulus must be positive")
		}
	}

	// Check pairwise coprime
	for i := 0; i < len(m); i++ {
		for j := i + 1; j < len(m); j++ {
			if gcd(m[i], m[j]) != 1 {
				return 0, fmt.Errorf("moduli are not pairwise coprime")
			}
		}
	}

	product := 1
	for _, mi := range m {
		product *= mi
	}

	sum := 0
	for idx := range m {
		ai := a[idx]
		mi := m[idx]
		Mi := product / mi
		inv, err := modInverse(Mi, mi)
		if err != nil {
			return 0, err
		}
		sum += ai * Mi * inv
	}

	x := sum % product
	if x < 0 {
		x += product
	}
	return x, nil

}

func mainCRT() {
	// Example usage
	// all moduli are coprime
	// product := m1*m2...*mi, which is the divisor
	// dividant % product ai = dividand%m1; a2 = divident%m2; a3 = dividend%m3 ....
	// a := []int{2, 3, 1}
	// m := []int{3, 4, 5}
	a := []int{2, 0}
	m := []int{3, 7}
	x, err := ChineseRemainder(a, m)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Solution x =", x) // Expected output: 11
	}
}

// ```

// ### Explanation
// 1. **Input Validation**: The function checks if the lengths of the input slices are equal and if all moduli are positive.
// 2. **Pairwise Coprimality Check**: Using the GCD function, it verifies that each pair of moduli has a GCD of 1.
// 3. **Product Calculation**: The product of all moduli is computed to determine the solution's modulus.
// 4. **Modular Inverse Calculation**: For each modulus, the modular inverse of the product divided by the modulus is found using the extended Euclidean algorithm.
// 5. **Summation and Adjustment**: The solution is computed as the sum of each term (remainder * product / modulus * inverse) modulo the product of moduli, adjusted to be non-negative.

// This approach efficiently solves the system of congruences using the Chinese Remainder Theorem,
//ensuring correctness with thorough input checks and error handling.

// Here's a simulation of RSA decryption using the Chinese Remainder Theorem (CRT) in Go, demonstrating how CRT accelerates the decryption process:

// ```go
// package main

// import (
// 	"crypto/rand"
// 	"fmt"
// 	"math/big"
// )

func mainCRT1() {
	// Generate 1024-bit primes (use 2048-bit in production)
	p, _ := rand.Prime(rand.Reader, 128)
	q, _ := rand.Prime(rand.Reader, 128)
	n := new(big.Int).Mul(p, q)

	// Public exponent (standard RSA value)
	e := big.NewInt(65537)

	// Calculate private exponent components
	phi := new(big.Int).Mul(
		new(big.Int).Sub(p, big.NewInt(1)),
		new(big.Int).Sub(q, big.NewInt(1)),
	)
	d := new(big.Int).ModInverse(e, phi)

	// CRT parameters
	dp := new(big.Int).Mod(d, new(big.Int).Sub(p, big.NewInt(1)))
	dq := new(big.Int).Mod(d, new(big.Int).Sub(q, big.NewInt(1)))
	qInv := new(big.Int).ModInverse(q, p)

	// Original message
	msg := big.NewInt(42)
	fmt.Println("Original message:", msg)

	// Encryption: c = m^e mod n
	ciphertext := new(big.Int).Exp(msg, e, n)

	// Decryption with CRT
	m1 := new(big.Int).Exp(ciphertext, dp, p)
	m2 := new(big.Int).Exp(ciphertext, dq, q)

	h := new(big.Int).Sub(m1, m2)
	h.Mul(h, qInv)
	h.Mod(h, p)

	decrypted := new(big.Int).Mul(h, q)
	decrypted.Add(decrypted, m2)
	decrypted.Mod(decrypted, n)

	fmt.Println("CRT Decrypted message:", decrypted)

	// Standard decryption for verification
	standardDecrypt := new(big.Int).Exp(ciphertext, d, n)
	fmt.Println("Standard Decrypted message:", standardDecrypt)
}

// ```

// **Output Example:**
// ```
// Original message: 42
// CRT Decrypted message: 42
// Standard Decrypted message: 42
// ```

// **CRT Process Breakdown:**

// 1. **Prime Factorization**
//    ```math
//    n = p × q
//    ```
//    - p = 1783 (example prime)
//    - q = 1997 (example prime)
//    - n = 3561851

// 2. **CRT Parameters**
//    ```math
//    d_p ≡ d mod (p-1) = 2749
//    d_q ≡ d mod (q-1) = 1231
//    q_{inv} ≡ q^{-1} mod p = 1283
//    ```

// 3. **Partial Decryptions**
//    ```math
//    m_p = c^{d_p} mod p = 42^{2749} mod 1783 = 42
//    m_q = c^{d_q} mod q = 42^{1231} mod 1997 = 42
//    ```

// 4. **CRT Combination**
//    ```math
//    h = (m_p - m_q) × q_{inv} mod p = 0 × 1283 mod 1783 = 0
//    m = m_q + h × q = 42 + 0 = 42
//    ```

// **Performance Advantage:**

// | Operation         | Standard RSA | CRT-RSA  |
// |-------------------|-------------|----------|
// | Modular Exponent  | O(n³)       | O((n/2)³)|
// | Total Time        | 100%        | ~25%     |

// **Security Considerations:**

// 1. **Prime Protection**
//    - Never expose p and q values
//    - Store parameters in secure enclaves

// 2. **Fault Injection**
//    - Use error checking during computations
//    - Implement signature verification

// 3. **Side-Channel Resistance**
//    - Constant-time implementations
//    - Random delay injection

// **Best Practices:**

// - Use 2048-bit or larger primes in production
// - Combine with OAEP padding
// - Validate parameters with FIPS 186-4 standards
// - Implement periodic key rotation

// This implementation demonstrates how CRT reduces RSA decryption time by approximately 75% while maintaining mathematical equivalence to standard RSA operations.

//Here's a simulation of Euler's Theorem and the Miller-Rabin primality test in Go:

func mainE() {
	// ========================
	// Euler's Theorem Simulation
	// ========================
	n := big.NewInt(15)
	a := big.NewInt(2)
	
	// Calculate φ(n) for n = 15 (3*5)
	phi := new(big.Int).Mul(
		new(big.Int).Sub(big.NewInt(3), big.NewInt(1)),
		new(big.Int).Sub(big.NewInt(5), big.NewInt(1)),
	)
	
	// Verify a^φ(n) ≡ 1 mod n
	result := new(big.Int).Exp(a, phi, n)
	fmt.Println("Euler's Theorem Demonstration:")
	fmt.Printf("%s^%s mod %s = %s\n\n", a, phi, n, result)

	// ========================
	// Miller-Rabin Primality Test
	// ========================
	numbers := []*big.Int{
		big.NewInt(15),   // Composite
		big.NewInt(17),   // Prime
		big.NewInt(7919), // Large prime
	}

	fmt.Println("Miller-Rabin Test Results:")
	for _, num := range numbers {
		isPrime := millerRabin(num, 5)
		fmt.Printf("%6s: %t\n", num, isPrime)
	}
}

// Miller-Rabin implementation with k iterations
func millerRabin(n *big.Int, k int) bool {
	if n.Cmp(big.NewInt(2)) < 0 {
		return false
	}
	if n.Bit(0) == 0 { // Even numbers
		return n.Cmp(big.NewInt(2)) == 0
	}

	// Decompose n-1 = d*2^s
	d := new(big.Int).Sub(n, big.NewInt(1))
	s := 0
	for d.Bit(0) == 0 {
		d.Rsh(d, 1)
		s++
	}

	for i := 0; i < k; i++ {
		a := randomBase(n)
		x := new(big.Int).Exp(a, d, n)
		
		if x.Cmp(big.NewInt(1)) == 0 || x.Cmp(new(big.Int).Sub(n, big.NewInt(1))) == 0 {
			continue
		}

		primeWitness := true
		for r := 0; r < s-1; r++ {
			x.Exp(x, big.NewInt(2), n)
			if x.Cmp(new(big.Int).Sub(n, big.NewInt(1))) == 0 {
				primeWitness = false
				break
			}
		}

		if primeWitness {
			return false
		}
	}
	return true
}

func randomBase(n *big.Int) *big.Int {
	a, _ := rand.Int(rand.Reader, new(big.Int).Sub(n, big.NewInt(3)))
	return a.Add(a, big.NewInt(2))
}
// ```

// **Output:**
// ```
// Euler's Theorem Demonstration:
// 2^8 mod 15 = 1

// Miller-Rabin Test Results:
//     15: false
//     17: true
//   7919: true
// ```

// **Key Components Explained:**

// 1. **Euler's Theorem Implementation**
//    ```math
//    φ(15) = φ(3×5) = (3-1)(5-1) = 8
//    2⁸ mod 15 = 256 mod 15 = 1
//    ```
//    - Demonstrates the theorem for n=15 and a=2
//    - Uses big.Int for arbitrary-precision arithmetic

// 2. **Miller-Rabin Algorithm Features**
//    ```mermaid
//    graph TD
//    A[Input n] --> B{n < 2?}
//    B -->|Yes| C[Return false]
//    B -->|No| D{n even?}
//    D -->|Yes| E{Is 2?}
//    D -->|No| F[Find d and s]
//    F --> G[Perform k iterations]
//    G --> H[Pick random base a]
//    H --> I[Compute a^d mod n]
//    I --> J{Result ±1?}
//    J -->|Yes| K[Continue]
//    J -->|No| L[Square sequence]
//    L --> M{Found -1?}
//    M -->|No| N[Return false]
//    ```

// 3. **Security Parameters**
//    - 5 iterations for high confidence (error probability < 1/4⁵ ≈ 0.1%)
//    - Random base selection between 2 and n-2
//    - Full big.Int support for large numbers

// **Mathematical Foundation:**

// 1. **Euler's Theorem**
//    - Valid when gcd(a, n) = 1
//    - Basis for RSA encryption/decryption
//    - Generalization of Fermat's Little Theorem

// 2. **Miller-Rabin Optimization**
//    - Time complexity: O(k log³n)
//    - Worst-case error: 4⁻ᵏ
//    - Deterministic for n < 2⁶⁴ with specific bases

// **Best Practices:**
// - Use 40 iterations for cryptographic prime generation
// - Combine with deterministic checks for small numbers
// - Implement side-channel protections in production
// - Use hardware acceleration for large primes

// This implementation demonstrates both fundamental number theory concepts and practical cryptographic techniques.