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

package ibmcloud

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	clustercommon "sigs.k8s.io/cluster-api/pkg/apis/cluster/common"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/util"
)

const ProviderName = "ibmcloud"
const (
	IBMCloudIPAnnotationKey = "ibmcloud-ip-address"
	IBMCloudIdAnnotationKey = "ibmcloud-resourceId"
)

func init() {
	clustercommon.RegisterClusterProvisioner(ProviderName, NewDeploymentClient())
}

type DeploymentClient struct{}

func NewDeploymentClient() *DeploymentClient {
	return &DeploymentClient{}
}

func (d *DeploymentClient) GetIP(cluster *clusterv1.Cluster, machine *clusterv1.Machine) (string, error) {
	if machine.ObjectMeta.Annotations != nil {
		if ip, ok := machine.ObjectMeta.Annotations[IBMCloudIPAnnotationKey]; ok {
			log.Printf("Returning IP from machine annotation %s", ip)
			return ip, nil
		}
	}

	return "", errors.New("could not get IP")
}

// GetKubeConfig gets a kubeconfig from the master.
func (d *DeploymentClient) GetKubeConfig(cluster *clusterv1.Cluster, master *clusterv1.Machine) (string, error) {
	ip, err := d.GetIP(cluster, master)
	if err != nil {
		return "", err
	}

	homeDir, ok := os.LookupEnv("HOME")
	if !ok {
		return "", fmt.Errorf("unable to use HOME environment variable to find SSH key: %v", err)
	}

	// FIXME: use ssh user defined in machine spec name later
	sshUserName := "ubuntu"
	// FIXME: use other predefined ssh keyname or make this global definition
	privateKey := "cluster-api-provider-ibmcloud"

	result := strings.TrimSpace(util.ExecCommand(
		"ssh", "-i", homeDir+"/.ssh/"+privateKey,
		"-o", "StrictHostKeyChecking no",
		"-o", "UserKnownHostsFile /dev/null",
		"-o", "BatchMode=yes",
		fmt.Sprintf("%s@%s", sshUserName, ip),
		"echo STARTFILE; sudo cat /etc/kubernetes/admin.conf"))
	parts := strings.Split(result, "STARTFILE")
	if len(parts) != 2 {
		return "", nil
	}
	return strings.TrimSpace(parts[1]), nil
}
