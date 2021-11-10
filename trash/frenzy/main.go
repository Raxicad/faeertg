package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/json-iterator/go"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"time"

	"github.com/bandar-monitors/monitors/core/http"
)

const WEBHOOK_URL = "https://discordapp.com/api/webhooks/735421276818112554/v-ZAFqJprnudMCkSJlHKZ7pWcASEL9CgfKkB81F3sjN9q3GgziN9D21Qr93DcBhMpq7I"

type WebhookPayload struct {
	Content  string `json:"content"`
	Username string `json:"username"`
}

func submitWebhook(found *[]int32, received *[]int32, client *http.HttpClient) {
	if len(*found) == len(*received) {
		return
	}

	t := time.Now().UnixNano()
	var b strings.Builder
	for i := 0; i < len(*received); i++ {
		var exist bool
		curr := (*received)[i]
		for j := 0; j < len(*found); j++ {
			exist = curr == (*found)[j]
			if exist {
				break
			}
		}

		if !exist {
			if i != 0 {
				fmt.Fprintf(&b, ",")
			}

			fmt.Fprintf(&b, "%d", curr)
		}
	}

	payload := WebhookPayload{
		Content:  "Products '" + b.String() + "' doesn't exist.",
		Username: "Product Existence Notifications Bot (golang)",
	}

	serialized, err := jsoniter.Marshal(&payload)
	if err != nil {
		panic(err)
	}

	serializedPayload := bytes.NewReader(serialized)
	client.SetBody(serializedPayload)
	_, _ = client.Execute()

	fmt.Printf("==== Webhook sent in '%d'ms====\n", (time.Now().UnixNano()-t)/1000_000)
}

func main() {
	database, _ := sql.Open("sqlite3", "./Products.gsqlite")
	createDbStatement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS products (id int PRIMARY KEY)")
	createDbStatement.Exec()
	var url = "https://frenzy.shopifyapps.com/api/flashsales"

	client := http.New(url, url)
	defer client.Release()

	webhookClient := http.New(WEBHOOK_URL, "")
	webhookClient.SetMethod("POST")
	webhookClient.SetHeader("Content-Type", "application/json")
	defer webhookClient.Release()

	var body []byte
	var logger strings.Builder
	for {
		totalTime := time.Now().UnixNano()
		fetchTime := time.Now().UnixNano()

		body, _ = client.Execute()

		fmt.Fprintf(&logger, "Got response in %dms\n", (time.Now().UnixNano()-fetchTime)/1000_000)

		jsonTime := time.Now().UnixNano()

		var ids []int32
		var b strings.Builder
		jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			idValue, _, _, _ := jsonparser.Get(value, "id")
			id, _ := jsonparser.ParseInt(idValue)
			ids = append(ids, int32(id))
			if b.Len() != 0 {
				fmt.Fprintf(&b, ",")
			}

			fmt.Fprintf(&b, "%d", id)
		}, "flashsales")
		fmt.Fprintf(&logger, "Extracted ids in %dms\n", (time.Now().UnixNano()-jsonTime)/1000_000)

		queryTime := time.Now().UnixNano()
		foundIDs := make([]int32, 0, 10)
		rows, _ := database.Query("select id from products where id IN (?)", b.String())

		var id int32
		for rows.Next() {
			rows.Scan(&id)
			foundIDs = append(foundIDs, id)
		}

		fmt.Fprintf(&logger, "Queried ids in %d\n", (time.Now().UnixNano()-queryTime)/1000_000)

		go submitWebhook(&foundIDs, &ids, webhookClient)
		fmt.Fprintf(&logger, "--Total time %d---\n\n\n", (time.Now().UnixNano()-totalTime)/1000_000)

		fmt.Println(logger.String())
		b.Reset()
		time.Sleep(time.Second)
	}
}
