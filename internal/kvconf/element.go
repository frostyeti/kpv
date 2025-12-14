package kvconf

const (
	EMPTY             = iota
	COMMENT           = 1
	SINGLE_LINE_VALUE = 2
	MULTI_LINE_VALUE  = 3
)

type ConfigElement struct {
	Key   string
	Type  int
	Value string
}

func (c *ConfigElement) IsEmpty() bool {
	return c.Type == EMPTY
}

func (c *ConfigElement) IsComment() bool {
	return c.Type == COMMENT
}

func (c *ConfigElement) IsSingleLineValue() bool {
	return c.Type == SINGLE_LINE_VALUE
}

func (c *ConfigElement) IsMultiLineValue() bool {
	return c.Type == MULTI_LINE_VALUE
}
