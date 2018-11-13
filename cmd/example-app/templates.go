package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/dexidp/dex/kubeclient"
)

var indexTmpl = template.Must(template.New("index.html").Parse(`<html>
<style>

body {
  background: #2d343d;
}

.login {
  margin: 20px auto;
  width: 300px;
  padding: 30px 25px;
  background: white;
  border: 1px solid #c4c4c4;
}

h1.login-title {
  margin: -28px -25px 25px;
  padding: 15px 25px;
  line-height: 30px;
  font-size: 25px;
  font-weight: 300;
  color: #ADADAD;
  text-align:center;
  background: #f7f7f7;
 
}

.login-input {
  width: 285px;
  height: 50px;
  margin-bottom: 25px;
  padding-left:10px;
  font-size: 15px;
  background: #fff;
  border: 1px solid #ccc;
  border-radius: 4px;
}
.login-input:focus {
    border-color:#6e8095;
    outline: none;
  }
.login-button {
  width: 100%;
  height: 50px;
  padding: 0;
  font-size: 20px;
  color: #fff;
  text-align: center;
  background: #f0776c;
  border: 0;
  border-radius: 5px;
  cursor: pointer; 
  outline:0;
}

.login-lost
{
  text-align:center;
  margin-bottom:0px;
}

.login-lost a
{
  color:#666;
  text-decoration:none;
  font-size:13px;
}



</style>
	<body>
		<form class="login" action="/login" method="post">
		<h1 class="login-title">Accees my cluster</h1>
       <input type="submit" class="login-button" value="Get kubeconfig">
    </form>
  </body>
</html>`))

func renderIndex(w http.ResponseWriter) {
	renderTemplate(w, indexTmpl, tokenTmplData{})
}

type tokenTmplData struct {
	IDToken      string
	RefreshToken string
	RedirectURL  string
	Claims       string
	Kubeconfig   string
	User         string
}

var tokenTmpl = template.Must(template.New("token.html").Parse(`<html>
  <head>
    <style>
/* make pre wrap */
pre {
 white-space: pre-wrap;       /* css-3 */
 white-space: -moz-pre-wrap;  /* Mozilla, since 1999 */
 white-space: -pre-wrap;      /* Opera 4-6 */
 white-space: -o-pre-wrap;    /* Opera 7 */
 word-wrap: break-word;       /* Internet Explorer 5.5+ */
}
.card {
	box-shadow: 0 4px 8px 0 rgba(0,0,0,0.2);
	transition: 0.3s;
	width: 40%;
}

.card:hover {
	box-shadow: 0 8px 16px 0 rgba(0,0,0,0.2);
}

.container {
	padding: 2px 16px;
}
    </style>
  </head>
	<body>
	<div class="card">
  <div class="container">
    <h4><b>Token</b></h4> 
    <p> Token: <pre><code>{{ .IDToken }}</code></pre></p>
	</div>
	</div>
	<div class="card">
  <div class="container">
    <h4><b>Claims</b></h4> 
    <p> Claims: <pre><code>{{ .Claims }}</code></pre></p>
  </div>
  </div>
  <div class="card">
  <div class="container">
    <h4><b>Kubeconfig</b></h4> 
    <p> Claims: <pre><code>{{ .Kubeconfig }}</code></pre></p>
  </div>
  </div>
  <a href="data:text/plain;charset=UTF-8,{{ .Kubeconfig }}" download>{{ .User }}-kubeconfig</a>
  </body>
</html>
`))

type MyClaims struct {
	Groups []string `json:"groups,omitempty"`
	Name   string   `json:"name"`
}

func renderToken(w http.ResponseWriter, redirectURL, idToken, refreshToken string, claims []byte) {
	claimsMap := MyClaims{}
	err := json.Unmarshal(claims, &claimsMap)
	if err != nil {
		log.Printf("\n Failed to unmarshal claims object: %v", err)
	}
	kubeconfig := kubeclient.PrintCSRs(claimsMap.Name, claimsMap.Groups)
	renderTemplate(w, tokenTmpl, tokenTmplData{
		IDToken:      idToken,
		RefreshToken: refreshToken,
		RedirectURL:  redirectURL,
		Claims:       string(claims),
		Kubeconfig:   kubeconfig,
		User:         claimsMap.Name,
	})
}

func renderTemplate(w http.ResponseWriter, tmpl *template.Template, data tokenTmplData) {
	err := tmpl.Execute(w, data)
	if err == nil {
		return
	}
	switch err := err.(type) {
	case *template.Error:
		// An ExecError guarantees that Execute has not written to the underlying reader.
		log.Printf("Error rendering template %s: %s", tmpl.Name(), err)

		// TODO(ericchiang): replace with better internal server error.
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	default:
		// An error with the underlying write, such as the connection being
		// dropped. Ignore for now.
	}
}
