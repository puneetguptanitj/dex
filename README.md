# Self service kubeconfig

![Alt text](doc.png?raw=true "Title")

1. User accesses the client app and asks for kubeconfig
2. Client app redirects to dex, where dex asks for user credentials
3. User credentials are relayed to IDP 
4. Dex returns the token with claim information 
5. Kubernetes mounts service account secrets on every pod in the namespaces. Client app uses that as the ca cert for kubeconfig
6. Client app uses cfssl to create private and CSR for user (using the right group claim for organization)
7. Client app then sends CSR for signing to API server and self approves it (client app service account has all the access to cert APIs)
8. Client uses CA cert, user key and cert to generate a kubeconfig for user to download
9. Client clears all footprint (no state remains after the kubeconfig is generated)
