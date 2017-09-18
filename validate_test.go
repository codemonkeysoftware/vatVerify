package main

import (
	"testing"
)

func TestProcessVAT(t *testing.T) {
	t.Run("invalid format", func(t *testing.T) {
		testVAT := "123"
		_, err := processVAT(testVAT)
		if err != ErrorNotEnoughLetters {
			t.Fatalf("expected ErrorNotEnoughLetters, received %s", err.Error())
		}
	})

	t.Run("invalid", func(t *testing.T) {
		testVAT := "CZ2"
		result, err := processVAT(testVAT)
		if err != nil {
			t.Fatal(err)
		}
		if result != "Invalid" {
			t.Fatalf("expected Invalid, received %s", result)
		}
	})

	t.Run("valid", func(t *testing.T) {
		testVAT := "CZ28987373"
		result, err := processVAT(testVAT)
		if err != nil {
			t.Fatal(err)
		}
		if result != "Valid" {
			t.Fatalf("expected Valid, received %s", result)
		}
	})
}

func TestSplitVAT(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		testVAT := "123"
		_, err := splitVAT(testVAT)
		if err != ErrorNotEnoughLetters {
			t.Fatal("Did not receive expected error")
		}
	})

	t.Run("invalid", func(t *testing.T) {
		testVAT := "CZ28987373"
		expected := VAT{
			countryCode: "CZ",
			number:      "28987373",
		}
		result, err := splitVAT(testVAT)
		if err != nil {
			t.Fatal("Received unexpected error")
		}
		if result != expected {
			t.Fatalf("Did not receive expected result. Expected %#v, received %#v", expected, result)
		}
	})
}

func TestIsAllLetters(t *testing.T) {
	t.Run("all letters", func(t *testing.T) {
		input := "abc"
		expected := true
		result := isAllLetters(input)
		if result != expected {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	})

	t.Run("includes number", func(t *testing.T) {
		input := "ab3"
		expected := false
		result := isAllLetters(input)
		if result != expected {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		input := ""
		expected := false
		result := isAllLetters(input)
		if result != expected {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	})
}

func TestBuildRequest(t *testing.T) {
	vat := VAT{
		countryCode: "CZ",
		number:      "28987373",
	}
	request, err := buildRequest(vat)
	if err != nil {
		t.Fatal("received unexpected error")
	}
	if request == nil {
		t.Fatal("expected request, got nil")
	}
	if request.Header.Get("content-type") != "text/xml" {
		t.Errorf("got incorrect content-type: expected text/xml, received %s", request.Header.Get("content-type"))
	}
	if request.Body == nil {
		t.Error("body was nil")
	}
}

func TestProcessRequest(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		vat := VAT{
			countryCode: "CZ",
			number:      "28987373",
		}
		request, err := buildRequest(vat)
		if err != nil {
			t.Fatal("received unexpected error")
		}

		result, err := processRequest(request)
		if err != nil {
			t.Fatal("received unexpected error")
		}

		if result != "Valid" {
			t.Fatalf("expected valid, received %s", result)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		vat := VAT{
			countryCode: "CZ",
			number:      "2",
		}
		request, err := buildRequest(vat)
		if err != nil {
			t.Fatal("received unexpected error")
		}

		result, err := processRequest(request)
		if err != nil {
			t.Fatalf("received unexpected error: %s", err.Error())
		}

		if result != "Invalid" {
			t.Fatal("did not receive a valid result")
		}
	})
}
