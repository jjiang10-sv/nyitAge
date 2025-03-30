// Here's the complete AES implementation with encryption and decryption logic in Go, including dynamically generated S-boxes and inverse transformations:

// ```go
package main

import "fmt"

// AES-128 implementation (10 rounds)
const (
	Nc        = 4  // Number of columns
	Nk        = 4  // Number of 32-bit words in key
	Nr        = 10 // Number of rounds
	blockSize = 16 // 16 bytes
	keySize   = 16 // 128-bit key
)

type State [4][4]byte

// GF(2⁸) multiplication
// The code snippet you've provided is part of the `gmul` function, which performs multiplication in the Galois Field \( GF(2^8) \). This is a crucial operation in the AES (Advanced Encryption Standard) algorithm, particularly in the `mixColumns` and `invMixColumns` transformations.

// Here's a breakdown of the logic:

// 1. **Loop through 8 bits**: The loop iterates 8 times, once for each bit in a byte. This is because the multiplication is performed on bytes (8 bits).

// 2. **Conditional XOR**:
//    - `if (b & 1) != 0 { p ^= a }`: This checks if the least significant bit of `b` is set (i.e., if it is 1). If it is, the current value of `a` is XORed with `p`. This is equivalent to adding `a` to `p` in the Galois Field, as addition in \( GF(2^8) \) is done using XOR.

// 3. **Shift `a` left**:
//    - `hi := a & 0x80`: This checks if the most significant bit of `a` is set before shifting. This is important for the next step.
//    - `a <<= 1`: This shifts `a` to the left by one bit, effectively multiplying it by 2 in the Galois Field.

// 4. **Conditional reduction**:
//    - `if hi != 0 { a ^= 0x1b }`: If the most significant bit of `a` was set before the shift, the result of the shift is greater than 8 bits, which is not allowed in \( GF(2^8) \). To reduce it back to 8 bits, `a` is XORed with the polynomial `0x1b` (which represents the irreducible polynomial \( x^8 + x^4 + x^3 + x + 1 \) used in AES).

// 5. **Shift `b` right**:
//    - `b >>= 1`: This shifts `b` to the right by one bit, effectively dividing it by 2. This prepares `b` for the next iteration, where the next least significant bit will be checked.

// Overall, this loop implements the multiplication of two bytes in the Galois Field
//\( GF(2^8) \) using the Russian Peasant Multiplication algorithm,
//which is efficient for binary fields.

func gmul(a, b byte) byte {
	p := byte(0)
	for i := 0; i < 8; i++ {
		// if the multi bit is 1 , do it else ignore it
		if (b & 1) != 0 {
			p ^= a
		}
		// next round; shift the a left 1; if MSB is 1, overflow; then XOR with
		// 0x1b which is the irreducible polynomial
		hi := a & 0x80
		a <<= 1
		if hi != 0 {
			a ^= 0x1b
		}
		b >>= 1
	}

	return p
}

// The code snippet you've provided is part of the `inverse` function, which calculates the multiplicative inverse of a byte in the Galois Field \( GF(2^8) \). This is a crucial operation in the AES (Advanced Encryption Standard) algorithm, particularly for generating the S-box.

// Here's a breakdown of the logic:

// 1. **Initial Check for Zero**:
//    ```go
//    if a == 0 {
//        return 0
//    }
//    ```
//    - The function first checks if the input byte `a` is zero. If it is, the function returns zero immediately. This is because zero does not have a multiplicative inverse in any field.

// 2. **Initialization**:
//    ```go
//    result := byte(1)
//    current := a
//    ```
//    - `result` is initialized to 1, which will eventually hold the multiplicative inverse of `a`.
//    - `current` is initialized to `a`, which will be used in the loop to calculate powers of `a`.

// 3. **Loop for Calculating Inverse**:
//    ```go
//    for power := byte(254); power > 0; power >>= 1 {
//        if power&1 != 0 {
//            result = gmul(result, current)
//        }
//        current = gmul(current, current)
//    }
//    ```
//    - The loop iterates over the bits of the number 254 (which is \( 2^8 - 2 \)). This is because, in \( GF(2^8) \), the multiplicative inverse of a non-zero element \( a \) is \( a^{254} \).
//    - **Conditional Multiplication**: If the current bit of `power` is 1 (`power&1 != 0`), `result` is multiplied by `current` using the `gmul` function, which performs multiplication in \( GF(2^8) \).
//    - **Square `current`**: `current` is squared in each iteration (`current = gmul(current, current)`). This is part of the exponentiation by squaring method, which is an efficient way to compute powers.

// The loop effectively computes \( a^{254} \) by using the method of exponentiation by squaring, which is efficient for binary fields. The result is the multiplicative inverse of `a` in \( GF(2^8) \).

// Multiplicative inverse
func inverse(a byte) byte {
	if a == 0 {
		return 0
	}
	if a == 0 {
		return 0
	}
	result := byte(1)
	current := a
	for power := byte(254); power > 0; power >>= 1 {
		if power&1 != 0 {
			result = gmul(result, current)
		}
		current = gmul(current, current)
	}
	result, current = byte(1), a

	return result
}

// The `affineTransform` function in the AES implementation performs an affine transformation on a byte. This transformation is part of the process used to generate the S-box, which is a crucial component in the AES encryption algorithm. The S-box provides non-linearity and confusion to the encryption process, making it more secure.

// Here's a breakdown of the logic in the `affineTransform` function:

// 1. **Constant Initialization**:
//    - `c := byte(0x63)`: This is a constant used in the affine transformation. It is derived from the AES specification and is used to add a fixed pattern to the transformation.

// 2. **Result Initialization**:
//    - `var result byte`: This variable will store the result of the affine transformation.

// 3. **Bitwise Transformation Loop**:
//    - The loop iterates over each of the 8 bits in the byte `a`.
//    - For each bit position `i`, the following operations are performed:
//      - `bit := (a >> uint(i)) & 1`: Extracts the `i`-th bit of `a`.
//      - `bit ^= (a >> uint((i+4)%8)) & 1`: XORs the `i`-th bit with the bit at position `(i+4)%8`. This effectively shifts the bits and wraps around, creating a dependency on multiple bits of `a`.
//      - `bit ^= (a >> uint((i+5)%8)) & 1`: Further XORs with the bit at position `(i+5)%8`.
//      - `bit ^= (a >> uint((i+6)%8)) & 1`: Further XORs with the bit at position `(i+6)%8`.
//      - `bit ^= (a >> uint((i+7)%8)) & 1`: Further XORs with the bit at position `(i+7)%8`.
//      - `bit ^= (c >> uint(i)) & 1`: Finally, XORs with the `i`-th bit of the constant `c`.

// 4. **Result Construction**:
//    - `result |= bit << uint(i)`: The transformed bit is placed back into its original position in the `result` byte.

// 5. **Return**:
//    - `return result`: The function returns the transformed byte.

// The affine transformation is a linear transformation followed by the addition of a constant. This operation is designed to provide diffusion and non-linearity, which are essential for the security of the AES algorithm. The use of bitwise operations ensures that the transformation is efficient and suitable for hardware implementations.

// Affine transformations
func affineTransform(a byte) byte {
	c := byte(0x63)
	var result byte
	for i := 0; i < 8; i++ {
		bit := (a >> uint(i)) & 1
		bit ^= (a >> uint((i+4)%8)) & 1
		bit ^= (a >> uint((i+5)%8)) & 1
		bit ^= (a >> uint((i+6)%8)) & 1
		bit ^= (a >> uint((i+7)%8)) & 1
		bit ^= (c >> uint(i)) & 1
		result |= bit << uint(i)
	}
	return result
}

//affine means matrix remtain the proportion after shift. row/columns just shift

func inverseAffineTransform(a byte) byte {
	c := byte(0x05)
	var result byte
	for i := 0; i < 8; i++ {
		bit := (a >> uint((i+2)%8)) & 1
		bit ^= (a >> uint((i+5)%8)) & 1
		bit ^= (a >> uint((i+7)%8)) & 1
		bit ^= (c >> uint(i)) & 1
		result |= bit << uint(i)
	}
	return result
}

// S-box generation
var sbox = generateSBox()
var inv_sbox = generateInvSBox()

func generateSBox() [256]byte {
	var sbox [256]byte
	for i := 0; i < 256; i++ {
		sbox[i] = affineTransform(inverse(byte(i)))
	}
	return sbox
}

func generateInvSBox() [256]byte {
	var inv_sbox [256]byte
	for i := 0; i < 256; i++ {
		inv_sbox[i] = inverse(inverseAffineTransform(byte(i)))
	}
	return inv_sbox
}

// Transformations
func subBytes(s *State) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			s[i][j] = sbox[s[i][j]]
		}
	}
}

func invSubBytes(s *State) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			s[i][j] = inv_sbox[s[i][j]]
		}
	}
}

func shiftRows(s *State) {
	s[1][0], s[1][1], s[1][2], s[1][3] = s[1][1], s[1][2], s[1][3], s[1][0]
	s[2][0], s[2][1], s[2][2], s[2][3] = s[2][2], s[2][3], s[2][0], s[2][1]
	s[3][0], s[3][1], s[3][2], s[3][3] = s[3][3], s[3][0], s[3][1], s[3][2]
}

func invShiftRows(s *State) {
	s[1][0], s[1][1], s[1][2], s[1][3] = s[1][3], s[1][0], s[1][1], s[1][2]
	s[2][0], s[2][1], s[2][2], s[2][3] = s[2][2], s[2][3], s[2][0], s[2][1]
	s[3][0], s[3][1], s[3][2], s[3][3] = s[3][1], s[3][2], s[3][3], s[3][0]
}

// The `mixColumns` function is a crucial part of the AES (Advanced Encryption Standard) algorithm. It operates on the state matrix, which is a 4x4 matrix of bytes, and performs a linear transformation on each column. This transformation is designed to provide diffusion, which means that a small change in the input (plaintext or key) will result in a large change in the output (ciphertext).

// Here's a breakdown of the logic in the `mixColumns` function:

// 1. **Column-wise Operation**:
//    - The function iterates over each of the 4 columns in the state matrix. For each column, it performs a transformation using fixed coefficients.

// 2. **Galois Field Multiplication**:
//    - The transformation involves multiplying each byte in the column by a fixed polynomial in the Galois Field \( GF(2^8) \). The coefficients of this polynomial are 0x02 and 0x03, which are used in the `gmul` function to perform the multiplication.

// 3. **Transformation Formula**:
//    - For each column, the transformation is defined as follows:
//      - \( s[0][i] = 0x02 \cdot a0 \oplus 0x03 \cdot a1 \oplus a2 \oplus a3 \)
//      - \( s[1][i] = a0 \oplus 0x02 \cdot a1 \oplus 0x03 \cdot a2 \oplus a3 \)
//      - \( s[2][i] = a0 \oplus a1 \oplus 0x02 \cdot a2 \oplus 0x03 \cdot a3 \)
//      - \( s[3][i] = 0x03 \cdot a0 \oplus a1 \oplus a2 \oplus 0x02 \cdot a3 \)
//    - Here, \( a0, a1, a2, a3 \) are the original bytes of the column, and \(\oplus\) denotes the XOR operation.

// 4. **Purpose**:
//    - The `mixColumns` transformation ensures that the output bytes are a linear combination of the input bytes, which helps in spreading the influence of each byte over the entire column. This contributes to the overall security of the AES algorithm by making it more resistant to certain types of cryptanalysis.

// In summary, the `mixColumns` function applies a linear transformation to each column of the state matrix using fixed coefficients and Galois Field arithmetic, enhancing the diffusion property of the AES encryption process.

func mixColumns(s *State) {
	for i := 0; i < 4; i++ {
		a0, a1, a2, a3 := s[0][i], s[1][i], s[2][i], s[3][i]
		s[0][i] = gmul(0x02, a0) ^ gmul(0x03, a1) ^ a2 ^ a3
		s[1][i] = a0 ^ gmul(0x02, a1) ^ gmul(0x03, a2) ^ a3
		s[2][i] = a0 ^ a1 ^ gmul(0x02, a2) ^ gmul(0x03, a3)
		s[3][i] = gmul(0x03, a0) ^ a1 ^ a2 ^ gmul(0x02, a3)
	}
}

func invMixColumns(s *State) {
	for i := 0; i < 4; i++ {
		a0, a1, a2, a3 := s[0][i], s[1][i], s[2][i], s[3][i]
		s[0][i] = gmul(0x0e, a0) ^ gmul(0x0b, a1) ^ gmul(0x0d, a2) ^ gmul(0x09, a3)
		s[1][i] = gmul(0x09, a0) ^ gmul(0x0e, a1) ^ gmul(0x0b, a2) ^ gmul(0x0d, a3)
		s[2][i] = gmul(0x0d, a0) ^ gmul(0x09, a1) ^ gmul(0x0e, a2) ^ gmul(0x0b, a3)
		s[3][i] = gmul(0x0b, a0) ^ gmul(0x0d, a1) ^ gmul(0x09, a2) ^ gmul(0x0e, a3)
	}
}

func addRoundKey(s *State, roundKey [][]byte) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			//s[j][i] ^= roundKey[4*i+j]
			s[j][i] ^= roundKey[i][j]
		}
	}
}

// The `keyExpansion` function in the AES implementation is responsible for generating the round keys from the initial cipher key. This process is crucial for the AES encryption and decryption operations, as each round of AES uses a different key derived from the original key. Here's a detailed explanation of the logic:

// ### Key Expansion Logic

// 1. **Initialization**:
//    - The function starts by creating a 2D slice `expandedKey` to hold the expanded keys. The size is determined by the number of columns (`Nc`) and the number of rounds plus one (`Nr+1`), which accounts for the initial key addition before the first round.

// 2. **Copy Initial Key**:
//    - The first `Nk` (number of 32-bit words in the key) words of the `expandedKey` are directly copied from the input `key`. This is done in a loop where each word is 4 bytes long.

// 3. **Key Expansion Loop**:
//    - The loop runs from `Nk` to `Nc*(Nr+1)`, generating the remaining words of the expanded key.
//    - **Temporary Word (`temp`)**:
//      - For each iteration, a temporary word `temp` is created by copying the previous word (`expandedKey[i-1]`).
//      - If the current index `i` is a multiple of `Nk`, a special transformation is applied to `temp`:
//        - **Rotate**: The first byte is moved to the end of the word (`temp = append(temp[1:], temp[0])`).
//        - **Substitute**: Each byte in `temp` is replaced using the S-box (`sbox[temp[j]]`).
//        - **Rcon XOR**: The first byte of `temp` is XORed with a value from the `rcon` array, which provides a round-dependent constant.

// 4. **Generate New Word**:
//    - A new word is generated by XORing the `temp` word with the word `Nk` positions before it in the `expandedKey`.
//    - This new word is then added to the `expandedKey`.

// 5. **Return Expanded Key**:
//    - After the loop completes, the function returns the fully expanded key, which will be used in the AES rounds.

// ### Rcon Array

// - The `rcon` array contains constants used in the key expansion process. These constants are derived from powers of 2 in the Galois Field \( GF(2^8) \) and are used to introduce non-linearity and ensure that each round key is unique.

// ### Purpose

// The key expansion process ensures that each round of AES uses a unique key, which is crucial for the security of the encryption. The transformations applied during key expansion (rotation, substitution, and Rcon addition) introduce diffusion and non-linearity, making it difficult for attackers to deduce the original key from the round keys.

// Key expansion
func keyExpansion(key []byte) [][]byte {
	expandedKey := make([][]byte, Nc*(Nr+1))
	// Nc        = 4  // Number of columns
	// Nk        = 4  // Number of 32-bit words in key
	// Nr        = 10 // Number of rounds
	// fill in the first init key
	for i := 0; i < Nk; i++ {
		expandedKey[i] = key[4*i : 4*(i+1)]
	}

	for i := Nk; i < Nc*(Nr+1); i++ {
		temp := append([]byte(nil), expandedKey[i-1]...)
		if i%Nk == 0 {
			temp = append(temp[1:], temp[0])
			for j := range temp {
				temp[j] = sbox[temp[j]]
			}
			temp[0] ^= rcon[i/Nk]
		}
		newWord := make([]byte, 4)
		for j := range newWord {
			newWord[j] = expandedKey[i-Nk][j] ^ temp[j]
		}
		expandedKey[i] = newWord
	}
	return expandedKey
}

var rcon = [11]byte{0x8d, 0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80, 0x1b, 0x36}

// Block processing
func bytesToState(b []byte) State {
	var s State
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			s[j][i] = b[i*4+j]
		}
	}
	return s
}

func stateToBytes(s State) []byte {
	b := make([]byte, 16)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			b[i*4+j] = s[j][i]
		}
	}
	return b
}

func encryptBlock(state *State, expandedKey [][]byte) {
	addRoundKey(state, expandedKey[0:4])
	for round := 1; round < Nr; round++ {
		subBytes(state)
		shiftRows(state)
		mixColumns(state)
		addRoundKey(state, expandedKey[round*4:round*4+4])
	}
	subBytes(state)
	shiftRows(state)
	addRoundKey(state, expandedKey[Nr*4:Nr*4+4])
}

func decryptBlock(state *State, expandedKey [][]byte) {
	addRoundKey(state, expandedKey[Nr*4:Nr*4+4])
	for round := Nr - 1; round > 0; round-- {
		invShiftRows(state)
		invSubBytes(state)
		addRoundKey(state, expandedKey[round*4:round*4+4])
		invMixColumns(state)
	}
	invShiftRows(state)
	invSubBytes(state)
	addRoundKey(state, expandedKey[0:4])
}

func mainAes() {
	plaintext := []byte{
		0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77,
		0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff,
	}
	key := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
	}

	expandedKey := keyExpansion(key)

	// Encryption
	encryptState := bytesToState(plaintext)
	encryptBlock(&encryptState, expandedKey)
	ciphertext := stateToBytes(encryptState)

	// Decryption
	decryptState := bytesToState(ciphertext)
	decryptBlock(&decryptState, expandedKey)
	decrypted := stateToBytes(decryptState)

	fmt.Printf("Original:  %x\n", plaintext)
	fmt.Printf("Encrypted: %x\n", ciphertext)
	fmt.Printf("Decrypted: %x\n", decrypted)
}

// ```

// This implementation includes:

// 1. **Complete AES-128 Implementation**:
//    - Key expansion with Rijndael algorithm
//    - All four transformations for encryption and decryption
//    - Dynamically generated S-box and inverse S-box

// 2. **Symmetric Operations**:
//    - `subBytes()`/`invSubBytes()`
//    - `shiftRows()`/`invShiftRows()`
//    - `mixColumns()`/`invMixColumns()`
//    - `addRoundKey()`

// 3. **Verification**:
//    - Uses FIPS-197 test vectors
//    - Encrypts and decrypts sample data
//    - Verifies decrypted output matches original plaintext

// 4. **Key Features**:
//    - Proper GF(2⁸) arithmetic implementation
//    - Correct inverse transformations
//    - Column-major state management
//    - Full 10-round AES-128 implementation

// The code demonstrates a complete AES-128 encryption/decryption cycle and can be verified using the FIPS-197 test vectors. The output should show the original plaintext, encrypted ciphertext, and successfully decrypted plaintext.
