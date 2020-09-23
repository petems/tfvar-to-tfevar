package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixturePath(t *testing.T, fixture string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("problems recovering caller information")
	}

	return filepath.Join(filepath.Dir(filename), "testdata/", fixture)
}

func loadFixture(t *testing.T, fixture string) string {
	content, err := ioutil.ReadFile(fixturePath(t, fixture))
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}

func TestPlain(t *testing.T) {
	os.Args = strings.Fields("tfvar testdata")

	var actual bytes.Buffer
	cmd, sync := New(&actual, "dev")
	defer sync()

	require.NoError(t, cmd.Execute())
	expected := loadFixture(t, "plain.golden")

	assert.Equal(t, expected, actual.String())
}

func TestWorkspaceOrg(t *testing.T) {
	os.Args = strings.Fields("tfvar testdata --workspace=cool_workspace --org=cool_org")

	var actual bytes.Buffer
	cmd, sync := New(&actual, "dev")
	defer sync()

	require.NoError(t, cmd.Execute())
	expected := loadFixture(t, "org_workspace_arg.golden")

	assert.Equal(t, expected, actual.String())
}

func TestIgnoreDefault(t *testing.T) {
	os.Args = strings.Fields("tfvar testdata --ignore-default")

	var actual bytes.Buffer
	cmd, sync := New(&actual, "dev")
	defer sync()

	require.NoError(t, cmd.Execute())
	expected := loadFixture(t, "ignore_default.golden")

	assert.Equal(t, expected, actual.String())
}

func TestAutoAssign(t *testing.T) {
	os.Args = strings.Fields("tfvar testdata -a")
	os.Setenv("TF_VAR_image_id", "abc123")

	var actual bytes.Buffer
	cmd, sync := New(&actual, "dev")
	defer sync()

	require.NoError(t, cmd.Execute())
	expected := loadFixture(t, "auto_assign.golden")

	assert.Equal(t, expected, actual.String())
}

func TestVar(t *testing.T) {
	os.Args = strings.Fields("tfvar testdata -a --var='image_id=abc123' --var='unknown=xxx'")

	var actual bytes.Buffer
	cmd, sync := New(&actual, "dev")
	defer sync()

	require.NoError(t, cmd.Execute())
	expected := loadFixture(t, "var_args.golden")

	assert.Equal(t, expected, actual.String())
}

func TestVarError(t *testing.T) {
	os.Args = strings.Fields("tfvar testdata -a --var='unknown'")

	var actual bytes.Buffer
	cmd, sync := New(&actual, "dev")
	defer sync()

	assert.Error(t, cmd.Execute())
	assert.Contains(t, actual.String(), `Error: tfvar: bad var string ''unknown''`)
}

func TestVarFile(t *testing.T) {
	os.Args = strings.Fields("tfvar testdata --var-file testdata/my.tfvars")

	var actual bytes.Buffer
	cmd, sync := New(&actual, "dev")
	defer sync()

	require.NoError(t, cmd.Execute())
	expected := loadFixture(t, "var_file_args.golden")

	assert.Equal(t, expected, actual.String())
}

func TestVarFileError(t *testing.T) {
	os.Args = strings.Fields("tfvar testdata --var-file testdata/bad.tfvars")

	var actual bytes.Buffer
	cmd, sync := New(&actual, "dev")
	defer sync()

	assert.Error(t, cmd.Execute())
	assert.Contains(t, actual.String(), `Error: tfvar: failed to parse 'testdata/bad.tfvars'`)
}
