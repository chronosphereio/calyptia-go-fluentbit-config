package property

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

// MarshalJSON implements json.Marshaler interface
// to marshall a sorted list of properties into an object.
func (pp Properties) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	if err := buf.WriteByte('{'); err != nil {
		return nil, err
	}

	enc := json.NewEncoder(&buf)
	for i, p := range pp {
		if i != 0 {
			if err := buf.WriteByte(','); err != nil {
				return nil, err
			}
		}

		if err := enc.Encode(p.Key); err != nil {
			return nil, err
		}

		if err := buf.WriteByte(':'); err != nil {
			return nil, err
		}

		if err := enc.Encode(p.Value); err != nil {
			return nil, err
		}
	}

	if err := buf.WriteByte('}'); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalJSON implements json.Unmarshaler interface
// to unmarshal an object into a sorted list of properties.
func (pp *Properties) UnmarshalJSON(data []byte) error {
	var m map[string]any
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	t, err := dec.Token()
	if err != nil {
		return err
	}

	if t != json.Delim('{') {
		return errors.New("expected start of object")
	}

	for {
		t, err := dec.Token()
		if err != nil {
			return err
		}

		if t == json.Delim('}') {
			break
		}

		key, ok := t.(string)
		if !ok {
			return errors.New("expected object key to be a string")
		}

		*pp = append(*pp, Property{
			Key:   key,
			Value: m[key],
		})

		// ignored value
		if err := skipJSONValue(dec); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}
	}

	return nil
}

var errJSONEnd = errors.New("invalid end of json array or object")

func skipJSONValue(dec *json.Decoder) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}

	switch t {
	case json.Delim('['), json.Delim('{'):
		for {
			if err := skipJSONValue(dec); err != nil {
				if errors.Is(err, errJSONEnd) {
					break
				}
				return err
			}
		}
	case json.Delim(']'), json.Delim('}'):
		return errJSONEnd
	}

	return nil
}
