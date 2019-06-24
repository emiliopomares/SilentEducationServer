package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"time"
	"io"
	"os"
	"strconv"
	"io/ioutil"
)

type BuildInfo struct {
	Build	int	`json:"build"`
}

type ConfigInfo struct {
	ConfigFile	string
	CheckURL	string
	LocalPath	string
}

var latestBuild int
var config *ConfigInfo

func MakeDefaultConfig() *ConfigInfo {
	NewConfig := &ConfigInfo{}
	NewConfig.ConfigFile = "./config.json"
	NewConfig.LocalPath = "./SilentDevice"
	NewConfig.CheckURL = "https://www.digitaldreamsinteractive.com/SEUpdater"
	//@TODO add environment variables check
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
	fmt.Println("1")
	for {
		fmt.Println("Current build version: ", latestBuild)
		fmt.Println("Checking for updates...")
		res, err := http.Get(config.CheckURL + "/currentBuildNumber")
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
					fmt.Println("New version available: build " + strconv.Itoa(newBuildInfo.Build) + ". Downloading and updating...")
					latestBuild = newBuildInfo.Build
					DownloadFile(config.CheckURL + "/currentBinary", config.LocalPath)
					saveConfigFile()
				} else {
					fmt.Println("No new versions available at the time...")
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
	CheckForUpdates(1 * time.Minute)
}
