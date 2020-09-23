package main

import (
	"context"
	"fmt"
	"github.com/digitalocean/godo"
	"net"
)

type Configuration struct {
	Key    string `yaml:"key"`
	Domain string `yaml:"domain"`
	Record string `yaml:"record"`
}

func main() {

	client := godo.NewFromToken("396d2b3afce375a5f820a176704b88552d611a944498a3380e7cb4815677ed85")

	ctx := context.TODO()

	ipv4Domains, response, err := client.Domains.RecordsByTypeAndName(ctx, "utahcon.com", "A", "home.utahcon.com", &godo.ListOptions{})
	ipv6Domains, response, err := client.Domains.RecordsByTypeAndName(ctx, "utahcon.com", "AAAA", "home.utahcon.com", &godo.ListOptions{})
	if err != nil {
		fmt.Printf("Error getting domains list: %v\n", err)
	}

	fmt.Printf("Response: %v\n", response)

	for _, domain := range ipv4Domains {
		fmt.Printf("Domain: %v\n", domain)
	}

	for _, domain := range ipv6Domains {
		fmt.Printf("Domain: %v\n", domain)
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("error: %s", err)
	}

	for _, iface := range interfaces {
		addresses, err := iface.Addrs()
		if err != nil {
			fmt.Printf("error: %s\n", err)
		}
		for _, addr := range addresses {
			isPrivate, err := CheckIpIsPrivate(addr)
			if err != nil {
				fmt.Printf("error: %s\n", err)
			}
			if isPrivate {
				drer := &godo.DomainRecordEditRequest{
					Type: "A",
					Name: "home.utahcon.com",
					Data: addr.String(),
					TTL:  300,
				}

			}
			fmt.Printf("%s is Private: %t\n", addr.String(), isPrivate)
		}
	}
}

func CheckIpIsPrivate(address net.Addr) (bool, error) {
	addr, _, err := net.ParseCIDR(address.String())
	if err != nil {
		return false, err
	}

	if net.ParseIP(addr.String()).IsLoopback() {
		return true, nil
	}

	_, ipv6Private, err := net.ParseCIDR("fe80::/10")
	if err != nil {
		return false, err
	}

	if ipv6Private.Contains(addr) {
		return true, err
	}

	_, ipv4PrivateNetwork192, err := net.ParseCIDR("192.168.0.0/16")
	if err != nil {
		return false, err
	}

	if ipv4PrivateNetwork192.Contains(addr) {
		return true, nil
	}

	_, ipv4PrivateNetwork172, err := net.ParseCIDR("172.16.0.0/12")
	if err != nil {
		return false, err
	}

	if ipv4PrivateNetwork172.Contains(addr) {
		return true, nil
	}

	_, ipv4PrivateNetwork10, err := net.ParseCIDR("10.0.0.0/8")
	if err != nil {
		return false, err
	}

	if ipv4PrivateNetwork10.Contains(addr) {
		return true, nil
	}

	return false, nil
}
