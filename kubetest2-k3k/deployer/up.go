/*
Copyright 2021 The Kubernetes Authors.

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

package deployer

import (
	"os"
	"strings"

	"k8s.io/klog/v2"

	"sigs.k8s.io/kubetest2/pkg/exec"
	"sigs.k8s.io/kubetest2/pkg/metadata"
	"sigs.k8s.io/kubetest2/pkg/process"
)

func (d *deployer) IsUp() (up bool, err error) {
	// naively assume that if the api server reports nodes, the cluster is up
	lines, err := exec.CombinedOutputLines(
		exec.Command("kubectl", "get", "nodes", "-o=name"),
	)
	if err != nil {
		return false, metadata.NewJUnitError(err, strings.Join(lines, "\n"))
	}
	return len(lines) > 0, nil
}

func (d *deployer) Up() error {
	args := []string{
		"cluster", "create",
		"--kubeconfig", d.KubeconfigPath,
		"--namespace", d.Namespace,
		"--name", d.Name,
		"--servers", d.Servers,
		"--agents", d.Agents,
		"--token", d.Token,
		"--cluster-cidr", d.ClusterCIDR,
		"--service-cidr", d.ServiceCIDR,
		"--persistence-type", d.PersistenceType,
		"--storage-class-name", d.StorageClassName,
		"--server-args", d.ServerArgs,
		"--agent-args", d.AgentArgs,
		"--version", d.Version,
		"--mode", d.Mode,
	}

	klog.V(0).Infof("Up(): creating K3K cluster...\n")
	klog.V(0).Infof("all the args are...", args)
	// we want to see the output so use process.ExecJUnit
	return process.ExecJUnit("k3kcli", args, os.Environ())
}
