package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/kong/semver/v4"
	rbacv1 "k8s.io/api/rbac/v1"
)

func renderHelperTemplate(semverVersions map[string]semver.Version, templateName, rawTemplate string) ([]byte, error) {
	versions := make(map[string]string, 0)
	for c := range semverVersions {
		versions[c] = convertConstraintName(c)
	}

	tpl, err := template.New(templateName).Funcs(sprig.TxtFuncMap()).Parse(rawTemplate)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err = tpl.Execute(buf, helperTemplateData{
		Versions: versions,
	}); err != nil {
		return nil, err
	}

	return format.Source(buf.Bytes())
}

func renderDoc(filename string) error {
	return os.WriteFile(filename, []byte(docTemplate), 0644)
}

func renderTemplate(clusterRoles []*rbacv1.ClusterRole, constraint string, templateName string, rawTemplate string) ([]byte, error) {
	tpl, err := template.New(templateName).Funcs(sprig.TxtFuncMap()).Parse(rawTemplate)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err = tpl.Execute(buf, templateData{
		Roles:      clusterRoles,
		Version:    convertConstraintName(constraint),
		Constraint: constraint,
	}); err != nil {
		return nil, err
	}

	return format.Source(buf.Bytes())
}

// convertConstraintName is an helper that replaces the basic constraint symbols
// with textual ones, to be used as suffixes on files and functions naming.
// The complete list of constraint symbols can be found here:
// https://github.com/Masterminds/semver#basic-comparisons
func convertConstraintName(constraint string) string {
	constraint = strings.ReplaceAll(constraint, " ", "")
	constraint = strings.ReplaceAll(constraint, ",", "_")
	constraint = strings.ReplaceAll(constraint, ".", "_")
	constraint = strings.ReplaceAll(constraint, "<=", "le")
	constraint = strings.ReplaceAll(constraint, ">=", "ge")
	constraint = strings.ReplaceAll(constraint, "<", "lt")
	constraint = strings.ReplaceAll(constraint, ">", "gt")
	constraint = strings.ReplaceAll(constraint, "!=", "ne")
	constraint = strings.ReplaceAll(constraint, "=", "eq")

	return constraint
}

// filesEqual returns a bool variable to express whether the path of the file
// given as parameter is equal to the buffer parameter.
func filesEqual(path string, buffer []byte) (bool, error) {
	oldBuffer, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return false, err
		}
	}

	return bytes.Equal(buffer, oldBuffer), nil
}

// mkdir creates a folder at the given path if it does not exist yet.
func mkdir(path string) error {
	err := os.Mkdir(path, os.ModePerm)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}
	return nil
}

func rmDirs(paths ...string) error {
	for _, p := range paths {
		if err := os.RemoveAll(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

func updateFile(path string, buffer []byte) error {
	return os.WriteFile(path, buffer, os.ModePerm)
}

func exitOnErr(err error, msg ...string) {
	if err != nil {
		if len(msg) == 1 {
			fmt.Printf("ERROR: %s: %v\n", msg[0], err)
		} else {
			fmt.Printf("ERROR: %v\n", err)
		}
		os.Exit(1)
	}
}

func buildFileName(dirName, namePrefix, version string) string {
	return fmt.Sprintf("%s_%s.go", path.Join(dirName, namePrefix), version)
}
