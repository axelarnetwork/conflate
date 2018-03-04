package conflate

import (
	"errors"
	"github.com/stretchr/testify/assert"
	pkgurl "net/url"
	"testing"
)

func testFiledataNew(t *testing.T, data []byte, path string) (filedata, error) {
	url, err := pkgurl.Parse(path)
	assert.Nil(t, err)
	return newFiledata(data, *url)
}

func testFiledataNewAssert(t *testing.T, data []byte, path string) filedata {
	fd, err := testFiledataNew(t, data, path)
	assert.Nil(t, err)
	return fd
}

func TestFiledata_WrapErrorNil(t *testing.T) {
	fd, err := testFiledataNew(t, testMarshalJSON, "myurl")
	assert.Nil(t, err)
	err = fd.wrapError(nil)
	assert.Nil(t, err)
}

func TestFiledata_WrapError(t *testing.T) {
	fd, err := testFiledataNew(t, testMarshalJSON, "myurl")
	assert.Nil(t, err)
	err = errors.New("My Error")
	err = fd.wrapError(err)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "My Error")
	assert.Contains(t, err.Error(), "Error processing myurl")
}

func TestFiledata_WrapErrorBlank(t *testing.T) {
	fd, err := testFiledataNew(t, testMarshalJSON, "")
	assert.Nil(t, err)
	err1 := errors.New("My Error")
	err2 := fd.wrapError(err1)
	assert.NotNil(t, err2)
	assert.Equal(t, err1, err2)
}

func TestFiledata_JSONAsAny(t *testing.T) {
	fd, err := testFiledataNew(t, testMarshalJSON, "file")
	assert.Nil(t, err)
	assert.Equal(t, fd.obj, testMarshalData)
}

func TestFiledata_JSONAsUnknown(t *testing.T) {
	fd, err := testFiledataNew(t, testMarshalJSON, "file.unknown")
	assert.Nil(t, err)
	assert.Equal(t, fd.obj, testMarshalData)
}

func TestFiledata_JSONAsJSON(t *testing.T) {
	fd, err := testFiledataNew(t, testMarshalJSON, "file.json")
	assert.Nil(t, err)
	assert.Equal(t, fd.obj, testMarshalData)
}

func TestFiledata_JSONAsJSN(t *testing.T) {
	fd, err := testFiledataNew(t, testMarshalJSON, "file.jsn")
	assert.Nil(t, err)
	assert.Equal(t, fd.obj, testMarshalData)
}

func TestFiledata_JSONAsTOML(t *testing.T) {
	fd, err := testFiledataNew(t, testMarshalJSON, "file.toml")
	assert.NotNil(t, err)
	assert.Nil(t, fd.obj)
}

func TestFiledata_YAMLAsAny(t *testing.T) {
	fd, err := testFiledataNew(t, testMarshalYAML, "file")
	assert.Nil(t, err)
	assert.Equal(t, fd.obj, testMarshalData)
}

func TestFiledata_YAMLAsUnknown(t *testing.T) {
	fd, err := testFiledataNew(t, testMarshalYAML, "file.unknown")
	assert.Nil(t, err)
	assert.Equal(t, fd.obj, testMarshalData)
}

func TestFiledata_YAMLAsYAML(t *testing.T) {
	fd, err := testFiledataNew(t, testMarshalYAML, "file.yaml")
	assert.Nil(t, err)
	assert.Equal(t, fd.obj, testMarshalData)
}

func TestFiledata_YAMLAsYML(t *testing.T) {
	fd, err := testFiledataNew(t, testMarshalYAML, "file.yml")
	assert.Nil(t, err)
	assert.Equal(t, fd.obj, testMarshalData)
}

func TestFiledata_YAMLAsTOML(t *testing.T) {
	fd, err := testFiledataNew(t, testMarshalYAML, "file.toml")
	assert.NotNil(t, err)
	assert.Nil(t, fd.obj)
}

func TestFiledata_TOMLAsAny(t *testing.T) {
	fd, err := testFiledataNew(t, testMarshalTOML, "file")
	assert.Nil(t, err)
	assert.Equal(t, fd.obj, testMarshalData)
}

func TestFiledata_TOMLAsUnknown(t *testing.T) {
	fd, err := testFiledataNew(t, testMarshalTOML, "file.unknown")
	assert.Nil(t, err)
	assert.Equal(t, fd.obj, testMarshalData)
}

func TestFiledata_TOMLAsTOML(t *testing.T) {
	fd, err := testFiledataNew(t, testMarshalTOML, "file.toml")
	assert.Nil(t, err)
	assert.Equal(t, fd.obj, testMarshalData)
}

func TestFiledata_TOMLAsTML(t *testing.T) {
	fd, err := testFiledataNew(t, testMarshalTOML, "file.tml")
	assert.Nil(t, err)
	assert.Equal(t, fd.obj, testMarshalData)
}

func TestFiledata_TOMLAsJSON(t *testing.T) {
	fd, err := testFiledataNew(t, testMarshalTOML, "file.json")
	assert.NotNil(t, err)
	assert.Nil(t, fd.obj)
}

func TestFiledata_NoIncludes(t *testing.T) {
	fd, err := testLoader.wrapFiledata([]byte(`{"x": 1}`))
	assert.Nil(t, err)
	assert.Nil(t, fd.obj["includes"])
	assert.Equal(t, fd.obj, map[string]interface{}{"x": 1.0})
}

func TestFiledata_BlankIncludes(t *testing.T) {
	fd, err := testLoader.wrapFiledata([]byte(`{"includes":[], "x": 1}`))
	assert.Nil(t, err)
	assert.Nil(t, fd.obj["includes"])
	assert.Equal(t, fd.obj, map[string]interface{}{"x": 1.0})
}

func TestFiledata_NullIncludes(t *testing.T) {
	fd, err := testLoader.wrapFiledata([]byte(`{"includes":null, "x": 1}`))
	assert.Nil(t, err)
	assert.Nil(t, fd.obj["includes"])
	assert.Equal(t, fd.obj, map[string]interface{}{"x": 1.0})
}

func TestFiledata_Includes(t *testing.T) {
	fd, err := testLoader.wrapFiledata([]byte(`{"includes":["test1", "test2"], "x": 1}`))
	assert.Nil(t, err)
	assert.Equal(t, fd.includes, []string{"test1", "test2"})
	assert.Nil(t, fd.obj["includes"])
	assert.Equal(t, fd.obj, map[string]interface{}{"x": 1.0})
}

func TestFiledata_IncludesError(t *testing.T) {
	_, err := testLoader.wrapFiledata([]byte(`{"includes": "not array"}`))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Could not extract includes")
}

func TestFiledatas_Unmarshal(t *testing.T) {
	fds := filedatas{
		testFiledataNewAssert(t, testMarshalJSON, "file.json"),
		testFiledataNewAssert(t, testMarshalYAML, "file.yaml"),
		testFiledataNewAssert(t, testMarshalTOML, "file.toml"),
	}
	assert.Equal(t, fds.objs(), []interface{}{testMarshalData, testMarshalData, testMarshalData})
}
