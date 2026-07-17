package valkeystore

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/mldsa"

	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
	"github.com/Fantom-foundation/go-opera/valkeystore/encryption"
)

type SignerI interface {
	Sign(pubkey validatorpk.PubKey, digest []byte) ([]byte, error)
}

type Signer struct {
	backend KeystoreI
}

func NewSigner(backend KeystoreI) *Signer {
	return &Signer{
		backend: backend,
	}
}

func (s *Signer) Sign(pubkey validatorpk.PubKey, digest []byte) ([]byte, error) {
	if pubkey.Type != validatorpk.Types.MLDSA {
		return nil, encryption.ErrNotSupportedType
	}
	key, err := s.backend.GetUnlocked(pubkey)
	if err != nil {
		return nil, err
	}

	// only mldsa supported for now
	mldsaSk := key.Decoded.(*mldsa.PrivateKey)

	sig, err := crypto.SignPQC(digest, mldsaSk)
	if err != nil {
		return nil, err
	}
	return sig, err
}
