package pkg

type Config struct {
	ApiVersion     string           `json:"apiVersion" yaml:"apiVersion"`
	Kind           string           `json:"kind" yaml:"kind"`
	ManagedCluster int              `json:"managedCluster" yaml:"managedCluster"`
	KubeconfigOpts KubeconfigOption `json:"kubeconfigOpts" yaml:"kubeconfigOpts"`
	HelmOpts       HelmOpts         `json:"helmOpts" yaml:"helmOpts"`
	Registries     Registry         `json:"registries" yaml:"registries"`
	Storage        Storage          `json:"storage" yaml:"storage"`
	Token          string           `json:"token" yaml:"token"`
}

type KubeconfigOption struct {
	Output            string `json:"output" yaml:"output"`
	UpdateEnvironment bool   `json:"updateEnvironment" yaml:"updateEnvironment"`
}

type HelmOpts struct {
	Type      string `json:"type" yaml:"type"`
	ChartPath string `json:"chartPath" yaml:"chartPath"`
	Version   string `json:"version" yaml:"version"`
}

type Storage struct {
	Endpoint string `json:"endpoint" yaml:"endpoint"`
	CAFile   string `json:"ca_file" yaml:"ca_file"`
	CertFile string `json:"cert_file" yaml:"cert_file"`
	KeyFile  string `json:"key_file" yaml:"key_file"`
}

// Copied from k3d (https://github.com/k3d-io/k3d/blob/3b0e6990c4e501d4424d1abe0d67ae54feed3650/pkg/types/k3s/registry.go) to avoid viper Unmarshal bug
// Copyright Here
/*
Copyright © 2020-2022 The k3d Author(s)
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

// Registry is for customize image registry download
type Registry struct {
	// Mirrors are namespace to mirror mapping for all namespaces.
	Mirrors map[string]Mirror `toml:"mirrors" yaml:"mirrors"`
	// Configs are configs for each registry.
	// The key is the FDQN or IP of the registry.
	Configs map[string]RegistryConfig `toml:"configs" yaml:"configs"`

	// Auths are registry endpoint to auth config mapping. The registry endpoint must
	// be a valid url with host specified.
	// DEPRECATED: Use Configs instead. Remove in containerd 1.4.
	Auths map[string]AuthConfig `toml:"auths" yaml:"auths"`
}

type Mirror struct {
	Endpoint []string `toml:"endpoint" yaml:"endpoint"`
}

// AuthConfig contains the config related to authentication to a specific registry
type AuthConfig struct {
	// Username is the username to login the registry.
	Username string `toml:"username" yaml:"username"`
	// Password is the password to login the registry.
	Password string `toml:"password" yaml:"password"`
	// Auth is a base64 encoded string from the concatenation of the username,
	// a colon, and the password.
	Auth string `toml:"auth" yaml:"auth"`
	// IdentityToken is used to authenticate the user and get
	// an access token for the registry.
	IdentityToken string `toml:"identitytoken" yaml:"identity_token"`
}

// TLSConfig contains the CA/Cert/Key used for a registry
type TLSConfig struct {
	CAFile             string `toml:"ca_file" yaml:"ca_file"`
	CertFile           string `toml:"cert_file" yaml:"cert_file"`
	KeyFile            string `toml:"key_file" yaml:"key_file"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify" yaml:"insecure_skip_verify"`
}

// RegistryConfig contains configuration used to communicate with the registry.
type RegistryConfig struct {
	// Auth contains information to authenticate to the registry.
	Auth *AuthConfig `toml:"auth" yaml:"auth"`
	// TLS is a pair of CA/Cert/Key which then are used when creating the transport
	// that communicates with the registry.
	TLS *TLSConfig `toml:"tls" yaml:"tls"`
}
