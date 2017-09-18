package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"unicode"
)

const xmlRequest = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:ns0="urn:ec.europa.eu:taxud:vies:services:checkVat:types"
    xmlns:ns1="http://schemas.xmlsoap.org/soap/envelope/"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/">
<SOAP-ENV:Header/>
<ns1:Body><ns0:checkVat><ns0:countryCode>%s</ns0:countryCode><ns0:vatNumber>%s</ns0:vatNumber></ns0:checkVat>
</ns1:Body>
</SOAP-ENV:Envelope>`

const serviceURI = `http://ec.europa.eu/taxation_customs/vies/services/checkVatService`

var (
	ErrorNotEnoughLetters = errors.New("the VAT should begin with at least 2 letters")
)

type VAT struct {
	countryCode string
	number      string
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: vatverify VATID")
		os.Exit(1)
	}
	toValidate := os.Args[1]
	result, err := processVAT(toValidate)
	if err != nil {
		fmt.Printf("Received an error while validating VAT: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Println(result)
}

func processVAT(pending string) (string, error) {
	vat, err := splitVAT(pending)
	if err != nil {
		return "", err
	}
	request, err := buildRequest(vat)
	if err != nil {
		return "", err
	}
	result, err := processRequest(request)
	return result, nil
}

func splitVAT(pending string) (VAT, error) {
	result := VAT{countryCode: pending[:2],
		number: pending[2:]}
	if !isAllLetters(result.countryCode) {
		return result, ErrorNotEnoughLetters
	}
	return result, nil
}

func isAllLetters(input string) bool {
	if input == "" {
		return false
	}

	for _, v := range input {
		if !unicode.IsLetter(v) {
			return false
		}
	}
	return true
}

func buildRequest(vat VAT) (*http.Request, error) {
	body := bytes.NewReader([]byte(fmt.Sprintf(xmlRequest, vat.countryCode, vat.number)))
	request, err := http.NewRequest(http.MethodPost, serviceURI, body)
	if err != nil {
		return nil, err
	}
	request.Header.Add("content-type", "text/xml")
	return request, nil
}

func processRequest(request *http.Request) (string, error) {
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	result, err := parseResponse(response)
	return result, err
}

func parseResponse(response *http.Response) (string, error) {
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("expected status 200, received %v", response.StatusCode)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	responseBody := new(Envelope)
	err = xml.Unmarshal(body, responseBody)
	if err != nil {
		return "", err
	}

	if responseBody.Body.Fault != nil {
		return responseBody.Body.Fault.FaultString, nil
	}

	if responseBody.Body.CheckVATResponse.Valid == "true" {
		return "Valid", nil
	}
	return "Invalid", nil

}

type Envelope struct {
	XMLName xml.Name
	Body    Body
}

type Body struct {
	CheckVATResponse *CheckVATResponse `xml:"checkVatResponse"`
	Fault            *Fault
}

type CheckVATResponse struct {
	XMLName     xml.Name
	CountryCode string `xml:"countryCode"`
	VatNumber   string `xml:"vatNumber"`
	RequestDate string `xml:"requestDate"`
	Valid       string `xml:"valid"`
	Name        string `xml:"name"`
	Address     string `xml:"address"`
}

type XMLField struct {
	XMLName xml.Name
}

type Fault struct {
	FaultCode   string `xml:"faultcode"`
	FaultString string `xml:"faultstring"`
}
