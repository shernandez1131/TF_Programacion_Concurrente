package main

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type Product struct {
	ID   string `json:"ID"`
	Name string `json:"Name"`
}

type Block struct {
	timestamp    time.Time
	transactions []Product
	prevhash     []byte
	Hash         []byte
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
			Name: attribute[1],
		})

		//fmt.Printf("%s %s \n", attribute[0], attribute[3])
	}

	//productsJson, _ := json.Marshal(products)
	//fmt.Println(string(productsJson))

	contBlock := 1
	var auxPHash []byte

	for _, p := range products {
		abc := []Product{p}
		xyz := Blocks(abc, auxPHash)
		fmt.Println("This is out Block", contBlock)
		Print(xyz)
		auxPHash = xyz.Hash
		contBlock++
	}

	/*abc := []string{" A sent 50 coins to BC"}
	xyz := Blocks(abc, []byte{})
	fmt.Println("This is out First Block")
	Print(xyz)

	pqrs := []string{" PQ sent 10 coins to BC"}
	klmn := Blocks(pqrs, xyz.Hash)
	fmt.Println("This is out Seconf Block")
	Print(klmn)*/
}

func Blocks(transactions []Product, prevhash []byte) *Block {
	currentTime := time.Now()
	return &Block{
		timestamp:    currentTime,
		transactions: transactions,
		prevhash:     prevhash,
		Hash:         NewHash(currentTime, transactions, prevhash),
	}
}

func NewHash(time time.Time, transactions []Product, prevhash []byte) []byte {
	input := append(prevhash, time.String()...)
	for transactions := range transactions {
		input = append(input, string(rune(transactions))...)
	}
	hash := sha256.Sum256(input)
	return hash[:]
}

func Print(block *Block) {
	fmt.Printf("\t time: %s \n", block.timestamp.String())
	fmt.Printf("\t prevhash: %x \n", block.prevhash)
	fmt.Printf("\t hash: %x \n", block.Hash)
	Transaction(block)
}

func Transaction(block *Block) {
	fmt.Println("\t Transactions: ")
	for i, transaction := range block.transactions {
		//fmt.Printf("\t\t %v: %q \n", i, transaction)
		fmt.Printf("\t\t %v = ID: %s - Name: %s \n", i, transaction.ID, transaction.Name)
	}
}
