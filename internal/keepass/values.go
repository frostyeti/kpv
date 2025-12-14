package keepass

import (
	"github.com/tobischo/gokeepasslib/v3"
	"github.com/tobischo/gokeepasslib/v3/wrappers"
)

type KeyValue struct {
	data *gokeepasslib.ValueData
}

func NewKeyValue(key string, value string) *KeyValue {
	return &KeyValue{
		data: &gokeepasslib.ValueData{
			Key:   key,
			Value: gokeepasslib.V{Content: value},
		},
	}
}

func NewKeyProtectedValue(key string, value string) *KeyValue {
	return &KeyValue{
		data: &gokeepasslib.ValueData{
			Key:   key,
			Value: gokeepasslib.V{Content: value, Protected: wrappers.NewBoolWrapper(true)},
		},
	}
}

func (kv *KeyValue) Key() string {
	return kv.data.Key
}

func (kv *KeyValue) Value() string {
	return kv.data.Value.Content
}

func (kv *KeyValue) SetValue(value string) {
	kv.data.Value = gokeepasslib.V{Content: value}
}

func (kv *KeyValue) SetProtectedValue(value string) {
	kv.data.Value = gokeepasslib.V{Content: value, Protected: wrappers.NewBoolWrapper(true)}
}

func (kv *KeyValue) ToValueData() *gokeepasslib.ValueData {
	return kv.data
}
