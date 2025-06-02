package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/labstack/gommon/log"
)

func main() {
	commit := getGitCommit()
	date := time.Now().Format("2006-01-02")
	version := "8.20"

	code := fmt.Sprintf(`package main

func init() {
	buildVersion = "%s"
	buildDate    = "%s"
	buildCommit  = "%s"
}
`, version, date, commit)

	err := os.WriteFile("build_info.go", []byte(code), 0644)
	if err != nil {
		log.Error("failed to write build info")
	}
}

func getGitCommit() string {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "N/A"
	}

	return strings.TrimSpace(string(out))
}
