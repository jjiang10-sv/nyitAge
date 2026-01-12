package main

import (
	"fmt"
	"log"

	"github.com/vishvananda/netlink"
)

func main() {
	// 1. Get a list of all links (interfaces)
	log.Println("Getting list of links...")
	links, err := netlink.LinkList()
	if err != nil {
		log.Fatal(err)
	}

	// 2. Iterate and print details
	for _, link := range links {
		attrs := link.Attrs()
		fmt.Printf("Name: %s | Type: %s | State: %s\n",
			attrs.Name, link.Type(), attrs.OperState)
	}
}
