package datasources

import (
	"encoding/json"
	"math"
	"net/http"

	"github.com/gertjaap/dlcoracle/logging"
)

type EurBtcRounded struct {
}

func (ds *EurBtcRounded) Id() uint64 {
	return 1
}

func (ds *EurBtcRounded) Name() string {
	return "Euro"
}

func (ds *EurBtcRounded) Description() string {
	return "Publishes the value of EUR denominated in 1/100000000th units of BTC (satoshi) in multitudes of 100"
}

func (ds *EurBtcRounded) Interval() uint64 {
	return 300 // every 5 minutes
}

func (ds *EurBtcRounded) Value() (uint64, error) {
	req, err := http.NewRequest("GET", "https://min-api.cryptocompare.com/data/price?fsym=EUR&tsyms=BTC", nil)
	if err != nil {
		logging.Error.Println("EurBtcRounded.Value - NewRequest", err)
		return 0, err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		logging.Error.Println("EurBtcRounded.Value - Do: ", err)
		return 0, err
	}

	defer resp.Body.Close()

	var record MinApiCryptoCompareBTCResponse

	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		logging.Error.Println("EurBtcRounded.Value - Json decode failed: ", err)
		return 0, err
	}

	satoshiValue := uint64(math.Floor((record.Value*1000000)+0.5)) * 100
	return satoshiValue, nil
}
