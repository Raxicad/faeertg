package footlocker

import (
	"fmt"
	"log"
	"strings"
	"time"
)

func StartFootlocker() {
	//database, _ := sql.Open("sqlite3", "./Products.gsqlite")
	//createDbStatement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS products (id int PRIMARY KEY)")
	//createDbStatement.Exec()
	ctx := NewContext(
		"https://www.footlocker.es/en/p/jordan-1-low-women-shoes-85963?v=315345596902#!searchCategory=all",
		"products-footlocker-315345596902",
		"")

	defer ctx.Release()

	//publisher, _ := CreateQueuePublisher(ctx)
	//defer publisher.Release()
	var logger strings.Builder
	for {
		totalTime := time.Now().UnixNano()
		product, err := FetchProduct(ctx, &logger)
		if err != nil {
			panic(err)
		}

		if product.IsAvailable {
			//e := publisher.PublishProduct(product)
			//if e != nil {
			//	log.Fatal("PUBLISHER FATAL: " + e.Error())
			//}
		} else {
			log.Print("Not yet... " + product.DisplayName)
		}

		//go submitWebhook(&foundIDs, &ids)
		fmt.Fprintf(&logger, "--Total time %d---\n\n\n", (time.Now().UnixNano()-totalTime)/1000_000)

		fmt.Println(logger.String())
		//b.Reset()
		//break
		time.Sleep(time.Second)
	}
}
