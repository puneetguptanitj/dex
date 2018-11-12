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
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	v1beta1 "k8s.io/api/certificates/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func PrintCSRs(user string, groups []string) string {
	dir, _ := os.Getwd()
	orgStr := strings.Join(groups, ",")
	genReq := exec.Command("./easyrsa", "--batch", "--req-cn="+user, "--req-email=", "--dn-mode=org", "--req-org="+orgStr, "gen-req", user, "nopass")
	genReq.Dir = dir + "/easy-rsa/easyrsa3/"
	fmt.Print("Command to be executed ", genReq)
	err := genReq.Run()
	if err != nil {
		log.Printf("\n%v", err.Error())
	}
	// creates the in-cluster config
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("\n%v", err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		log.Printf("\n%v", err.Error())
	}
	config := `{"apiVersion":"v1","clusters":[{"cluster":{"certificate-authority-data":"CACERT",
	"server":"https://10.105.16.158:6443"},"name":"myK8sCluster"}],"contexts":
	[{"context":{"cluster":"myK8sCluster","user":"usera"},"name":"myK8sCluster"}],
	"current-context":"myK8sCluster","kind":"Config","preferences":{},"users":
	[{"name":"usera","user":{"client-certificate-data":"CLIENT_CERT","client-key-data":"CLIENT_KEY"}}]}`

	csr, err := clientset.Certificates().CertificateSigningRequests().List(metav1.ListOptions{})
	if err != nil {
		log.Printf("\n%v", err.Error())
	}
	request, err := ioutil.ReadFile(dir + "/easy-rsa/easyrsa3/pki/reqs/" + user + ".req")
	if err != nil {
		log.Printf("\n%v", err.Error())
	}
	fmt.Printf("\nCSRs in the cluster %v", csr.Items)
	csrObject := v1beta1.CertificateSigningRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CertificateSigningRequest",
			APIVersion: "certificates.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: user + "-csr",
		},
		Spec: v1beta1.CertificateSigningRequestSpec{
			Request: request,
		},
	}
	csrReturedOject, err := clientset.Certificates().CertificateSigningRequests().Create(&csrObject)
	if err != nil {
		log.Printf("\n%v", err.Error())
	}
	_, err = clientset.Certificates().CertificateSigningRequests().UpdateApproval(csrReturedOject)
	if err != nil {
		log.Printf("\n%v", err.Error())
	}
	csrSignedOject, err := clientset.Certificates().CertificateSigningRequests().Get(user+"-csr", metav1.GetOptions{})
	clienCert := base64.StdEncoding.EncodeToString(csrSignedOject.Status.Certificate)

	caBytes, err := ioutil.ReadFile("/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		log.Printf("\n%v", err.Error())
	}
	config1 := strings.Replace(config, "CACERT", base64.StdEncoding.EncodeToString(caBytes), -1)
	clientKey, err := ioutil.ReadFile(dir + "/easy-rsa/easyrsa3/pki/private/" + user + ".key")
	if err != nil {
		log.Printf("\n%v", err.Error())
	}
	config2 := strings.Replace(config1, "CLIENT_CERT", clienCert, -1)
	config3 := strings.Replace(config2, "CLIENT_KEY", base64.StdEncoding.EncodeToString(clientKey), -1)
	return config3

}
