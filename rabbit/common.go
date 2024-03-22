package rabbit

import "encoding/json"

func NewJsonSerializer() *JsonSerializer {
	return new(JsonSerializer)
}

type JsonSerializer struct {
}

func (ser *JsonSerializer) Serialize(item any) ([]byte, error) {
	return json.Marshal(item)
}

func (ser *JsonSerializer) Deserialize(payload []byte, item any) error {
	return json.Unmarshal(payload, item)
}
