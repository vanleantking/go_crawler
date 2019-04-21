package main

import (
	"log"
	"os"
)

func main() {
	//create your file with desired read/write permissions
	f, err := os.OpenFile("./log/news.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	//defer to close when you're done with it, not because you think it's idiomatic!
	defer f.Close()

	//set output of logs to f
	log.SetOutput(f)

	//test case
	log.Println("check to make sure it works")
}
