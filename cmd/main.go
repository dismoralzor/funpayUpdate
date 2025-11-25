package main

import (
	SeleniumTebeka "FunpayUpdater"
	"flag"
	"fmt"
	"log"
	"time"
)

type Args struct {
	Username string
	Password string
}

func main() {
	args := Args{}

	flag.StringVar(&args.Username, "username", "", "Username for VK\n")
	flag.StringVar(&args.Password, "password", "", "Password for VK\n")

	flag.Parse()

	if args.Username == "" {
		log.Fatal("arg -username is empty")
	}

	if args.Password == "" {
		log.Fatal("arg -password is empty")
	}

	for {
		SeleniumTebeka.FunpayUpdate(args.Username, args.Password)
		time.Sleep(15480 * time.Second)
		today := time.Now()
		fmt.Printf("Update! %s", today)
	}
}
