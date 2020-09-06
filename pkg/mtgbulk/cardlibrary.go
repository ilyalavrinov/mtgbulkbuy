package mtgbulk

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Library interface {
	CardAliases(string) (map[string]bool, error)
}

type InMemoryLibrary struct {
	cardIDtoNames map[string]map[string]bool
	cardNameToID  map[string]string
}

type Card struct {
	ID        string `json:"id"`
	OracleID  string `json:"oracle_id"`
	Name      string `json:"name"`
	LocalName string `json:"printed_name"`
	Lang      string `json:"lang"`
	URI       string `json:"uri"`
}

func NewInMemoryLibrary(dumpPath string) (Library, error) {
	f, err := os.Open(dumpPath)
	if err != nil {
		return nil, fmt.Errorf("Cannot open file with dump: %w", err)
	}
	logger.Debugw("decoding dump",
		"path", dumpPath)
	dec := json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		return nil, fmt.Errorf("Cannot tokenize file with dump: %w", err)
	}

	lib := &InMemoryLibrary{
		cardIDtoNames: make(map[string]map[string]bool),
		cardNameToID:  make(map[string]string),
	}

	for dec.More() {
		var c Card
		err := dec.Decode(&c)
		if err != nil {
			return nil, fmt.Errorf("Cannot decode file with dump: %w", err)
		}
		if c.LocalName == "" {
			c.LocalName = c.Name
		}
		c.LocalName = strings.ToLower(c.LocalName)
		names, found := lib.cardIDtoNames[c.OracleID]
		if !found {
			lib.cardIDtoNames[c.OracleID] = make(map[string]bool)
			names = lib.cardIDtoNames[c.OracleID]
		}
		names[c.LocalName] = true
		lib.cardNameToID[c.LocalName] = c.OracleID
		/*
				names := []string{c.Name, c.LocalName}
				if strings.Contains(c.Name, "//") {
					names = []string{}
					names = append(names, strings.Split(c.Name, " // ")...)
					names = append(names, strings.Split(c.LocalName, " // ")...)
				}
				for _, n := range names {
					n := strings.ToLower(n)
					_, found := h.cardsByName[n]
					if found {
						if c.Lang == "en" {
							h.cardsByName[n] = c
						}
					} else {
						h.cardsByName[n] = c
					}
				}
			}
		*/
	}
	_, err = dec.Token()
	if err != nil {
		return nil, fmt.Errorf("Cannot advance to next token at dump: %w", err)
	}

	logger.Debugw("decoding done")
	return lib, nil
}

func (lib *InMemoryLibrary) CardAliases(cardname string) (map[string]bool, error) {
	cardname = strings.ToLower(cardname)
	id, found := lib.cardNameToID[cardname]
	if !found {
		return nil, fmt.Errorf("no Oracle ID for card: %s", cardname)
	}
	return lib.cardIDtoNames[id], nil
}
