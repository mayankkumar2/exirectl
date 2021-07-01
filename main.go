package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
)

var (
	InvalidHash = errors.New("error: hash is invalid")
	UnexpectedError = errors.New("error: unexpected error")
	)

type Conf struct {
	LeakServer string `toml:"leak_server_address"`
}
var conf Conf

func ValConfig(d *Conf) bool {
	signalServerAddress, _ := regexp.Compile("^[a-zA-Z0-9\\-\\.]+:{0,1}\\d{0,5}$")
	if !signalServerAddress.MatchString(d.LeakServer) {
		color.Set(color.FgRed)
		log.Println("error: leak_server_address invalid")
		color.Unset()
		os.Exit(1)
	}
	return true
}
func unblockEndpoint(hash string) error {
	listEndpoint, _ := url.Parse(fmt.Sprintf("http://%s/unblock_endpoint", conf.LeakServer))
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
	listEndpoint, _ := url.Parse(fmt.Sprintf("http://%s/unblock_endpoint", conf.LeakServer))
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

	homePath := os.Getenv("HOME")
	if homePath == "" {
		color.Set(color.FgRed)
		log.Println("error: HOME variable not found")
		color.Unset()
	}

	appDir := homePath + "/.securum_exire"
	_, err := os.Stat(appDir)
	if os.IsNotExist(err) {
		color.Set(color.FgRed)
		log.Println("error: .securum_exire (app directory) doesn't exit")
		color.Unset()
		err = os.MkdirAll(appDir, os.ModePerm)
		if err != nil {
			color.Set(color.FgRed)
			log.Println("error: unable to create a directory")
			color.Unset()
			os.Exit(1)
		}
	}
	configFilePath := appDir + "/secexctl.config.toml"
	_, err = os.Stat(configFilePath)
	if os.IsNotExist(err) {
		color.Set(color.FgRed)
		log.Println("error: secexctl.config.toml doesn't exit")
		color.Unset()
		_, err = os.Create(configFilePath)
		if err != nil {
			color.Set(color.FgRed)
			log.Println("error: unable to create signal_server.config.toml")
			color.Unset()
			os.Exit(1)
		}
	}

	if _, err := toml.DecodeFile(configFilePath, &conf); err != nil {
		color.Set(color.FgRed)
		log.Println("error: unable to open signal_server.config.toml")
		color.Unset()
		os.Exit(1)
	}
	ValConfig(&conf)
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