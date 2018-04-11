package crypto

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/adiabat/btcutil/hdkeychain"
	"github.com/awnumar/memguard"
	"github.com/gertjaap/dlcoracle/logging"
	"github.com/mit-dci/lit/coinparam"
)

var safeKey *memguard.LockedBuffer

func StoreKey(key *[32]byte) error {
	newKey, err := memguard.NewImmutableFromBytes(key[:])
	if err != nil {
		fmt.Println(err)
		memguard.SafeExit(1)
		return err
	}
	safeKey = newKey
	return nil
}

func RetrieveKey() *[32]byte {
	key := new([32]byte)
	copy(key[:], safeKey.Buffer())
	return key
}

func GetPubKey() (*[33]byte, error) {
	result := new([33]byte)
	key := RetrieveKey()
	masterKey, err := hdkeychain.NewMaster(key[:], &coinparam.BitcoinParams)
	if err != nil {
		logging.Error.Printf("Create master key %s\n", err.Error())
		return result, err
	}
	key = nil
	pubKey, err := masterKey.ECPubKey()
	if err != nil {
		logging.Error.Printf("get pubkey %s\n", err.Error())
		return result, err
	}

	copy(result[:], pubKey.SerializeCompressed()[:])
	return result, nil

}

func DeriveK(datasourceId uint64, timestamp uint64) ([32]byte, error) {
	key := RetrieveKey()
	id := keyDerivationPayload(datasourceId, timestamp)

	var flatKey [32]byte
	copy(flatKey[:], key[:])
	key = nil
	k, _ := deriveK(flatKey, id)
	return k, nil
}

func DeriveR(datasourceId uint64, timestamp uint64) ([33]byte, error) {
	key := RetrieveKey()
	id := keyDerivationPayload(datasourceId, timestamp)

	var flatKey [32]byte
	copy(flatKey[:], key[:])
	_, R := deriveK(flatKey, id)
	return R, nil
}

func keyDerivationPayload(datasourceId uint64, timestamp uint64) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, datasourceId)
	binary.Write(&buf, binary.BigEndian, timestamp)
	return buf.Bytes()
}
