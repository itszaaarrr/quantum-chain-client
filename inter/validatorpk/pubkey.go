package validatorpk

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/mldsa"
	"github.com/pkg/errors"
)

const (
	FakePassword = "fakepassword"
)

type PubKey struct {
	Type uint8
	Raw  []byte
}

var Types = struct {
	MLDSA uint8
}{
	MLDSA: 0x01, // MLDSA / PQC type identifier
}

func (pk PubKey) Empty() bool {
	return len(pk.Raw) == 0 && pk.Type == 0
}

func (pk PubKey) String() string {
	return "0x" + common.Bytes2Hex(pk.Bytes())
}

func (pk PubKey) Bytes() []byte {
	return append([]byte{pk.Type}, pk.Raw...)
}

func (pk PubKey) Copy() PubKey {
	return PubKey{
		Type: pk.Type,
		Raw:  common.CopyBytes(pk.Raw),
	}
}

func FromString(str string) (PubKey, error) {
	return FromBytes(common.FromHex(str))
}

func FromBytes(b []byte) (PubKey, error) {
	if len(b) == 0 {
		return PubKey{}, errors.New("empty pubkey")
	}

	keyType := b[0]
	raw := b[1:]

	//  Enforce MLDSA-only
	if keyType != Types.MLDSA {
		return PubKey{}, errors.Errorf("unsupported public key type 0x%x", keyType)
	}

	//  Validate length matches MLDSA public key size
	if len(raw) != mldsa.MLDSA_44_PUBLIC_KEY_SIZE {
		return PubKey{}, errors.Errorf("invalid MLDSA public key size: got %d, expected %d", len(raw), mldsa.MLDSA_44_PUBLIC_KEY_SIZE)
	}

	return PubKey{Type: keyType, Raw: raw}, nil
}

// MarshalText returns the hex representation of a.
func (pk *PubKey) MarshalText() ([]byte, error) {
	return []byte(pk.String()), nil
}

// UnmarshalText parses a hash in hex syntax.
func (pk *PubKey) UnmarshalText(input []byte) error {
	res, err := FromString(string(input))
	if err != nil {
		return err
	}
	*pk = res
	return nil
}

//
// ===== ✅ Type-Safe Access to MLDSA PublicKey =====
//

// ToMLDSAPubKey converts validator PubKey → mldsa.PublicKey
func (pk PubKey) ToMLDSAPubKey() (*mldsa.PublicKey, error) {
	if pk.Type != Types.MLDSA {
		return nil, errors.New("invalid key type: not MLDSA")
	}
	if len(pk.Raw) != mldsa.MLDSA_44_PUBLIC_KEY_SIZE {
		return nil, errors.New("invalid MLDSA public key size")
	}

	mlKey := &mldsa.PublicKey{}
	if err := mlKey.FromBytes(pk.Raw); err != nil {
		return nil, err
	}
	return mlKey, nil
}
