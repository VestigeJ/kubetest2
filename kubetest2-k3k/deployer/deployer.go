/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package deployer implements the kubetest2 kind deployer
package deployer

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"sigs.k8s.io/kubetest2/pkg/artifacts"
	"sigs.k8s.io/kubetest2/pkg/types"
)

// Name is the name of the deployer
const Name = "k3k"

var GitTag string

// New implements deployer.New for kind
func New(opts types.Options) (types.Deployer, *pflag.FlagSet) {
	// create a deployer object and set fields that are not flag controlled
	d := &deployer{
		commonOptions: opts,
		logsDir:       filepath.Join(artifacts.BaseDir(), "logs"),
	}
	// register flags and return
	return d, bindFlags(d)
}

// assert that New implements types.NewDeployer
var _ types.NewDeployer = New

type deployer struct {
	// generic parts
	commonOptions types.Options
	// k3k specific details
	KubeconfigPath   string `flag:"kubeconfig" desc:"--kubeconfig Kubeconfig path for k3k create cluster"`
	Namespace        string `flag:"namespace" desc:"--namespace Namespace to create the k3k cluster in"`
	Name             string `flag:"name" desc:"--name Name of the k3k cluster"`
	Servers          string `default:"1"`
	Agents           string `default:"0"`
	Token            string `flag:"token" desc:"--token Token to use for k3k cluster creation"`
	ClusterCIDR      string `flag:"cluster-cidr" desc:"--cluster-cidr Cluster CIDR to use for k3k cluster creation"`
	ServiceCIDR      string `flag:"service-cidr" desc:"--service-cidr Service CIDR to use for k3k cluster creation"`
	PersistenceType  string `default:"ephemeral" desc:"--persistence-type Persistence mode for the nodes (ephermal, static, dynamic)"`
	StorageClassName string `flag:"storage-class-name" desc:"--storage-class-name Storage class name for dynamic persistence type"`
	ServerArgs       string `flag:"server-args" desc:"--server-args Additional arguments to pass to the k3k server nodes"`
	AgentArgs        string `flag:"agent-args" desc:"--agent-args Additional arguments to pass to the k3k agent nodes"`
	Mode             string `default:"shared" desc:"--mode Mode to run k3k in (shared, virtual)"`
	Version          string `default:"v1.32.1" desc:"--version Version of k3s to install"`
	// K3kConfigPath    string `flag:"k3k-config" desc:" $HOME/$Name-kubeconfig.yaml"`

	logsDir string `default:"/tmp/k3k/logs"`
}

func (d *deployer) Kubeconfig() (string, error) {
	if d.KubeconfigPath != "" {
		return d.KubeconfigPath, nil
	}

	if kconfig, ok := os.LookupEnv("KUBECONFIG"); ok {
		return kconfig, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, d.Name+"-kubeconfig.yaml"), nil
}

func (d *deployer) GetVersion() string {
	// return GitTag
	return d.Version
}

func bindFlags(d *deployer) *pflag.FlagSet {
	flags, err := gpflag.Parse(d)
	if err != nil {
		klog.Fatal("unable to generate flags from deployer")
		return nil
	}

	klog.InitFlags(nil)
	flags.AddGoFlagSet(flag.CommandLine)

	return flags
}

var _ types.DeployerWithKubeconfig = &deployer{}
