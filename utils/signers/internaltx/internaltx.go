package internaltx

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/mldsa"
)

func IsInternal(tx *types.Transaction) bool {
	sig := tx.RawPQCSignatureValues()

	// 1) Empty or nil signature => internal (genesis/system tx)
	if len(sig) == 0 {
		return true
	}

	// 2) Try decoding MLDSA signature
	signature, err := mldsa.BytesToSig(sig)
	if err != nil {
		return false
	}

	// 3) Internal if the signature part is empty
	return len(signature.Sig) == 0
}

func InternalSender(tx *types.Transaction) common.Address {
	sig := tx.RawPQCSignatureValues()

	// 1) No signature => zero address (system tx)
	if len(sig) == 0 {
		return common.Address{}
	}

	// 2) Parse MLDSA
	signature, err := mldsa.BytesToSig(sig)
	if err != nil || len(signature.Sig) != 0 {
		return common.Address{}
	}

	// 3) Internal tx → derive from embedded public key
	return crypto.PQCPubkeyToAddress(&signature.Pk)
}

func Sender(signer types.Signer, tx *types.Transaction) (common.Address, error) {
	if !IsInternal(tx) {
		return types.Sender(signer, tx)
	}
	return InternalSender(tx), nil
}
