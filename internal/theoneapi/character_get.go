package theoneapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func (c *Client) GetCharacterByName(characterName string) (Character, error) {
	characterNameLower := strings.ToLower(characterName)

	if val, ok := c.cache.Get("character:" + characterNameLower); ok {
		characterID := string(val)
		return c.fetchCharacterByID(characterID)
	}

	charResp, err := c.ListCharacters()
	if err != nil {
		return Character{}, err
	}

	for _, character := range charResp.Docs {
		if strings.ToLower(character.Name) == characterNameLower {
			c.cache.Add("character:"+strings.ToLower(character.Name), []byte(character.ID))
			return c.fetchCharacterByID(character.ID)
		}
	}

	return Character{}, errors.New("Character not found")
}

func (c *Client) fetchCharacterByID(characterID string) (Character, error) {
	url := baseURL + "/character/" + characterID

	if val, ok := c.cache.Get(url); ok {
		charResp := CharacterResponse{}
		err := json.Unmarshal(val, &charResp)
		if err != nil {
			return Character{}, err
		}
		if len(charResp.Docs) > 0 {
			return charResp.Docs[0], nil
		}
		return Character{}, errors.New("Character not found in cached data")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Character{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Character{}, err
	}
	defer resp.Body.Close()

	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return Character{}, err
	}

	if len(dat) == 0 {
		return Character{}, errors.New("received empty response from API")
	}

	charResp := CharacterResponse{}
	err = json.Unmarshal(dat, &charResp)
	if err != nil {
		return Character{}, err
	}

	if len(charResp.Docs) > 0 {
		return charResp.Docs[0], nil
	}

	return Character{}, errors.New("Book not found in API response")
}
