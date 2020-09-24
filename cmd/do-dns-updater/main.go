package main

import (
	"context"
	"fmt"
	"github.com/digitalocean/godo"
	"github.com/utahcon/regex"
	"net"
	"strings"
)

/*

IPv6 CIDR:
^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))\/[0-9]{1,2}$

IPv6:
^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$

IPv4:
^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$

IPV4 CIDR:
^(?:[0-9]{1,3}\.){3}[0-9]{1,3}\/[0-9]{1,2}$

*/

type Configuration struct {
	Key    string `yaml:"key"`
	Domain string `yaml:"domain"`
	Record string `yaml:"record"`
}

func main() {

	config := &Configuration{
		Key: "396d2b3afce375a5f820a176704b88552d611a944498a3380e7cb4815677ed85",
		Domain: "utahcon.com",
		Record: "home",
	}

	client := godo.NewFromToken(config.Key)

	ctx := context.TODO()

	ipv4Domains, _, err := client.Domains.RecordsByTypeAndName(ctx, config.Domain, "A", strings.Join([]string{config.Record, config.Domain}, "."), &godo.ListOptions{})
	if err != nil {
		fmt.Printf("Error getting IPv4 domains list: %v\n", err)
	}

	//ipv6Domains, _, err := client.Domains.RecordsByTypeAndName(ctx, config.Domain, "AAAA", strings.Join([]string{config.Record, config.Domain}, "."), &godo.ListOptions{})
	//if err != nil {
	//	fmt.Printf("Error getting IPv6 domains list: %v\n", err)
	//}

	//for _, domain := range ipv6Domains {
	//	fmt.Printf("Domain: %v\n", domain)
	//}

	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("error: %s", err)
	}

	for _, iface := range interfaces {
		addresses, err := iface.Addrs()
		if err != nil {
			fmt.Printf("error: %s\n", err)
		}

		for _, address := range addresses {
			addr, _, err := net.ParseCIDR(address.String())
			if err != nil {
				return false, err
			}

			isPrivate, err := CheckIpIsPrivate(addr.String())
			if err != nil {
				fmt.Printf("error: %s\n", err)
			}

			if !isPrivate {
				if ipv4.Validate(addr.String()){
					fmt.Println("IP Address is IPv4")
					for _, domain := range ipv4Domains {
						fmt.Printf("Domain: %v\n", domain)
						drer := &godo.DomainRecordEditRequest{
							Type: "A",
							Name: config.Record,
							Data: net.IP(addr,
							TTL: 300,
						}
						fmt.Printf("Request: %v", drer)
						newRecord, response, err := client.Domains.EditRecord(ctx, config.Domain, domain.ID, drer)
						if err != nil {
							fmt.Printf("New Record: %v\n", newRecord)
							fmt.Printf("Response: %v\n", response)
							fmt.Printf("Error updating record: %v\n", err)
						}

					}
				}
			}
			fmt.Printf("%s is Private: %t\n", addr.String(), isPrivate)
		}
	}
}

func CheckIpIsPrivate(address string) (bool, error) {
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
