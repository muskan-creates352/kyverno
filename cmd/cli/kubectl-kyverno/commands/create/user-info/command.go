package userinfo

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	kyvernov2 "github.com/kyverno/kyverno/api/kyverno/v2"
	"github.com/kyverno/kyverno/cmd/cli/kubectl-kyverno/command"
	"github.com/kyverno/kyverno/cmd/cli/kubectl-kyverno/commands/create/templates"
	"github.com/spf13/cobra"
	authenticationv1 "k8s.io/api/authentication/v1"
)

type options struct {
	Name string
	kyvernov2.RequestInfo
}

// Command returns a cobra.Command for creating a Kyverno user-info file.
func Command() *cobra.Command {
	var path string
	var name string
	var username string
	var roles, clusterRoles, groups []string
	cmd := &cobra.Command{
		Use:          "user-info",
		Short:        command.FormatDescription(true, websiteUrl, false, description...),
		Long:         command.FormatDescription(false, websiteUrl, false, description...),
		Example:      command.FormatExamples(examples...),
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("name must not be empty")
			}
			tmpl, err := template.New("userinfo").Parse(templates.UserInfoTemplate)
			if err != nil {
				return err
			}
			values := options{
				Name: strings.TrimSpace(name),
				RequestInfo: kyvernov2.RequestInfo{
					Roles:        roles,
					ClusterRoles: clusterRoles,
					AdmissionUserInfo: authenticationv1.UserInfo{
						Username: username,
						Groups:   groups,
					},
				},
			}
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, values); err != nil {
				return err
			}
			if path != "" {
				return writeOutputFile(path, buf.Bytes())
			}
			_, err = cmd.OutOrStdout().Write(buf.Bytes())
			return err
		},
	}
	cmd.Flags().StringVarP(&path, "output", "o", "", "Output path (uses standard console output if not set)")
	cmd.Flags().StringVarP(&name, "name", "n", "user-info", "User info resource name")
	cmd.Flags().StringVarP(&username, "username", "u", "", "User name")
	cmd.Flags().StringSliceVarP(&roles, "role", "r", nil, "Role")
	cmd.Flags().StringSliceVarP(&clusterRoles, "cluster-role", "c", nil, "Cluster role")
	cmd.Flags().StringSliceVarP(&groups, "group", "g", nil, "Group")
	return cmd
}

// writeOutputFile writes data safely to destination path using a temporary file in the same directory,
// preserving any pre-existing permissions and symlink targets.
func writeOutputFile(path string, data []byte) error {
	targetPath := filepath.Clean(path)
	if fi, err := os.Lstat(targetPath); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		resolved, err := filepath.EvalSymlinks(targetPath)
		if err != nil {
			return err
		}
		targetPath = resolved
	}

	dir := filepath.Dir(targetPath)
	tmpFile, err := os.CreateTemp(dir, "user-info-*.tmp") // #nosec G304
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpName) // #nosec G304
	}()

	if fi, err := os.Stat(targetPath); err == nil {
		if err := tmpFile.Chmod(fi.Mode().Perm()); err != nil {
			_ = tmpFile.Close()
			return err
		}
	}
	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, targetPath) // #nosec G304
}
