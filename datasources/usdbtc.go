package datasources

import (
	"encoding/json"
	"net/http"

	"github.com/gertjaap/dlcoracle/logging"
)

type UsdBtc struct {
}

func (ds *UsdBtc) Id() uint64 {
	return 1
}

func (ds *UsdBtc) Name() string {
	return "USD / BTC"
}

func (ds *UsdBtc) Description() string {
	return "Publishes the value of USD denominated in 1/100000000th units of BTC (satoshi)"
}

type MinApiCryptoCompareBTCResponse struct {
	Value float64 `json:"BTC"`
}

func (ds *UsdBtc) Value() (uint64, error) {
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

	satoshiValue := uint64(record.Value * 100000000)
	return satoshiValue, nil
}
