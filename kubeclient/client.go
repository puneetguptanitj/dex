/*
Copyright 2016 The Kubernetes Authors.

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

// Note: the example only works with the code within the same release/branch.
package kubeclient

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func PrintCSRs(user string, groups []string) {
	dir, _ := os.Getwd()
	command := dir + "/easy-rsa/easyrsa3/easyrsa"

	genReq := exec.Command(command, "--batch", "--req-cn="+user, "--req-email=", "--dn-mode=org", "--req-org="+userInfo.Group, "gen-req", userInfo.Name, "nopass")
	fmt.Print("Command to be executed ", genReq)
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("\n%v", err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("\n%v", err.Error())
	}
	csr, err := clientset.Certificates().CertificateSigningRequests().List(metav1.ListOptions{})
	if err != nil {
		log.Printf("\n%v", err.Error())
	}
	fmt.Printf("\nCSRs in the cluster %v", csr.Items)
}
