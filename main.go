package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type AzureIpRange struct {
	ChangeNumber int     `json:"changeNumber"`
	Cloud        string  `json:"cloud"`
	Values       []Value `json:"values"`
}

type Value struct {
	Name       string     `json:"name"`
	ID         string     `json:"id"`
	Properties Properties `json:"properties"`
}

type Properties struct {
	ChangeNumber    int      `json:"changeNumber"`
	Region          string   `json:"region"`
	RegionID        int      `json:"regionId"`
	Platform        string   `json:"platform"`
	SystemService   string   `json:"systemService"`
	AddressPrefixes []string `json:"addressPrefixes"`
	NetworkFeatures []string `json:"networkFeatures"`
}

func outputFilename(cWeek int, cYear int, region string, service string) *string {
	filename := ""
	filename += "ip-ranges-w"
	filename += strconv.Itoa(cWeek)
	filename += "y"
	filename += strconv.Itoa(cYear)
	filename += "-"

	if region == "" {
		region = "no-region"
	}

	if service == "" {
		service = "no-service"
	}

	filename += region
	filename += "-"
	filename += service
	filename += ".txt"
	return &filename
}

func writeToFile(content string, f os.File) {
	f.WriteString(content)
}

// This function removes the subnet mask from the IPv4 Address
func formatIpv4(ip string) string {
	return (strings.Split(ip, "/")[0] + "\n")
}

func matchSystemFilter(filter string) {

}

func main() {
	// Parse command line arguments
	region := flag.String("region", "", "filter on Azure region")
	service := flag.String("service", "", "filter on Azure service")

	flag.Parse()

	tn := time.Now().UTC()

	currentYear, currentWeek := tn.ISOWeek()
	url := "https://www.microsoft.com/en-us/download/confirmation.aspx?id=56519"

	filename := outputFilename(currentWeek, currentYear, *region, *service)
	file, err := os.Create(*filename)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	var client http.Client
	resp, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)

		if err != nil {
			log.Fatal(err)
		}

		// Parse HTTP response to HTML object
		doc, err := html.Parse(strings.NewReader(string(bodyBytes)))

		var f func(*html.Node)
		f = func(n *html.Node) {
			// Search in HTML Object for all <a> tags
			if n.Type == html.ElementNode && n.Data == "a" {
				for _, a := range n.Attr {
					// Fetch the href attribute from all <a> tags
					if a.Key == "href" && strings.Contains(a.Val, "ServiceTags_Public") {
						// Make HTTP request to the href value of the <a>
						res, err := http.Get(a.Val)

						body, err := ioutil.ReadAll(res.Body)
						if err != nil {
							panic(err.Error())
						}

						//Write output to a JSON File
						err = ioutil.WriteFile("AzurePublicIp.json", body, 0644)

						fileContent, err := os.Open("./AzurePublicIp.json")
						if err != nil {
							log.Fatal(err)
							return
						}

						defer fileContent.Close()

						byteResult, _ := ioutil.ReadAll(fileContent)

						var ipRanges AzureIpRange

						json.Unmarshal([]byte(byteResult), &ipRanges)

						for i := 0; i < len(ipRanges.Values); i++ {
							// Only add values that match our region filter
							if ipRanges.Values[i].Properties.Region == *region {
								for j := 0; j < len(ipRanges.Values[i].Properties.AddressPrefixes); j++ {

									if *service != "" {
										if ipRanges.Values[i].Properties.SystemService == *service {
											// Create file content string
											writeToFile(formatIpv4(ipRanges.Values[i].Properties.AddressPrefixes[j]), *file)
										}
									} else {
										// Create file content string
										writeToFile(formatIpv4(ipRanges.Values[i].Properties.AddressPrefixes[j]), *file)
									}

								}
							}
						}
						break
					}
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
		f(doc)
	}
}
