package app

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/rancher/go-rancher-metadata/metadata"
)

type StringifyError struct {
	message string
}

func (e *StringifyError) Error() string {
	return e.message
}

func IPCommand() cli.Command {
	return cli.Command{
		Name:  "ip",
		Usage: "Get IP information",
		Subcommands: []cli.Command{
			{
				Name:   "stringify",
				Usage:  "Prints a joined list of IPs",
				Action: ipStringifyAction,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "delimiter",
						Usage: "Delimiter to use between entries",
						Value: ",",
					},
					cli.StringFlag{
						Name:  "prefix",
						Usage: "Prepend Entries with this value",
						Value: "",
					},
					cli.StringFlag{
						Name:  "suffix",
						Usage: "Add this value to the end of each entry.",
						Value: "",
					},
					cli.StringFlag{
						Name:  "source",
						Usage: "Source to lookup IPs. [metadata, dns]",
						Value: "metadata",
					},
					cli.BoolFlag{
						Name:  "use-agent-ips",
						Usage: "Use agent ips instead of rancher ips, only works with metadata source",
					},
					cli.BoolFlag{
						Name:  "use-agent-names",
						Usage: "Use agent name instead of rancher ips, only works with metadata source",
					},
				},
			},{
				Name:  "myip",
				Usage: "Prints the containers Rancher managed IP",
				Action: ipMyIpAction,
			},
		},
	}
}

func ipMyIpAction(c *cli.Context) {
	mdClient, _ := metadata.NewClientAndWait(metadataURL)

	selfContainer, err := mdClient.GetSelfContainer()
	if err != nil {
		logrus.Fatalf("Failed to find IP: %v", err)
	}
	fmt.Print(selfContainer.PrimaryIp)
}

func ipStringifyAction(c *cli.Context) {
	str := ""
	var err error

	if c.String("source") == "dns" {
		str, err = ipStringifyDNS(c)
		if err != nil {
			logrus.Fatalf("Failed to generate string: %v", err)
		}
	} else {
		str, err = ipStringifyMetadata(c)
		if err != nil {
			logrus.Fatalf("Failed to generate string: %v", err)
		}
	}

	fmt.Print(str)
}

func ipStringifyDNS(c *cli.Context) (string, error) {
	if len(c.Args()) <= 0 {
		return "", nil
	}
	ips, err := getDnsContainerIPs(c.Args().First())
	rString := joinString(
		c.String("prefix"),
		c.String("suffix"),
		c.String("delimiter"),
		ips,
	)
	return rString, err
}

func getDnsContainerIPs(host string) ([]string, error) {
	ips := []string{}
	err := &StringifyError{"Could not resolve Host: " + host}

	if testDNSResolves(host) {
		if ipBytes, errIP := net.LookupIP(host); errIP == nil {
			for _, ip := range ipBytes {
				ips = append(ips, ip.String())
			}
			return ips, errIP
		}
	}

	return ips, err
}

func testDNSResolves(host string) bool {
	retVal := false
	ticker := time.NewTicker(time.Millisecond * 500)
	timer := time.NewTimer(time.Second * 60)
	resolves := make(chan bool)

	go func() {
		for range ticker.C {
			if _, err := net.LookupIP(host); err == nil {
				retVal = true
				resolves <- true
			}
		}
	}()

	for {
		select {
		case <-timer.C:
			ticker.Stop()
			return retVal
		case <-resolves:
			ticker.Stop()
			timer.Stop()
			return retVal
		}
	}
}

func ipStringifyMetadata(c *cli.Context) (string, error) {
	split := []string{}
	rString := ""
	var err error

	if len(c.Args()) > 0 {
		split = strings.SplitN(c.Args().First(), "/", 2)
	} else {
		split, err = getSelfStackServiceName()
	}

	getMetaIPMethod := getMetadataContainerIPs
	if c.Bool("use-agent-ips") {
		getMetaIPMethod = getMetadataAgentIPs
	}

	if c.Bool("use-agent-names") {
		getMetaIPMethod = getMetadataAgentNames
	}

	if len(split) == 2 {
		ips, err := getMetaIPMethod(split[0], split[1])
		if err != nil {
			return rString, err
		}
		rString = joinString(
			c.String("prefix"),
			c.String("suffix"),
			c.String("delimiter"),
			ips,
		)
		err = nil
	} else {
		err = &StringifyError{"Not enough arguements supplied. Need stack/service or this container is not part of service"}
	}

	return rString, err
}

func getSelfStackServiceName() ([]string, error) {
	mdClient, _ := metadata.NewClientAndWait(metadataURL)

	selfContainer, err := mdClient.GetSelfContainer()
	if err != nil {
		return nil, err
	}

	return []string{selfContainer.StackName, selfContainer.ServiceName}, err
}

func getMetadataContainerIPs(stack string, service string) ([]string, error) {
	rIPs := []string{}
	mdClient, _ := metadata.NewClientAndWait(metadataURL)

	containers, err := mdClient.GetServiceContainers(service, stack)
	if err != nil {
		return rIPs, err
	}

	for _, container := range containers {
		rIPs = append(rIPs, container.PrimaryIp)
	}

	return rIPs, nil
}

func getMetadataAgentIPs(stack string, service string) ([]string, error) {
	return getMetadataAgentInfoStrings(stack, service, "AgentIP")
}

func getMetadataAgentNames(stack string, service string) ([]string, error) {
	return getMetadataAgentInfoStrings(stack, service, "Name")
}

func getMetadataAgentInfoStrings(stack, service, property string) ([]string, error) {
	rInfo := []string{}
	mdClient, _ := metadata.NewClientAndWait(metadataURL)

	containers, err := mdClient.GetServiceContainers(service, stack)
	if err != nil {
		return rInfo, err
	}

	for _, container := range containers {
		host, err := mdClient.GetHost(container.HostUUID)
		if err != nil {
			return rInfo, err
		}
		rInfo = append(rInfo, getHostInfoProperty(&host, property))
	}

	return rInfo, nil
}

func getHostInfoProperty(host *metadata.Host, property string) string {
	switch {
	case property == "Name":
		return host.Name
	case property == "AgentIP":
		return host.AgentIP
	default:
		return ""
	}
}

func joinString(pfx string, suffix string, delim string, list []string) string {
	intList := []string{}
	for _, item := range list {
		intermediate := pfx + item + suffix
		intList = append(intList, strings.Repeat(intermediate, 1))
	}
	return strings.Join(intList, delim)
}
