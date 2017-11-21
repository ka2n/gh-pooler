package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
	"golang.org/x/net/context"
)

func main() {
	os.Exit(mainCLI())
}

func mainCLI() int {
	var (
		interval time.Duration
	)
	flag.DurationVar(&interval, "interval", time.Second, "interval")
	flag.Parse()
	log.SetOutput(os.Stderr)

	args := flag.Args()
	if len(args) < 2 {
		return 1
	}

	repo := strings.SplitN(args[0], "/", 2)
	ref := args[1]

	log.Printf("watch repo: %v/%v#%v\n", repo[0], repo[1], ref)

	var hc *http.Client
	token := os.Getenv("GITHUB_API_TOKEN")
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		hc = oauth2.NewClient(context.Background(), ts)
	}

	gh := github.NewClient(hc)
	last, _, err := gh.Repositories.GetBranch(context.Background(), repo[0], repo[1], ref)
	if err != nil {
		log.Println(err)
		return 1
	}
	log.Println("last SHA1: ", last.Commit.GetSHA())

	timer := time.NewTicker(interval)
	defer timer.Stop()

	for {
		<-timer.C

		t := time.Now()
		new, resp, err := gh.Repositories.GetBranch(context.Background(), repo[0], repo[1], ref)
		if err != nil {
			log.Println(err)
			return 1
		}

		log.Printf("limit: %v, rate: %v, remaining: %v, reset: %v\n", resp.Limit, resp.Rate, resp.Remaining, resp.Reset)

		lastSha1 := last.Commit.GetSHA()
		newSha1 := new.Commit.GetSHA()
		log.Printf("%v: last SHA1: %v, new SHA1: %v\n", t, lastSha1, newSha1)
		if lastSha1 != newSha1 {
			log.Printf("%s", newSha1)
			return 0
		}
	}
}
