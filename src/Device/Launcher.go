package main

import (
        "fmt"
        "encoding/json"
        "net/http"
        "time"
        "io"
        "os"
	"os/exec" 
       "strconv"
        "io/ioutil"
)

// schema for the prototype: simple PSK, better than nothing!
const SilentEducationPSK = "4baUV/2T=1a4nGrDS43FGnv6100asRNa35+shd/2b42300aNUFHsdn2m3iUJ86B/d2"

type BuildInfo struct {
        Build   int     `json:"build"`
}

type ConfigInfo struct {
        ConfigFile      string
        BaseURL        string
        LocalPath       string
	APIRoot		string
	HotReload	bool
}

var latestBuild int
var config *ConfigInfo

func MakeDefaultConfig() *ConfigInfo {
        NewConfig := &ConfigInfo{}
        NewConfig.ConfigFile = "./version.json"
        NewConfig.LocalPath = "./Device"
        NewConfig.BaseURL = "https://www.digitaldreamsinteractive.com"
	NewConfig.APIRoot = "/SEUpdater"
	NewConfig.HotReload = true
	if(os.Getenv("UPDATER_BASE_URL") != "") {
		NewConfig.BaseURL = os.Getenv("UPDATER_BASE_URL")
	}
	if(os.Getenv("HOT_RELOAD") == "FALSE") {
		NewConfig.HotReload = false
	} 

        return NewConfig
}

func DownloadFile(url string, localPath string) error {
        res, errGet := http.Get(url)
        if errGet != nil {
                fmt.Println("Error: " + errGet.Error())
                return errGet
        } else {
                defer res.Body.Close()
                out, errFile := os.Create(localPath)
                if errFile != nil {
                        fmt.Println("Error: " + errFile.Error())
                        return errFile
                } else {
                        defer out.Close()
                        _, copyErr := io.Copy(out, res.Body)
                        return copyErr
                }
        }
        return nil
}

func CheckForUpdates(interval time.Duration) {
        for {
                //fmt.Println("Current build version: ", latestBuild)
                //fmt.Println("Checking for updates...")
                res, err := http.Get(config.BaseURL + "/" + config.APIRoot + "/currentBuildNumber")
                if(err == nil) {
                        defer res.Body.Close()
                        body, _ := ioutil.ReadAll(res.Body)
                        fmt.Println("This came from the server: ", body)
                        var newBuildInfo BuildInfo
                        unmarshalErr := json.Unmarshal(body, &newBuildInfo)
                        if unmarshalErr != nil {
                                fmt.Println("Error: " + err.Error())
                        } else {
                                if newBuildInfo.Build > latestBuild {
                                        //fmt.Println("New version available: build " + strconv.Itoa(newBuildInfo.Build) + ". Downloading and updating...")
                                        latestBuild = newBuildInfo.Build
                                        DownloadFile(config.BaseURL + "/" + config.APIRoot + "/currentBinary", config.LocalPath)
                                        fmt.Println("")
					if(config.HotReload) {
						fmt.Println("      ####### New version downloaded and installed. Rebooting device not. #######       ")
						client := &http.Client{}
						req, _ := http.NewRequest("GET", "http://localhost:8000/kill", nil)
						req.Header.Set("psk", SilentEducationPSK)
						_, _ = client.Do(req)
					} else {
						fmt.Println("      ####### New version downloaded and installed. Please reboot device. #######       ")
					}
					fmt.Println("")
					saveConfigFile()
                                } else {
                                        //fmt.Println("No new versions available at the time...")
                                }
                        }
                } else {
                        fmt.Println("Error: " + err.Error())
                }
                time.Sleep(interval)
        }
}

func loadConfigFile() {
        file, err := ioutil.ReadFile(config.ConfigFile)
        var curBuild BuildInfo
        if err != nil {
                fmt.Println("Error: no configuration file found")
                latestBuild = 0
        } else {
                unmarshalErr := json.Unmarshal(file, &curBuild)
                if unmarshalErr != nil {
                        fmt.Println("Error: " + unmarshalErr.Error())
                } else {
                        latestBuild = curBuild.Build
                        fmt.Println("Latest build " + strconv.Itoa(latestBuild) + " read from file")
                }
        }
}

func saveConfigFile() {
        var saveBuild BuildInfo
        saveBuild.Build = latestBuild
        marshaledData, err := json.Marshal(saveBuild)
        if err != nil {
                fmt.Println("Error: " + err.Error())
        } else {
                fileErr := ioutil.WriteFile(config.ConfigFile, marshaledData, 0644)
                if fileErr != nil {
                        fmt.Println("Error: " + err.Error())
                } else {
                        fmt.Println("Build config file save successfully")
                }
        }
}

func main() {
	config = MakeDefaultConfig()
        loadConfigFile()
        go CheckForUpdates(1 * time.Minute)
	for {
		fmt.Println("Starting device...")
		cmd := exec.Command("./Device")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	}
}
