package k8s

import (
	"bytes"

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
