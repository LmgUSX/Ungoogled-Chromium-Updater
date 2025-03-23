package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gocolly/colly/v2"
)

type VersionInfo struct {
	Platform string
	Version  string
}

func main() {
	whereAmI := whereAmI()
	fmt.Println(whereAmI)

	err := delInstall(whereAmI)
	if err != nil {
		fmt.Printf("Error deleting file: %v\n", err)
	}

	versionInstalled := getVersionInstalled()

	versionInstalled = "134.0.6998"
	newVS := crawler()

	newVS_2 := newVS
	sha256 := crawler_sha(newVS)

	if strings.Contains(newVS, "-") {

		index := strings.Index(newVS, "-")
		newVS = newVS[:index]
	}

	downloadURL := "https://github.com/ungoogled-software/ungoogled-chromium-windows/releases/download/" + newVS_2 + ".1/ungoogled-chromium_" + newVS_2 + ".1_installer_x64.exe"
	install_path := whereAmI + `\ungoogled-chromium_installer.exe`

	if versionInstalled == newVS {
		fmt.Println("Up to Date")
		os.Exit(0)
	} else {
		err := downloadFile(downloadURL)
		if err != nil {
			println("Error downloading file:", err.Error())
			return
		}

		println("Download completed successfully.")

		hash, err := calculateSHA256(install_path)

		if err != nil {
			fmt.Println("Fehler beim Berechnen des SHA-256 Hashes:", err)
			return
		}


		if hash == sha256 {
			fmt.Println("Hello World")
		} else {
			os.Exit(1)
		}

		cmd := exec.Command(install_path)
		fmt.Println(cmd)
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error executing command: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(out)
	}
	err = delInstall(whereAmI)
	if err != nil {
		fmt.Printf("Error deleting file: %v\n", err)
	}

}


func downloadFile(url string) error {

	client := http.DefaultClient
	fmt.Println(url)


	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("failed rsp")
		return err
	}

	defer resp.Body.Close()


	if resp.StatusCode != 200 {
		fmt.Println(resp.StatusCode)
		return io.EOF
	}


	outputFile, err := os.Create("ungoogled-chromium_installer.exe")
	if err != nil {
		fmt.Println("can't create file")
		return err
	}
	defer outputFile.Close()


	_, err = io.Copy(outputFile, resp.Body)
	if err != nil {
		fmt.Println("fill exe")
		return err
	}

	return nil
}


func crawler() string {

	url := "https://ungoogled-software.github.io/ungoogled-chromium-binaries/"


	var windowsVersions []VersionInfo

	c := colly.NewCollector(
		colly.AllowedDomains("ungoogled-software.github.io"),
	)

	c.OnHTML("table tbody tr", func(e *colly.HTMLElement) {
		platform := e.ChildText("td:first-child strong a")
		version := e.ChildText("td:nth-child(2) a")
		if platform != "" && strings.Contains(platform, "Windows") {
			windowsVersions = append(windowsVersions, VersionInfo{
				Platform: platform,
				Version:  version,
			})
		}
	})


	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Request URL: %s failed with response: %v\n", r.Request.URL, r)
		log.Printf("Error: %v", err)
	})

	var str string

	c.OnScraped(func(r *colly.Response) {

		if len(windowsVersions) == 0 {
			fmt.Println("No Windows versions found")
			return
		}

		for _, info := range windowsVersions {
			if strings.Contains(string(info.Platform), "64-bit") && !strings.Contains(string(info.Platform), "ARM") {
				str = info.Version
			}
		}
	})

	err := c.Visit(url)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(windowsVersions)
	return str
}

func crawler_sha(str string) string {
	url_installer_page := "https://ungoogled-software.github.io/ungoogled-chromium-binaries/releases/windows/64bit/" + str
	c := colly.NewCollector(
		colly.AllowedDomains("ungoogled-software.github.io"),
	)
	var SHA256 []string

	c.OnHTML("h2:nth-of-type(2) + ul > li", func(e *colly.HTMLElement) {
		SHA256 = append(SHA256, e.ChildText("ul li:nth-child(3) code"))
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Request URL: %s failed with response: %v\n", r.Request.URL, r)
		log.Printf("Error: %v", err)
	})
	err := c.Visit(url_installer_page)
	if err != nil {
		log.Fatal(err)
	}

	return SHA256[0]
}

func getVersionInstalled() string {
	
	chromePath := `C:\Users\admin\AppData\Local\Chromium\Application\chrome.exe`
	chromePath = strings.ReplaceAll(chromePath, `\`, `\\`)
	
	cmd := exec.Command("wmic",
		"datafile",
		"where",
		fmt.Sprintf(`name='%s'`, chromePath), 
		"get",
		"version",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}

	version := parseWMICOutput(string(out))
	if version == "" {
		fmt.Println("Failed to parse Chrome version")
		os.Exit(1)
	}

	return version
}

func parseWMICOutput(output string) string {
	lines := strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n")
	for _, line := range lines {
		cleanLine := strings.TrimSpace(line)
		if cleanLine != "" && cleanLine != "Version" {
			return cleanLine
		}
	}
	return ""
}

func whereAmI() string {

	execPath := os.Args[0]
	absPath, err := filepath.Abs(execPath)
	if err != nil {
		absPath = execPath
	}

	executionDirectory := filepath.Dir(absPath)

	return executionDirectory
}

func delInstall(directoryPath string) error {
	filePath := filepath.Join(directoryPath, "ungoogled-chromium_installer.exe")
	_, err := os.Stat(filePath)
	if err != nil {
		return nil
	}
	err = os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}
	return nil
}


func calculateSHA256(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	hashInBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashInBytes), nil
}
