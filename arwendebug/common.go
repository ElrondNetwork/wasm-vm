package arwendebug

import (
	"encoding/hex"
	"math/big"
)

func decodeArguments(arguments []string) ([][]byte, error) {
	result := make([][]byte, len(arguments))

	for i := 0; i < len(arguments); i++ {
		decoded, err := hex.DecodeString(arguments[i])
		if err != nil {
			return nil, ErrInvalidArgumentEncoding
		}

		result[i] = decoded
	}

	return result, nil
}

func parseValue(value string) (*big.Int, error) {
	valueAsBigInt := big.NewInt(0)

	if len(value) > 0 {
		_, ok := valueAsBigInt.SetString(value, 10)
		if !ok {
			return nil, NewRequestError("invalid value (erd)")
		}
	}

	return valueAsBigInt, nil
}
