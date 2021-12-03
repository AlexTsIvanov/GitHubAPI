package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
)

type Client struct {
	apiURL string
	client *http.Client
}

type UserUnmarshal struct {
	Login       string `json:"login"`
	PublicRepos int    `json:"public_repos"`
	Followers   int    `json:"followers"`
}

type RepoUnmarshal struct {
	Forks int    `json:"forks"`
	Name  string `json:"name"`
}

type Repo struct {
	Forks    int
	TopLangs []string
}

type kv struct {
	Key   string
	Value int
}

type Data struct {
	UserUnmarshal
	Repo
}

func processFile(fname *string) []string {
	file, err := os.Open(*fname)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	users := make([]string, 0)
	for scanner.Scan() {
		users = append(users, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return users
}

func NewClient() *Client {
	return &Client{apiURL: "https://api.github.com/", client: http.DefaultClient}
}

func (c *Client) GetUsers(users ...string) ([]UserUnmarshal, error) {
	usersdata := make([]UserUnmarshal, 0, len(users))
	for _, user := range users {
		resp, err := c.client.Get(c.apiURL + "users/" + user)
		if err != nil {
			return nil, err
		}
		var userdata UserUnmarshal
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&userdata)
		if err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()
		usersdata = append(usersdata, userdata)
	}
	return usersdata, nil
}

func (c *Client) GetRepos(users ...string) ([]Repo, error) {
	reposdata := make([]Repo, 0, len(users))
	for _, user := range users {
		resp, err := http.Get(c.apiURL + "users/" + user + "/repos")
		if err != nil {
			return nil, err
		}
		var repodata []RepoUnmarshal
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&repodata)
		if err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		var RepoEntry Repo
		RepoEntry.Forks = forksCount(repodata)
		RepoEntry.TopLangs, err = c.GetRepoLangs(user, repodata)
		if err != nil {
			return nil, err
		}

		reposdata = append(reposdata, RepoEntry)

	}
	return reposdata, nil
}

func (c *Client) GetRepoLangs(user string, repodata []RepoUnmarshal) ([]string, error) {
	// var langstat map[string]int
	langstat := make(map[string]int, 0)
	var totalsum int
	for _, reponame := range repodata {
		resp, err := http.Get(c.apiURL + "repos/" + user + "/" + reponame.Name + "/languages")
		if err != nil {
			return nil, err
		}
		var langdata map[string]int
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&langdata)
		if err != nil {
			resp.Body.Close()
			return nil, err
		}

		for k, v := range langdata {
			totalsum += int(v)
			if val, ok := langstat[k]; ok {
				langstat[k] = val + v
			} else {
				langstat[k] = v
			}
		}
	}
	toplangs := topLangsCalc(langstat, totalsum)
	return toplangs, nil
}

func forksCount(repodata []RepoUnmarshal) int {
	counter := 0
	for _, r := range repodata {
		counter += r.Forks
	}
	return counter
}

func topLangsCalc(langstat map[string]int, totalsum int) []string {
	var topLangs []string
	otherLangs := float32(totalsum)
	ss := make([]kv, 0, len(langstat)+5)
	for k, v := range langstat {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})
	slice := 5
	if len(langstat) < 5 {
		slice = len(langstat)
	}
	for _, kv := range ss[:slice] {
		otherLangs -= float32(kv.Value)
		entry := fmt.Sprintf("%s: %.2f%%", kv.Key, float32(kv.Value)*100.0/float32(totalsum))
		topLangs = append(topLangs, entry)
	}
	other := fmt.Sprintf("Other: %.2f%%", float32(otherLangs)*100.0/float32(totalsum))
	topLangs = append(topLangs, other)
	return topLangs
}

func tablePrint(usersdata []UserUnmarshal, reposdata []Repo) {
	fmt.Println("Name\tNofRepos\tDistofLang\tFollowers\tForks")
	for i := 0; i < len(usersdata); i++ {
		printData := dataObject(usersdata[i], reposdata[i])
		fmt.Println(printData.Login, printData.PublicRepos, printData.TopLangs, printData.Followers, printData.Forks)
	}
}

func dataObject(userdata UserUnmarshal, repodata Repo) Data {
	var printData Data
	printData.Login = userdata.Login
	printData.PublicRepos = userdata.PublicRepos
	printData.TopLangs = repodata.TopLangs
	printData.Followers = userdata.Followers
	printData.Forks = repodata.Forks
	return printData
}

var (
	fname string
)

func main() {
	fmt.Print("Enter file name (with .txt): ")
	fmt.Scanln(&fname)
	users := processFile(&fname)
	c := NewClient()
	usersdata, err := c.GetUsers(users...)
	if err != nil {
		log.Fatal(err)
	}
	reposdata, err := c.GetRepos(users...)
	if err != nil {
		log.Fatal(err)
	}
	tablePrint(usersdata, reposdata)
}
