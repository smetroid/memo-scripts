package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"memo/sunbeam"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// getLastShellCommand retrieves the last executed command from the shell history
func getLastShellCommand() (string, error) {
	// Get the shell's history file (assuming Bash or Zsh)
	historyFile := os.Getenv("HISTFILE")
	if historyFile == "" {
		historyFile = os.ExpandEnv("$HOME/.bash_history") // Default to Bash history
	}

	// Read the history file
	data, err := os.ReadFile(historyFile)
	if err != nil {
		return "", fmt.Errorf("failed to read history file: %w", err)
	}

	// Split the history file into lines and return the last non-empty line
	lines := strings.Split(string(data), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return lines[i], nil
		}
	}

	return "", fmt.Errorf("no commands found in history file")
}

// UpdateMemo updates an existing memo entry via the Memo API.
func UpdateMemo(apiURL string, memoName string, updatedData sunbeam.Memo, apiKey string) (sunbeam.Memo, error) {
	// Construct the URL with the memo ID
	url := fmt.Sprintf("%s/api/v1/%s", apiURL, memoName)

	// Serialize the updated data to JSON
	payload, err := json.Marshal(updatedData)
	if err != nil {
		fmt.Printf("Error serializing JSON: %v\n", err)
		os.Exit(1)
	}

	// Create an HTTP PUT request
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(payload))
	if err != nil {
		return sunbeam.Memo{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return sunbeam.Memo{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return sunbeam.Memo{}, fmt.Errorf("unexpected response status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse the response body
	var updatedMemo sunbeam.Memo
	err = json.NewDecoder(resp.Body).Decode(&updatedMemo)
	if err != nil {
		return sunbeam.Memo{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return updatedMemo, nil
}

func updateMemo(apiURL string, name *string, apiKey string) {
	file := fmt.Sprintf("/tmp/%s.md", *name)
	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", content, err)
		os.Exit(1)
	}

	//cmd := exec.Command("sunbeam", "paste")
	updateData := sunbeam.Memo{
		Content: string(content),
		Name:    string(*name),
	}
	_, err = UpdateMemo(apiURL, *name, updateData, apiKey)
	if err != nil {
		fmt.Printf("Error updating memo: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Updated memo id : %s\n", *name)

}
func postMemo(clipboard *bool, shellCommand *bool, tags *string, apiURL string, apiKey string) {
	var content string
	if *clipboard {
		cmd := exec.Command("sunbeam", "paste")
		out, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error reading clipboard: %v\n", err)
			os.Exit(1)
		}
		content = string(out)
	} else if *shellCommand {
		lastCommand, err := getLastShellCommand()
		if err != nil {
			fmt.Printf("Error retrieving last shell command: %v\n", err)
			os.Exit(1)
		}
		content = lastCommand
	} else {
		fmt.Print("Unable to post memo ... please specify --clipboard or --shellCommand")
		os.Exit(1)

	}
	// Ensure the API URL ends with `/api/memo`
	if !strings.HasSuffix(apiURL, "/api/v1/memos") {
		apiURL = strings.TrimRight(apiURL, "/") + "/api/v1/memos"
	}

	// Trim any trailing newline
	content = strings.TrimSpace(content)

	// If the last command is empty, exit
	if content == "" {
		fmt.Println("No last shell command found.")
		os.Exit(1)
	}

	// Prompt for additional tags
	fmt.Printf("content: %s \n", content)
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter additional tags (comma-separated): ")
	additionalTags, _ := reader.ReadString('\n')
	additionalTags = strings.TrimSpace(additionalTags)

	// Combine tags from flag and prompt
	var allTags []string
	if *tags != "" {
		allTags = append(allTags, strings.Split(*tags, ",")...)
	}
	if additionalTags != "" {
		allTags = append(allTags, strings.Split(additionalTags, ",")...)
	}

	// Extract the first word as a tag
	firstWord := strings.Split(content, " ")[0]

	// Prefix tags with # and format as Markdown
	var hashtags []string

	//add default tags
	hashtags = append(hashtags, "#cmds")
	hashtags = append(hashtags, "#"+firstWord)

	for _, tag := range allTags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			hashtags = append(hashtags, "#"+tag)
		}
	}

	tagsMarkdown := ""
	if len(hashtags) > 0 {
		tagsMarkdown = "\n\n**Tags:**\n" + strings.Join(hashtags, " ")
	}

	// Create Markdown content
	markdownContent := fmt.Sprintf("```bash\n%s\n```%s", content, tagsMarkdown)

	// Create memo payload
	memo := map[string]interface{}{
		"content":    markdownContent,
		"visibility": "PUBLIC", // Default visibility
	}
	if len(allTags) > 0 {
		memo["tags"] = allTags
	}

	payload, err := json.Marshal(memo)
	if err != nil {
		fmt.Printf("Error creating request payload: %v\n", err)
		os.Exit(1)
	}

	// Create HTTP client and request
	client := &http.Client{}
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("Error creating HTTP request: %v\n", err)
		os.Exit(1)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Output response
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Memo posted successfully!")
		//fmt.Printf("Response: %s\n", string(body))
	} else {
		fmt.Printf("Failed to post memo. Status: %s\n", resp.Status)
		fmt.Printf("Response: %s\n", string(body))
		os.Exit(1)
	}
}
