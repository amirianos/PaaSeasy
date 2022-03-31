package functions

import (
	"fmt"
	"io/ioutil"
	"os"
)

func Updater(path string, secret string, execute_commands string, work_directory string) {
	bagheie, err := ioutil.ReadFile("./configfiles/hooks.json")
	if err != nil {
		fmt.Println(err)
	}
	var jsonfiles = `    {
	"id": "` + path + `",
	"execute-command": "` + execute_commands + `",
	"command-working-directory": "` + work_directory + `",
	"pass-arguments-to-command": [
	  {
		"source": "payload",
		"name": "repository.clone_url"
	  }
	],
	"trigger-rule": {
		"match":
	{
		  "type": "payload-hmac-sha1",
		  "secret": "` + secret + `",
	  "parameter":
		  {
			"source": "header",
			"name": "X-Hub-Signature"
		  }
		}
	  }
  },`
	datasource := string(bagheie)
	data := "[\n" + jsonfiles + "\n" + datasource[1:]
	fmt.Println(data)
	f, err := os.Create("./configfiles/hooks.json")
	if err != nil {
		fmt.Println(err)
	}
	// close the file with defer
	defer f.Close()
	f.WriteString(data)
}
