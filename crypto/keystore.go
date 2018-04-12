package crypto

import (
	"fmt"

	"github.com/adiabat/btcd/btcec"

	"github.com/awnumar/memguard"
)

type KeyType int

const (
	KeyTypeA KeyType = iota
	KeyTypeB
	KeyTypeQ
)

var safeA *memguard.LockedBuffer
var safeB *memguard.LockedBuffer
var safeQ *memguard.LockedBuffer

func StoreKeys(key *[96]byte) error {
	newA, err := memguard.NewImmutableFromBytes(key[:32])
	if err != nil {
		fmt.Println(err)
		memguard.SafeExit(1)
		return err
	}
	safeA = newA

	newB, err := memguard.NewImmutableFromBytes(key[33:64])
	if err != nil {
		fmt.Println(err)
		memguard.SafeExit(1)
		return err
	}
	safeB = newB

	newQ, err := memguard.NewImmutableFromBytes(key[65:])
	if err != nil {
		fmt.Println(err)
		memguard.SafeExit(1)
		return err
	}
	safeQ = newQ
	return nil
}

func RetrieveKey(keyType KeyType) *[32]byte {
	key := new([32]byte)
	switch keyType {
	case KeyTypeA:
		copy(key[:], safeA.Buffer())
	case KeyTypeB:
		copy(key[:], safeB.Buffer())
	case KeyTypeQ:
		copy(key[:], safeQ.Buffer())
	}
	return key
}

func GetPubKey(keyType KeyType) (*[33]byte, error) {
	result := new([33]byte)
	key := RetrieveKey(keyType)
	_, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), key[:])
	key = nil
	copy(result[:], pubKey.SerializeCompressed()[:])
	return result, nil
}
