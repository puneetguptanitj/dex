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
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
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
	cfsslJsonString := `{"CN":"CNNAME","names":[{"O":"GROUPS"}],"key":{"algo":"ecdsa","size":256}}`
	cfsslJsonString1 := strings.Replace(cfsslJsonString, "CNNAME", user, -1)
	cfsslJsonString2 := strings.Replace(cfsslJsonString1, "GROUPS", orgStr, -1)

	file, err := ioutil.TempFile("", "csr")
	if err != nil {
		log.Printf("\n%v", err.Error())
	}
	ioutil.WriteFile(file.Name(), []byte(cfsslJsonString2), os.FileMode(os.O_RDONLY))
	cfssl := exec.Command("./cfssl", "genkey", file.Name())
	cfssl.Dir = dir + "/dex/cfssl"
	fmt.Print("Command to be executed ", cfssl)

	cfssljson := exec.Command("./cfssljson", "-bare", "server")
	cfssljson.Dir = dir + "/dex/cfssl"

	r, w := io.Pipe()
	cfssl.Stdout = w
	cfssljson.Stdin = r

	var b2 bytes.Buffer
	cfssljson.Stdout = &b2

	cfssl.Start()
	cfssljson.Start()
	cfssl.Wait()
	w.Close()
	cfssljson.Wait()
	io.Copy(os.Stdout, &b2)

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
	request, err := ioutil.ReadFile("/dex/cfssl/ca.csr")
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
			Groups:  groups,
		},
	}
	_, err = clientset.Certificates().CertificateSigningRequests().Create(&csrObject)
	if err != nil {
		log.Printf("\n%v", err.Error())
	}
	_, err = clientset.Certificates().CertificateSigningRequests().UpdateApproval(&csrObject)
	if err != nil {
		log.Printf("\nError approving the request %v", err.Error())
	}
	csrSignedOject, err := clientset.Certificates().CertificateSigningRequests().Get(user+"-csr", metav1.GetOptions{})
	clienCert := base64.StdEncoding.EncodeToString(csrSignedOject.Status.Certificate)

	caBytes, err := ioutil.ReadFile("/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		log.Printf("\n%v", err.Error())
	}
	config1 := strings.Replace(config, "CACERT", base64.StdEncoding.EncodeToString(caBytes), -1)
	clientKey, err := ioutil.ReadFile("/dex/cfssl/ca-key.pem")
	if err != nil {
		log.Printf("\n%v", err.Error())
	}
	config2 := strings.Replace(config1, "CLIENT_CERT", clienCert, -1)
	config3 := strings.Replace(config2, "CLIENT_KEY", base64.StdEncoding.EncodeToString(clientKey), -1)
	return config3

}
