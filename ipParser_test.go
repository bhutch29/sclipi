package main

import (
	"net"
	"testing"
)

func TestParseIpv4(t *testing.T) {
	testData := []net.IP{net.ParseIP("1.2.3.4"), net.ParseIP("1.2.99.99"), net.ParseIP("5.11.99.99"), net.ParseIP("5.12.99.99")}
	result := parseIpv4(testData)
	if len(result.Children) != 2 {
		t.Error("Wrong number of first level children", result.Children)
	}
	if result.Children[0].Children[0].Content != "2" {
		t.Error("Wrong element found in second level first node first child", result.Children[0].Children)
	}
	if len(result.Children[1].Children) != 2 {
		t.Error("Wrong number of second level second node children", result.Children[1].Children)
	}
	if result.Children[1].Children[1].Content != "12" {
		t.Error("Wrong element found in second level second node second child", result.Children[1].Children)
	}
}
