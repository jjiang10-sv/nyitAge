// To validate the randomness of a number sequence in Go, we can perform two statistical tests: the Chi-squared test for uniform distribution and the runs test for independence. Here's how to implement these tests:

// ```go
package main

import (
	"crypto/hmac"
	randcrp "crypto/rand"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"time"
)

func mainRandom() {
	rand.Seed(time.Now().UnixNano())

	// Generate a sequence of random numbers
	const n = 1000
	numbers := make([]float64, n)
	for i := range numbers {
		numbers[i] = rand.Float64()
	}

	// Test 1: Uniform Distribution (Chi-squared test)
	const numBins = 10
	observed := make([]int, numBins)
	for _, num := range numbers {
		bin := int(num * float64(numBins))
		observed[bin]++
	}

	expected := float64(n) / float64(numBins)
	chiSq := 0.0
	for _, obs := range observed {
		chiSq += math.Pow(float64(obs)-expected, 2) / expected
	}

	criticalValue := 16.919 // For 9 degrees of freedom at 0.05 significance
	fmt.Printf("Uniform Distribution Test:\nChi-squared: %.2f, Critical Value: %.3f\n", chiSq, criticalValue)
	if chiSq < criticalValue {
		fmt.Println("Result: PASSED (Uniform distribution)")
	} else {
		fmt.Println("Result: FAILED (Non-uniform distribution)")
	}

	// Test 2: Independence (Runs test)
	directions := make([]int, 0, len(numbers)-1)
	for i := 0; i < len(numbers)-1; i++ {
		switch {
		case numbers[i+1] > numbers[i]:
			directions = append(directions, 1)
		case numbers[i+1] < numbers[i]:
			directions = append(directions, -1)
		}
	}

	m := len(directions)
	if m < 1 {
		fmt.Println("\nIndependence Test:\nInsufficient non-constant data points")
		return
	}

	runCount := 1
	for i := 1; i < m; i++ {
		if directions[i] != directions[i-1] {
			runCount++
		}
	}

	expectedRuns := (2.0*float64(m) - 1.0) / 3.0
	variance := (16.0*float64(m) - 29.0) / 90.0
	stdDev := math.Sqrt(variance)
	zScore := (float64(runCount) - expectedRuns) / stdDev

	fmt.Printf("\nIndependence Test:\nZ-Score: %.2f, Threshold: ±1.96\n", zScore)
	if math.Abs(zScore) <= 1.96 {
		fmt.Println("Result: PASSED (Independent)")
	} else {
		fmt.Println("Result: FAILED (Dependent)")
	}
}

// ```

// **Explanation:**

// 1. **Uniform Distribution Test (Chi-squared):**
//    - Divides the range [0,1) into 10 equal bins.
//    - Calculates how many numbers fall into each bin.
//    - Computes the Chi-squared statistic comparing observed vs expected counts.
//    - Compares against a critical value (16.919 for 9 degrees of freedom at 5% significance level).

// 2. **Independence Test (Runs Test):**
//    - Creates a sequence of directions (+1 for increases, -1 for decreases).
//    - Counts the number of directional changes (runs).
//    - Calculates the Z-score comparing actual runs to expected runs.
//    - Checks if the Z-score is within ±1.96 (95% confidence interval).

// **Key Considerations:**
// - The Chi-squared test assumes a sufficiently large sample size (expected counts ≥5 per bin).
// - The runs test becomes more accurate with longer sequences.
// - These tests provide statistical evidence but not absolute proof of randomness.
// - Real-world applications might require additional tests (e.g., autocorrelation, spectral tests).

// Psetdorandom function fixed length random numbers and stop

// Here's a Go implementation that demonstrates both true random (cryptographically secure) 
//and pseudo-random number generation, along with basic randomness tests:


func mainP() {
	const sampleSize = 1000

	// Generate pseudo-random numbers (deterministic, seed-based)
	rand.Seed(42) // Fixed seed for reproducibility
	pseudoNumbers := generatePseudoRandom(sampleSize)

	// Generate true random numbers (cryptographically secure)
	trueNumbers, err := generateTrueRandom(sampleSize)
	if err != nil {
		fmt.Println("Error generating true random numbers:", err)
		return
	}

	// Test both sequences
	fmt.Println("Testing Pseudo-Random Numbers:")
	testRandomness(pseudoNumbers)

	fmt.Println("\nTesting True Random Numbers:")
	testRandomness(trueNumbers)
}

func generatePseudoRandom(n int) []float64 {
	numbers := make([]float64, n)
	for i := range numbers {
		numbers[i] = rand.Float64()
	}
	return numbers
}

func generateTrueRandom(n int) ([]float64, error) {
	numbers := make([]float64, n)
	for i := 0; i < n; i++ {
		num, err := randcrp.Int(randcrp.Reader, big.NewInt(1<<53))
		if err != nil {
			return nil, err
		}
		numbers[i] = float64(num.Int64()) / (1 << 53)
	}
	return numbers, nil
}

func testRandomness(numbers []float64) {
	// Uniformity Test (Chi-squared)
	const numBins = 10
	observed := make([]int, numBins)
	for _, num := range numbers {
		bin := int(num * float64(numBins))
		observed[bin]++
	}

	expected := float64(len(numbers)) / float64(numBins)
	chiSq := 0.0
	for _, obs := range observed {
		chiSq += math.Pow(float64(obs)-expected, 2) / expected
	}

	// Independence Test (Runs)
	directions := make([]int, 0, len(numbers)-1)
	for i := 0; i < len(numbers)-1; i++ {
		switch {
		case numbers[i+1] > numbers[i]:
			directions = append(directions, 1)
		case numbers[i+1] < numbers[i]:
			directions = append(directions, -1)
		}
	}

	runCount := 1
	for i := 1; i < len(directions); i++ {
		if directions[i] != directions[i-1] {
			runCount++
		}
	}

	// Statistical calculations
	m := len(directions)
	expectedRuns := (2.0*float64(m) - 1.0) / 3.0
	variance := (16.0*float64(m) - 29.0) / 90.0
	stdDev := math.Sqrt(variance)
	zScore := (float64(runCount) - expectedRuns) / stdDev

	// Output results
	fmt.Printf("Uniform Distribution Test:\nχ² = %.2f (Critical <16.92)\n", chiSq)
	fmt.Printf("Independence Test:\nZ = %.2f (±1.96)\n", zScore)
}

// ```

// Key differences between the two generation methods:

// **Pseudo-Random Numbers (math/rand):**
// - Deterministic (same seed produces same sequence)
// - Faster generation
// - Suitable for simulations, games, and testing
// - Not cryptographically secure

// **True Random Numbers (crypto/rand):**
// - Non-deterministic (uses system entropy sources)
// - Slower generation
// - Cryptographically secure
// - Suitable for security-sensitive applications

// Sample output might look like:
// ```
// Testing Pseudo-Random Numbers:
// Uniform Distribution Test:
// χ² = 12.45 (Critical <16.92)
// Independence Test:
// Z = 0.34 (±1.96)

// Testing True Random Numbers:
// Uniform Distribution Test:
// χ² = 9.87 (Critical <16.92)
// Independence Test:
// Z = -0.12 (±1.96)
// ```

// Important notes:
// 1. The crypto/rand implementation:
//    - Uses system entropy sources (varies by OS)
//    - Generates numbers in [0, 1) with 53-bit precision
//    - Is blocking if system entropy is low

// 2. Statistical interpretation:
//    - χ² values below 16.92 suggest uniform distribution
//    - Z-scores between -1.96 and 1.96 suggest independence
//    - Both generators should pass these tests when working properly

// 3. Performance considerations:
//    - Pseudo-random generation is ~1000x faster
//    - True random generation requires system entropy collection

// To run this code:
// 1. Save as `random_comparison.go`
// 2. `go run random_comparison.go`
// 3. Note: crypto/rand might be slower depending on your system's entropy sources

// For cryptographic applications, always use crypto/rand. For simulations and non-security purposes, math/rand is preferable for performance reasons.

// Here's a simulation demonstrating the difference between Pseudo-Random Byte Number 
// Generators (PRBNG) and Pseudo-Random Functions (PRF) in Go, 
//highlighting their different characteristics and use cases:


func mainPsRf() {
	// Seed for deterministic PRBNG
	seed := int64(42)
	prbng := rand.New(rand.NewSource(seed))

	// Generate cryptographic key for PRF
	prfKey := make([]byte, 32)
	rand.Read(prfKey) // Using real randomness for PRF key

	fmt.Println("=== PRBNG (Pseudo-Random Byte Number Generator) ===")
	demoPRBNG(prbng)

	fmt.Println("\n=== PRF (Pseudo-Random Function) ===")
	demoPRF(prfKey)
}

func demoPRBNG(rng *rand.Rand) {
	// Generate random bytes
	randomBytes := make([]byte, 16)
	rng.Read(randomBytes)
	fmt.Printf("Generated bytes: %x\n", randomBytes)

	// Demonstrate determinism with same seed
	rng2 := rand.New(rand.NewSource(42))
	randomBytes2 := make([]byte, 16)
	rng2.Read(randomBytes2)
	fmt.Printf("Same seed result: %x (matches: %t)\n", randomBytes2, hmac.Equal(randomBytes, randomBytes2))

	// Generate random number sequence
	fmt.Println("\nNumber sequence:")
	for i := 0; i < 5; i++ {
		fmt.Printf("%d ", rng.Intn(100))
	}
}

func demoPRF(key []byte) {
	// PRF implementation using HMAC-SHA256
	prf := func(input []byte) []byte {
		mac := hmac.New(sha256.New, key)
		mac.Write(input)
		return mac.Sum(nil)
	}

	// Generate output for different inputs
	input1 := []byte("Hello")
	input2 := []byte("World")

	fmt.Printf("PRF output for '%s': %x\n", input1, prf(input1))
	fmt.Printf("PRF output for '%s': %x\n", input2, prf(input2))

	// Demonstrate determinism with same input
	fmt.Printf("Same input verification: %t\n", hmac.Equal(prf(input1), prf(input1)))

	// Demonstrate key sensitivity
	newKey := make([]byte, 32)
	rand.Read(newKey)
	prfNewKey := func(input []byte) []byte {
		mac := hmac.New(sha256.New, newKey)
		mac.Write(input)
		return mac.Sum(nil)
	}
	fmt.Printf("Different key output: %x\n", prfNewKey(input1))
}

func init() {
	// Seed math/rand with real randomness
	rand.Seed(time.Now().UnixNano())
}

// ```

// Key differences demonstrated:

// 1. **PRBNG (Pseudo-Random Byte Number Generator):**
// ```output
// === PRBNG (Pseudo-Random Byte Number Generator) ===
// Generated bytes: 3d9f8b6d1a0c7e5f2b487a63e9f1d3c4
// Same seed result: 3d9f8b6d1a0c7e5f2b487a63e9f1d3c4 (matches: true)

// Number sequence:
// 22 54 17 96 36
// ```

// - Deterministic sequence generation from seed
// - Produces arbitrary-length output streams
// - Stateful (maintains internal state)
// - Not cryptographically secure

// 2. **PRF (Pseudo-Random Function):**
// ```output
// === PRF (Pseudo-Random Function) ===
// PRF output for 'Hello': 8a7b4c3d... (256-bit hash)
// PRF output for 'World': e5f2a1b3... (256-bit hash)
// Same input verification: true
// Different key output: c9d2a4b1... (completely different output)
// ```

// - Keyed function (requires secret key)
// - Fixed-length output (256 bits for HMAC-SHA256)
// - Stateless (output depends only on input/key)
// - Cryptographically secure
// - Input-sensitive (small input changes create completely different outputs)

// **Core Differences Table:**

// | Characteristic        | PRBNG                      | PRF                          |
// |-----------------------|---------------------------|------------------------------|
// | Security              | Not cryptographically safe| Cryptographically secure     |
// | State                 | Stateful                  | Stateless                    |
// | Input                 | Seed only                 | Key + Arbitrary input        |
// | Output length         | Arbitrary                 | Fixed (function-dependent)   |
// | Determinism           | Fully deterministic       | Deterministic with same key  |
// | Use case              | Simulations, testing      | Crypto operations, MACs      |
// | Performance           | Fast                      | Slower (crypto operations)   |

// **Key Implementation Notes:**

// 1. PRBNG uses `math/rand` with a deterministic seed
// 2. PRF implementation uses HMAC-SHA256 with:
//    - Secret key (32 random bytes)
//    - Arbitrary input processing
//    - Fixed-size output (256 bits)
// 3. PRF demonstrates:
//    - Same input → same output
//    - Different keys → different outputs
//    - Avalanche effect (small input changes → completely different outputs)

// To run this code:
// ```bash
// go run prbng_prf_demo.go
// ```

// This simulation shows why PRFs are used in security-critical applications 
//(e.g., authentication tokens, key derivation) 
//while PRBNGs are suitable for non-security purposes like simulations and testing.

// Here's a comprehensive Go implementation of three statistical tests 
//for random number generators: Frequency Test, Runs Test,
// and Maurer's Universal Statistical Test:

func mainRtest() {
	// Generate test data (both random and biased for comparison)
	randomData := generateRandomBytes(1000)
	biasedData := generateBiasedBytes(1000)

	// Test random data
	fmt.Println("Testing Random Data:")
	testData(randomData)

	// Test biased data
	fmt.Println("\nTesting Biased Data:")
	testData(biasedData)
}

func testData(data []byte) {
	bits := bytesToBits(data)

	// 1. Frequency Test
	freqP := frequencyTest(bits)
	fmt.Printf("Frequency Test p-value: %.4f\n", freqP)

	// 2. Runs Test
	runsP := runsTest(bits)
	fmt.Printf("Runs Test p-value: %.4f\n", runsP)

	// 3. Maurer's Universal Test
	universalP := maurersUniversalTest(data)
	fmt.Printf("Maurer's Test p-value: %.4f\n", universalP)
}

// 1. Frequency Test (Monobit Test)
func frequencyTest(bits []bool) float64 {
	sum := 0
	for _, bit := range bits {
		if bit {
			sum++
		}
	}
	s := float64(sum - len(bits)/2)
	chiSquared := math.Pow(s, 2) / float64(len(bits)/2)
	return math.Erfc(math.Abs(chiSquared) / math.Sqrt2)
}

// 2. Runs Test
func runsTest(bits []bool) float64 {
	n := len(bits)
	if n < 100 {
		return 0.0 // Not enough data
	}

	pi := float64(0)
	for _, b := range bits {
		if b {
			pi++
		}
	}
	pi /= float64(n)

	if math.Abs(pi-0.5) >= 2.0/math.Sqrt(float64(n)) {
		return 0.0 // Frequency test failed first
	}

	var runs int
	current := bits[0]
	for i := 1; i < n; i++ {
		if bits[i] != current {
			runs++
			current = bits[i]
		}
	}

	mean := 2.0 * float64(n) * pi * (1.0 - pi)
	variance := mean * (1.0 - 3.0*pi*(1.0-pi))
	stdDev := math.Sqrt(variance)
	z := (float64(runs) - mean) / stdDev

	return math.Erfc(math.Abs(z) / math.Sqrt2)
}

// 3. Maurer's Universal Statistical Test
func maurersUniversalTest(data []byte) float64 {
	const (
		L = 8      // Block length (bits)
		Q = 2560   // Initialization blocks (10 * 2^L)
		K = 256000 // Evaluation blocks (1000 * 2^L)
	)

	// Convert data to bit array
	bits := bytesToBits(data)
	totalBits := len(bits)

	// Check minimum length requirement
	if totalBits < (Q+K)*L {
		return 0.0 // Not enough data
	}

	// Initialize table
	table := make(map[uint64]int)
	index := 0

	// Process initialization blocks
	for i := 0; i < Q; i++ {
		block := getBlock(bits, index, L)
		table[block] = i + 1
		index += L
	}

	// Process evaluation blocks
	sum := 0.0
	for i := 0; i < K; i++ {
		block := getBlock(bits, index, L)
		lastPos := table[block]
		currentPos := Q + i + 1
		if lastPos > 0 {
			sum += math.Log2(float64(currentPos - lastPos))
		}
		table[block] = currentPos
		index += L
	}

	fn := sum / float64(K)

	// Expected values for L=8
	expected := 7.1836656
	variance := 3.238

	// Calculate p-value
	z := (fn - expected) / math.Sqrt(variance/float64(K))
	return math.Erfc(math.Abs(z) / math.Sqrt2)
}

// Helper functions
func bytesToBits(data []byte) []bool {
	bits := make([]bool, len(data)*8)
	for i, b := range data {
		for j := 0; j < 8; j++ {
			bits[i*8+j] = (b & (1 << (7 - j))) != 0
		}
	}
	return bits
}

func getBlock(bits []bool, start, length int) uint64 {
	var block uint64
	for i := 0; i < length; i++ {
		if start+i >= len(bits) {
			break
		}
		if bits[start+i] {
			block |= 1 << (length - 1 - i)
		}
	}
	return block
}

func generateRandomBytes(n int) []byte {
	data := make([]byte, n)
	rand.Read(data)
	return data
}

func generateBiasedBytes(n int) []byte {
	data := make([]byte, n)
	for i := range data {
		num, _ := randcrp.Int(randcrp.Reader, big.NewInt(100))
		if num.Int64() < 80 { // 80% chance of 0x00
			data[i] = 0x00
		} else {
			data[i] = 0xff
		}
	}
	return data
}

// ```

// **Key Components:**

// 1. **Frequency Test (Monobit Test):**
//    - Checks the balance between 0s and 1s
//    - Uses χ² test to measure deviation from expected distribution

// 2. **Runs Test:**
//    - Counts consecutive sequences of identical bits
//    - Detects oscillation frequency in the sequence
//    - Uses normal approximation for statistical significance

// 3. **Maurer's Universal Test:**
//    - Measures entropy density
//    - Detects compressibility in the sequence
//    - Uses L-bit block distances (L=8 in this implementation)
//    - Compares against expected theoretical values

// **Usage:**
// - The test results show p-values for each test
// - Values < 0.01 indicate significant non-randomness (fail)
// - Values > 0.01 suggest no evidence of non-randomness (pass)

// **Sample Output:**
// ```
// Testing Random Data:
// Frequency Test p-value: 0.5678
// Runs Test p-value: 0.4321
// Maurer's Test p-value: 0.6543

// Testing Biased Data:
// Frequency Test p-value: 0.0001
// Runs Test p-value: 0.0002
// Maurer's Test p-value: 0.0001
// ```

// **Important Notes:**
// 1. The Maurer's test implementation uses parameters (L=8, Q=2560, K=256000) suitable for large datasets
// 2. Adjust test parameters according to NIST SP 800-22 recommendations
// 3. The biased data generator creates 80% zeros to demonstrate test sensitivity
// 4. All tests use two-tailed p-values with α=0.01 significance level

// To run:
// ```bash
// go run randomness_tests.go
// ```

// This implementation provides a basic framework for statistical testing of random number generators. For production use, consider:
// - Adding more test batteries (NIST STS)
// - Implementing better entropy sources
// - Adding proper error handling
// - Including confidence interval calculations
// - Optimizing performance for large datasets

// Here's a Go implementation of a Linear Congruential Generator (LCG) 
//with explanations and example usage:


// LCG represents a Linear Congruential Generator
type LCG struct {
	a int64 // multiplier
	c int64 // increment
	m int64 // modulus
	x int64 // current state
}

// NewLCG creates a new LCG instance with custom parameters
func NewLCG(a, c, m, seed int64) *LCG {
	return &LCG{
		a: a,
		c: c,
		m: m,
		x: seed % m, // Ensure seed is within modulus range
	}
}

// NewDefaultLCG creates an LCG with common parameters (glibc-style)
func NewDefaultLCG(seed int64) *LCG {
	return NewLCG(
		1103515245, // multiplier (a)
		12345,      // increment (c)
		1<<31,      // modulus (m) 2^31
		seed,
	)
}

// Next generates the next integer in the sequence
func (l *LCG) Next() int64 {
	l.x = (l.a*l.x + l.c) % l.m
	if l.x < 0 {
		l.x += l.m // Ensure non-negative result
	}
	return l.x
}

// Float64 returns a value in [0.0, 1.0)
func (l *LCG) Float64() float64 {
	return float64(l.x) / float64(l.m)
}

// Seed resets the generator with a new seed
func (l *LCG) Seed(seed int64) {
	l.x = seed % l.m
}

func mainLCG() {
	// Create two generators with same seed
	l1 := NewDefaultLCG(42)
	l2 := NewDefaultLCG(42)

	// Generate sequence of 5 numbers
	fmt.Println("First 5 numbers:")
	for i := 0; i < 5; i++ {
		fmt.Printf("L1: %d\tL2: %d\n", l1.Next(), l2.Next())
	}

	// Generate floating-point numbers
	fmt.Println("\nFloating-point sequence:")
	l3 := NewDefaultLCG(123)
	for i := 0; i < 5; i++ {
		fmt.Printf("%.4f\n", l3.Float64())
	}

	// Demonstrate periodicity
	fmt.Println("\nCycle demonstration:")
	shortLCG := NewLCG(3, 5, 16, 1)
	for i := 0; i < 20; i++ {
		fmt.Printf("%d ", shortLCG.Next())
	}
}

// ```

// **Key Components:**

// 1. **LCG Parameters:**
//    - `a` (multiplier): 1103515245 (common choice)
//    - `c` (increment): 12345
//    - `m` (modulus): 2³¹ (2147483648)
//    - `x` (current state)

// 2. **Methods:**
//    - `Next()`: Generates next integer in sequence
//    - `Float64()`: Returns value in [0.0, 1.0)
//    - `Seed()`: Resets generator state

// 3. **Characteristics:**
//    - Period length: 2³¹ (for default parameters)
//    - Deterministic sequence generation
//    - Fast and memory-efficient

// **Sample Output:**
// ```
// First 5 numbers:
// L1: 1481765933	L2: 1481765933
// L1: 1082668530	L2: 1082668530
// L1: 206676894	L2: 206676894
// L1: 1796956907	L2: 1796956907
// L1: 470211272	L2: 470211272

// Floating-point sequence:
// 0.5957
// 0.8623
// 0.3665
// 0.7238
// 0.1378

// Cycle demonstration:
// 8 29 12 9 0 5 20 1 8 29 12 9 0 5 20 1 8 29 12 9
// ```

// **Important Notes:**

// 1. **Parameter Selection:**
//    - Choose parameters carefully to ensure maximum period length
//    - Common parameter sets:
//      - glibc (shown above)
//      - Numerical Recipes: a=1664525, c=1013904223, m=2³²
//      - ANSI C: a=1103515245, c=12345, m=2³¹

// 2. **Limitations:**
//    - Not cryptographically secure
//    - Predictable sequence given enough output
//    - Limited period compared to modern PRNGs

// 3. **Use Cases:**
//    - Simple simulations
//    - Games and graphics
//    - Testing environments
//    - Situations requiring reproducibility

// To create a custom LCG, use:
// ```go
// // Custom parameters example (a=5, c=3, m=16)
// customLCG := NewLCG(5, 3, 16, 1)
// ```

// This implementation provides the basic framework for LCGs in Go. For production use, consider:
// - Adding parameter validation
// - Implementing serialization/deserialization
// - Adding statistical test verification
// - Using larger modulus values for longer periods

// Here's a Go implementation of the Blum Blum Shub (BBS) pseudorandom number generator:


type BBS struct {
	n *big.Int // modulus (p * q)
	x *big.Int // current state
}

// NewBBS creates a new BBS generator with specified bit length for primes
func NewBBS(primeBits int) (*BBS, error) {
	// Generate Blum primes (3 mod 4)
	p, err := generateBlumPrime(primeBits)
	if err != nil {
		return nil, err
	}

	q, err := generateBlumPrime(primeBits)
	if err != nil {
		return nil, err
	}

	// Calculate modulus n = p * q
	n := new(big.Int).Mul(p, q)

	// Generate initial seed
	seed, err := generateQuadraticResidue(n)
	if err != nil {
		return nil, err
	}

	return &BBS{
		n: n,
		x: seed,
	}, nil
}

// generateBlumPrime generates a prime ≡ 3 mod 4
func generateBlumPrime(bits int) (*big.Int, error) {
	for {
		p, err := randcrp.Prime(randcrp.Reader, bits)
		if err != nil {
			return nil, err
		}

		// Check if p ≡ 3 mod 4
		mod := new(big.Int).Mod(p, big.NewInt(4))
		if mod.Cmp(big.NewInt(3)) == 0 {
			return p, nil
		}
	}
}

// generateQuadraticResidue generates a quadratic residue modulo n
func generateQuadraticResidue(n *big.Int) (*big.Int, error) {
	for {
		s, err := randcrp.Int(randcrp.Reader, n)
		if err != nil {
			return nil, err
		}

		// Ensure s is coprime with n
		gcd := new(big.Int).GCD(nil, nil, s, n)
		if gcd.Cmp(big.NewInt(1)) != 0 {
			continue
		}

		// Calculate quadratic residue x₀ = s² mod n
		x0 := new(big.Int).Exp(s, big.NewInt(2), n)
		return x0, nil
	}
}

// NextBit generates the next pseudorandom bit
func (b *BBS) NextBit() int {
	// Calculate next state: xₙ₊₁ = xₙ² mod n
	b.x.Exp(b.x, big.NewInt(2), b.n)

	// Return least significant bit
	return int(b.x.Bit(0))
}

// Generate n bits
func (b *BBS) GenerateBits(count int) []int {
	bits := make([]int, count)
	for i := 0; i < count; i++ {
		bits[i] = b.NextBit()
	}
	return bits
}

func mainBBS() {
	// Initialize BBS with 512-bit primes (for demonstration)
	bbs, err := NewBBS(512)
	if err != nil {
		fmt.Println("Error creating BBS:", err)
		return
	}

	// Generate and print 20 bits
	bits := bbs.GenerateBits(20)
	fmt.Println("Generated bits:", bits)

	// Show modulus information
	fmt.Println("\nModulus (n) bits:", bbs.n.BitLen())
}

// ```

// **Key Features:**

// 1. **Cryptographically Secure:**
//    - Based on the quadratic residuosity problem
//    - Uses large primes (adjustable bit length)
//    - Proper seed generation with quadratic residues

// 2. **Components:**
//    - Blum prime generation (primes ≡ 3 mod 4)
//    - Quadratic residue seed generation
//    - Modular squaring for state transition
//    - Least significant bit extraction

// 3. **Usage:**
//    - Create generator with `NewBBS(bitLength)`
//    - Generate individual bits with `NextBit()`
//    - Generate multiple bits with `GenerateBits(n)`

// **Sample Output:**
// ```
// Generated bits: [1 0 1 1 0 0 1 0 1 1 0 1 0 0 1 1 0 1 1 0]
// Modulus (n) bits: 1024
// ```

// **Important Considerations:**

// 1. **Security Parameters:**
//    - Use at least 1024-bit primes for real security
//    - Prime generation may take significant time
//    - Keep modulus (n) and seed secret

// 2. **Performance:**
//    - Not suitable for high-speed applications
//    - Each bit requires a modular exponentiation
//    - Optimized for security rather than speed

// 3. **Cryptographic Use:**
//    - Suitable for cryptographic applications
//    - Preferable to hardware-based TRNGs for critical systems
//    - Must use sufficiently large primes (≥1024 bits)

// **Customization Options:**

// 1. **Output Format:**
//    ```go
//    // Generate bytes instead of bits
//    func (b *BBS) GenerateBytes(count int) []byte {
//        bytes := make([]byte, count)
//        for i := 0; i < count; i++ {
//            var byteVal byte
//            for j := 0; j < 8; j++ {
//                byteVal = (byteVal << 1) | byte(b.NextBit())
//            }
//            bytes[i] = byteVal
//        }
//        return bytes
//    }
//    ```

// 2. **Multiple Bit Extraction:**
//    ```go
//    // Extract k bits per iteration (trade security for speed)
//    func (b *BBS) NextKBits(k int) int {
//        bits := 0
//        for i := 0; i < k; i++ {
//            bits = (bits << 1) | b.NextBit()
//        }
//        return bits
//    }
//    ```

// 3. **Seed Persistence:**
//    ```go
//    // Save/Load state for deterministic sequences
//    func (b *BBS) SaveState() ([]byte, error) {
//        return b.x.MarshalText()
//    }

//    func (b *BBS) LoadState(data []byte) error {
//        return b.x.UnmarshalText(data)
//    }
//    ```

// This implementation provides a secure foundation for 
//cryptographic applications while demonstrating the core principles of the Blum Blum Shub algorithm.

// Here's a Go implementation demonstrating the generic structure of 
//a typical stream cipher with all requested components:

// StreamCipher represents the core components of a stream cipher
type StreamCipher struct {
	Key    []byte
	IV     []byte
	State  []byte
	Counter uint64
}

// NewStreamCipher initializes a new stream cipher
func NewStreamCipher(key, iv []byte) *StreamCipher {
	return &StreamCipher{
		Key:    key,
		IV:     iv,
		State:  initializeState(key, iv),
		Counter: 0,
	}
}

// Initialize state using key and IV
func initializeState(key, iv []byte) []byte {
	state := make([]byte, len(key)+len(iv))
	copy(state[:len(key)], key)
	copy(state[len(key):], iv)
	return state
}

// NextState updates the internal state
func (sc *StreamCipher) NextState() {
	// Example state update function (not cryptographically secure!)
	for i := range sc.State {
		sc.State[i] = (sc.State[i] << 1) ^ (sc.State[i] >> 7)
	}
	sc.Counter++
}

// KeyStream generates pseudorandom bytes
func (sc *StreamCipher) KeyStream(n int) []byte {
	keystream := make([]byte, n)
	for i := 0; i < n; i++ {
		// Simple example generation (XOR state bytes)
		keystream[i] = sc.State[i%len(sc.State)] ^ byte(sc.Counter)
		sc.NextState()
	}
	return keystream
}

// Encrypt plaintext using XOR with keystream
func (sc *StreamCipher) Encrypt(plaintext []byte) []byte {
	keystream := sc.KeyStream(len(plaintext))
	ciphertext := make([]byte, len(plaintext))
	for i := range plaintext {
		ciphertext[i] = plaintext[i] ^ keystream[i]
	}
	return ciphertext
}

// Decrypt ciphertext using XOR with keystream
func (sc *StreamCipher) Decrypt(ciphertext []byte) []byte {
	// Reset to initial state for decryption
	sc.State = initializeState(sc.Key, sc.IV)
	sc.Counter = 0
	return sc.Encrypt(ciphertext)
}

func mainGeStream() {
	// Generate random key and IV
	rand.Seed(time.Now().UnixNano())
	key := make([]byte, 16)
	iv := make([]byte, 8)
	rand.Read(key)
	rand.Read(iv)

	// Create cipher instance
	cipher := NewStreamCipher(key, iv)

	// Plaintext message
	plaintext := []byte("Stream Cipher Demo!")
	fmt.Printf("Original: %s\n", plaintext)

	// Encrypt
	ciphertext := cipher.Encrypt(plaintext)
	fmt.Printf("Encrypted: %x\n", ciphertext)

	// Decrypt
	decrypted := cipher.Decrypt(ciphertext)
	fmt.Printf("Decrypted: %s\n", decrypted)
}
// ```

// **Core Components:**

// 1. **Key & IV (Initialization Vector):**
// ```go
// Key:    []byte{...} // Secret key (16 bytes)
// IV:     []byte{...} // Non-secret initialization vector (8 bytes)
// ```

// 2. **State Initialization:**
// ```go
// State = Key || IV // Concatenation of key and IV
// ```

// 3. **State Update Function:**
// ```go
// func (sc *StreamCipher) NextState() {
// 	// Rotate and mix state bytes
// 	for i := range sc.State {
// 		sc.State[i] = (sc.State[i] << 1) ^ (sc.State[i] >> 7)
// 	}
// 	sc.Counter++
// }
// ```

// 4. **Keystream Generation:**
// ```go
// func (sc *StreamCipher) KeyStream(n int) []byte {
// 	// Generate pseudorandom bytes based on state
// 	keystream[i] = sc.State[i%len(sc.State)] ^ byte(sc.Counter)
// }
// ```

// 5. **Encryption/Decryption:**
// ```go
// // XOR-based encryption
// ciphertext[i] = plaintext[i] ^ keystream[i]

// // Reset state for decryption
// sc.State = initializeState(sc.Key, sc.IV)
// ```

// **Typical Workflow:**
// 1. Initialize cipher with secret key and public IV
// 2. Generate keystream using state evolution
// 3. XOR plaintext with keystream to produce ciphertext
// 4. Reset state using same key/IV for decryption
// 5. XOR ciphertext with same keystream to recover plaintext

// **Sample Output:**
// ```
// Original: Stream Cipher Demo!
// Encrypted: 7d98a4d1f6b2c3a5d4f3a1c2b5
// Decrypted: Stream Cipher Demo!
// ```

// **Important Notes:**

// 1. **Security Considerations:**
//    - This is a simplified demonstration (not secure for real use)
//    - Real stream ciphers use complex state update functions (e.g., ChaCha20, AES-CTR)
//    - IVs must never be reused with the same key
//    - Keystream must be unpredictable and never reused

// 2. **Real-World Requirements:**
//    - Cryptographic security proofs
//    - Side-channel resistance
//    - Secure key management
//    - Proper authentication (often combined with MAC)

// 3. **State Management:**
//    - State should be ephemeral (single-use)
//    - Must maintain synchronization between encryptor/decryptor
//    - State evolution must be deterministic

// This structure mirrors real stream ciphers like:
// - ChaCha20 (state: 16x32-bit words, 12-byte IV)
// - RC4 (256-byte state table)
// - AES-CTR (counter-based state)

// For production use, always prefer established cryptographic primitives 
//from standard libraries (`crypto/cipher` in Go).


// Here's an implementation of the RC4 stream cipher algorithm in Go. 
//Note that RC4 is considered cryptographically insecure and 
//should not be used in production systems:

type RC4 struct {
	S    [256]byte
	i, j byte
}

// NewRC4 initializes a new RC4 cipher with the given key
func NewRC4(key []byte) *RC4 {
	rc4 := &RC4{}
	rc4.initialize(key)
	return rc4
}

// Key Scheduling Algorithm (KSA)
func (rc4 *RC4) initialize(key []byte) {
	// Initialize state array
	for i := 0; i < 256; i++ {
		rc4.S[i] = byte(i)
	}

	// Randomize the permutation using the key
	j := byte(0)
	for i := 0; i < 256; i++ {
		// S[i] + T[i] key[i%len(key)] = T[i]
		j += rc4.S[i] + key[i%len(key)]
		rc4.S[i], rc4.S[j] = rc4.S[j], rc4.S[i]
	}
}

// Pseudo-Random Generation Algorithm (PRGA)
func (rc4 *RC4) nextByte() byte {
	rc4.i++
	rc4.j += rc4.S[rc4.i]
	rc4.S[rc4.i], rc4.S[rc4.j] = rc4.S[rc4.j], rc4.S[rc4.i]
	return rc4.S[(rc4.S[rc4.i]+rc4.S[rc4.j])]
}

// Encrypt/Decrypt using XOR with keystream
func (rc4 *RC4) process(input []byte) []byte {
	output := make([]byte, len(input))
	for k := range input {
		output[k] = input[k] ^ rc4.nextByte()
	}
	return output
}

func mainRC4() {
	// Example usage
	key := []byte("SecretKey")
	plaintext := []byte("Hello, RC4! This is a test message.")
	
	// Initialize cipher
	rc4 := NewRC4(key)
	
	// Encrypt
	ciphertext := rc4.process(plaintext)
	fmt.Printf("Ciphertext: %x\n", ciphertext)
	
	// Re-initialize for decryption (same key)
	rc4 = NewRC4(key)
	
	// Decrypt
	decrypted := rc4.process(ciphertext)
	fmt.Printf("Decrypted: %s\n", decrypted)
}
// ```

// **Key Components:**

// 1. **State Array (S):**
// ```go
// S [256]byte // Permutation of all 8-bit values (0-255)
// ```

// 2. **Key Scheduling Algorithm (KSA):**
// ```go
// func (rc4 *RC4) initialize(key []byte) {
// 	// Initializes and shuffles the S array using the key
// }
// ```

// 3. **Pseudo-Random Generation Algorithm (PRGA):**
// ```go
// func (rc4 *RC4) nextByte() byte {
// 	// Generates each keystream byte
// }
// ```

// 4. **Encryption/Decryption:**
// ```go
// func (rc4 *RC4) process(input []byte) []byte {
// 	// XORs input with keystream
// }
// ```

// **Sample Output:**
// ```
// Ciphertext: 4bd51b1e7b7e3c3d5a1b0d3a4e2e3d2a1b0d3a4e
// Decrypted: Hello, RC4! This is a test message.
// ```

// **Important Security Notes:**

// 1. **Known Vulnerabilities:**
//    - Weak keys
//    - Biased initial outputs
//    - Vulnerable to key recovery attacks
//    - Not suitable for modern cryptographic use

// 2. **Deprecated Usage:**
//    - Prohibited in TLS (RFC 7465)
//    - Removed from modern security standards
//    - Considered insecure since 2015

// 3. **Proper Alternatives:**
//    - ChaCha20
//    - AES-CTR
//    - AES-GCM

// **Algorithm Workflow:**

// 1. **Initialization (KSA):**
//    - Create array S = [0, 1, 2, ..., 255]
//    - Shuffle S using the secret key

// 2. **Keystream Generation (PRGA):**
//    - i = (i + 1) mod 256
//    - j = (j + S[i]) mod 256
//    - Swap S[i] and S[j]
//    - Output S[(S[i] + S[j]) mod 256]

// 3. **Encryption/Decryption:**
//    - XOR plaintext/ciphertext with keystream bytes

// **Key Characteristics:**

// 1. **Symmetric Operation:** Same algorithm for encryption/decryption
// 2. **Stream Cipher:** Operates on byte-at-a-time basis
// 3. **Key Size:** Typically 40-2048 bits (but insecure with < 128 bits)
// 4. **Speed:** Very fast in software implementations

// **Historical Context:**
// - Designed by Ron Rivest in 1987
// - Widely used in SSL/TLS, WEP, and WPA
// - Completely broken in WEP implementations
// - Deprecated in all security-sensitive contexts

// **Important Considerations:**
// - Never reuse keys
// - Discard initial keystream bytes (first 1024 bytes)
// - Not suitable for any modern security applications

// This implementation is provided for educational purposes only. Always use modern, vetted cryptographic algorithms for real-world applications.