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

				pubB, err := crypto.GetPubKey(crypto.KeyTypeB)
				if err != nil {
					logging.Error.Println("Could not get pub key B", err)
					return err
				}

				pubQ, err := crypto.GetPubKey(crypto.KeyTypeQ)
				if err != nil {
					logging.Error.Println("Could not get pub key Q", err)
					return err
				}

				var B, Q [33]byte
				copy(B[:], pubB[:])
				copy(Q[:], pubQ[:])

				rPoint, err := crypto.ComputeR(B, Q, ds.Id(), time)
				if err != nil {
					logging.Error.Println("Could not get R-point for data source %d and timestamp %d", ds.Id(), timeNow)
					return err
				}

				publishedAlready, err := store.IsPublished(rPoint)
				if err != nil {
					logging.Error.Printf("Error determining if this is already published: %s", err.Error())
					continue
				}

				if publishedAlready {
					logging.Info.Printf("Already published for data source %d and timestamp %d", ds.Id(), timeNow)
					continue
				}

				valueToPublish, err := ds.Value()
				if err != nil {
					logging.Error.Printf("Could not retrieve value for data source %d: %s", ds.Id(), err.Error())
					continue
				}

				var a, b, q [32]byte
				copy(a[:], crypto.RetrieveKey(crypto.KeyTypeA)[:])
				copy(b[:], crypto.RetrieveKey(crypto.KeyTypeB)[:])
				copy(q[:], crypto.RetrieveKey(crypto.KeyTypeQ)[:])

				k, err := crypto.ComputeK(q, b, ds.Id(), time)
				if err != nil {
					logging.Error.Printf("Could not derive signing key for data source %d and timestamp %d : %s", ds.Id(), time, err.Error())
					continue
				}

				q = [32]byte{}
				b = [32]byte{}

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

				store.Publish(rPoint, valueToPublish, signature)
			}
		}
	}

	lastPublished = timeNow
	return nil
}
