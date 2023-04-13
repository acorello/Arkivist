package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

type GoogleBooksResponse struct {
	TotalItems int `json:"totalItems"`
	Items      []struct {
		VolumeInfo struct {
			Title         string   `json:"title"`
			Authors       []string `json:"authors"`
			PublishedDate string   `json:"publishedDate"`
			Description   string   `json:"description" yaml:",flow"`
		} `json:"volumeInfo"`
	} `json:"items"`
}

func main() {
	var conf = struct{ books struct{ apikey string } }{}
	f, err := os.Open("_tmp/googleapi.toml")
	if err != nil {
		log.Fatal(err)
	}
	toml.NewDecoder(f).Decode(&conf)

	isbn := "9781449310509"

	bookDetails, err := getBookDetails(isbn, conf.books.apikey)
	if err != nil {
		log.Fatal(err)
	}
	yaml.NewEncoder(os.Stdout)
	book, err := yaml.Marshal(&bookDetails.Items[0])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(book))
}

func getBookDetails(isbn, apiKey string) (*GoogleBooksResponse, error) {
	booksEndpoint, err := url.Parse("https://www.googleapis.com/books/v1/volumes")
	if err != nil {
		log.Fatal(err)
	}
	query := booksEndpoint.Query()
	query.Add("q", "isbn:"+isbn)
	query.Add("key", apiKey)
	booksEndpoint.RawQuery = query.Encode()
	log.Println("URL:", booksEndpoint.String())

	resp, err := http.Get(booksEndpoint.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch book details: %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var bookDetails GoogleBooksResponse
	err = json.Unmarshal(body, &bookDetails)
	if err != nil {
		return nil, err
	}

	return &bookDetails, nil
}
