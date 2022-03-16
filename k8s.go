/*
Package k8s is focused on the most common utility functions and structures needed for rapid applications development for Kubernetes. For more robust and substantial production applications consider using the Kubernetes packages directly. Many may prefer to simply pilfer from this package and paste code into their own.
*/
package k8s

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
	kyaml "sigs.k8s.io/yaml"
)

// NormYAML (originally from the Kind project) round trips YAML bytes
// through sigs.k8s.io/yaml to normalize them versus other Kubernetes
// ecosystem YAML output. This is particularly useful when dealing with
// one tool that uses yaml.v3 while Kubernetes itself has remained with
// yaml.v2 (and has had issues with the YAML project because of breaking
// changes even to yaml.v2). Most will note the difference between
// indentation and wrapping between v2 and v3, which has caused some
// project tests to fail in the past (although testing for specific YAML
// output is probably unwise).
func NormYAML(y []byte) ([]byte, error) {
	buf := new(any)
	if err := kyaml.Unmarshal(y, buf); err != nil {
		return nil, err
	}
	encoded, err := kyaml.Marshal(buf)
	if err != nil {
		return nil, err
	}
	// special case: don't write anything when empty
	if bytes.Equal(encoded, []byte("{}\n")) {
		return []byte{}, nil
	}
	return encoded, nil
}

// https://github.com/kubernetes/client-go/blob/0bdba2f9188006fc64057c2f6d82a0f9ee0ee422/tools/clientcmd/api/v1/types.go

// KubeConfig represents an opinionated subset of KUBECONFIG primarily
// for use by apps that are manipulating the file (and inspired by
// Kind). While there are some similarities to client-go the structure
// is unique (contexts as slices, not maps, for example).  The special
// O map captures anything that could potentially be added prevents the
// risk of compatibility changes over time and allows authors to create
// their own unique sections without worry (although this should be done
// carefully to ensure there is never any incompatibility with future
// KUBECONFIG API changes). If something new is added, it will always be
// available under O until (and if) it is graduated to having its own
// reference. See the client-go types under clientcmd/api for more.
type KubeConfig struct {
	Clusters []*NCluster    `yaml:"clusters,omitempty"`
	Contexts []*NContext    `yaml:"contexts,omitempty"`
	Users    []*NUser       `yaml:"users,omitempty"` // AuthInfos
	Current  string         `yaml:"current-context,omitempty"`
	O        map[string]any `yaml:",inline,omitempty"`
}

// Load configuration from a specific file at path.
func (c *KubeConfig) Load(path string) error {
	buf, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(buf, c)
}

func (c KubeConfig) String() string {
	buf, _ := yaml.Marshal(c)
	return string(buf)
}

func (c *KubeConfig) Write(path string) error {
	buf, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, buf, 0600)
}

// NCluster associates a name with a cluster.
type NCluster struct {
	Name    string
	Cluster *Cluster
}

// Cluster contains information about a Kubernetes cluster.
type Cluster struct {
	Server        string         `yaml:"server"`
	TLSServerName string         `yaml:"tls-server-name,omitempty"`
	SkipTLSVerify bool           `yaml:"insecure-skip-tls-verify,omitempty"`
	CertAuthority string         `yaml:"certificate-authority-data,omitempty"`
	Proxy         string         `yaml:"proxy-url,omitempty"`
	O             map[string]any `yaml:",inline,omitempty"`
}

// NUser associates a name with a user.
type NUser struct {
	Name string
	User *User
}

// User (officially "AuthInfo") contains information that describes
// identity information. This struct is significantly different than
// its client-go ancestor and assumes that all configuration data
// resides within a single KUBECONFIG file. Names have been shortened to
// reasonable lengths.
type User struct {
	Cert     string         `yaml:"client-certificate-data,omitempty"`
	Key      string         `yaml:"client-key-data,omitempty"`
	Token    string         `yaml:"token,omitempty"`
	As       string         `yaml:"act-as,omitempty"`
	AsUID    string         `yaml:"act-as-uid,omitempty"`
	AsGroups []string       `yaml:"act-as-groups,omitempty"`
	Name     string         `yaml:"username,omitempty"`
	Pass     string         `yaml:"password,omitempty"`
	Auth     *AuthProvider  `yaml:"auth-provider,omitempty"`
	O        map[string]any `yaml:",inline,omitempty"`
}

// NContext associates a name with a context.
type NContext struct {
	Name    string
	Context *Context
}

// Context is mostly cluster, user, and namespace.
type Context struct {
	Cluster   string         `yaml:"cluster"`
	User      string         `yaml:"user"`
	Namespace string         `yaml:"namespace,omitempty"`
	O         map[string]any `yaml:",inline,omitempty"`
}

// AuthProvider holds the configuration for a specified auth provider.
type AuthProvider struct {
	Name   string            `yaml:"name"`
	Config map[string]string `yaml:"config,omitempty"`
	O      map[string]any    `yaml:",inline,omitempty"`
}

// GoString implements fmt.GoStringer and sanitizes sensitive fields of
// AuthProvider to prevent accidental leaking via logs.
func (c AuthProvider) GoString() string { return c.String() }

// String implements fmt.Stringer and sanitizes sensitive fields of
// AuthProvider to prevent accidental leaking via logs.
func (c AuthProvider) String() string {
	cfg := "<nil>"
	if c.Config != nil {
		cfg = "--- REDACTED ---"
	}
	return fmt.Sprintf("api.AuthProvider{Name: %q, Config: map[string]string{%s}}",
		c.Name, cfg)
}
