package main

import (
	"context"
	"encoding/json"
	"github.com/bandar-monitors/monitors/sites/core"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
)

type targetFetcher struct {
	http       *http.Client
	productUrl string
}

func (t *targetFetcher) FetchStatus(ctx context.Context) *core.StatusFetchResult {
	r, _ := http.NewRequestWithContext(ctx, "GET", t.productUrl, nil)
	response, err := t.http.Do(r)
	return core.ProcessFetchResult(t.http, r, response, err, func(result *core.StatusFetchResult) error {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}

		var responseStruct targetResponse
		err = json.Unmarshal(body, &responseStruct)
		if err != nil {
			return err
		}

		available := responseStruct.Data.Product.Fulfillment.ShippingOptions.AvailabilityStatus == "IN_STOCK"

		title := responseStruct.Data.Product.Item.ProductDescription.Title
		pic := responseStruct.Data.Product.Item.Enrichment.Images.PrimaryImageUrl
		data := &responseStruct

		result.RawResponse = body
		result.AddTarget(t.productUrl, title, pic, available, data)
		return nil
	})
}

type discordPayload struct {
	Content  string `json:"content"`
	Username string `json:"username"`
}

type targetResponse struct {
	Data struct {
		Product struct {
			Tcin string `json:"tcin"`
			Item struct {
				Dpci                        string `json:"dpci"`
				AssignedSellingChannelsCode string `json:"assigned_selling_channels_code"`
				PrimaryBarcode              string `json:"primary_barcode"`
				ProductClassification       struct {
					ProductTypeName string `json:"product_type_name"`
				} `json:"product_classification"`
				ProductDescription struct {
					Title                 string   `json:"title"`
					DownstreamDescription string   `json:"downstream_description"`
					BulletDescriptions    []string `json:"bullet_descriptions"`
					SoftBullets           struct {
						Title   string   `json:"title"`
						Bullets []string `json:"bullets"`
					} `json:"soft_bullets"`
				} `json:"product_description"`
				Compliance struct {
					IsProposition65 bool `json:"is_proposition_65"`
				} `json:"compliance"`
				Enrichment struct {
					BuyUrl string `json:"buy_url"`
					Images struct {
						PrimaryImageUrl    string   `json:"primary_image_url"`
						AlternateImageUrls []string `json:"alternate_image_urls"`
						ContentLabels      []struct {
							ImageUrl string `json:"image_url"`
						} `json:"content_labels"`
					} `json:"images"`
					Videos []struct {
						VideoCaptions []struct {
							CaptionUrl string `json:"caption_url"`
						} `json:"video_captions"`
						VideoTitle string `json:"video_title"`
						VideoFiles []struct {
							VideoUrl string `json:"video_url"`
						} `json:"video_files"`
						VideoPosterImage   string `json:"video_poster_image"`
						VideoLengthSeconds string `json:"video_length_seconds"`
					} `json:"videos"`
				} `json:"enrichment"`
				RelationshipTypeCode string `json:"relationship_type_code"`
				Fulfillment          struct {
					PurchaseLimit             int      `json:"purchase_limit"`
					IsShipInOriginalContainer bool     `json:"is_ship_in_original_container"`
					PoBoxProhibitedMessage    string   `json:"po_box_prohibited_message"`
					ShippingExclusionCodes    []string `json:"shipping_exclusion_codes"`
				} `json:"fulfillment"`
				ProductVendors []struct {
					Id         string `json:"id"`
					VendorName string `json:"vendor_name"`
				} `json:"product_vendors"`
				MerchandiseTypeAttributes []struct {
					Id     string `json:"id"`
					Name   string `json:"name"`
					Values []struct {
						Id   string `json:"id,omitempty"`
						Name string `json:"name"`
					} `json:"values"`
				} `json:"merchandise_type_attributes"`
				WellnessMerchandiseAttributes []struct {
					BadgeUrl            string `json:"badge_url"`
					ParentId            string `json:"parent_id"`
					ParentName          string `json:"parent_name"`
					ValueId             string `json:"value_id"`
					ValueName           string `json:"value_name"`
					WellnessDescription string `json:"wellness_description"`
				} `json:"wellness_merchandise_attributes"`
				MmbvContent struct {
					StreetDate string `json:"street_date"`
				} `json:"mmbv_content"`
				EligibilityRules struct {
				} `json:"eligibility_rules"`
				Handling struct {
					BuyUnitOfMeasure             string `json:"buy_unit_of_measure"`
					ImportDesignationDescription string `json:"import_designation_description"`
				} `json:"handling"`
				PackageDimensions struct {
					Weight                 float64 `json:"weight"`
					WeightUnitOfMeasure    string  `json:"weight_unit_of_measure"`
					Width                  float64 `json:"width"`
					Depth                  float64 `json:"depth"`
					Height                 float64 `json:"height"`
					DimensionUnitOfMeasure string  `json:"dimension_unit_of_measure"`
				} `json:"package_dimensions"`
				EnvironmentalSegmentation struct {
					IsHazardousMaterial bool `json:"is_hazardous_material"`
				} `json:"environmental_segmentation"`
				FormattedReturnMethod      string  `json:"formatted_return_method"`
				ReturnPoliciesGuestMessage string  `json:"return_policies_guest_message"`
				ReturnPolicyUrl            string  `json:"return_policy_url"`
				CartAddOnThreshold         float64 `json:"cart_add_on_threshold"`
			} `json:"item"`
			Price struct {
				CurrentRetail             float64 `json:"current_retail"`
				DefaultPrice              bool    `json:"default_price"`
				FormattedCurrentPrice     string  `json:"formatted_current_price"`
				FormattedCurrentPriceType string  `json:"formatted_current_price_type"`
				IsCurrentPriceRange       bool    `json:"is_current_price_range"`
				LocationId                int     `json:"location_id"`
				Msrp                      float64 `json:"msrp"`
				RegRetail                 float64 `json:"reg_retail"`
			} `json:"price"`
			FreeShipping struct {
				Enabled bool `json:"enabled"`
			} `json:"free_shipping"`
			RatingsAndReviews struct {
				Statistics struct {
					QuestionCount int `json:"question_count"`
					Rating        struct {
						Average float64 `json:"average"`
						Count   int     `json:"count"`
					} `json:"rating"`
					ReviewCount int `json:"review_count"`
				} `json:"statistics"`
			} `json:"ratings_and_reviews"`
			Promotions []struct {
				PdpMessage             string  `json:"pdp_message"`
				PlpMessage             string  `json:"plp_message"`
				SubscriptionType       string  `json:"subscription_type"`
				ThresholdType          string  `json:"threshold_type"`
				ThresholdValue         float64 `json:"threshold_value"`
				AppliedLocationId      int     `json:"applied_location_id"`
				Channel                string  `json:"channel"`
				LegalDisclaimerText    string  `json:"legal_disclaimer_text"`
				PromotionId            string  `json:"promotion_id"`
				PromotionClass         string  `json:"promotion_class"`
				GlobalSubscriptionFlag bool    `json:"global_subscription_flag"`
				CircleOffer            bool    `json:"circle_offer"`
			} `json:"promotions"`
			Esp struct {
				EspGroupId         string `json:"esp_group_id"`
				Tcin               string `json:"tcin"`
				ProductDescription struct {
					Title              string   `json:"title"`
					BulletDescriptions []string `json:"bullet_descriptions"`
				} `json:"product_description"`
				Enrichment struct {
					Images struct {
						PrimaryImageUrl string `json:"primary_image_url"`
					} `json:"images"`
				} `json:"enrichment"`
				Price struct {
					CurrentRetail             float64 `json:"current_retail"`
					DefaultPrice              bool    `json:"default_price"`
					FormattedCurrentPrice     string  `json:"formatted_current_price"`
					FormattedCurrentPriceType string  `json:"formatted_current_price_type"`
					IsCurrentPriceRange       bool    `json:"is_current_price_range"`
					LocationId                int     `json:"location_id"`
					Msrp                      float64 `json:"msrp"`
					RegRetail                 float64 `json:"reg_retail"`
				} `json:"price"`
			} `json:"esp"`
			StoreCoordinates []struct {
				Aisle int     `json:"aisle"`
				Block string  `json:"block"`
				Floor string  `json:"floor"`
				X     float64 `json:"x"`
				Y     float64 `json:"y"`
			} `json:"store_coordinates"`
			Fulfillment struct {
				ShippingOptions struct {
					AvailabilityStatus         string        `json:"availability_status"`
					LoyaltyAvailabilityStatus  string        `json:"loyalty_availability_status"`
					AvailableToPromiseQuantity float64       `json:"available_to_promise_quantity"`
					MinimumOrderQuantity       float64       `json:"minimum_order_quantity"`
					ReasonCode                 string        `json:"reason_code"`
					Services                   []interface{} `json:"services"`
				} `json:"shipping_options"`
				StoreOptions []struct {
					LocationName                       string  `json:"location_name"`
					LocationId                         string  `json:"location_id"`
					SearchResponseStoreType            string  `json:"search_response_store_type"`
					LocationAvailableToPromiseQuantity float64 `json:"location_available_to_promise_quantity"`
					OrderPickup                        struct {
						AvailabilityStatus string `json:"availability_status"`
					} `json:"order_pickup"`
					Curbside struct {
						AvailabilityStatus string `json:"availability_status"`
					} `json:"curbside"`
					InStoreOnly struct {
						AvailabilityStatus string `json:"availability_status"`
					} `json:"in_store_only"`
					ShipToStore struct {
						AvailabilityStatus string `json:"availability_status"`
					} `json:"ship_to_store"`
				} `json:"store_options"`
				ScheduledDelivery struct {
					AvailabilityStatus                 string  `json:"availability_status"`
					LocationAvailableToPromiseQuantity float64 `json:"location_available_to_promise_quantity"`
				} `json:"scheduled_delivery"`
			} `json:"fulfillment"`
			Taxonomy struct {
				Category struct {
					Name   string `json:"name"`
					NodeId string `json:"node_id"`
				} `json:"category"`
				Breadcrumbs []struct {
					Name   string `json:"name"`
					NodeId string `json:"node_id"`
				} `json:"breadcrumbs"`
			} `json:"taxonomy"`
			NotifyMeEnabled bool   `json:"notify_me_enabled"`
			AdPlacementUrl  string `json:"ad_placement_url"`
		} `json:"product"`
	} `json:"data"`
}

func (t *targetResponse) ToJSON() []byte {
	var content string
	if t != nil {
		content = "AVAILABLE " + t.Data.Product.Item.ProductDescription.Title
	} else {
		content = "AVAILABLE: <NO PRODUCTS DETAILS>"
	}
	payload := &discordPayload{
		Content:  content,
		Username: "FAKE_MONITOR",
	}

	jsonBytes, _ := jsoniter.Marshal(payload)
	return jsonBytes
}
