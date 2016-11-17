package main

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func init() {
	go StartServer(8000)
	time.Sleep(time.Second)
}

func TestCommand(t *testing.T) {
	response, err := http.Get("http://127.0.0.1:8000/command?command=ls")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatal("status != 200")
	}

	bs, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	body := string(bs)
	if len(body) == 0 {
		t.Fatal("len(body) == 0")
	}
}

func TestMissingCommand(t *testing.T) {
	response, err := http.Get("http://127.0.0.1:8000/command")
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusBadRequest {
		t.Fatal("status != 400")
	}
}
