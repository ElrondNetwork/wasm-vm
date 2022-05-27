package factory

import (
	"github.com/ElrondNetwork/arwen-wasm-vm/v1_5/crypto"
	"github.com/ElrondNetwork/arwen-wasm-vm/v1_5/crypto/hashing"
	"github.com/ElrondNetwork/arwen-wasm-vm/v1_5/crypto/signing/bls"
	"github.com/ElrondNetwork/arwen-wasm-vm/v1_5/crypto/signing/ed25519"
	"github.com/ElrondNetwork/arwen-wasm-vm/v1_5/crypto/signing/secp256k1"
)

// NewVMCrypto returns a composite struct containing VMCrypto functionality implementations
func NewVMCrypto() crypto.VMCrypto {
	return struct {
		crypto.Hasher
		crypto.Ed25519
		crypto.BLS
		crypto.Secp256k1
	}{
		Hasher:    hashing.NewHasher(),
		Ed25519:   ed25519.NewEd25519Signer(),
		BLS:       bls.NewBLS(),
		Secp256k1: secp256k1.NewSecp256k1(),
	}
}
