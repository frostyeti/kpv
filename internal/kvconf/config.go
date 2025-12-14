package kvconf

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	items          []ConfigElement
	dict           map[string]int
	allowedSymbols []rune
	path           string
}

type ConfigParams struct {
	AllowedSymbols []rune
	Path           string
}
type ConfigOption func(*ConfigParams)

func NewConfig(options ...ConfigOption) *Config {
	params := &ConfigParams{
		AllowedSymbols: []rune{'-', '_', '.', ':'},
	}
	for _, option := range options {
		option(params)
	}

	return &Config{
		dict:           make(map[string]int),
		allowedSymbols: params.AllowedSymbols,
		path:           params.Path,
	}
}

func (c *Config) Keys() []string {
	keys := make([]string, 0, len(c.dict))
	for key := range c.dict {
		keys = append(keys, key)
	}
	return keys
}

func (c *Config) Parse(input string) {
	// use buf scanner to read line by line
	scanner := bufio.NewScanner(strings.NewReader(input))
	inMultiLine := false
	currentKey := ""
	multiLineValue := strings.Builder{}
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if inMultiLine {
			if trimmed == "EOF" {
				c.Set(currentKey, multiLineValue.String())
				inMultiLine = false
				currentKey = ""
				multiLineValue.Reset()
			} else {
				if multiLineValue.Len() > 0 {
					multiLineValue.WriteString("\n")
				}
				multiLineValue.WriteString(line)
			}
			continue
		}

		if trimmed == "" {
			c.AddLine()
		} else if strings.HasPrefix(trimmed, "#") {
			comment := strings.TrimPrefix(trimmed, "#")
			comment = strings.TrimSpace(comment)
			c.AddComment(comment)
		} else if eqIndex := strings.Index(line, "="); eqIndex != -1 {
			key := strings.TrimSpace(line[:eqIndex])
			value := strings.TrimSpace(line[eqIndex+1:])

			if strings.HasSuffix(value, "=EOF") {
				value = strings.TrimSuffix(value, "=EOF")
				value = strings.TrimSpace(value)
				c.Set(key, value)
			} else if value == "EOF" {
				inMultiLine = true
				currentKey = key
			} else {
				c.Set(key, value)
			}
		}
	}
}

func (c *Config) Load(path string) error {
	if c == nil {
		c = NewConfig()
	}

	c.path = path

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	c.Parse(string(data))
	return nil
}

func (c *Config) Save() error {
	if c.path == "" {
		return nil
	}
	data := c.String()
	return os.WriteFile(c.path, []byte(data), 0644)
}

func (c *Config) SaveAs(path string) error {
	data := c.String()
	return os.WriteFile(path, []byte(data), 0644)
}

func (c *Config) Set(key string, value string) {
	kind := SINGLE_LINE_VALUE
	if strings.ContainsAny(value, "\n\r") {
		kind = MULTI_LINE_VALUE
	}

	if index, exists := c.dict[key]; exists {
		c.items[index].Type = kind
		c.items[index].Value = value
	} else {
		c.items = append(c.items, ConfigElement{
			Key:   key,
			Type:  kind,
			Value: value,
		})
		c.dict[key] = len(c.items) - 1
	}
}

func (c *Config) Get(key string) (string, bool) {
	if index, exists := c.dict[key]; exists {
		return c.items[index].Value, true
	}
	return "", false
}

func (c *Config) AddLine() {
	c.items = append(c.items, ConfigElement{
		Type: EMPTY,
	})
}

func (c *Config) AddComment(comment string) {
	c.items = append(c.items, ConfigElement{
		Type:  COMMENT,
		Value: comment,
	})
}

func (c *Config) Add(element ConfigElement) {
	c.items = append(c.items, element)
}

func (c *Config) AddValue(key string, value string) {
	kind := SINGLE_LINE_VALUE
	if strings.ContainsAny(value, "\n\r") {
		kind = MULTI_LINE_VALUE
	}

	c.items = append(c.items, ConfigElement{
		Key:   key,
		Type:  kind,
		Value: value,
	})
	c.dict[key] = len(c.items) - 1
}

func (c *Config) Remove(key string) {
	if index, exists := c.dict[key]; exists {
		// Remove from items slice
		c.items = append(c.items[:index], c.items[index+1:]...)
		// Remove from dict
		delete(c.dict, key)
		// Update indices in dict
		for k, v := range c.dict {
			if v > index {
				c.dict[k] = v - 1
			}
		}
	}
}

func (c *Config) RemoveAt(index int) {
	if index >= 0 && index < len(c.items) {
		key := c.items[index].Key
		// Remove from items slice
		c.items = append(c.items[:index], c.items[index+1:]...)

		if key != "" {
			// Remove from dict
			delete(c.dict, key)
		}

		// Update indices in dict
		for k, v := range c.dict {
			if v > index {
				c.dict[k] = v - 1
			}
		}
	}
}

func (c *Config) String() string {
	var builder strings.Builder
	for _, item := range c.items {
		switch item.Type {
		case EMPTY:
			builder.WriteString("\n")
		case COMMENT:
			builder.WriteString("# " + item.Value + "\n")
		case SINGLE_LINE_VALUE:
			builder.WriteString(item.Key + "=" + item.Value + "\n")
		case MULTI_LINE_VALUE:
			builder.WriteString(item.Key + "=EOF\n" + item.Value + "\nEOF\n")
		}
	}
	return builder.String()
}
