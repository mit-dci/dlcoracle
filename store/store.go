package store

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/adiabat/btcd/btcec"
	"github.com/boltdb/bolt"
	"github.com/gertjaap/dlcoracle/logging"
)

var db *bolt.DB

func Init() error {
	database, err := bolt.Open("data/oracle.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		logging.Error.Fatal(err)
		return err
	}

	db = database

	// Ensure buckets exist that we need
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte("Keys"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("Publications"))
		return err
	})

	return err
}

func GetRPoint(datasourceId, timestamp uint64) ([33]byte, error) {
	var pubKey [33]byte

	privKey, err := GetK(datasourceId, timestamp)
	if err != nil {
		logging.Error.Print(err)
		return pubKey, err
	}

	_, pk := btcec.PrivKeyFromBytes(btcec.S256(), privKey[:])

	copy(pubKey[:], pk.SerializeCompressed())
	return pubKey, nil
}

func GetK(datasourceId, timestamp uint64) ([32]byte, error) {
	var privKey [32]byte
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Keys"))
		key := makeStorageKey(datasourceId, timestamp)

		priv := b.Get(key)
		if priv == nil {
			_, err := rand.Read(privKey[:])
			if err != nil {
				return err
			}
			err = b.Put(key, privKey[:])
			return err
		} else {
			copy(privKey[:], priv)
		}
		return nil
	})

	if err != nil {
		logging.Error.Print(err)
		return privKey, err
	}
	return privKey, nil
}

func Publish(rPoint [33]byte, value uint64, signature [32]byte) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Publications"))

		v := b.Get(rPoint[:])
		if v != nil { // only add when not yet exists
			return fmt.Errorf("There is already a value published for this rpoint")
		}

		var buf bytes.Buffer
		binary.Write(&buf, binary.BigEndian, value)
		buf.Write(signature[:])

		err := b.Put(rPoint[:], buf.Bytes())
		return err
	})

	if err != nil {
		logging.Error.Print(err)
		return err
	}

	return nil
}

func IsPublished(rPoint [33]byte) (bool, error) {
	published := false
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Publications"))
		v := b.Get(rPoint[:])
		if v != nil { // only add when not yet exists
			published = true
		}

		return nil
	})

	if err != nil {
		logging.Error.Print(err)
		return false, err
	}

	return published, nil
}

func GetPublication(rPoint [33]byte) (uint64, [32]byte, error) {
	value := uint64(0)
	signature := [32]byte{}

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Publications"))
		v := b.Get(rPoint[:])
		if v == nil { // only add when not yet exists
			return fmt.Errorf("Publication not found")
		}

		buf := bytes.NewBuffer(v)

		err := binary.Read(buf, binary.BigEndian, &value)
		if err != nil {
			return err
		}
		copy(signature[:], buf.Next(32))
		return nil
	})

	if err != nil {
		logging.Error.Print(err)
		return 0, [32]byte{}, err
	}

	return value, signature, nil
}

func makeStorageKey(datasourceId uint64, timestamp uint64) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, timestamp)
	binary.Write(&buf, binary.BigEndian, datasourceId)
	return buf.Bytes()
}
