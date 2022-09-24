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
	ChangeNumber int    `json:"changeNumber"`
	Cloud        string `json:"cloud"`
	Values       []struct {
		Name       string `json:"name"`
		ID         string `json:"id"`
		Properties struct {
			ChangeNumber    int         `json:"changeNumber"`
			Region          string      `json:"region"`
			RegionID        int         `json:"regionId"`
			Platform        string      `json:"platform"`
			SystemService   string      `json:"systemService"`
			AddressPrefixes []string    `json:"addressPrefixes"`
			NetworkFeatures interface{} `json:"networkFeatures"`
		} `json:"properties"`
	} `json:"values"`
}

func outputFilename(cWeek int, cYear int, region string) *string {
	filename := ""
	filename += "ip-ranges-w"
	filename += strconv.Itoa(cWeek)
	filename += "y"
	filename += strconv.Itoa(cYear)
	filename += "-"

	if region == "" {
		region = "no-region"
	}

	filename += region
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

func main() {
	// Parse command line arguments
	region := flag.String("region", "", "filter on Azure region")
	//systemService := flag.String("service", "", "filter on Azure service")

	if *region == "" {

	}
	flag.Parse()

	tn := time.Now().UTC()

	currentYear, currentWeek := tn.ISOWeek()
	url := "https://www.microsoft.com/en-us/download/confirmation.aspx?id=56519"

	filename := outputFilename(currentWeek, currentYear, *region)
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
					if a.Key == "href" {

						// Check if the href attr contains the "ServiceTags_Public" value
						if strings.Contains(a.Val, "ServiceTags_Public") {
							// Make HTTP request
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
								if ipRanges.Values[i].Properties.Region == *region || ipRanges.Values[i].Properties.Region == "" {
									for j := 0; j < len(ipRanges.Values[i].Properties.AddressPrefixes); j++ {

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
