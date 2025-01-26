package main

import (
	"flag"
	"fmt"
	"memo/sunbeam"
	"os"
	"path/filepath"
)

func main() {
	// Example path to the Sunbeam configuration file
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "sunbeam", "sunbeam.json")
	// Retrieve memo preferences
	preferences, err := sunbeam.ReadSunbeamConfig(configPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	apiKey := ""
	apiURL := ""
	if len(preferences.MemoToken) == 0 || len(preferences.MemoURL) == 0 {
		fmt.Printf("Error: no values found in sunbeam memo extension configuration ... trying environment variables")

		apiKey = os.Getenv("USEMEMOS_API_KEY")
		apiURL = os.Getenv("USEMEMOS_API_URL")
	} else {
		apiKey = preferences.MemoToken
		apiURL = preferences.MemoURL
	}

	if apiKey == "" || apiURL == "" {
		fmt.Println("Environment variables USEMEMOS_API_KEY and USEMEMOS_API_URL must be set.")
		os.Exit(1)
	}

	tags := flag.String("tags", "", "Comma-separated list of tags for the memo (e.g., 'shell,commands')")
	clipboard := flag.Bool("clipboard", false, "Create a memo using the contents of the clipboard")
	update := flag.Bool("update", false, "Update memo using the contents of the clipboard")
	name := flag.String("name", "", "id of memo to update")
	flag.Parse()

	if *update {
		if *name == "" {
			fmt.Println("Please provide a memo id to update")
			os.Exit(1)
		} else {
			updateMemo(apiURL, name, apiKey)
		}
	} else if *clipboard {
		postMemo(clipboard, tags, apiURL, apiKey)
	} else {
		// by default get all memos or filter by tags
		getMemos(tags, apiKey, apiURL)
	}
}
