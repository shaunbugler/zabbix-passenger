package main

import (
	"encoding/xml"
	"fmt"
	"golang.org/x/net/html/charset"
	"gopkg.in/alecthomas/kingpin.v2"
	"launchpad.net/xmlpath"
	"log"
	"os"
)

func read_xml() *xmlpath.Node {
	// TODO Execute passenger-status
	reader, err := os.Open("ppm.xml")
	if err != nil {
		log.Fatal(err)
	}

	// Stuff to handle the iso-8859-1 xml encoding
	// http://stackoverflow.com/a/32224438/606167
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	xmlData, err := xmlpath.ParseDecoder(decoder)
	if err != nil {
		log.Fatal(err)
	}

	// Check version
	version_path := xmlpath.MustCompile("/info/@version")
	if version, ok := version_path.String(xmlData); !ok || version != "3" {
		log.Fatal("Unsupported Passenger version (xml version ", version, ")")
	}

	return xmlData
}

func print_simple_selector(pathString string) {
	path := xmlpath.MustCompile(pathString)

	if value, ok := path.String(read_xml()); ok {
		fmt.Println(value)
	}
}

func print_app_groups_json() {
	path := xmlpath.MustCompile("//supergroup/name")

	app_iter := path.Iter(read_xml())

	fmt.Println("{\"data\": [")
	for app_iter.Next() {
		fmt.Printf("{\"{#NAME}\": \"%v\"},\n", app_iter.Node().String())
	}
	fmt.Println("]}")
}

var (
	app     = kingpin.New("zabbix-passenger", "A utility to parse passenger-status output for usage with Zabbix")
	appPath = app.Flag("app", "Full path to application (for app* commands)").String()

	appGroupsJson      = app.Command("app-groups-json", "Get list of application groups in JSON format (for LLD)")
	globalQueue        = app.Command("global-queue", "Get number of requests in global queue")
	globalCapacityUsed = app.Command("global-capacity-used", "Get global capacity used")

	appQueue        = app.Command("app-queue", "Get number of requests in application queue")
	appCapacityUsed = app.Command("app-capacity-used", "Get application capacity used")
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	case appGroupsJson.FullCommand():
		print_app_groups_json()
	case globalQueue.FullCommand():
		print_simple_selector("//info/get_wait_list_size")
	case appQueue.FullCommand():
		print_simple_selector(fmt.Sprintf("//supergroup[name='%v']/get_wait_list_size", *appPath))
	case globalCapacityUsed.FullCommand():
		print_simple_selector("//info/capacity_used")
	case appCapacityUsed.FullCommand():
		print_simple_selector(fmt.Sprintf("//supergroup[name='%v']/capacity_used", *appPath))
	}
}