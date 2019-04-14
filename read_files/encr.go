package main

import (
	"fmt"

	"../utils"
)

func main() {
	key := "vanle@tctav.com"
	str := []byte("this is a long text haha 5645654 8903 #$%")
	encr := utils.Encrypt(str, key)
	fmt.Println(encr)
	fmt.Println(string(utils.Decrypt([]byte(encr), key)))
	fmt.Println(string(str))
}
