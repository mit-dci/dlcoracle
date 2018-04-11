package store

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

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
		_, err := tx.CreateBucketIfNotExists([]byte("Subscriptions"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("Publications"))
		return err
	})

	return err
}

func Subscribe(datasourceId, timestamp uint64) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Subscriptions"))
		key := makeStorageKey(datasourceId, timestamp)

		v := b.Get(key)
		if v == nil { // only add when not yet exists
			err := b.Put(key, []byte{})
			return err
		}
		return nil
	})

	if err != nil {
		logging.Error.Print(err)
		return err
	}

	return nil
}

func Publish(datasourceId, timestamp, value uint64, signature [32]byte) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Publications"))
		key := makeStorageKey(datasourceId, timestamp)

		v := b.Get(key)
		if v != nil { // only add when not yet exists
			return fmt.Errorf("There is already a value published for this datasource and timestamp")
		}

		var buf bytes.Buffer
		binary.Write(&buf, binary.BigEndian, value)
		buf.Write(signature[:])

		err := b.Put(key, buf.Bytes())
		return err
		return nil
	})

	if err != nil {
		logging.Error.Print(err)
		return err
	}

	return nil
}

func IsPublished(datasourceId, timestamp uint64) (bool, error) {
	published := false
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Publications"))
		key := makeStorageKey(datasourceId, timestamp)

		v := b.Get(key)
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

func GetPublication(datasourceId, timestamp uint64) (uint64, [32]byte, error) {
	value := uint64(0)
	signature := [32]byte{}

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Publications"))
		key := makeStorageKey(datasourceId, timestamp)

		v := b.Get(key)
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

func GetLastPublishedTimestamp() (uint64, error) {
	var latest uint64
	err := db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("Publications"))

		c := b.Cursor()
		k, _ := c.Last()

		if k == nil { // No keys, so return - it will return 0 from the parent function
			return nil
		}
		buf := bytes.NewBuffer(k)

		err := binary.Read(buf, binary.BigEndian, &latest)

		return err
	})

	if err != nil {
		return 0, err
	}
	return latest, nil
}

type Subscription struct {
	DatasourceId uint64
	Timestamp    uint64
}

func GetSubscriptionsInPeriod(start uint64, end uint64) ([]Subscription, error) {
	subscriptions := []Subscription{}
	err := db.View(func(tx *bolt.Tx) error {
		// Assume our events bucket exists and has RFC3339 encoded time keys.
		c := tx.Bucket([]byte("Subscriptions")).Cursor()

		// Our time range spans the 90's decade.
		min := makeStorageKey(0, start)
		max := makeStorageKey(^uint64(0), end)

		// Iterate over the 90's.
		for k, _ := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, _ = c.Next() {
			sub, err := makeSubscription(k)
			if err != nil {
				return err
			}
			subscriptions = append(subscriptions, sub)
		}

		return nil
	})

	if err != nil {
		return subscriptions, err
	}
	return subscriptions, nil
}

func makeStorageKey(datasourceId uint64, timestamp uint64) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, timestamp)
	binary.Write(&buf, binary.BigEndian, datasourceId)
	return buf.Bytes()
}

func makeSubscription(storageKey []byte) (Subscription, error) {
	buf := bytes.NewBuffer(storageKey)
	result := Subscription{}
	err := binary.Read(buf, binary.BigEndian, &result.Timestamp)
	if err != nil {
		return result, err
	}
	err = binary.Read(buf, binary.BigEndian, &result.DatasourceId)
	if err != nil {
		return result, err
	}
	return result, nil
}
