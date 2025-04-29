package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type DictionaryResponse []struct {
	Phonetics []struct {
		Audio string `json:"audio,omitempty"`
	} `json:"phonetics"`

	Meanings []struct {
		PartOfSpeech string   `json:"partOfSpeech,omitempty"`
		Synonyms     []string `json:"synonyms,omitempty"`
		Antonyms     []string `json:"antonyms,omitempty"`

		Definitions []struct {
			Definition string   `json:"definition"`
			Synonyms   []string `json:"synonyms,omitempty"`
			Antonyms   []string `json:"antonyms,omitempty"`
			Example    string   `json:"example,omitempty"`
		} `json:"definitions"`
	} `json:"meanings"`
}

func autoCorrect(word string) string {
	token := os.Getenv("TEXTGEARS_API")
	url := "https://api.textgears.com/correct?text=" + word + "&language=en-GB&key=" + token

	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Request failed with status code: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatal(err)
	}

	// Pull out auto-corrected value
	if respMap, ok := data["response"].(map[string]interface{}); ok {
		if corrected, ok := respMap["corrected"].(string); ok {
			return corrected
		}
	}

	return word
}

// func contains(slice []string, item string) bool {
// 	for _, val := range slice {
// 		if val == item {
// 			return true
// 		}
// 	}
// 	return false
// }

func dictionaryApi(word string) {

	url := "https://api.dictionaryapi.dev/api/v2/entries/en/" + word

	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(response.Body)

	if err != nil {
		log.Fatalln(err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Request failed with status code: %d", response.StatusCode)
	}

	// mapping
	var result DictionaryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalln(err)
	}

	// display audio/pronunciation link
	seen := make(map[string]bool)
	for _, entry := range result {
		for _, phonetic := range entry.Phonetics {
			audio := phonetic.Audio
			if audio != "" && !seen[audio] {
				fmt.Println("Audio:", audio)
				seen[audio] = true
			}
		}
	}

	// first source is enough and its more relevant
	for _, meaning := range result[0].Meanings {
		fmt.Println("")
		fmt.Println("Part of Speech:", meaning.PartOfSpeech)

		// limit definitions to 3
		defs := meaning.Definitions
		if len(defs) > 2 {
			defs = defs[:2]
		}
		fmt.Println("Definition(s):")
		for count, def := range defs {
			fmt.Printf("%d. %s\n", count+1, def.Definition)
		}

		// need to display max 1 example for each PartOfSpeech
		for _, def := range meaning.Definitions {
			if len(def.Example) > 1 {
				fmt.Println("")
				fmt.Println("Example: ", def.Example)
				break
			}
		}

		if len(meaning.Synonyms) > 0 {
			fmt.Println("Synonym(s): ", strings.Join(meaning.Synonyms, ", "))
		}
		if len(meaning.Antonyms) > 0 {
			fmt.Println("Antonym(s): ", strings.Join(meaning.Antonyms, ", "))
		}

	}

	fmt.Println("")

}

func main() {
	var input string

	if len(os.Args) > 1 {
		input = os.Args[1]
	} else {
		fmt.Print("Enter one word: ")
		fmt.Scan(&input)
	}

	validString := regexp.MustCompile(`^[a-zA-Z]+$`).MatchString

	if !(validString(input)) {
		fmt.Println("Please enter string only!")
		os.Exit(1)
	}

	word := autoCorrect(input)

	fmt.Println("Searching for: " + word)

	dictionaryApi(strings.ToUpper(word))
}
