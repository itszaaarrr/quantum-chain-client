package inter

import "github.com/ethereum/go-ethereum/crypto/mldsa"

// only mldsa supported for now
const SigSize = mldsa.MLDSA_44_SIGNATURE_SIZE + mldsa.MLDSA_44_PUBLIC_KEY_SIZE

// Signature is a mldsa signature of size 3732
type Signature [SigSize]byte

func (s Signature) Bytes() []byte {
	return s[:]
}

func BytesToSignature(b []byte) (sig Signature) {
	if len(b) != SigSize {
		panic("invalid signature length")
	}
	copy(sig[:], b)
	return sig
}
