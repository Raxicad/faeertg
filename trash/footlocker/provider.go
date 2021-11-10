package footlocker

import (
	"fmt"
	"github.com/bandar-monitors/monitors/core/http"
	"github.com/bandar-monitors/monitors/sites/domain"
	"github.com/buger/jsonparser"
	"github.com/moovweb/gokogiri"
	"github.com/moovweb/gokogiri/html"
	"github.com/moovweb/gokogiri/xml"
	"strings"
	"time"
)

const SizesUrlXpath = "//div[@data-ajaxcontent='fl-productDetailsSizeSelection']"

func FetchProduct(ctx *ProcessingContext, logger *strings.Builder) (*domain.Product, error) {
	fetchTime := time.Now().UnixNano()

	// Perform the request
	body, err := ctx.pageClient.Execute()
	if err != nil {
		fmt.Printf("Client get failed: %s\n", err)
		return nil, err
	}

	fmt.Fprintf(logger, "Got response in %dms\n", (time.Now().UnixNano()-fetchTime)/1000_000)
	parseTime := time.Now().UnixNano()

	product := domain.Product{
		Url: ctx.cfg.GetUrl(),
	}
	doc, _ := gokogiri.ParseHtml(body)
	root := doc.Root()
	product.IsAvailable = isAvailable(root)
	if !product.IsAvailable {
		return &product, nil
	}

	parseNameAndPrice(root, &product)
	sizesDoc := parseSizes(root, ctx.sizesClient, &product)

	sizesDoc.Free()
	doc.Free()

	fmt.Fprintf(logger, "Parsed in %dms\n", (time.Now().UnixNano()-parseTime)/1000_000)
	return &product, nil
}

func isAvailable(root *xml.ElementNode) bool {
	return true
}

func parseNameAndPrice(root *xml.ElementNode, product *domain.Product) {
	nameNodes, _ := root.Search("/html/body/main/div/div[4]/div/div/div[1]/h1/span")
	name := nameNodes[0].Content()
	priceNodes, _ := root.Search("/html/body/main/div/div[4]/div/div/div[3]/form/div/div[1]/div[1]/div/span[1]/span")
	price := priceNodes[0].Content()

	product.DisplayName = name
	product.Price = price
}

func parseSizes(root *xml.ElementNode, client *http.HttpClient, product *domain.Product) *html.HtmlDocument {
	sizesNodeUrlNodes, _ := root.Search(SizesUrlXpath)
	sizesUrl := sizesNodeUrlNodes[0].Attr("data-ajaxcontent-url")
	client.SetUrl(sizesUrl)

	sizesRespBytes, _ := client.Execute()

	sizesRawHtml, _ := jsonparser.GetString(sizesRespBytes, "content")
	sizesDoc, _ := gokogiri.ParseHtml([]byte(sizesRawHtml))
	sizesRoot := sizesDoc.Root()
	sizesContainerNodes, _ := sizesRoot.Search("//*[@data-product-size-select]") // 314214106404
	sku := sizesContainerNodes[0].Attr("data-product-size-select")
	euSizeNodes, _ := sizesContainerNodes[0].Search("//section[1]//button[contains(@class, 'fl-product-size--item')]")
	var sizes []domain.ProductSize

	for _, node := range euSizeNodes {
		cls := node.Attr("class")
		if strings.Contains(cls, "fl-product-size--item__not-available") {
			continue
		}

		size := domain.ProductSize{
			Id:          node.Attr("data-product-size-select-item"),
			DisplayName: strings.Trim(node.Content(), "\n"),
		}

		sizes = append(sizes, size)
	}

	product.PictureUrl = fmt.Sprintf("https://images.footlocker.com/is/image/FLEU/%s_01?wid=763&hei=538&fmt=png-alpha", sku)
	product.Sizes = &sizes

	return sizesDoc
}
