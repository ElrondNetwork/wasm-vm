package vmhost

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/multiversx/mx-chain-vm-go/math"
)

// Zero is the big integer 0
var Zero = big.NewInt(0)

// One is the big integer 1
var One = big.NewInt(1)

// CustomStorageKey generates a storage key of a specific type.
func CustomStorageKey(keyType string, associatedKey []byte) []byte {
	return append([]byte(keyType), associatedKey...)
}

// BooleanToInt returns 1 if the given bool is true, 0 otherwise
func BooleanToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// GuardedGetBytesSlice extracts a subslice from a given slice, guarding against overstepping the bounds.
func GuardedGetBytesSlice(data []byte, offset int32, length int32) ([]byte, error) {
	dataLength := uint32(len(data))
	isOffsetTooSmall := offset < 0
	isOffsetTooLarge := uint32(offset) > dataLength
	requestedEnd := math.AddInt32(offset, length)
	isRequestedEndTooLarge := uint32(requestedEnd) > dataLength
	isLengthNegative := length < 0

	if isOffsetTooSmall || isOffsetTooLarge || isRequestedEndTooLarge {
		return nil, fmt.Errorf("GuardedGetBytesSlice: bad bounds")
	}

	if isLengthNegative {
		return nil, fmt.Errorf("GuardedGetBytesSlice: negative length")
	}

	result := data[offset:requestedEnd]
	return result, nil
}

// PadBytesLeft adds a specified number of zeros to the left of a byte slice.
func PadBytesLeft(data []byte, size int) []byte {
	if data == nil {
		return nil
	}
	if len(data) == 0 {
		return []byte{}
	}
	padSize := math.SubInt(size, len(data))
	if padSize <= 0 {
		return data
	}

	paddedBytes := make([]byte, 0, padSize)
	paddedBytes = append(paddedBytes, data...)
	return paddedBytes
}

// InverseBytes reverses the order of a byte slice.
func InverseBytes(data []byte) []byte {
	length := len(data)
	invBytes := make([]byte, length)
	for i := 0; i < length; i++ {
		invBytes[length-i-1] = data[i]
	}
	return invBytes
}

// GetSCCode retrieves the bytecode of a WASM module from a file.
func GetSCCode(fileName string) []byte {
	code, err := os.ReadFile(filepath.Clean(fileName))
	if err != nil {
		panic(fmt.Sprintf("GetSCCode(): %s", fileName))
	}
	return code
}

// GetTestSCCode retrieves the bytecode of a WASM testing module.
func GetTestSCCode(scName string, prefixToTestSCs string) []byte {
	pathToSC := prefixToTestSCs + "test/contracts/" + scName + "/output/" + scName + ".wasm"
	return GetSCCode(pathToSC)
}

// U64ToLEB128 encodes an uint64 using LEB128 (Little Endian Base 128), used in WASM bytecode
// See https://en.wikipedia.org/wiki/LEB128
// Copied from https://github.com/filecoin-project/go-leb128/blob/master/leb128.go
func U64ToLEB128(n uint64) (out []byte) {
	more := true
	for more {
		b := byte(n & 0x7F)
		n >>= 7
		if n == 0 {
			more = false
		} else {
			b |= 0x80
		}
		out = append(out, b)
	}
	return
}

// IfNil tests if the provided interface pointer or underlying object is nil
func IfNil(checker nilInterfaceChecker) bool {
	if checker == nil {
		return true
	}
	return checker.IsInterfaceNil()
}

type nilInterfaceChecker interface {
	IsInterfaceNil() bool
}
