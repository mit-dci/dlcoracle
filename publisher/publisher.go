package publisher

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/gertjaap/dlcoracle/datasources"

	"github.com/gertjaap/dlcoracle/crypto"
	"github.com/gertjaap/dlcoracle/logging"
	"github.com/gertjaap/dlcoracle/store"
)

func Init() {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			Process()
		}
	}()
}

var lastPublished = uint64(0)

func Process() error {
	if lastPublished == 0 {
		lp, err := store.GetLastPublishedTimestamp()
		if err != nil {
			logging.Error.Fatal("Cannot retrieve last published timestamp, skipping process loop", err)
			return err
		}
		lastPublished = lp
	}

	timeNow := uint64(time.Now().Unix())

	subscriptions, err := store.GetSubscriptionsInPeriod(lastPublished, timeNow)
	if err != nil {
		logging.Error.Println("Could not retrieve subscriptions", err)
		return err
	}
	for _, sub := range subscriptions {
		ds, err := datasources.GetDatasource(sub.DatasourceId)
		if err != nil {
			logging.Error.Printf("Found subscription to undefined data source %d", sub.DatasourceId)
			continue
		}

		publishedAlready, err := store.IsPublished(sub.DatasourceId, sub.Timestamp)
		if err != nil {
			logging.Error.Printf("Error determining if this is already published: %s", err.Error())
			continue
		}

		if publishedAlready {
			continue
		}

		valueToPublish, err := ds.Value()
		if err != nil {
			logging.Error.Printf("Could not retrieve value for data source %d: %s", sub.DatasourceId, err.Error())
			continue
		}

		privateKey := crypto.RetrieveKey()
		signingKey, err := crypto.DeriveK(sub.DatasourceId, sub.Timestamp)
		if err != nil {
			logging.Error.Printf("Could not derive signing key for data source %d and timestamp %d : %s", sub.DatasourceId, sub.Timestamp, err.Error())
			continue
		}

		// Zero pad the value before signing. Sign expects a [32]byte message
		var buf bytes.Buffer
		binary.Write(&buf, binary.BigEndian, uint64(0))
		binary.Write(&buf, binary.BigEndian, uint64(0))
		binary.Write(&buf, binary.BigEndian, uint64(0))
		binary.Write(&buf, binary.BigEndian, valueToPublish)

		var msg [32]byte
		var priv [32]byte
		var k [32]byte
		copy(msg[:], buf.Bytes())
		copy(priv[:], privateKey[:])
		copy(k[:], signingKey[:])
		signature, err := crypto.RSign(msg, priv, k)
		if err != nil {
			logging.Error.Printf("Could not sign the message: %s", err.Error())
			continue
		}

		store.Publish(sub.DatasourceId, sub.Timestamp, valueToPublish, signature)
		lastPublished = sub.Timestamp
	}
	lastPublished = timeNow
	return nil
}
