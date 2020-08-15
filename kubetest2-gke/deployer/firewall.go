/*
Copyright 2020 The Kubernetes Authors.

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
	"fmt"
	"sort"
	"strings"
	"time"

	"k8s.io/klog"

	"sigs.k8s.io/kubetest2/pkg/exec"
)

func (d *deployer) ensureFirewall(hostProject, curtProject, cluster, network string) error {
	klog.V(1).Infof("Ensuring firewall rules for cluster %s in %s", cluster, curtProject)
	if network == "default" {
		return nil
	}
	firewall := d.getClusterFirewall(curtProject, cluster)
	if runWithNoOutput(exec.Command("gcloud", "compute", "firewall-rules", "describe", firewall,
		"--project="+hostProject,
		"--format=value(name)")) == nil {
		// Assume that if this unique firewall exists, it's good to go.
		return nil
	}
	klog.V(1).Infof("Couldn't describe firewall '%s', assuming it doesn't exist and creating it", firewall)

	tagOut, err := exec.Output(exec.Command("gcloud", "compute", "instances", "list",
		"--project="+curtProject,
		"--filter=metadata.created-by:*"+d.instanceGroups[curtProject][cluster][0].path,
		"--limit=1",
		"--format=get(tags.items)"))
	if err != nil {
		return fmt.Errorf("instances list failed: %s", execError(err))
	}
	tag := strings.TrimSpace(string(tagOut))
	if tag == "" {
		return fmt.Errorf("instances list returned no instances (or instance has no tags)")
	}

	if err := runWithOutput(exec.Command("gcloud", "compute", "firewall-rules", "create", firewall,
		"--project="+hostProject,
		"--network="+network,
		"--allow="+e2eAllow,
		"--target-tags="+tag)); err != nil {
		return fmt.Errorf("error creating e2e firewall: %v", err)
	}
	return nil
}

func (d *deployer) getClusterFirewall(project, cluster string) string {
	// We want to ensure that there's an e2e-ports-* firewall rule
	// that maps to the cluster nodes, but the target tag for the
	// nodes can be slow to get. Use the hash from the lexically first
	// node pool instead.
	return "e2e-ports-" + d.instanceGroups[project][cluster][0].uniq
}

// This function ensures that all firewall-rules are deleted from specific network.
// We also want to keep in logs that there were some resources leaking.
func (d *deployer) cleanupNetworkFirewalls(hostProject, network string) (int, error) {
	// Do not delete firewall rules for the default network.
	if network == "default" {
		return 0, nil
	}

	klog.V(1).Infof("Cleaning up network firewall rules for network %s in %s", network, hostProject)
	fws, err := exec.Output(exec.Command("gcloud", "compute", "firewall-rules", "list",
		"--format=value(name)",
		"--project="+hostProject,
		"--filter=network:"+network))
	if err != nil {
		return 0, fmt.Errorf("firewall rules list failed: %s", execError(err))
	}
	if len(fws) > 0 {
		fwList := strings.Split(strings.TrimSpace(string(fws)), "\n")
		klog.V(1).Infof("Network %s has %v undeleted firewall rules %v", network, len(fwList), fwList)
		commandArgs := []string{"compute", "firewall-rules", "delete", "-q"}
		commandArgs = append(commandArgs, fwList...)
		commandArgs = append(commandArgs, "--project="+hostProject)
		errFirewall := runWithOutput(exec.Command("gcloud", commandArgs...))
		if errFirewall != nil {
			return 0, fmt.Errorf("error deleting firewall: %v", errFirewall)
		}
		// It looks sometimes gcloud exits before the firewall rules are actually deleted,
		// so sleep 10 seconds to wait for the firewall rules being deleted completely.
		// TODO(chizhg): change to a more reliable way to check if they are deleted or not.
		time.Sleep(10 * time.Second)
	}
	return len(fws), nil
}

func (d *deployer) getInstanceGroups() error {
	// If instanceGroups has already been populated, return directly.
	if d.instanceGroups != nil {
		return nil
	}

	// Initialize project instance groups structure
	d.instanceGroups = map[string]map[string][]*ig{}

	location, err := d.location()
	if err != nil {
		return err
	}

	for _, project := range d.projects {
		d.instanceGroups[project] = map[string][]*ig{}

		for _, cluster := range d.projectClustersLayout[project] {
			igs, err := exec.Output(exec.Command("gcloud", d.containerArgs("clusters", "describe", cluster,
				"--format=value(instanceGroupUrls)",
				"--project="+project,
				location)...))
			if err != nil {
				return fmt.Errorf("instance group URL fetch failed: %s", execError(err))
			}
			igURLs := strings.Split(strings.TrimSpace(string(igs)), ";")
			if len(igURLs) == 0 {
				return fmt.Errorf("no instance group URLs returned by gcloud, output %q", string(igs))
			}
			sort.Strings(igURLs)

			// Inialize cluster instance groups
			d.instanceGroups[project][cluster] = make([]*ig, 0)

			for _, igURL := range igURLs {
				m := poolRe.FindStringSubmatch(igURL)
				if len(m) == 0 {
					return fmt.Errorf("instanceGroupUrl %q did not match regex %v", igURL, poolRe)
				}
				d.instanceGroups[project][cluster] = append(d.instanceGroups[project][cluster], &ig{path: m[0], zone: m[1], name: m[2], uniq: m[3]})
			}
		}
	}

	return nil
}
