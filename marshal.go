package conflate

import (
	"bytes"
	"encoding/json"

	"github.com/BurntSushi/toml"
	"github.com/ghodss/yaml"
)

func jsonMarshalAll(data ...interface{}) ([][]byte, error) {
	var outs [][]byte
	for i, datum := range data {
		out, err := jsonMarshal(datum)
		if err != nil {
			return nil, wrapError(err, "Could not marshal data %v", i)
		}
		outs = append(outs, out)
	}
	return outs, nil
}

func jsonMarshalUnmarshal(in interface{}, out interface{}) error {
	data, err := jsonMarshal(in)
	if err != nil {
		return err
	}
	return JSONUnmarshal(data, out)
}

func jsonPostUnmarshalConvertNumber(raw interface{}) interface{} {
	switch raw.(type) {
	case *interface{}:
		rawInterface := *raw.(*interface{})
		return jsonPostUnmarshalConvertNumber(rawInterface)
	case *map[string]interface{}:
		rawMap := *raw.(*map[string]interface{})
		return jsonPostUnmarshalConvertNumber(rawMap)
	case *[]interface{}:
		rawSlice := *raw.(*[]interface{})
		return jsonPostUnmarshalConvertNumber(rawSlice)
	case map[string]interface{}:
		rawMap := raw.(map[string]interface{})
		for key, val := range rawMap {
			rawMap[key] = jsonPostUnmarshalConvertNumber(val)
		}
		return rawMap
	case []interface{}:
		rawSlice := raw.([]interface{})
		for index, val := range rawSlice {
			rawSlice[index] = jsonPostUnmarshalConvertNumber(val)
		}
		return rawSlice
	case json.Number:
		rawJSONNumber := raw.(json.Number)
		if intValue, err := rawJSONNumber.Int64(); err == nil {
			return intValue
		} else if floatValue, err := rawJSONNumber.Float64(); err == nil {
			return floatValue
		}
	}

	return raw
}

// JSONUnmarshal unmarshals the data as JSON
func JSONUnmarshal(data []byte, out interface{}) error {
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	decoder.UseNumber()

	err := decoder.Decode(out)
	if err != nil {
		return wrapError(err, "The data could not be unmarshalled as json")
	}

	out = jsonPostUnmarshalConvertNumber(out)

	return nil
}

// YAMLUnmarshal unmarshals the data as YAML
func YAMLUnmarshal(data []byte, out interface{}) error {
	err := yaml.Unmarshal(data, out)
	if err != nil {
		return wrapError(err, "The data could not be unmarshalled as yaml")
	}
	return nil
}

// TOMLUnmarshal unmarshals the data as TOML
func TOMLUnmarshal(data []byte, out interface{}) error {
	err := toml.Unmarshal(data, out)
	if err != nil {
		return wrapError(err, "The data could not be unmarshalled as toml")
	}
	return nil
}

func jsonMarshal(data interface{}) ([]byte, error) {
	buffer := bytes.Buffer{}
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(data)
	if err != nil {
		return nil, wrapError(err, "The data could not be marshalled to json")
	}
	return buffer.Bytes(), nil
}

func yamlMarshal(in interface{}) ([]byte, error) {
	data, err := yaml.Marshal(in)
	if err != nil {
		return nil, wrapError(err, "The data could not be marshalled to yaml")
	}
	return data, nil
}

func tomlMarshal(in interface{}) (out []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = wrapError(makeError("%v", r), "The data could not be marshalled to toml")
		}
	}()
	buf := bytes.Buffer{}
	enc := toml.NewEncoder(&buf)

	err = enc.Encode(in)
	if err != nil {
		return nil, wrapError(err, "The data could not be marshalled to toml")
	}
	out = buf.Bytes()
	return out, nil
}
