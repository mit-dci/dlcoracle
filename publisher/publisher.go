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

var lastPublished = uint64(0)

func Init() {
	//TO DO: Store in database and retrieve from there on start up
	lastPublished = uint64(time.Now().Unix())

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			Process()
		}
	}()
}

func Process() error {
	timeNow := uint64(time.Now().Unix())
	for time := lastPublished + 1; time <= timeNow; time++ {
		for _, ds := range datasources.GetAllDatasources() {
			if time%ds.Interval() == 0 {

				logging.Info.Printf("Publishing data source %d [ts: %d]\n", ds.Id(), time)

				valueToPublish, err := ds.Value()
				if err != nil {
					logging.Error.Printf("Could not retrieve value for data source %d: %s", ds.Id(), err.Error())
					continue
				}

				var a [32]byte
				copy(a[:], crypto.RetrieveKey(crypto.KeyTypeA)[:])

				k, err := store.GetK(ds.Id(), time)
				if err != nil {
					logging.Error.Printf("Could not get signing key for data source %d and timestamp %d : %s", ds.Id(), time, err.Error())
					continue
				}

				R, err := store.GetRPoint(ds.Id(), time)
				if err != nil {
					logging.Error.Printf("Could not get pubkey for data source %d and timestamp %d : %s", ds.Id(), time, err.Error())
					continue
				}

				publishedAlready, err := store.IsPublished(R)
				if err != nil {
					logging.Error.Printf("Error determining if this is already published: %s", err.Error())
					continue
				}

				if publishedAlready {
					logging.Info.Printf("Already published for data source %d and timestamp %d", ds.Id(), time)
					continue
				}

				// Zero pad the value before signing. Sign expects a [32]byte message
				var buf bytes.Buffer
				binary.Write(&buf, binary.BigEndian, uint64(0))
				binary.Write(&buf, binary.BigEndian, uint64(0))
				binary.Write(&buf, binary.BigEndian, uint64(0))
				binary.Write(&buf, binary.BigEndian, valueToPublish)

				signature, err := crypto.ComputeS(a, k, buf.Bytes())
				if err != nil {
					logging.Error.Printf("Could not sign the message: %s", err.Error())
					continue
				}

				store.Publish(R, valueToPublish, signature)
			}
		}
	}

	lastPublished = timeNow
	return nil
}
