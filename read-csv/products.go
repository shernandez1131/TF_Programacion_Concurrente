package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
)

type Product struct {
	ID   string `json:"ID"`
	Name string `json:"Name"`
}

func main() {
	bytes, _ := os.Open("sample_products.csv")

	r := csv.NewReader(bytes)

	var products []Product

	for {
		attribute, err := r.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		products = append(products, Product{
			ID:   attribute[0],
			Name: attribute[3],
		})

		//fmt.Printf("%s %s \n", attribute[0], attribute[3])
	}

	//productsJson, _ := json.Marshal(products)
	//fmt.Println(string(productsJson))

	for _, p := range products {
		fmt.Printf("ID: %s - Name: %s \n", p.ID, p.Name)
	}
}
