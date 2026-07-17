package inter

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/utils/cser"
)

var ErrUnknownTxType = errors.New("unknown tx type")

// redundant functions for pqc tx type

// func encodeSig(r, s *big.Int) (sig [64]byte) {
// 	copy(sig[0:], cser.PaddedBytes(r.Bytes(), 32)[:32])
// 	copy(sig[32:], cser.PaddedBytes(s.Bytes(), 32)[:32])
// 	return sig
// }

// func decodeSig(sig [64]byte) (r, s *big.Int) {
// 	r = new(big.Int).SetBytes(sig[:32])
// 	s = new(big.Int).SetBytes(sig[32:64])
// 	return
// }

func TransactionMarshalCSER(w *cser.Writer, tx *types.Transaction) error {
	if tx.Type() != types.PQCTxTypeV1 {
		return ErrUnknownTxType
	}

	// Marker for non-standard tx (6 bits = 0)
	w.BitsW.Write(6, 0)
	// Write tx type (PQCTxTypeV1)
	w.U8(tx.Type())

	// Basic fields
	w.U64(tx.Nonce())
	w.U64(tx.Gas())
	w.BigInt(tx.GasTipCap())
	w.BigInt(tx.GasFeeCap())
	w.BigInt(tx.Value())

	// To address (nullable)
	w.Bool(tx.To() != nil)
	if tx.To() != nil {
		w.FixedBytes(tx.To().Bytes())
	}

	// Data (arbitrary length)
	w.SliceBytes(tx.Data())

	// Signature (raw PQC sig bytes)
	sig := tx.RawPQCSignatureValues()
	w.FixedBytes(sig[:])

	// Chain ID
	w.BigInt(tx.ChainId())

	// Access list
	w.U32(uint32(len(tx.AccessList())))
	for _, tuple := range tx.AccessList() {
		w.FixedBytes(tuple.Address.Bytes())
		w.U32(uint32(len(tuple.StorageKeys)))
		for _, h := range tuple.StorageKeys {
			w.FixedBytes(h.Bytes())
		}
	}

	return nil
}

func TransactionUnmarshalCSER(r *cser.Reader) (*types.Transaction, error) {
	fmt.Printf("first few bits are %v \n", r.BytesR.Bytes()[:10])
	// Read tx type
	if r.BitsR.View(6) != 0 {
		fmt.Printf("tx doens't have 6 init bits \n")
		return nil, ErrUnknownTxType
	}
	r.BitsR.Read(6)
	txType := r.U8()
	if txType != types.PQCTxTypeV1 {
		fmt.Printf("tx type is %v \n", txType)
		return nil, ErrUnknownTxType
	}

	// Basic fields
	nonce := r.U64()
	gasLimit := r.U64()
	gasTipCap := r.BigInt()
	gasFeeCap := r.BigInt()
	amount := r.BigInt()

	// To address
	toExists := r.Bool()
	var to *common.Address
	if toExists {
		var _to common.Address
		r.FixedBytes(_to[:])
		to = &_to
	}

	// Data
	data := r.SliceBytes(ProtocolMaxMsgSize)

	// Signature
	var sig [SigSize]byte
	r.FixedBytes(sig[:])

	// Chain ID
	chainID := r.BigInt()

	// Access list
	accessListLen := r.U32()
	if accessListLen > ProtocolMaxMsgSize/24 {
		return nil, cser.ErrTooLargeAlloc
	}
	accessList := make(types.AccessList, accessListLen)
	for i := range accessList {
		r.FixedBytes(accessList[i].Address[:])
		keysLen := r.U32()
		if keysLen > ProtocolMaxMsgSize/32 {
			return nil, cser.ErrTooLargeAlloc
		}
		accessList[i].StorageKeys = make([]common.Hash, keysLen)
		for j := range accessList[i].StorageKeys {
			r.FixedBytes(accessList[i].StorageKeys[j][:])
		}
	}

	// Construct PQC transaction
	return types.NewTx(&types.PQCTxV1{
		ChainID:    chainID,
		Nonce:      nonce,
		GasTipCap:  gasTipCap,
		GasFeeCap:  gasFeeCap,
		Gas:        gasLimit,
		To:         to,
		Value:      amount,
		Data:       data,
		AccessList: accessList,
		Sig:        sig[:],
	}), nil
}
