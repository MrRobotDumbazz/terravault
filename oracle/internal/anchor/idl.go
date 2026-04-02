package anchor

import (
	"encoding/json"
	"fmt"
	"os"
)

// IDL represents a minimal Anchor IDL structure needed for instruction building.
type IDL struct {
	Version      string        `json:"version"`
	Name         string        `json:"name"`
	Instructions []IDLInstruction `json:"instructions"`
	Accounts     []IDLAccount  `json:"accounts"`
	Types        []IDLType     `json:"types"`
	Errors       []IDLError    `json:"errors"`
}

type IDLInstruction struct {
	Name     string       `json:"name"`
	Accounts []IDLAccount `json:"accounts"`
	Args     []IDLArg     `json:"args"`
}

type IDLAccount struct {
	Name        string `json:"name"`
	IsMut       bool   `json:"isMut"`
	IsSigner    bool   `json:"isSigner"`
	IsOptional  bool   `json:"isOptional,omitempty"`
	Docs        []string `json:"docs,omitempty"`
}

type IDLArg struct {
	Name string      `json:"name"`
	Type interface{} `json:"type"`
}

type IDLType struct {
	Name string      `json:"name"`
	Type interface{} `json:"type"`
}

type IDLError struct {
	Code    int    `json:"code"`
	Name    string `json:"name"`
	Msg     string `json:"msg"`
}

// LoadIDL reads and parses an Anchor IDL JSON file.
func LoadIDL(path string) (*IDL, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading IDL file %s: %w", path, err)
	}
	var idl IDL
	if err := json.Unmarshal(data, &idl); err != nil {
		return nil, fmt.Errorf("parsing IDL: %w", err)
	}
	return &idl, nil
}

// GetInstruction finds an instruction definition by name.
func (idl *IDL) GetInstruction(name string) (*IDLInstruction, error) {
	for i, ix := range idl.Instructions {
		if ix.Name == name {
			return &idl.Instructions[i], nil
		}
	}
	return nil, fmt.Errorf("instruction %q not found in IDL", name)
}
