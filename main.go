package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

var (
	InvalidHash = errors.New("error: hash is invalid")
	UnexpectedError = errors.New("error: unexpected error")
	)


func unblockEndpoint(hash string) error {
	listEndpoint, _ := url.Parse("http://localhost:8080/unblock_endpoint")
	client := http.Client{
		Timeout: time.Hour,
	}
	request := &http.Request{
		URL: listEndpoint,
		Method: http.MethodGet,
		Header: map[string][]string{
			"endpoint_hash": {hash},
		},
	}
	response, err := client.Do(request)
	if err != nil {
		log.Fatalf("error: not able to connect to server")
		return UnexpectedError
	}
	if response.StatusCode == http.StatusNotFound {
		return InvalidHash
	} else if response.StatusCode == http.StatusInternalServerError {
		return UnexpectedError
	}
	return nil
}

func listAllEndpoint() (map[string] string, error) {
	listEndpoint, _ := url.Parse("http://localhost:8080/get_all_blocked_endpoints")
	response, _ := http.Get(listEndpoint.String())
	var responseMap = make(map[string] string)
	err := json.NewDecoder(response.Body).Decode(&responseMap)
	if err != nil {
		log.Fatal("error:", err)
		return nil, err
	}
	return responseMap, nil
}

func main(){
	list := flag.Bool("list", false, "list all the blocked endpoint")
	unblockFlag := flag.String("unblock", "", "unblock the endpoint with the specified hash value")
	flag.Parse()
	if *list {
		endpoint, err := listAllEndpoint()
		if err != nil {
			return
		}
		fmt.Printf("%-44s %s\n", "HASH", "ENDPOINT")
		for k, v := range endpoint {
			fmt.Printf("%-44s %s\n", k, v)
		}
	} else if *unblockFlag != "" {
		err := unblockEndpoint(*unblockFlag)
		if err == UnexpectedError {
			log.Fatalf("error: unexpected error occured")
		} else if err == InvalidHash {
			log.Fatalf("error: not endpoint with the specified hash was found")
		} else {
			log.Println("sucessful! endpoint with hash value [%s] removed", *unblockFlag)
		}
	} else {
		log.Println("check -h flag for usage")
	}
}