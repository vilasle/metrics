package rest

import (
	"net/http"
	"testing"
)

func Test_allowedMethods(t *testing.T) {
	type args struct {
		method []string
	}
	tests := []struct {
		name string
		args args
		want func(h http.Handler) http.Handler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

		})
	}
}

func Test_allowedContentType(t *testing.T) {
	type args struct {
		contentTypes []string
	}
	tests := []struct {
		name string
		args args
		want func(h http.Handler) http.Handler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

		})
	}
}
