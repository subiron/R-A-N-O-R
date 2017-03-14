package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

//"container/list"
const nsfMux = `/nfs{rest:.*}`
const rpcMux = `/rpc{rest:.*}`

var rootDir = "" //set this from args or config file
var currentDir, _ = os.Getwd()

func main() {
	r := mux.NewRouter()
	r.PathPrefix("/web/").Handler(http.StripPrefix("/web/", http.FileServer(rice.MustFindBox("web").HTTPBox())))
	r.HandleFunc(nsfMux, nfsGet).Methods("GET")
	r.HandleFunc(nsfMux, nfsPOST).Methods("POST")
	r.HandleFunc(nsfMux, nfsDELETE).Methods("DELETE")

	r.HandleFunc(rpcMux, rpc).Methods("POST")

	http.Handle("/", r)

	log.Println("R_A_N_O_R is listening")
	http.ListenAndServe(":3000", nil)
}

func rpc(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	cmd := r.Form.Get("cmd")
	arg := r.Form.Get("arg")

	out, err := exec.Command(cmd, arg).Output() //TODO need fix for native commands(not executables)like eg dir, del etc .... try to store 'cmd' in a batch script and then run it.
	if err != nil {
		log.Fatal(err)
		fmt.Fprintf(w, "RPC FAILED: "+err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		fmt.Fprintf(w, string(out))
	}
}

func basicAuth(w http.ResponseWriter, r *http.Request) bool {
	s := r.Header.Get("Authorization")
	if len(s) > 0 {
		s = s[6:]
		decoded, err := base64.StdEncoding.DecodeString(s)
		s2 := string(decoded)
		if err == nil && s2 == "admin:admin" {
			return true
		}
	}

	w.Header().Add("WWW-Authenticate", "Basic realm=\"Authentication Required\"")
	w.WriteHeader(http.StatusUnauthorized)
	return false

}

func nfsGet(w http.ResponseWriter, r *http.Request) {
	if !basicAuth(w, r) {
		return
	}
	dirName := rootDir
	if dirName == "" {
		dirName = currentDir
	}

	urlPath := r.URL.Path[4:]
	if len(urlPath) > 0 {
		dirName = dirName + urlPath
	}

	fileInfoList, err := ioutil.ReadDir(dirName)
	if err != nil {
		fmt.Fprintf(w, " :( "+err.Error(), r.URL.Path[1:])
		return
	}

	files := make([]CommanderFile, 0)
	for _, dir := range fileInfoList {
		files = append(files, CommanderFile{dir.Name(), dir.Size(), dir.IsDir()})
	}

	j, err := json.Marshal(files)
	if err != nil {
		fmt.Fprintf(w, " :( Hi there, I love %s!", r.URL.Path[1:])
	} else {
		fmt.Fprintf(w, "%s\n", j)
	}

}

func nfsPOST(w http.ResponseWriter, r *http.Request) {

}

func nfsDELETE(w http.ResponseWriter, r *http.Request) {

}

type CommanderFile struct {
	Title    string `json:"title"`
	Size     int64  `json:"size"`
	IsFolder bool   `json:"isFolder"`
}
