package conflate

import (
	gocontext "context"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestData struct {
	Includes      []string
	ParentOnly    string `json:"parent_only"`
	ChildOnly     string `json:"child_only"`
	SiblingOnly   string `json:"sibling_only"`
	ParentSibling string `json:"parent_sibling"`
	ParentChild   string `json:"parent_child"`
	SiblingChild  string `json:"sibling_child"`
	All           string `json:"all"`
}

func TestFromFiles(t *testing.T) {
	c, err := FromFiles("testdata/valid_parent.json")
	assert.Nil(t, err)
	assert.NotNil(t, c)
	var testData TestData
	err = c.Unmarshal(&testData)
	assert.Nil(t, err)
	assert.Equal(t, "child", testData.ChildOnly)
	assert.Equal(t, "parent", testData.ParentOnly)
	assert.Equal(t, "sibling", testData.SiblingOnly)
	assert.Equal(t, "parent", testData.ParentChild)
	assert.Equal(t, "parent", testData.ParentSibling)
	assert.Equal(t, "sibling", testData.SiblingChild)
	assert.Equal(t, "parent", testData.All)
}

func TestFromFiles_IncludesRemoved(t *testing.T) {
	c, err := FromFiles("testdata/valid_parent.json")
	assert.Nil(t, err)
	assert.NotNil(t, c)
	var testData map[string]interface{}
	err = c.Unmarshal(&testData)
	assert.Nil(t, err)
	assert.Nil(t, testData[Includes])
}

func TestAddData_Expand(t *testing.T) {
	c := New()
	c.Expand(true)
	os.Setenv("X", "123")
	os.Setenv("Y", "str")
	inJSON := []byte(`{ "x": $X, "y": "$Y", "z": "$Z"}`)
	err := c.AddData(inJSON)
	assert.Nil(t, err)
	outJSON, err := c.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, `{
  "x": 123,
  "y": "str",
  "z": "$Z"
}
`, string(outJSON))
}

func TestAddData_NoExpand(t *testing.T) {
	c := New()
	c.Expand(false)
	os.Setenv("X", "123")
	os.Setenv("Y", "str")
	inJSON := []byte(`{ "x": "$X" }`)
	err := c.AddData(inJSON)
	assert.Nil(t, err)
	outJSON, err := c.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, `{
  "x": "$X"
}
`, string(outJSON))
}

func TestFromFilesRemote(t *testing.T) {

	// we simulate that access tokens are passed to relative paths in 'includes' list
	dummyQueryString := "accessToken=123"

	fileServer := http.FileServer(http.Dir("./testdata"))

	var wg sync.WaitGroup
	wg.Add(1)
	server := &http.Server{
		Addr: ":9999",
		Handler: http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.RawQuery != dummyQueryString {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				fileServer.ServeHTTP(w, r)
			}),
	}
	go func() {
		defer wg.Done()
		server.ListenAndServe()
	}()
	defer func() {
		server.Shutdown(gocontext.Background())
	}()

	testWaitForURL(t, "http://0.0.0.0:9999")

	c, err := FromFiles("http://0.0.0.0:9999/valid_parent.json?" + dummyQueryString)
	assert.Nil(t, err)
	assert.NotNil(t, c)
	var testData TestData
	c.Unmarshal(&testData)
	assert.Equal(t, "child", testData.ChildOnly)
	assert.Equal(t, "parent", testData.ParentOnly)
	assert.Equal(t, "sibling", testData.SiblingOnly)
	assert.Equal(t, "parent", testData.ParentChild)
	assert.Equal(t, "parent", testData.ParentSibling)
	assert.Equal(t, "sibling", testData.SiblingChild)
	assert.Equal(t, "parent", testData.All)
}

func TestFromURLs(t *testing.T) {
	url, err := toURL(nil, "testdata/valid_parent.json")
	assert.Nil(t, err)
	c, err := FromURLs(url)
	assert.Nil(t, err)
	assert.NotNil(t, c)
}

func TestFromURLs_Error(t *testing.T) {
	url, err := toURL(nil, "missing file")
	assert.Nil(t, err)
	c, err := FromURLs(url)
	assert.NotNil(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "Failed to load url")
}

func TestFromFiles_Error(t *testing.T) {
	c, err := FromFiles("missing file")
	assert.NotNil(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "Failed to load url")
}

func TestFromFiles_WorkingDirError(t *testing.T) {
	oldGetwd := getwd
	getwd = func() (dir string, err error) {
		return "", makeError("No root error")
	}
	defer func() { getwd = oldGetwd }()
	_, err := FromFiles("some file")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No root error")
}

func TestFromFiles_ToUrlsError(t *testing.T) {
	_, err := FromFiles("testdata/bad_url_in_include.json")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Could not parse path")
}

func TestFromFiles_ExpandError(t *testing.T) {
	_, err := FromFiles("testdata/missing_file_in_include.json")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Failed to load url")
}

func TestFromFiles_ValidationNoSchemaError(t *testing.T) {
	c, err := FromFiles("testdata/valid_child.json")
	assert.Nil(t, err)
	err = c.Validate(nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Schema is not set")
}

func TestFromFiles_ValidationError(t *testing.T) {
	c, err := FromFiles("testdata/valid_child.json")
	assert.Nil(t, err)
	s, err := NewSchemaFile("testdata/test.schema.json")
	assert.Nil(t, err)
	err = c.Validate(s)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Schema validation failed")
}

func TestFromFiles_ValidationOk(t *testing.T) {
	c, err := FromFiles("testdata/valid_parent.json")
	assert.Nil(t, err)
	s, err := NewSchemaFile("testdata/test.schema.json")
	assert.Nil(t, err)
	err = c.Validate(s)
	assert.Nil(t, err)
}

func TestFromFiles_ApplyDefaultsNoSchema(t *testing.T) {
	c, err := FromFiles("testdata/valid_parent.json")
	assert.Nil(t, err)
	err = c.ApplyDefaults(nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Schema is not set")
}

func TestFromFiles_ApplyDefaultsError(t *testing.T) {
	c, err := FromFiles("testdata/valid_parent.json")
	assert.Nil(t, err)
	s, err := NewSchemaFile("testdata/test.schema.json")
	assert.Nil(t, err)
	s.s = []interface{}{"not a map"}
	err = c.ApplyDefaults(s)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "The defaults could not be applied")
	assert.Contains(t, err.Error(), "Schema section is not a map")
}

func TestFromFiles_ApplyDefaults(t *testing.T) {
	c, err := FromFiles()
	assert.Nil(t, err)
	s, err := NewSchemaFile("testdata/test.schema.json")
	assert.Nil(t, err)
	err = c.ApplyDefaults(s)
	assert.Nil(t, err)
	testData := TestData{}
	err = c.Unmarshal(&testData)
	assert.Nil(t, err)
	assert.Equal(t, "child", testData.ChildOnly)
	assert.Equal(t, "parent", testData.ParentOnly)
	assert.Equal(t, "sibling", testData.SiblingOnly)
	assert.Equal(t, "parent", testData.ParentChild)
	assert.Equal(t, "parent", testData.ParentSibling)
	assert.Equal(t, "sibling", testData.SiblingChild)
	assert.Equal(t, "parent", testData.All)
}

func TestFromData(t *testing.T) {
	c, err := FromData([]byte(`{"x": 1}`))
	assert.Nil(t, err)
	assert.NotNil(t, c)
}

func TestFromData_Error(t *testing.T) {
	c, err := FromData([]byte("{bad data"))
	assert.NotNil(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "Could not unmarshal data")
}

func TestFromData_MergeToError(t *testing.T) {
	_, err := FromData([]byte(`{ "x": [1]}`), []byte(`{ "x": {"y": 1}}`))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Failed to merge object property")
}

func TestFromGo(t *testing.T) {
	x := struct{ X []int }{X: []int{1}}
	y := struct{ X []int }{X: []int{2}}
	c, err := FromGo(x, y)
	assert.Nil(t, err)
	z := struct{ X []int }{}
	err = c.Unmarshal(&z)
	assert.Nil(t, err)
	assert.Equal(t, z.X, []int{1, 2})
}

func TestFromGo_MarshalError(t *testing.T) {
	_, err := FromGo(make(chan int))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "The data could not be marshalled to json")
}

func TestFromGo_MergeToError(t *testing.T) {
	x := struct{ X []int }{X: []int{1}}
	y := struct{ X map[string]int }{X: map[string]int{"Y": 1}}
	_, err := FromGo(x, y)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Failed to merge object property")
}

func TestConflate_MarshalJSON(t *testing.T) {
	c, err := FromData(testMarshalJSON)
	assert.Nil(t, err)
	data, err := c.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, string(testMarshalJSON), string(data))
}

func TestConflate_MarshalYAML(t *testing.T) {
	c, err := FromData(testMarshalYAML)
	assert.Nil(t, err)
	data, err := c.MarshalYAML()
	assert.Nil(t, err)
	assert.Equal(t, testMarshalYAML, data)
}

func TestConflate_MarshalTOML(t *testing.T) {
	c, err := FromData(testMarshalTOML)
	assert.Nil(t, err)
	data, err := c.MarshalTOML()
	assert.Nil(t, err)
	assert.Equal(t, testMarshalTOML, data)
}

func TestConflate_addDataError(t *testing.T) {
	c := New()
	err := c.AddData([]byte(`{"includes": ["missing"]}`))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Failed to load url")
}

func TestConflate_mergeDataError(t *testing.T) {
	c := New()
	err := c.AddData([]byte(`"x": {}`), []byte(`"x": []`))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Failed to merge")
}

func TestConflate_IncludesWithMapArray(t *testing.T) {
	c, err := FromFiles("testdata/merge_includes_with_map_array/data.json")
	assert.NoError(t, err)
	assert.NotNil(t, c)

	if err != nil {
		return
	}

	var data interface{}
	err = c.Unmarshal(&data)
	assert.NoError(t, err)

	marshalledData, err := c.MarshalTOML()
	assert.NoError(t, err)
	assert.NotEmpty(t, marshalledData)

	expectedData, err := ioutil.ReadFile("testdata/merge_includes_with_map_array/expected.toml")
	assert.NoError(t, err)
	assert.Equal(t, string(expectedData), string(marshalledData))
}

func TestConflate_JSONIntegerToTOMLInteger(t *testing.T) {
	c, err := FromFiles("testdata/json_to_toml_numbers/data.json")
	assert.NoError(t, err)
	assert.NotNil(t, c)

	if err != nil {
		return
	}

	var data interface{}
	err = c.Unmarshal(&data)
	assert.NoError(t, err)

	marshalledData, err := c.MarshalTOML()
	assert.NoError(t, err)
	assert.NotEmpty(t, marshalledData)

	expectedData, err := ioutil.ReadFile("testdata/json_to_toml_numbers/expected.toml")
	assert.NoError(t, err)
	assert.Equal(t, string(expectedData), string(marshalledData))
}
