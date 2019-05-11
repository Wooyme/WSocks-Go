package Wsocks

import (
	"encoding/json"
	"fmt"
	"github.com/getlantern/systray"
	"log"
	"net/http"
	"os"
	"os/exec"
	user2 "os/user"
	"path/filepath"
	"runtime"
	"strconv"
)

type Configuration struct {
	Host string
	User string
	Pass string
}

var configs = make([]Configuration, 0)
var client *Client = nil
var TrayState *systray.MenuItem = nil
func Tray() {
	localConfigServer()
	systray.Run(func() {
		icon, err := Asset("assets/icon.jpg")
		if err != nil {
			fmt.Printf("Err: %v \n",err)
		}
		systray.SetIcon(icon)
		systray.SetTitle("Wsocks-Go")
		TrayState = systray.AddMenuItem("Waiting","State")
		mEdit := systray.AddMenuItem("Edit", "Edit configuration")
		mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
		go func() {
			for {
				select {
				case <-mQuit.ClickedCh:
					os.Exit(0)
				case <-mEdit.ClickedCh:
					openBrowser("http://localhost:1082")
				}
			}
		}()
	}, nil)
}

func init() {
	user, _ := user2.Current()
	configPath := filepath.Join(user.HomeDir, ".wsocks/")
	if _, err := os.Stat(filepath.Join(configPath, "config.json")); os.IsNotExist(err) {
		err := os.MkdirAll(configPath, 0777)
		if err != nil {
			fmt.Printf("Err: %v \n", err)
			return
		}
	} else {
		f, err := os.Open(filepath.Join(configPath, "config.json"))
		if err != nil {
			fmt.Printf("Err: %v \n", err)
			return
		}
		bytes := make([]byte, 1024)
		l, _ := f.Read(bytes)
		if l > 1 {
			json.Unmarshal(bytes[:l], &configs)
		}
		f.Close()
	}
	openBrowser("http://localhost:1082")
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func localConfigServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		state:="Waiting"
		_log:=""
		if client!=nil {
			state = client.State
			_log = client.Log

		}
		if host, ok := r.URL.Query()["host"]; ok && len(host) > 0 {
			fmt.Printf("Value: %v \n", host[0])
			user, _ := r.URL.Query()["user"]
			pass, _ := r.URL.Query()["pass"]
			configs = append(configs, Configuration{
				Host: host[0],
				User: user[0],
				Pass: pass[0],
			})
			_user, _ := user2.Current()
			f, err := os.Create(filepath.Join(_user.HomeDir, ".wsocks/", "config.json"))
			if err != nil {
				fmt.Printf("Err: %v \n", err)
				return
			}
			configJson, err := json.Marshal(configs)
			if err != nil {
				log.Fatal("Cannot encode to JSON ", err)
			}
			_, _ = f.Write(configJson)
			f.Close()
			state="Saved"
		}
		if del,ok:=r.URL.Query()["delete"]; ok && len(del) >0 {
			config, _ := r.URL.Query()["config"]
			fmt.Printf("Config: %v \n", config)
			i, _ := strconv.Atoi(config[0])
			configs = append(configs[:i], configs[i+1:]...)
			_user, _ := user2.Current()
			f, err := os.Create(filepath.Join(_user.HomeDir, ".wsocks/", "config.json"))
			if err != nil {
				fmt.Printf("Err: %v \n", err)
				return
			}
			configJson, err := json.Marshal(configs)
			if err != nil {
				log.Fatal("Cannot encode to JSON ", err)
			}
			_, _ = f.Write(configJson)
			f.Close()
			state="Saved"
		}
		if confirm,ok:=r.URL.Query()["confirm"]; ok && len(confirm)>0 {
			config, _ := r.URL.Query()["config"]
			fmt.Printf("Config: %v \n", config)
			i, _ := strconv.Atoi(config[0])
			if client == nil {
				client = NewClient(configs[i].Host, configs[i].User, configs[i].Pass)
				go client.Start()
			} else {
				client.EditRemote(configs[i].Host, configs[i].User, configs[i].Pass)
			}
			state=fmt.Sprintf("Current configuration is: %v:%v@%v, Connecting...", configs[i].User, configs[i].Pass, configs[i].Host)
		}

		var str = ""
		for i := 0; i < len(configs); i++ {
			str += fmt.Sprintf("<option value=\"%v\">%v:%v@%v</option>", i, configs[i].User, configs[i].Pass, configs[i].Host)
		}
		fmt.Fprint(w, fmt.Sprintf(`
			<html>
			<head>
				<title>Wsocks-Go Configuration</title>
			</head>
			<body>
				<div>
				State: %v
				<form action="/"><input type="submit" value="refresh"></form>
				<div>
				<form action="/">
					<div>
						<label>Configuration</label>
						<select name="config">
							%v
						</select>
					</div>
					<div><input type="submit" name="confirm" value="Confirm"><input type="submit" name="delete" value="Delete"></div>
				</form>
				<form action="/">
				<div><label>Host:</label><input type="text" name="host"></div>
				<div><label>User:</label><input type="text" name="user"></div>
				<div><label>Pass:</label><input type="text" name="pass"></div>
				<div><input type="submit" value="Save"></div>
				</form>
				<div>
				Log
				<div><textarea rows="10" cols="50">%v</textarea></div>
				</div>
			</body>
			</html>
		`, state,str,_log))
	})
	go http.ListenAndServe(":1082", nil)
}
