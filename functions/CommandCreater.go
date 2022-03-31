package functions

import (
	"fmt"
	"os"
)

func CommandCreator(work_directory string, path string, dbuser string, dbpassword string, dbname string, gitrepo string) {
	bashfile := `#!/bin/bash

	repo="$1"

	
	lastver="$(sudo docker exec -t mysql-server2 mysql -u ` + dbuser + ` -p` + dbpassword + ` -D ` + dbname + ` -e 'select lastversion from lastversion where path="` + path + `"' | tail -n2 | head -n1 |awk '{print $2}')"

	rm -rf ` + work_directory + `/src && mkdir ` + work_directory + `/src
	

	git clone ${repo} ` + work_directory + `/src
	sudo docker build -t ` + path + `:$(($lastver+1))  . 
	sudo docker exec -t mysql-server2 mysql -u ` + dbuser + ` -p` + dbpassword + ` -e 'update ` + dbname + `.lastversion set lastversion='"$(($lastver+1))"' where path="` + path + `"'
	read -d '' manifest << EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ` + path + `-deployment
  labels:
	name: nginxdeploymentlabel
spec:
  replicas: 1
  selector:
	matchLabels:
	  name: ` + path + `podlabel
  template:
	metadata:
	  labels:
		name: ` + path + `podlabel
	spec:
	  containers:
	  - name: ` + path + `containername
		image: nginx:$(($lastver+1))
		ports:
		- containerPort: 80

---
apiVersion: v1
kind: Service
metadata:
  name: ` + path + `service
  labels:
	name: ` + path + `service
spec:
  type: LoadBalancer
  ports:
	- port: 80
	  nodePort: 30080
	  name: http
  selector:
	name: ` + path + `podlabel

EOF
    rm -rf ` + work_directory + `/manifestsrc && mkdir ` + work_directory + `/manifestsrc
	git clone ` + gitrepo + ` ` + work_directory + `/manifestsrc
    cd ` + work_directory + `/manifestsrc 

	rm -f ` + path + `.yaml
    echo "$manifest" > ` + path + `.yaml
	git add .
	git commit -m "add new manifest"
	git push ` + gitrepo + ` --all

	`

	Dockerfile := `FROM nginx:alpine

	COPY src/ /usr/share/nginx/html`
	f, err := os.Create(work_directory + "/Dockerfile")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	f.WriteString(Dockerfile)
	bash, _ := os.Create(work_directory + "/commands.sh")
	bash.WriteString(bashfile)
	bash.Close()
	err = os.Chmod(work_directory+"/commands.sh", 0777)
	if err != nil {
		fmt.Println("err in permission setting", err)
	}
	//fmt.Println(bashfile)
}
