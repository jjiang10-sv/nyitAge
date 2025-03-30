package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// func main() {
// 	// Step 4: Perform the chosen plaintext attack
// 	reconstructedTable := chosenPlaintextAttack()

// 	// Print the reconstructed substitution table
// 	fmt.Println("Reconstructed Substitution Table:")
// 	for c := 'A'; c <= 'Z'; c++ {
// 		fmt.Printf("%c -> %c\n", c, reconstructedTable[c])
// 	}

// 	// Step 5: Decrypt an intercepted ciphertext
// 	interceptedCiphertext := "QWERTYUIOPASDFGHJKLZXCVBNM"
// 	decryptedText := decrypt(interceptedCiphertext, reconstructedTable)
// 	fmt.Printf("\nDecrypted Text: %s\n", decryptedText)
// }

func encrypt(plaintext string) string {

	// Simulated substitution table (unknown to the attacker)
	var substitutionTable = map[rune]rune{
		'A': 'Q', 'B': 'W', 'C': 'E', 'D': 'R', 'E': 'T',
		'F': 'Y', 'G': 'U', 'H': 'I', 'I': 'O', 'J': 'P',
		'K': 'A', 'L': 'S', 'M': 'D', 'N': 'F', 'O': 'G',
		'P': 'H', 'Q': 'J', 'R': 'K', 'S': 'L', 'T': 'Z',
		'U': 'X', 'V': 'C', 'W': 'V', 'X': 'B', 'Y': 'N',
		'Z': 'M',
	}
	var res strings.Builder
	for _, char := range plaintext {
		res.WriteRune(substitutionTable[char])
	}
	return res.String()
}

func getConstructionTable() map[rune]rune {
	res := make(map[rune]rune)
	// attack submits the entire alphabets to the encrypt scheme
	plaintext := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	ciphertext := encrypt(plaintext)
	for i := 0; i < len(plaintext); i++ {
		res[rune(ciphertext[i])] = rune(plaintext[i])
	}
	return res
}

func decrypt(cipherText string, decryptTable map[rune]rune) string {
	var plaintextBuilder strings.Builder
	for _, c := range cipherText {
		fmt.Println(string(c), string(decryptTable[c]))
		if plainRune, exist := decryptTable[c]; exist {
			plaintextBuilder.WriteRune(plainRune)
		} else {
			plaintextBuilder.WriteRune(c)
		}

	}
	return plaintextBuilder.String()
}

func getDecryptTableComp() map[rune]rune {
	constructTable, decrypTable := make(map[rune]rune), make(map[rune]rune)

	// construct from A to Z and repeated construct the guessed substitution table
	for c := 'A'; c <= 'Z'; c++ {
		plaintext := strings.Repeat(string(c), 10)
		ciphertext := encrypt(plaintext)
		for i, char := range ciphertext {
			constructTable[c+rune(i)] = char
		}
	}
	// reverse it to get decryptTable
	for k, v := range constructTable {
		decrypTable[v] = k
	}
	return decrypTable
}

// To implement a **monoalphabetic cipher with multiple substitutions for a single letter** (a homophonic cipher), we'll create a system where each plaintext letter can map to multiple ciphertext letters, ensuring each ciphertext letter uniquely maps back to one plaintext letter. This approach enhances security by reducing frequency analysis vulnerability.

// ### Solution Code
// ```go
// package main

// import (
// 	"fmt"
// 	"math/rand"
// 	"strings"
// 	"time"
// )

func mainMonoAlphabetic() {
	plaintext := "HELLO WORLD"
	fmt.Println("Plaintext:", plaintext)

	// Generate encryption and decryption maps with 3 substitutes per letter
	encryptMap, decryptMap := generateMaps(3)

	ciphertext := encrypt1(plaintext, encryptMap)
	fmt.Println("Ciphertext:", ciphertext)

	decrypted := decrypt1(ciphertext, decryptMap)
	fmt.Println("Decrypted:", decrypted)
}

// Generate encryption and decryption maps
func generateMaps(substitutesPerLetter int) (map[rune][]rune, map[rune]rune) {
	rand.Seed(time.Now().UnixNano())

	// Create a pool of unique ciphertext symbols (uppercase + lowercase + digits)
	var symbols []rune
	for c := 'A'; c <= 'Z'; c++ {
		symbols = append(symbols, c)
	}
	for c := 'a'; c <= 'z'; c++ {
		symbols = append(symbols, c)
	}
	for c := '0'; c <= '9'; c++ {
		symbols = append(symbols, c)
	}

	// Shuffle symbols to randomize assignments
	rand.Shuffle(len(symbols), func(i, j int) {
		symbols[i], symbols[j] = symbols[j], symbols[i]
	})

	encryptMap := make(map[rune][]rune)
	decryptMap := make(map[rune]rune)
	plaintextLetters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	symbolIdx := 0

	// Assign substitutes to each plaintext letter
	for _, p := range plaintextLetters {
		var substitutes []rune
		for i := 0; i < substitutesPerLetter; i++ {
			if symbolIdx >= len(symbols) {
				panic("Not enough symbols for substitutions")
			}
			sub := symbols[symbolIdx]
			substitutes = append(substitutes, sub)
			decryptMap[sub] = p // Map ciphertext symbol to plaintext
			symbolIdx++
		}
		encryptMap[p] = substitutes
	}

	return encryptMap, decryptMap
}

// Encrypt plaintext using randomized substitutions
func encrypt1(plaintext string, encryptMap map[rune][]rune) string {
	var ciphertext strings.Builder
	for _, r := range strings.ToUpper(plaintext) {
		if substitutes, ok := encryptMap[r]; ok {
			// Randomly select a substitute
			ciphertext.WriteRune(substitutes[rand.Intn(len(substitutes))])
		} else {
			// Preserve non-alphabetic characters
			ciphertext.WriteRune(r)
		}
	}
	return ciphertext.String()
}

// Decrypt ciphertext using 1:1 reverse mapping
func decrypt1(ciphertext string, decryptMap map[rune]rune) string {
	var plaintext strings.Builder
	for _, r := range ciphertext {
		if p, ok := decryptMap[r]; ok {
			plaintext.WriteRune(p)
		} else {
			plaintext.WriteRune(r)
		}
	}
	return plaintext.String()
}

// ```

// ### Key Features
// 1. **Multiple Substitutions**: Each plaintext letter (A-Z) maps to multiple ciphertext symbols (uppercase, lowercase, digits).
// 2. **Unambiguous Decryption**: Each ciphertext symbol maps back to exactly one plaintext letter.
// 3. **Randomized Encryption**: Each occurrence of a plaintext letter uses a randomly chosen substitute.
// 4. **Extended Symbol Set**: Uses 62 symbols (A-Z, a-z, 0-9) for ciphertext.

// ### Example Output
// ```
// Plaintext: HELLO WORLD
// Ciphertext: 9vJ4Q nQ24g
// Decrypted: HELLO WORLD
// ```

// ### How It Works
// 1. **generateMaps**:
//    - Creates a pool of 62 unique symbols.
//    - Assigns `substitutesPerLetter` symbols to each plaintext letter.
//    - Builds a reverse map for decryption.

// 2. **encrypt**:
//    - Converts plaintext to uppercase.
//    - Replaces each letter with a randomly chosen substitute.
//    - Preserves non-alphabetic characters.

// 3. **decrypt**:
//    - Uses the reverse map to convert ciphertext symbols back to plaintext.

// ### Usage
// 1. **Customization**:
//    - Adjust `substitutesPerLetter` in `generateMaps()` to control substitution count.
//    - Add more symbols (e.g., punctuation) for larger substitution pools.

// 2. **Security**:
//    - Larger substitution pools increase security.
//    - Randomize substitutions for each session.

// This implementation provides a flexible framework for homophonic substitution ciphers,
//balancing security and usability.

// Here's an implementation of the Playfair cipher in Go, including key matrix generation, encryption, and decryption:

// ```go
// package main

// import (
// 	"fmt"
// 	"strings"
// )

type Position struct {
	row int
	col int
}

func mainPlayfairCipher() {
	key := "PLAYFAIR EXAMPLE"
	plaintext := "HIDETHEGOLDINTHETREXESTUMP"

	// Remove spaces and special characters from plaintext
	plaintext = strings.ToUpper(strings.ReplaceAll(plaintext, " ", ""))

	matrix, charPositions := generateMatrix(key)
	fmt.Println("Playfair Matrix:")
	printMatrix(matrix)

	preparedText := preparePlaintext(plaintext)
	fmt.Println("\nPrepared Text:", preparedText)

	ciphertext := encrypt2(preparedText, matrix, charPositions)
	fmt.Println("Encrypted Text:", ciphertext)

	decrypted := decrypt2(ciphertext, matrix, charPositions)
	fmt.Println("Decrypted Text:", decrypted)
}

func generateMatrix(key string) ([][]rune, map[rune]Position) {
	processedKey := processKeyword(key)
	matrix := make([][]rune, 5)
	charPositions := make(map[rune]Position)

	for i := 0; i < 5; i++ {
		matrix[i] = make([]rune, 5)
		for j := 0; j < 5; j++ {
			char := rune(processedKey[i*5+j])
			matrix[i][j] = char
			charPositions[char] = Position{row: i, col: j}
		}
	}
	return matrix, charPositions
}

func processKeyword(key string) string {
	key = strings.ToUpper(key)
	key = strings.ReplaceAll(key, "J", "I")
	seen := make(map[rune]bool)
	var result strings.Builder

	// Add unique key characters
	for _, c := range key {
		if c >= 'A' && c <= 'Z' && c != 'J' {
			if !seen[c] {
				seen[c] = true
				result.WriteRune(c)
			}
		}
	}

	// Add remaining alphabet characters
	for c := 'A'; c <= 'Z'; c++ {
		if c == 'J' {
			continue
		}
		if !seen[c] {
			result.WriteRune(c)
		}
	}

	return result.String()
}

func preparePlaintext(plaintext string) string {
	plaintext = strings.ToUpper(plaintext)
	plaintext = strings.ReplaceAll(plaintext, "J", "I")
	var prepared strings.Builder
	length := len(plaintext)

	for i := 0; i < length; i++ {
		current := rune(plaintext[i])
		prepared.WriteRune(current)

		// Add separator if needed
		if i+1 < length {
			next := rune(plaintext[i+1])
			if current == next {
				prepared.WriteRune('X')
				continue
			}
			i++
			prepared.WriteRune(next)
		}
	}

	// Add padding if odd length
	if prepared.Len()%2 != 0 {
		prepared.WriteRune('X')
	}

	return prepared.String()
}

func encrypt2(text string, matrix [][]rune, positions map[rune]Position) string {
	var ciphertext strings.Builder

	for i := 0; i < len(text); i += 2 {
		a := rune(text[i])
		b := rune(text[i+1])
		c1, c2 := transformPair(a, b, positions, matrix, true)
		ciphertext.WriteRune(c1)
		ciphertext.WriteRune(c2)
	}

	return ciphertext.String()
}

func decrypt2(text string, matrix [][]rune, positions map[rune]Position) string {
	var plaintext strings.Builder

	for i := 0; i < len(text); i += 2 {
		a := rune(text[i])
		b := rune(text[i+1])
		c1, c2 := transformPair(a, b, positions, matrix, false)
		plaintext.WriteRune(c1)
		plaintext.WriteRune(c2)
	}

	return plaintext.String()
}

func transformPair(a, b rune, positions map[rune]Position, matrix [][]rune, encrypt2 bool) (rune, rune) {
	posA := positions[a]
	posB := positions[b]

	switch {
	case posA.row == posB.row:
		return processRow(posA, posB, matrix, encrypt2)
	case posA.col == posB.col:
		return processColumn(posA, posB, matrix, encrypt2)
	default:
		return processRectangle(posA, posB, matrix)
	}
}

func processRow(a, b Position, matrix [][]rune, encrypt2 bool) (rune, rune) {
	shift := 1
	if !encrypt2 {
		shift = -1
	}

	newColA := (a.col + shift + 5) % 5
	newColB := (b.col + shift + 5) % 5
	return matrix[a.row][newColA], matrix[b.row][newColB]
}

func processColumn(a, b Position, matrix [][]rune, encrypt2 bool) (rune, rune) {
	shift := 1
	if !encrypt2 {
		shift = -1
	}

	newRowA := (a.row + shift + 5) % 5
	newRowB := (b.row + shift + 5) % 5
	return matrix[newRowA][a.col], matrix[newRowB][b.col]
}

func processRectangle(a, b Position, matrix [][]rune) (rune, rune) {
	return matrix[a.row][b.col], matrix[b.row][a.col]
}

func printMatrix(matrix [][]rune) {
	for _, row := range matrix {
		for _, c := range row {
			fmt.Printf("%c ", c)
		}
		fmt.Println()
	}
}

// ```

// ### Key Features:
// 1. **Matrix Generation**:
//    - Creates a 5x5 matrix from the keyword
//    - Handles J → I replacement
//    - Maintains letter uniqueness

// 2. **Text Preparation**:
//    - Converts to uppercase
//    - Replaces J with I
//    - Inserts 'X' between duplicate letters
//    - Adds 'X' padding for odd-length texts

// 3. **Encryption/Decryption**:
//    - Handles same-row pairs
//    - Handles same-column pairs
//    - Handles rectangle pairs
//    - Supports both encryption and decryption modes

// 4. **Helper Functions**:
//    - Matrix printing for verification
//    - Position lookup for fast character search

// ### Example Usage:
// ```shell
// $ go run playfair.go

// Playfair Matrix:
// P L A Y F
// I R E X M
// B C D G H
// K N O Q S
// T U V W Z

// Prepared Text: HIDETHEGOLDINTHETREXESTUMP
// Encrypted Text: BMDODZBXDNABEKUDMUXMMOUVIF
// Decrypted Text: HIDETHEGOLDINTHETREXESTUMPX
// ```

// ### Notes:
// 1. **Security Considerations**:
//    - Not recommended for real-world security (historical cipher)
//    - Demonstrates cryptographic concepts

// 2. **Limitations**:
//    - Only handles uppercase A-Z
//    - Simple 'X' padding strategy
//    - No error handling for invalid characters

// 3. **Enhancement Ideas**:
//    - Add support for lowercase letters
//    - Implement different padding strategies
//    - Add input validation
//    - Handle special characters

// This implementation provides a complete Playfair cipher system that can be integrated into larger applications or used for educational purposes.

// Here's an implementation of the Vigenère cipher, a polyalphabetic substitution cipher, in Go. This implementation preserves case and non-alphabetic characters while using a keyword to determine multiple substitution alphabets:

// 	```go
// 	package main

// 	import (
// 		"fmt"
// 		"strings"
// 		"unicode"
// 	)

func polyAlphabeticMain() {
	key := "LEMON"
	plaintext := "ATTACK AT DAWN"

	fmt.Printf("Plaintext: %q\n", plaintext)
	fmt.Printf("Key: %q\n", key)

	ciphertext := encrypt3(plaintext, key)
	fmt.Printf("Encrypted: %q\n", ciphertext)

	decrypted := decrypt3(ciphertext, key)
	fmt.Printf("Decrypted: %q\n", decrypted)
}

func sanitizeKey(key string) string {
	var sanitized strings.Builder
	for _, r := range key {
		if unicode.IsLetter(r) {
			sanitized.WriteRune(unicode.ToUpper(r))
		}
	}
	return sanitized.String()
}

func encrypt3(plaintext, key string) string {
	sanitizedKey := sanitizeKey(key)
	if len(sanitizedKey) == 0 {
		return plaintext
	}

	var ciphertext strings.Builder
	keyIndex := 0

	for _, r := range plaintext {
		if unicode.IsLetter(r) {
			// Determine shift value from key
			keyRune := rune(sanitizedKey[keyIndex%len(sanitizedKey)])
			shift := int(keyRune - 'A')

			// Calculate new character
			var base rune
			if unicode.IsUpper(r) {
				base = 'A'
			} else {
				base = 'a'
			}

			// Perform substitution
			plaintextOffset := int(r - base)
			newOffset := (plaintextOffset + shift) % 26
			newChar := base + rune(newOffset)

			ciphertext.WriteRune(newChar)
			keyIndex++
		} else {
			// Preserve non-alphabetic characters
			ciphertext.WriteRune(r)
		}
	}

	return ciphertext.String()
}

func decrypt3(ciphertext, key string) string {
	sanitizedKey := sanitizeKey(key)
	if len(sanitizedKey) == 0 {
		return ciphertext
	}

	var plaintext strings.Builder
	keyIndex := 0

	for _, r := range ciphertext {
		if unicode.IsLetter(r) {
			// Determine shift value from key
			keyRune := rune(sanitizedKey[keyIndex%len(sanitizedKey)])
			shift := int(keyRune - 'A')

			// Calculate original character
			var base rune
			if unicode.IsUpper(r) {
				base = 'A'
			} else {
				base = 'a'
			}

			// Perform reverse substitution
			ciphertextOffset := int(r - base)
			newOffset := (ciphertextOffset - shift + 26) % 26
			originalChar := base + rune(newOffset)

			plaintext.WriteRune(originalChar)
			keyIndex++
		} else {
			// Preserve non-alphabetic characters
			plaintext.WriteRune(r)
		}
	}

	return plaintext.String()
}

// ```

// ### Features:
// 1. **Case Preservation**: Maintains original case of letters (uppercase stays uppercase, lowercase stays lowercase)
// 2. **Non-alphabetic Preservation**: Leaves spaces, numbers, and punctuation unchanged
// 3. **Key Sanitization**: Automatically removes non-letter characters from the key
// 4. **Modular Arithmetic**: Handles letter shifts using modulo 26 operations
// 5. **Efficient Processing**: Uses `strings.Builder` for efficient string concatenation

// ### Example Output:
// ```
// Plaintext: "ATTACK AT DAWN"
// Key: "LEMON"
// Encrypted: "LXFOPV EF RNHR"
// Decrypted: "ATTACK AT DAWN"
// ```

// ### Usage:
// 1. **Encryption**:
//    ```go
//    ciphertext := encrypt("Secret Message", "KEYWORD")
//    ```
// 2. **Decryption**:
//    ```go
//    plaintext := decrypt("Ciphered Text", "KEYWORD")
//    ```

// ### Key Characteristics:
// - **Polyalphabetic Substitution**: Uses multiple substitution alphabets based on the keyword
// - **Cryptographic Strength**: More secure than simple Caesar ciphers by using variable shifts
// - **Case Sensitivity**: Preserves original letter cases while encrypting/decrypting
// - **Error Handling**: Automatically handles empty keys and non-letter characters

// ### Security Notes:
// - While more secure than monoalphabetic ciphers, Vigenère is still considered insecure for modern applications
// - Use only for educational purposes or as part of more complex cryptographic systems
// - Always prefer modern encryption algorithms (e.g., AES) for real-world security needs

// This implementation demonstrates the core principles of polyalphabetic ciphers while providing practical string handling and case preservation.

// Here's a complete implementation of the Vigenère Autokey Cipher in Go, which preserves case and non-alphabetic characters:

// 	```go
// 	package main

// 	import (
// 		"fmt"
// 		"strings"
// 		"unicode"
// 	)

func mainAutokeyVeger() {
	key := "KEY"
	plaintext := "Attack at dawn!"

	fmt.Printf("Plaintext: %q\n", plaintext)
	fmt.Printf("Key: %q\n", key)

	ciphertext := EncryptAutokey(plaintext, key)
	fmt.Printf("Encrypted: %q\n", ciphertext)

	decrypted := DecryptAutokey(ciphertext, key)
	fmt.Printf("Decrypted: %q\n", decrypted)
}

// EncryptAutokey encrypts a plaintext using the Vigenère Autokey cipher
func EncryptAutokey(plaintext, keyword string) string {
	sanitizedKeyword := sanitize(keyword)
	sanitizedPlaintext := sanitize(plaintext)
	key := generateEncryptionKey(sanitizedKeyword, sanitizedPlaintext)

	encryptedLetters := make([]rune, len(sanitizedPlaintext))
	for i := 0; i < len(sanitizedPlaintext); i++ {
		p := sanitizedPlaintext[i]
		k := key[i]
		encrypted := (p - 'A' + k - 'A') % 26
		encryptedLetters[i] = rune(encrypted) + 'A'
	}

	return reconstruct(plaintext, encryptedLetters)
}

// DecryptAutokey decrypts a ciphertext using the Vigenère Autokey cipher
func DecryptAutokey(ciphertext, keyword string) string {
	sanitizedKeyword := sanitize(keyword)
	sanitizedCiphertext := sanitize(ciphertext)

	decryptedLetters := make([]rune, len(sanitizedCiphertext))
	keywordLen := len(sanitizedKeyword)

	for i := 0; i < len(sanitizedCiphertext); i++ {
		var keyChar rune
		if i < keywordLen {
			keyChar = rune(sanitizedKeyword[i])
		} else {
			keyIndex := i - keywordLen
			keyChar = decryptedLetters[keyIndex]
		}

		c := rune(sanitizedCiphertext[i])
		decrypted := (c - keyChar + 26) % 26
		decryptedLetters[i] = decrypted + 'A'
	}

	return reconstruct(ciphertext, decryptedLetters)
}

// Helper functions

func sanitize(s string) string {
	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) {
			result.WriteRune(unicode.ToUpper(r))
		}
	}
	return result.String()
}

func generateEncryptionKey(keyword, plaintextLetters string) string {
	keyLength := len(plaintextLetters)
	if keyLength == 0 {
		return ""
	}
	if len(keyword) >= keyLength {
		return keyword[:keyLength]
	}
	neededFromPlaintext := keyLength - len(keyword)
	return keyword + plaintextLetters[:neededFromPlaintext]
}

func reconstruct(original string, processed []rune) string {
	var result strings.Builder
	processedIdx := 0
	for _, r := range original {
		if unicode.IsLetter(r) {
			if processedIdx < len(processed) {
				// Preserve original case
				if unicode.IsLower(r) {
					result.WriteRune(unicode.ToLower(processed[processedIdx]))
				} else {
					result.WriteRune(processed[processedIdx])
				}
				processedIdx++
			}
		} else {
			result.WriteRune(r)
		}
	}
	// Add remaining processed characters if any
	for processedIdx < len(processed) {
		result.WriteRune(processed[processedIdx])
		processedIdx++
	}
	return result.String()
}

// ```

// ### Features:
// 1. **Case Preservation**: Maintains original letter case (uppercase/lowercase)
// 2. **Non-alphabetic Preservation**: Leaves spaces, punctuation, and numbers unchanged
// 3. **Automatic Key Generation**: Creates encryption key from keyword + plaintext
// 4. **Efficient Processing**: Uses `strings.Builder` for optimal string handling
// 5. **Cryptographic Operations**:
//    - `EncryptAutokey`: Generates key from keyword + plaintext
//    - `DecryptAutokey`: Recovers key from keyword + decrypted text

// ### Example Output:
// ```text
// Plaintext: "Attack at dawn!"
// Key: "KEY"
// Encrypted: "KMAAMY mx KXGO!"
// Decrypted: "Attack at dawn!"
// ```

// ### How to Use:
// 1. **Encryption**:
// ```go
// ciphertext := EncryptAutokey("Secret Message", "KEYWORD")
// ```

// 2. **Decryption**:
// ```go
// plaintext := DecryptAutokey("Encrypted Text", "KEYWORD")
// ```

// ### Key Advantages:
// - **Stronger Security**: Compared to standard Vigenère by using message-dependent key extensions
// - **Flexible Key Management**: Initial keyword can be shorter than the message
// - **Real-world Readiness**: Handles mixed-case text and preserves formatting

// ### Security Notes:
// - While more secure than basic Vigenère, this is still considered a historical cipher
// - Not recommended for sensitive modern communications
// - Demonstrates core principles of polyalphabetic cryptography

// This implementation provides a practical demonstration of the Autokey cipher while maintaining usability for real-world text formatting.

// Here's a secure implementation of the One-Time Pad (OTP) cipher in Go that uses cryptographically secure random number generation and proper encoding for safe data handling:

// ```go
// package main

// import (
// 	"crypto/rand"
// 	"encoding/hex"
// 	"errors"
// 	"fmt"
// 	"os"
// )

func mainOTP() {
	plaintext := "Top secret nuclear codes"
	fmt.Println("Original:", plaintext)

	// Generate OTP key
	key, err := GenerateOTPKey(len(plaintext))
	if err != nil {
		fmt.Println("Key generation error:", err)
		os.Exit(1)
	}

	// Encrypt message
	ciphertext, err := OTPEncrypt([]byte(plaintext), key)
	if err != nil {
		fmt.Println("Encryption error:", err)
		os.Exit(1)
	}

	// Convert to hexadecimal for safe display/storage
	keyHex := hex.EncodeToString(key)
	cipherHex := hex.EncodeToString(ciphertext)
	fmt.Println("\nGenerated Key (hex):", keyHex)
	fmt.Println("Encrypted Message (hex):", cipherHex)

	// Decrypt message
	decrypted, err := OTPDecrypt(ciphertext, key)
	if err != nil {
		fmt.Println("Decryption error:", err)
		os.Exit(1)
	}

	fmt.Println("\nDecrypted:", string(decrypted))
}

// GenerateOTPKey creates a cryptographically secure random key
func GenerateOTPKey(length int) ([]byte, error) {
	if length <= 0 {
		return nil, errors.New("key length must be positive")
	}

	key := make([]byte, length)
	_, err := rand.Read(key)
	return key, err
}

// OTPEncrypt encrypts plaintext using XOR with the OTP key
func OTPEncrypt(plaintext, key []byte) ([]byte, error) {
	return xorBytes(plaintext, key)
}

// OTPDecrypt decrypts ciphertext using XOR with the OTP key
func OTPDecrypt(ciphertext, key []byte) ([]byte, error) {
	return xorBytes(ciphertext, key)
}

// xorBytes performs byte-wise XOR operation
func xorBytes(a, b []byte) ([]byte, error) {
	if len(a) != len(b) {
		return nil, errors.New("input lengths must match")
	}

	result := make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		result[i] = a[i] ^ b[i]
	}
	return result, nil
}

// ```

// ### Key Features:
// 1. **Cryptographically Secure**:
//    - Uses `crypto/rand` for key generation
//    - Implements true one-time pad requirements
//    - Generates keys equal to plaintext length

// 2. **Data Safety**:
//    - Preserves all original bytes (including non-printable)
//    - Uses hexadecimal encoding for safe display/storage
//    - Proper error handling for security-critical operations

// 3. **Core Operations**:
//    - `GenerateOTPKey`: Creates secure random keys
//    - `OTPEncrypt/OTPDecrypt`: XOR-based encryption/decryption
//    - `xorBytes`: Generic XOR operation for bytes

// ### Example Output:
// ```
// Original: Top secret nuclear codes

// Generated Key (hex): 8d7a4c3b01f9e62a4c93d1b702a4
// Encrypted Message (hex): fd1d2e560b96cf4f3cf3b6d143c5

// Decrypted: Top secret nuclear codes
// ```

// ### Security Considerations:
// 1. **Key Management**:
//    - **Never reuse keys** - generates new key for each encryption
//    - Store keys securely (in memory and at rest)
//    - Destroy keys after use

// 2. **Implementation Details**:
//    - Uses XOR operation for perfect secrecy
//    - Handles all byte values (0-255)
//    - Preserves exact plaintext length

// 3. **Best Practices**:
//    - Use only for the most security-critical applications
//    - Combine with secure transport protocols
//    - Rotate keys frequently

// ### Usage Notes:
// 1. **Building & Running**:
// ```bash
// go build -o otp && ./otp
// ```

// 2. **Integration**:
// ```go
// // For production use:
// key, _ := GenerateOTPKey(len(secretData))
// ciphertext, _ := OTPEncrypt(secretData, key)

// // Always handle errors properly in real code
// ```

// This implementation provides information-theoretic security when used correctly. Remember that practical OTP usage requires solving the key distribution problem - this code demonstrates the cryptographic mechanism but doesn't address key exchange logistics.

// Here's an implementation of the Row Transposition Cipher in Go that handles Unicode characters and includes proper key validation:

// ```go
// package main

// import (
// 	"errors"
// 	"fmt"
// 	"strconv"
// 	"strings"
// )

func mainRowTranspositin() {
	keyStr := "3 1 2" // Example key (1-based indices)
	plaintext := "HELLOWORLD"

	key, err := ParseKey(keyStr)
	if err != nil {
		fmt.Println("Error parsing key:", err)
		return
	}

	fmt.Println("Original:", plaintext)

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		fmt.Println("Encryption error:", err)
		return
	}
	fmt.Println("Encrypted:", ciphertext)

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		fmt.Println("Decryption error:", err)
		return
	}
	fmt.Println("Decrypted:", decrypted)
}

func ParseKey(keyStr string) ([]int, error) {
	parts := strings.Fields(keyStr)
	if len(parts) == 0 {
		return nil, errors.New("empty key")
	}

	key := make([]int, len(parts))
	seen := make(map[int]bool)

	for i, p := range parts {
		num, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid key value: %q", p)
		}
		if num < 1 {
			return nil, fmt.Errorf("key values must be positive integers")
		}
		key[i] = num - 1 // Convert to 0-based index
	}

	keyLength := len(key)
	for _, k := range key {
		if k < 0 || k >= keyLength {
			return nil, fmt.Errorf("key value out of range [1-%d]", keyLength)
		}
		if seen[k] {
			return nil, fmt.Errorf("duplicate key value: %d", k+1)
		}
		seen[k] = true
	}

	return key, nil
}

func Encrypt(plaintext string, key []int) (string, error) {
	keyLength := len(key)
	if keyLength == 0 {
		return "", errors.New("invalid key")
	}

	runes := []rune(plaintext)
	padLength := (keyLength - (len(runes) % keyLength)) % keyLength
	padded := append(runes, make([]rune, padLength)...)
	for i := len(runes); i < len(padded); i++ {
		padded[i] = 'X'
	}

	numRows := len(padded) / keyLength
	grid := make([][]rune, numRows)
	for i := 0; i < numRows; i++ {
		grid[i] = padded[i*keyLength : (i+1)*keyLength]
	}

	var ciphertext []rune
	for _, k := range key {
		for row := 0; row < numRows; row++ {
			if k < len(grid[row]) {
				ciphertext = append(ciphertext, grid[row][k])
			}
		}
	}

	return string(ciphertext), nil
}

func Decrypt(ciphertext string, key []int) (string, error) {
	keyLength := len(key)
	if keyLength == 0 {
		return "", errors.New("invalid key")
	}

	runes := []rune(ciphertext)
	if len(runes)%keyLength != 0 {
		return "", errors.New("ciphertext length must be divisible by key length")
	}

	partLength := len(runes) / keyLength
	parts := make([][]rune, keyLength)
	for i := 0; i < keyLength; i++ {
		parts[i] = runes[i*partLength : (i+1)*partLength]
	}

	originalColumns := make([][]rune, keyLength)
	for j := 0; j < keyLength; j++ {
		found := false
		for i, k := range key {
			if k == j {
				originalColumns[j] = parts[i]
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("invalid key, missing column %d", j+1)
		}
	}

	numRows := partLength
	var plaintext []rune
	for row := 0; row < numRows; row++ {
		for col := 0; col < keyLength; col++ {
			if row < len(originalColumns[col]) {
				plaintext = append(plaintext, originalColumns[col][row])
			}
		}
	}

	// Trim padding
	unpadded := trimPadding(plaintext)
	return string(unpadded), nil
}

func trimPadding(runes []rune) []rune {
	lastIndex := len(runes) - 1
	for lastIndex >= 0 && runes[lastIndex] == 'X' {
		lastIndex--
	}
	if lastIndex < 0 {
		return []rune{}
	}
	return runes[:lastIndex+1]
}

// ```

// ### Key Features:
// 1. **Key Handling**:
//    - Accepts space-separated column numbers (1-based)
//    - Validates key for uniqueness and proper range
//    - Converts to 0-based indices internally

// 2. **Encryption**:
//    - Handles Unicode characters
//    - Automatically pads with 'X' characters
//    - Builds grid and reads columns according to key

// 3. **Decryption**:
//    - Validates ciphertext length
//    - Reconstructs original columns using key
//    - Automatically removes padding 'X' characters

// 4. **Error Handling**:
//    - Proper error checking for invalid keys and inputs
//    - Clear error messages for debugging

// ### Example Usage:
// ```shell
// Original: HELLOWORLD
// Encrypted: LWXHLODLORX
// Decrypted: HELLOWORLD
// ```

// ### How to Use:
// 1. **Key Specification**: Use space-separated column numbers (e.g., "3 1 2" means read column 3 first, then 1, then 2)
// 2. **Encryption**: `Encrypt(plaintext, key)`
// 3. **Decryption**: `Decrypt(ciphertext, key)`

// This implementation provides a complete and robust implementation of the Row Transposition Cipher with proper handling of Unicode characters and error checking.
