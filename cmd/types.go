/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

// CustomString represents a custom field with encryption metadata
type CustomString struct {
	Value     string `json:"value"`
	Encrypted bool   `json:"encrypted"`
}
