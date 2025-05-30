package actions

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"memo/sunbeam"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type Response struct {
	NextPageToken string         `json:"nextPageToken"`
	Memos         []sunbeam.Memo `json:"memos"`
}

// extractCommand parses the command from the shell code block
func extractCodeBlock(content string) string {
	// Match the content inside the shell code block
	re := regexp.MustCompile("(?s)```\\w*\\n(.*?)\\n```")
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractTags parses the tags from the Tags section
func extractTags(content string) []string {
	// Match the line starting with **Tags:** and extract hashtags
	re := regexp.MustCompile(`(?i)(#[a-zA-Z0-9-_]+(?:\s*#[a-zA-Z0-9-_]+)*)`)

	matches := re.FindStringSubmatch(content)
	//fmt.Println(matches)
	if len(matches) > 1 {
		tagsLine := matches[1]
		tags := strings.Fields(tagsLine) // Split by spaces
		for i, tag := range tags {
			tags[i] = strings.TrimPrefix(tag, "#") // Remove leading '#'
		}
		return tags
	}
	return nil
}

// Function to filter commands based on a tag
func filterCommandsByTag(resultSlice []map[string]string, tag string) map[string]string {
	// Create a map to hold the filtered results
	filteredResults := make(map[string]string)

	// Iterate through the slice and check if the "tags" contain the specified tag
	for _, result := range resultSlice {
		// Check if the "tags" contain the provided tag (case-sensitive)
		if strings.Contains(result["tags"], tag) {
			// Add the command to the filtered map
			filteredResults[result["name"]] = result["tags"]
		}
	}

	return filteredResults
}

func memos(token string, apiURL string) ([]sunbeam.Memo, error) {
	var allMemos []sunbeam.Memo
	url := apiURL

	// Loop to handle pagination
	for {
		// Create the request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}

		// Add Authorization header
		req.Header.Add("Authorization", "Bearer "+token)

		// Send the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to send request: %v", err)
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}

		// Check if the request was successful
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		// Parse the JSON response
		var response Response
		err = json.Unmarshal(body, &response)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response JSON: %v", err)
		}

		// Append the memos to the allMemos slice
		allMemos = append(allMemos, response.Memos...)

		// Check if there is a next page, and if so, update the URL
		if response.NextPageToken == "" {
			break // No more pages, exit the loop
		}
		url = fmt.Sprintf("%s&pageToken=%s", apiURL, response.NextPageToken)
		//fmt.Println(url)
	}

	return allMemos, nil
}

func GetMemos(tags *string, apiKey string, apiURL string) {

	// Split the tags into a slice
	tagList := strings.Split(*tags, ",")
	// Format tags into query parameter
	rawUrl := ""

	// Ensure the API URL ends with `/api/memos`
	if !strings.HasSuffix(apiURL, "/api/v1/memos") {
		apiURL = strings.TrimRight(apiURL, "/") + "/api/v1/memos"
	}

	// if no tags, the default is an empty string, which will be 1
	if len(tagList) == 1 && tagList[0] == "" {
		rawUrl = fmt.Sprintf("%s?", apiURL)
	} else {
		formattedTags := fmt.Sprintf("tag in ['%s']", strings.Join(tagList, "','"))
		urlEncodedTags := url.QueryEscape(formattedTags)
		// Construct the full URL with the tag filter
		rawUrl = fmt.Sprintf("%s?filter=%s", apiURL, urlEncodedTags)
	}

	//fmt.Println("URL: ", rawUrl)
	memos, err := memos(apiKey, rawUrl)
	if err != nil {
		log.Fatalf("Error retrieving memos: %v", err)
	}

	stringsMap := []map[string]string{}
	for _, memo := range memos {
		codeBlock := extractCodeBlock(memo.Content)
		tags := extractTags(memo.Content)
		itemMap := map[string]string{"cmd": codeBlock, "tags": strings.Join(tags, " "), "content": memo.Content, "name": memo.Name}
		stringsMap = append(stringsMap, itemMap)
	}

	jsonData, err := json.MarshalIndent(stringsMap, "", "  ")
	if err != nil {
		log.Fatalf("Error converting to JSON: %v", err)
	}

	// Print JSON to console
	fmt.Println(string(jsonData))
}
