package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
)

var client = req.C().
	SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36").
	SetTimeout(8 * time.Second).
	EnableInsecureSkipVerify().
	SetRedirectPolicy(func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse })

func main() {

	fmt.Println(`
██████   ██████   ██████ ██   ██ ███████ ██████       █████  ██████  ██     ██████   ██████ ███████ 
██   ██ ██    ██ ██      ██  ██  ██      ██   ██     ██   ██ ██   ██ ██     ██   ██ ██      ██      
██   ██ ██    ██ ██      █████   █████   ██████      ███████ ██████  ██     ██████  ██      █████   
██   ██ ██    ██ ██      ██  ██  ██      ██   ██     ██   ██ ██      ██     ██   ██ ██      ██      
██████   ██████   ██████ ██   ██ ███████ ██   ██     ██   ██ ██      ██     ██   ██  ██████ ███████ 
`)

	var (
		url string
		id  string
		cmd string
	)

	flag.StringVar(&url, "u", "", "url")
	flag.StringVar(&id, "i", "", "id")
	flag.StringVar(&cmd, "c", "", "cmd")
	flag.Parse()

	if strings.EqualFold(url, "") {
		flag.Usage()
		return
	}

	url = strings.TrimRight(url, "/")

	if strings.EqualFold(id, "") {
		getContainersJson(url)
		return
	}

	if !strings.EqualFold(cmd, "") {
		getContainersExec(url, id, cmd)
	} else {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print(">>> ")
			input, _ := reader.ReadString('\n')

			input = strings.TrimSpace(input)

			if input == "exit" {
				break
			}
			getContainersExec(url, id, input)
		}
	}

}

func getContainersJson(url string) {
	containers_url := url + "/containers/json"

	resp, err := client.R().Get(containers_url)
	if err != nil {
		fmt.Println(err)
		return
	}
	GetContainersld(resp.String())
}

func GetContainersld(resp string) {
	result := gjson.Parse(resp).Array()
	if len(result) > 0 {
		for _, obj := range result {

			id := obj.Get("Id").String()
			image := obj.Get("Image").String()
			name := obj.Get("Names.0").String()

			fmt.Printf("ID: %s\nImage: %s\t Name: %s\n\n", id, image, name)
		}
	} else {
		fmt.Println("未发现 Docker 容器")
	}

}

func getContainersExec(url, id, cmd string) {
	containers_exec_url := fmt.Sprintf("%s/containers/%s/exec", url, id)
	cmd_data := fmt.Sprintf(`{"AttachStdin":true,"AttachStdout":true,"AttachStderr":true, "Cmd": %s,"DetachKeys":"ctrl-p,ctrl-q","Privileged":true,"Tty":true}`, getCmd(cmd))

	resp, err := client.R().SetBodyJsonString(cmd_data).Post(containers_exec_url)
	if err != nil {
		fmt.Println(err)
		return
	}
	cmd_id := gjson.Get(resp.String(), "Id")

	exec_url := fmt.Sprintf("%s/exec/%s/start", url, cmd_id)
	exec_data := `{"Detach":false,"Tty":false}`

	resp, err = client.R().SetBodyJsonString(exec_data).Post(exec_url)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(removeValues(resp.Bytes(), 1, 0))
}

func getCmd(str string) string {
	splitStr := strings.Split(str, " ")
	jsonObj := map[string][]string{"Cmd": splitStr}
	cmdStr, _ := json.Marshal(jsonObj["Cmd"])
	return string(cmdStr)
}

func removeValues(bytes []byte, values ...byte) string {
	var cleaned []byte

	for _, b := range bytes {

		remove := false
		for _, v := range values {
			if b == v {
				remove = true
				break
			}
		}

		if !remove {
			cleaned = append(cleaned, b)
		}
	}

	return string(cleaned)
}
