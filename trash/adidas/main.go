package adidas

import (
	"github.com/bandar-monitors/monitors/core/http"
	"github.com/json-iterator/go"
	"log"
	"time"
)

type ProductInfo struct {
	ProductId       string  `json:"product_id"`
	ModelNumber     string  `json:"model_number"`
	OriginalPrice   float64 `json:"original_price"`
	DisplayCurrency string  `json:"display_currency"`
	Orderable       bool    `json:"orderable"`
	Embedded        Embed   `json:"_embedded"`
}

type Embed struct {
	Variations []Variation `json:"variations"`
}

type Variation struct {
	Size               string  `json:"size"`
	Orderable          bool    `json:"orderable"`
	AbsoluteSize       float64 `json:"absolute_size"`
	VariationProductId string  `json:"variation_product_id"`
	StockLevel         int     `json:"stock_level"`
}

func Start() {
	productId := "FW5926"
	client := http.New("https://api.3stripes.net/gw-api/v2/products/"+productId+"/availability?experiment_product_data=A", "")
	client.SetHeader("accept", "application/hal+json")
	client.SetHeader("x-signature", "B6B98CB707410DBDD4ED8D894121F3BC72D6BA08E0CEAF1A2998473EAEEE8B73")
	client.SetHeader("x-device-info", "app/adidas; os/iOS; os-version/14.0; app-version/3.33.2; buildnumber/2020.8.5.7.32; type/iPhone12,5; fingerprint/37B28BD9-FC4B-49DF-8746-4FA34DD5BB26")
	client.SetHeader("x-api-key", "m79qyapn2kbucuv96ednvh22")
	client.SetHeader("x-market", "ES")
	client.SetHeader("accept-language", "es-ES")
	client.SetHeader("user-agent", "adidas/2020.8.5.7.32 CFNetwork/1188 Darwin/20.0.0")
	client.SetHeader("cookie", "akacd_api_3stripes=3774777331~rv=98~id=0c8135ed373c011ce299035d52736045")
	defer client.Release()

	for {
		resp, _ := client.Execute()

		start := time.Now().UnixNano()
		var p ProductInfo
		err := jsoniter.Unmarshal(resp, &p)
		if err != nil {
			panic(err)
		}
		if p.Orderable {
			log.Printf("Fetched %d variations\n", len(p.Embedded.Variations))

			// notify
		}

		spent := time.Now().UnixNano() - start
		log.Printf("Parsed in %d\n", spent/1000)

		time.Sleep(time.Second)
	}
}
