package userinfo

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kyverno/kyverno/cmd/cli/kubectl-kyverno/apis/v1alpha1"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

// TestCommand verifies default command execution and default manifest structure.
func TestCommand(t *testing.T) {
	cmd := Command()
	assert.NotNil(t, cmd)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	err := cmd.Execute()
	assert.NoError(t, err)

	var info v1alpha1.UserInfo
	err = yaml.UnmarshalStrict(buf.Bytes(), &info)
	assert.NoError(t, err)
	assert.Equal(t, "cli.kyverno.io/v1alpha1", info.APIVersion)
	assert.Equal(t, "UserInfo", info.Kind)
	assert.Equal(t, "user-info", info.Name)
}

// TestCommandWithCustomFlags verifies that all flags are correctly set and serialized.
func TestCommandWithCustomFlags(t *testing.T) {
	cmd := Command()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{
		"-n", "custom-admin",
		"-u", "alice@example.com",
		"-r", "editor,admin",
		"-c", "cluster-viewer",
		"-g", "devs,ops",
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	var info v1alpha1.UserInfo
	err = yaml.UnmarshalStrict(buf.Bytes(), &info)
	assert.NoError(t, err)
	assert.Equal(t, "cli.kyverno.io/v1alpha1", info.APIVersion)
	assert.Equal(t, "UserInfo", info.Kind)
	assert.Equal(t, "custom-admin", info.Name)
	assert.Equal(t, "alice@example.com", info.RequestInfo.AdmissionUserInfo.Username)
	assert.Equal(t, []string{"editor", "admin"}, info.RequestInfo.Roles)
	assert.Equal(t, []string{"cluster-viewer"}, info.RequestInfo.ClusterRoles)
	assert.Equal(t, []string{"devs", "ops"}, info.RequestInfo.AdmissionUserInfo.Groups)
}

// TestCommandOutputToFile verifies writing the generated manifest to a file.
func TestCommandOutputToFile(t *testing.T) {
	tempDir := t.TempDir()
	outFile := filepath.Join(tempDir, "user-info.yaml")

	cmd := Command()
	cmd.SetArgs([]string{
		"-o", outFile,
		"-n", "file-user",
		"-u", "bob",
		"-g", "engineers",
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	content, err := os.ReadFile(outFile)
	assert.NoError(t, err)

	var info v1alpha1.UserInfo
	err = yaml.UnmarshalStrict(content, &info)
	assert.NoError(t, err)
	assert.Equal(t, "cli.kyverno.io/v1alpha1", info.APIVersion)
	assert.Equal(t, "UserInfo", info.Kind)
	assert.Equal(t, "file-user", info.Name)
	assert.Equal(t, "bob", info.RequestInfo.AdmissionUserInfo.Username)
	assert.Equal(t, []string{"engineers"}, info.RequestInfo.AdmissionUserInfo.Groups)
}

// TestCommandErrorDoesNotTruncateFile verifies that validation failures do not truncate pre-existing files.
func TestCommandErrorDoesNotTruncateFile(t *testing.T) {
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "existing.yaml")
	originalContent := []byte("original: content\ndo: not truncate\n")
	err := os.WriteFile(existingFile, originalContent, 0644)
	assert.NoError(t, err)

	cmd := Command()
	cmd.SetArgs([]string{
		"-o", existingFile,
		"-n", "  ",
	})
	err = cmd.Execute()
	assert.Error(t, err)

	content, err := os.ReadFile(existingFile)
	assert.NoError(t, err)
	assert.Equal(t, originalContent, content)
}

// TestCommandValidationErrors verifies validation error messages on invalid input.
func TestCommandValidationErrors(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		errorMsg string
	}{
		{
			name:     "EmptyName",
			args:     []string{"-n", ""},
			errorMsg: "name must not be empty",
		},
		{
			name:     "WhitespaceName",
			args:     []string{"-n", "   "},
			errorMsg: "name must not be empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := Command()
			cmd.SetArgs(tc.args)
			err := cmd.Execute()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.errorMsg)
		})
	}
}

// TestCommandWithArgs verifies that positional arguments return an error.
func TestCommandWithArgs(t *testing.T) {
	cmd := Command()
	assert.NotNil(t, cmd)
	cmd.SetArgs([]string{"foo"})
	err := cmd.Execute()
	assert.Error(t, err)
}

// TestCommandWithInvalidArg verifies error messages when invalid arguments are provided.
func TestCommandWithInvalidArg(t *testing.T) {
	cmd := Command()
	assert.NotNil(t, cmd)
	b := bytes.NewBufferString("")
	cmd.SetErr(b)
	cmd.SetArgs([]string{"foo"})
	err := cmd.Execute()
	assert.Error(t, err)
	out, err := io.ReadAll(b)
	assert.NoError(t, err)
	expected := `Error: unknown command "foo" for "user-info"`
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(string(out)))
}

// TestCommandWithInvalidFlag verifies error messages when unknown flags are provided.
func TestCommandWithInvalidFlag(t *testing.T) {
	cmd := Command()
	assert.NotNil(t, cmd)
	b := bytes.NewBufferString("")
	cmd.SetErr(b)
	cmd.SetArgs([]string{"--xxx"})
	err := cmd.Execute()
	assert.Error(t, err)
	out, err := io.ReadAll(b)
	assert.NoError(t, err)
	expected := `Error: unknown flag: --xxx`
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(string(out)))
}

// TestCommandHelp verifies that the help flag outputs command usage.
func TestCommandHelp(t *testing.T) {
	cmd := Command()
	assert.NotNil(t, cmd)
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	assert.NoError(t, err)
	out, err := io.ReadAll(b)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(out), cmd.Long))
}
