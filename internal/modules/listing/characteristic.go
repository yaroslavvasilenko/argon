package listing

import "encoding/json"

// Characteristic представляет собой карту характеристик, где ключ - это роль, а значение - это значение характеристики
type Characteristic map[string]interface{}

// CharacteristicItem представляет отдельную характеристику с ролью и значением
type CharacteristicItem struct {
	Role  string      `json:"role"`
	Value interface{} `json:"value"`
}

// MarshalJSON реализует интерфейс json.Marshaler для типа Characteristic
func (c Characteristic) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte("null"), nil
	}

	charItems := make([]CharacteristicItem, 0, len(c))
	for role, value := range c {
		charItems = append(charItems, CharacteristicItem{
			Role:  role,
			Value: value,
		})
	}

	return json.Marshal(charItems)
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler для типа Characteristic
func (c *Characteristic) UnmarshalJSON(data []byte) error {
	var charItems []CharacteristicItem
	if err := json.Unmarshal(data, &charItems); err != nil {
		return err
	}

	if *c == nil {
		*c = make(Characteristic)
	}

	for _, item := range charItems {
		(*c)[item.Role] = item.Value
	}

	return nil
}
