package main

import (
	"context"
	"fmt"
	"github.com/digitalocean/godo"
	"github.com/utahcon/regex/ipv4"
	"github.com/utahcon/regex/ipv6"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

type Configuration struct {
	Key    string `yaml:"key"`
	Domain string `yaml:"domain"`
	Record string `yaml:"record"`
}

func LoadConfiguration() (*Configuration, error) {
	config := &Configuration{}

	data, err := ioutil.ReadFile("/etc/do-dns-updater.yml")
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal([]byte(data), config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func main() {

	config, err := LoadConfiguration()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	client := godo.NewFromToken(config.Key)

	ctx := context.TODO()

	ipv4Domains, _, err := client.Domains.RecordsByTypeAndName(ctx, config.Domain, "A", strings.Join([]string{config.Record, config.Domain}, "."), &godo.ListOptions{})
	if err != nil {
		fmt.Printf("Error getting IPv4 domains list: %v\n", err)
		os.Exit(1)
	}

	ipv6Domains, _, err := client.Domains.RecordsByTypeAndName(ctx, config.Domain, "AAAA", strings.Join([]string{config.Record, config.Domain}, "."), &godo.ListOptions{})
	if err != nil {
		fmt.Printf("Error getting IPv6 domains list: %v\n", err)
		os.Exit(1)
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}

	for _, iface := range interfaces {
		addresses, err := iface.Addrs()
		if err != nil {
			fmt.Printf("error: %s\n", err)
			os.Exit(1)
		}

		for _, address := range addresses {
			addr, _, err := net.ParseCIDR(address.String())
			if err != nil {
				fmt.Printf("error parsing CIDR %s: %s\n", address.String(), err)
				os.Exit(1)
			}

			isPrivate, err := CheckIpIsPrivate(addr)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				os.Exit(1)
			}

			if !isPrivate {
				if ipv4.Validate(addr.String()) {
					for _, domain := range ipv4Domains {
						drer := &godo.DomainRecordEditRequest{
							Type: "A",
							Name: config.Record,
							Data: addr.String(),
							TTL:  300,
						}
						newRecord, response, err := client.Domains.EditRecord(ctx, config.Domain, domain.ID, drer)
						if err != nil {
							fmt.Printf("New Record: %v\n", newRecord)
							fmt.Printf("Response: %v\n", response)
							fmt.Printf("Error updating record: %v\n", err)
							os.Exit(1)
						}
					}
				}

				if ipv6.Validate(addr.String()) {
					for _, domain := range ipv6Domains {
						drer := &godo.DomainRecordEditRequest{
							Type: "AAAA",
							Name: config.Record,
							Data: addr.String(),
							TTL:  300,
						}
						fmt.Printf("Request: %v", drer)
						newRecord, response, err := client.Domains.EditRecord(ctx, config.Domain, domain.ID, drer)
						if err != nil {
							fmt.Printf("New Record: %v\n", newRecord)
							fmt.Printf("Response: %v\n", response)
							fmt.Printf("Error updating record: %v\n", err)
							os.Exit(1)
						}
					}
				}
			}
		}
	}
}

func CheckIpIsPrivate(addr net.IP) (bool, error) {
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
