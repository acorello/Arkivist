package main

import (
	"encoding/xml"
	"fmt"
	"time"
)

type Person struct {
	XMLName xml.Name  `xml:"person"`
	Name    string    `xml:"name"`
	DOB     time.Time `xml:"dob"`
}

func main() {
	p := Person{Name: "John Doe", DOB: time.Now()}
	xmlBytes, err := xml.MarshalIndent(p, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling to XML:", err)
		return
	}
	xmlString := string(xmlBytes)
	fmt.Println(xmlString)
}
