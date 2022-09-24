# Go Azure IP Ranges

Fetch Azure IP Ranges for a given week and write them to a file

![Build](https://github.com/jeroensmink98/go-azure-ip-ranges/actions/workflows/github-actions-go.yml/badge.svg)


# Getting started

Download one of the newest [releases](https://github.com/jeroensmink98/go-azure-ip-ranges/releases)

```powershell
az-ip-ranges.exe 
``` 
This will output the IPv4 addresses to a file named like `ip-ranges-weekNr-Year-region.txt`

for example `ip-ranges-w38-2022-westeurope.txt`

Or 

Fetching all IPv4 addresses for a specific region

```powershell
az-ip-ranges.exe --region=westeurope
``` 
If you don't specify a region it will only add the IPv4 addresses of Azure services that are not bound to a specifc region.

The filename will contain `no-region` to mark that file as not having specified a region parameter

## How does it work?

- The application makes an HTTP request to the download page of the weekly `.json` file Microsoft publishes
- in the HTML body it looks for a `href` attribute with the value of `ServiceTags_Public`. This is part of the name of the `.json` file
- It gets the link from the `<a>` tag and makes a new HTTP request to that url, It now retrieves the weekly uploaded `.json`
- It outputs the HTTP response to a new `.json` file
- Then it parses the `.json` to a struct we created named `AzureIpRange`
- Next it makes some checks based on the filter you have specified to add or skip a certain entry for the new file
- It creates a new `.txt` file with the IPv4 addresses based on your filters

> ⚠️ By default it will also add addresses of services that are not bound to a region or have RegionId 0 which also means the service does not belong to a specific Azure region.

# Todo

- [x] Add filter for Region and Azure Platform
- [ ] Make CLI tool from application
- [ ] Read from HTTP response instead of local file
- [ ] Add filter for specific Azure Service `systemService`
- [x] Add option to query all regions
- [ ] Add option to query all Azure services
