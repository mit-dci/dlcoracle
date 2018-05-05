package datasources

import (
	"encoding/json"
	"math"
	"net/http"

	"github.com/gertjaap/dlcoracle/logging"
)

type UsdBtcRounded struct {
}

func (ds *UsdBtcRounded) Id() uint64 {
	return 1
}

func (ds *UsdBtcRounded) Name() string {
	return "US Dollar"
}

func (ds *UsdBtcRounded) Description() string {
	return "Publishes the value of USD denominated in 1/100000000th units of BTC (satoshi) in multitudes of 100"
}

func (ds *UsdBtcRounded) Interval() uint64 {
	return 300 // every 5 minutes
}

func (ds *UsdBtcRounded) Value() (uint64, error) {
	req, err := http.NewRequest("GET", "https://min-api.cryptocompare.com/data/price?fsym=USD&tsyms=BTC", nil)
	if err != nil {
		logging.Error.Println("UsdBtc.Value - NewRequest", err)
		return 0, err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		logging.Error.Println("UsdBtc.Value - Do: ", err)
		return 0, err
	}

	defer resp.Body.Close()

	var record MinApiCryptoCompareBTCResponse

	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		logging.Error.Println("UsdBtc.Value - Json decode failed: ", err)
		return 0, err
	}

	satoshiValue := uint64(math.Floor((record.Value*1000000)+0.5)) * 100
	return satoshiValue, nil
}
