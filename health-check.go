package main

import (
       "crypto/tls"
       "io/ioutil"
       "net/http"
       "os"
       "os/exec"
       "log"
       "bytes"
       "time"
       "strconv"
       "strings"
       "encoding/json"
)

const HTTP_STATUS_OK = 200

var checkUrl string
var nginxConf string
var activeHost string
var passiveHost string
var interval int
var nginxCmd string

/************************/

type AppConfiguration struct {
    Check_url       string  `json:"check_url"`
    Nginx_conf      string  `json:"nginx_conf"`
    Active_host     string  `json:"active_host"`
    Passive_host    string  `json:"passive_host"`
    Interval_sec    string  `json:"interval_sec"`
    Nginx_cmd       string  `json:"nginx_cmd"`
}

/************************/
var (
    outfile, _ = os.Create(os.Args[2])
    AppLog      = log.New(outfile, "PREFIX: ", log.Ldate|log.Ltime|log.Lshortfile)
)

/************************/
func healthCheck() int{
  transCfg := &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired or self signed SSL certificates
  }
  client := &http.Client{Transport: transCfg}

  response, err := client.Head(checkUrl)

  if err != nil {
      AppLog.Printf("1. %v",err)
          os.Exit(1)
  }

  defer response.Body.Close()

  //AppLog.Printf("2. %v http status code: %v",checkUrl,response.StatusCode)
  return response.StatusCode
}

/************************/
func setActive(){


  inFile, err := ioutil.ReadFile(nginxConf)
  if err != nil {
     AppLog.Printf("3. %v",err)
     os.Exit(1)
  }

  currentActive := "server " + activeHost + ";"
  newActive := "server " + passiveHost + ";"
  currentPassive := "server " + passiveHost + " down;"
  newPassive := "server " + activeHost + " down;"

  setActive := bytes.Replace(inFile, []byte(currentActive), []byte(newActive), -1)
  setPassive := bytes.Replace(setActive, []byte(currentPassive), []byte(newPassive), -1)

  if err = ioutil.WriteFile(nginxConf, setPassive, 0644); err != nil {
    AppLog.Printf("4. %v",err)
    os.Exit(1)
  }
  AppLog.Printf("5. Change file: %v", nginxConf)

  temp := activeHost
  activeHost = passiveHost
  passiveHost = temp
  checkUrl = strings.Replace(checkUrl, activeHost, passiveHost, -1)

  AppLog.Printf("12. URL to check set to: %v ",checkUrl)
}

/************************/
func reloadConfig(){
  AppLog.Printf("6. Running %v and waiting for it to finish...",nginxCmd)
  cmd := exec.Command(nginxCmd,"reload")
  status := cmd.Run()
  AppLog.Printf("7. Command finished with error: %v", status)
}

/************************/
func runAppCustom(){

}

/************************/
func readConfig() (string, string, string, string, int, string){

  appConfigFile := os.Args[1]

  //filename is the path to the json config file
  configFile, err := ioutil.ReadFile(appConfigFile)
  if err != nil {
    AppLog.Printf("8: %v",err)
    os.Exit(1)
  }

  configOut := string(configFile)

  appconfig := AppConfiguration {}
  json.Unmarshal([]byte(configOut), &appconfig)

  interval, _ := strconv.Atoi(appconfig.Interval_sec)

  AppLog.Printf("10: %v %v %v %v %v %v",appconfig.Check_url, appconfig.Nginx_conf, appconfig.Active_host, appconfig.Passive_host, interval, appconfig.Nginx_cmd)
  AppLog.Printf("11. URL to check set to: %v ",appconfig.Check_url)
  return appconfig.Check_url, appconfig.Nginx_conf, appconfig.Active_host, appconfig.Passive_host, interval, appconfig.Nginx_cmd
}

/************************/
func init(){
  checkUrl , nginxConf, activeHost, passiveHost, interval, nginxCmd = readConfig()
}

/************************/
func main() {

   for ok := true; ok; ok = true {
     httpStatus := healthCheck()
     if httpStatus != HTTP_STATUS_OK {
       AppLog.Printf("13. %v http status code: %v",checkUrl,httpStatus)
       setActive()
       reloadConfig()
     }
     time.Sleep(time.Duration(interval) * time.Second)
   }


}
