package main

import (
	"encoding/json"
	"fmt"
	query2 "github.com/goark/pa-api/query"
	"strings"
	"time"
)

var InitHeaders = map[string]string{
	"host":             "webservices.amazon.com",
	"content-type":     "application/json; charset=utf-8",
	"x-amz-target":     "com.amazon.paapi5.v1.ProductAdvertisingAPIv1.GetItems",
	"content-encoding": "amz-1.0",
}

var reverse = false

func (apiCredential *ApiCredential) RequestData() {
	asins := make([]string, 1)

	apiCredential.Jobs.Range(func(key string, value []chan ProductData) bool {
		asins = append(asins, key)
		return true
	})

	query := query2.NewGetItems(
		apiCredential.Client.Marketplace(),
		apiCredential.Client.PartnerTag(),
		apiCredential.Client.PartnerType()).ASINs(asins).EnableOffers().EnableItemInfo().EnableImages()

	body, err := apiCredential.Client.Request(query)

	if strings.Contains(string(body), "provided in the request is invalid.") {
		go sendAsinError(apiCredential.PartnerTag, asins)
		apiCredential.Jobs = JobsMap{}
		return
	}

	if err != nil || strings.Contains(string(body), "Errors") {
		fmt.Println(apiCredential.PartnerTag, "Erroneous response, cooling down")
		go sendAccountError(apiCredential.PartnerTag, apiCredential.RequestsSinceError)

		/*apiCredential.OnCooldown = true
		time.AfterFunc(time.Hour, func() {
			apiCredential.OnCooldown = false
		})*/
		apiCredential.Jobs = JobsMap{}

		for _, asin := range asins {
			channels, _ := apiCredential.Jobs.Load(asin)
			for _, dataChannel := range channels {
				assignDataRequest(asin, dataChannel)
			}
		}

		apiCredential.RequestsSinceError = 0
		return
	}

	var response AmazonResponse
	json.Unmarshal(body, &response)

	for _, item := range response.ItemsResult.Items {
		channels, _ := apiCredential.Jobs.Load(item.ASIN)
		for _, dataReceiver := range channels {
			dataReceiver <- item
		}
		apiCredential.Jobs.Delete(item.ASIN)
	}

	apiCredential.RequestsSinceError++
}

func runRequestLoop(credential *ApiCredential) {
	ticker := time.Tick(time.Second / time.Duration(credential.RequestsPerSecond))

	for time := range ticker {
		empty := true
		credential.Jobs.Range(func(key string, value []chan ProductData) bool {
			if key != "" {
				empty = false
			}
			return false
		})

		if !empty {
			fmt.Println(time, credential.PartnerTag, "Requesting data for the following ASINs: ", credential.Jobs)
			credential.RequestData()
		}
	}
}

func assignDataRequest(asin string, dataChannel chan ProductData) {
	last := len(apiCredentials) - 1

	/*sort.Slice(apiCredentials, func(i, j int) bool {
		return len(apiCredentials[i].Jobs) > len(apiCredentials[j].Jobs)
	})*/

	for i := range apiCredentials {
		if _, ok := apiCredentials[i].Jobs.Load(asin); ok {
			channels, _ := apiCredentials[i].Jobs.Load(asin)
			apiCredentials[i].Jobs.Store(asin, append(channels, dataChannel))
			reverse = !reverse
			return
		}
	}

	for i := range apiCredentials {
		if reverse {
			i = last - i
		}

		if getJobLength(apiCredentials[i].Jobs) <= 10 && !apiCredentials[i].OnCooldown {
			channels, _ := apiCredentials[i].Jobs.Load(asin)
			apiCredentials[i].Jobs.Store(asin, append(channels, dataChannel))
			reverse = !reverse
			return
		}
	}

	time.AfterFunc(time.Millisecond*100, func() {
		assignDataRequest(asin, dataChannel)
	})
}

func getJobLength(jobsMap JobsMap) int {
	count := 0
	jobsMap.Range(func(key string, value []chan ProductData) bool {
		count++
		return true
	})

	return count
}

type AmazonResponse struct {
	ItemsResult struct {
		Items []ProductData `json:"Items"`
	} `json:"ItemsResult"`
}

type ProductData struct {
	ASIN   string
	Images struct {
		Primary struct {
			Large struct {
				Height int    `json:"Height"`
				URL    string `json:"URL"`
				Width  int    `json:"Width"`
			} `json:"Large"`
			Medium struct {
				Height int    `json:"Height"`
				URL    string `json:"URL"`
				Width  int    `json:"Width"`
			} `json:"Medium"`
			Small struct {
				Height int    `json:"Height"`
				URL    string `json:"URL"`
				Width  int    `json:"Width"`
			} `json:"Small"`
		} `json:"Primary"`
		Variants []struct {
			Large struct {
				Height int    `json:"Height"`
				URL    string `json:"URL"`
				Width  int    `json:"Width"`
			} `json:"Large"`
			Medium struct {
				Height int    `json:"Height"`
				URL    string `json:"URL"`
				Width  int    `json:"Width"`
			} `json:"Medium"`
			Small struct {
				Height int    `json:"Height"`
				URL    string `json:"URL"`
				Width  int    `json:"Width"`
			} `json:"Small"`
		} `json:"Variants"`
	} `json:"Images"`
	ItemInfo struct {
		ByLineInfo struct {
			Brand struct {
				DisplayValue string `json:"DisplayValue"`
				Label        string `json:"Label"`
				Locale       string `json:"Locale"`
			} `json:"Brand"`
			Manufacturer struct {
				DisplayValue string `json:"DisplayValue"`
				Label        string `json:"Label"`
				Locale       string `json:"Locale"`
			} `json:"Manufacturer"`
		} `json:"ByLineInfo"`
		Classifications struct {
			Binding struct {
				DisplayValue string `json:"DisplayValue"`
				Label        string `json:"Label"`
				Locale       string `json:"Locale"`
			} `json:"Binding"`
			ProductGroup struct {
				DisplayValue string `json:"DisplayValue"`
				Label        string `json:"Label"`
				Locale       string `json:"Locale"`
			} `json:"ProductGroup"`
		} `json:"Classifications"`
		ExternalIds struct {
			EANs struct {
				DisplayValues []string `json:"DisplayValues"`
				Label         string   `json:"Label"`
				Locale        string   `json:"Locale"`
			} `json:"EANs"`
			UPCs struct {
				DisplayValues []string `json:"DisplayValues"`
				Label         string   `json:"Label"`
				Locale        string   `json:"Locale"`
			} `json:"UPCs"`
		} `json:"ExternalIds"`
		Features struct {
			DisplayValues []string `json:"DisplayValues"`
			Label         string   `json:"Label"`
			Locale        string   `json:"Locale"`
		} `json:"Features"`
		ManufactureInfo struct {
			ItemPartNumber struct {
				DisplayValue string `json:"DisplayValue"`
				Label        string `json:"Label"`
				Locale       string `json:"Locale"`
			} `json:"ItemPartNumber"`
			Model struct {
				DisplayValue string `json:"DisplayValue"`
				Label        string `json:"Label"`
				Locale       string `json:"Locale"`
			} `json:"Model"`
		} `json:"ManufactureInfo"`
		ProductInfo struct {
			IsAdultProduct struct {
				DisplayValue bool   `json:"DisplayValue"`
				Label        string `json:"Label"`
				Locale       string `json:"Locale"`
			} `json:"IsAdultProduct"`
			ItemDimensions struct {
				Height struct {
					DisplayValue int    `json:"DisplayValue"`
					Label        string `json:"Label"`
					Locale       string `json:"Locale"`
					Unit         string `json:"Unit"`
				} `json:"Height"`
				Length struct {
					DisplayValue float64 `json:"DisplayValue"`
					Label        string  `json:"Label"`
					Locale       string  `json:"Locale"`
					Unit         string  `json:"Unit"`
				} `json:"Length"`
				Weight struct {
					DisplayValue float64 `json:"DisplayValue"`
					Label        string  `json:"Label"`
					Locale       string  `json:"Locale"`
					Unit         string  `json:"Unit"`
				} `json:"Weight"`
				Width struct {
					DisplayValue float64 `json:"DisplayValue"`
					Label        string  `json:"Label"`
					Locale       string  `json:"Locale"`
					Unit         string  `json:"Unit"`
				} `json:"Width"`
			} `json:"ItemDimensions"`
			ReleaseDate struct {
				DisplayValue time.Time `json:"DisplayValue"`
				Label        string    `json:"Label"`
				Locale       string    `json:"Locale"`
			} `json:"ReleaseDate"`
		} `json:"ProductInfo"`
		Title struct {
			DisplayValue string `json:"DisplayValue"`
			Label        string `json:"Label"`
			Locale       string `json:"Locale"`
		} `json:"Title"`
	} `json:"ItemInfo"`
	Offers struct {
		Listings []struct {
			Availability struct {
				Message          string `json:"Message"`
				MinOrderQuantity int    `json:"MinOrderQuantity"`
				Type             string `json:"Type"`
			} `json:"Availability"`
			Condition struct {
				SubCondition struct {
					Value string `json:"Value"`
				} `json:"SubCondition"`
				Value string `json:"Value"`
			} `json:"Condition"`
			DeliveryInfo struct {
				IsAmazonFulfilled      bool `json:"IsAmazonFulfilled"`
				IsFreeShippingEligible bool `json:"IsFreeShippingEligible"`
				IsPrimeEligible        bool `json:"IsPrimeEligible"`
			} `json:"DeliveryInfo"`
			ID         string `json:"Id"`
			Promotions []struct {
				Type            string  `json:"Type"`
				Amount          float64 `json:"Amount"`
				Currency        string  `json:"Currency"`
				DiscountPercent int     `json:"DiscountPercent"`
				PricePerUnit    float64 `json:"PricePerUnit"`
				DisplayAmount   string  `json:"DisplayAmount"`
			} `json:"Promotions"`
			IsBuyBoxWinner bool `json:"IsBuyBoxWinner"`
			MerchantInfo   struct {
				DefaultShippingCountry string  `json:"DefaultShippingCountry"`
				FeedbackCount          int     `json:"FeedbackCount"`
				FeedbackRating         float64 `json:"FeedbackRating"`
				ID                     string  `json:"Id"`
				Name                   string  `json:"Name"`
			} `json:"MerchantInfo"`
			Price struct {
				Amount        float64 `json:"Amount"`
				Currency      string  `json:"Currency"`
				DisplayAmount string  `json:"DisplayAmount"`
			} `json:"Price"`
			ProgramEligibility struct {
				IsPrimeExclusive bool `json:"IsPrimeExclusive"`
				IsPrimePantry    bool `json:"IsPrimePantry"`
			} `json:"ProgramEligibility"`
			ViolatesMAP bool `json:"ViolatesMAP"`
		} `json:"Listings"`
	} `json:"Offers"`
}
