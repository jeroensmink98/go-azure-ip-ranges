package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	// Azure Public IP Ranges Download page
	const region = "westeurope"
	const platform = "Azure"
	const url = "https://www.microsoft.com/en-us/download/confirmation.aspx?id=56519"

	file, err := os.Create("output.txt")

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

		bodyString := string(bodyBytes)

		// Parse HTTP response to HTML object
		doc, err := html.Parse(strings.NewReader(bodyString))

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
								// Only write the IPv4 addresses that are within our specified region
								// Todo: Add function to write all IP's instead of a single region
								if ipRanges.Values[i].Properties.Region == region && ipRanges.Values[i].Properties.Platform == platform {
									for j := 0; j < len(ipRanges.Values[i].Properties.AddressPrefixes); j++ {
										s := ipRanges.Values[i].Properties.AddressPrefixes[j]
										s += "\n"

										_, err := file.WriteString(s)

										if err != nil {
											panic(err)
										}

										fmt.Println(ipRanges.Values[i].Properties.AddressPrefixes[j])
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
