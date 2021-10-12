package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func waitFor(timeout time.Duration, url string, ch chan<- error) {
	cli := http.Client{
		Timeout: time.Second * 1,
	}
	end := time.After(timeout)
	for {
		select {
		case <-time.After(time.Second * 5):
			resp, err := cli.Get(fmt.Sprintf("%s/status", url))
			if err != nil {
				log.Printf("%s\n", err.Error())
				continue
			}
			if resp.StatusCode == http.StatusOK {
				status := &ChainStatus{}
				err = json.NewDecoder(resp.Body).Decode(&status)
				_ = resp.Body.Close()
				if err != nil {
					log.Printf("error decoding response %s\n", err.Error())
					continue
				}

				if block, err := strconv.Atoi(status.Result.SyncInfo.LatestBlockHeight); err == nil && block > 10 {
					log.Printf("chain %s is ready \n", url)
					ch <- nil
					return
				}
				log.Println("latest block: ", status.Result.SyncInfo.LatestBlockHeight)
			}
		case <-end:
			ch <- fmt.Errorf("timed out waiting for %s", url)
			return
		}
	}

}

func main() {

	chains := strings.Split(os.Getenv("PLUGIN_CHAIN_LIST"), ",")

	if len(chains) == 0 || len(os.Getenv("PLUGIN_CHAIN_LIST")) == 0 {
		log.Fatal("must provide at least one chain")
	}

	timeout, err := strconv.Atoi(os.Getenv("PLUGIN_TIMEOUT"))
	if err != nil {
		log.Fatal("must provide a valid timeout")
	}
	ch := make(chan error, len(chains))

	for _, c := range chains {
		log.Printf("wait for %s %d", c, timeout)
		go waitFor(time.Duration(timeout)*time.Second, c, ch)
	}

	jobs := len(chains)
	for err := range ch {
		if err != nil {
			log.Fatal(err)
		}
		jobs--
		if jobs == 0 {
			return
		}
	}
}
