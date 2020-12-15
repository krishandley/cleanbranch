package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey"
)

func printErr(err error) {
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
}

func gitDeleteBranch(branch string) {
	gitCheckout("master")
	fmt.Printf(" -> git branch -D %s", branch)
	cmd := exec.Command("git", "branch", "-D", branch)
	err := cmd.Run()
	printErr(err)
}

func gitUnsetUpstream(branch string) {
	fmt.Println(" -> git branch --unset-upstream")
	cmd := exec.Command("git", "branch", "--unset-upstream")
	err := cmd.Run()
	printErr(err)
}

func gitStatus() []byte {
	cmd := exec.Command("git", "status")
	out, err := cmd.Output()
	printErr(err)
	return out
}

func gitPull() {
	fmt.Println(" -> git pull")
	cmd := exec.Command("git", "pull")
	err := cmd.Run()
	printErr(err)
}

func gitResetHard(branch string) {
	org := fmt.Sprintf("origin/%s", branch)
	fmt.Printf(" -> git reset --hard %s\n", org)
	cmd := exec.Command("git", "reset", "--hard", org)
	err := cmd.Run()
	printErr(err)
}

func gitFetch() {
	fmt.Println(" -> git fetch")
	cmd := exec.Command("git", "fetch")
	err := cmd.Run()
	printErr(err)
}

func gitPrune() {
	fmt.Println(" -> git remote prune origin")
	cmd := exec.Command("git", "remote", "prune", "origin")
	err := cmd.Run()
	printErr(err)
}

func cleanBranchNames(lines []string) []string {
	ls := make([]string, 0)
	for _, v := range lines {
		v = strings.Replace(v, "* ", "", 1)
		v = strings.TrimSpace(v)
		if len(v) > 0 {
			ls = append(ls, v)
		}
	}
	return ls
}

func branches() []string {
	cmd := exec.Command("git", "branch")
	out, err := cmd.Output()
	printErr(err)
	win := strings.Replace(string(out), "\r\n", "\n", -1)
	lns := strings.Split(win, "\n")
	return cleanBranchNames(lns)
}

func gitCheckout(branch string) {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Run()
}

func createSelect(msg string, options []string) string {
	res := ""
	prompt := &survey.Select{
		Message: msg,
		Options: options,
	}
	survey.AskOne(prompt, &res)
	return res
}

func checkUpstream(str string, branch string) {
	r, _ := regexp.Compile("upstream is gone")
	if r.MatchString(str) {
		not := "Nothing"

		cmdDel := fmt.Sprintf("git branch -D %s", branch)
		del := fmt.Sprintf("Delete branch (%s)", cmdDel)

		cmdRem := "git branch --unset-upstream"
		rem := fmt.Sprintf("Remove remote (%s)", cmdRem)

		opt := []string{not, del, rem}
		msg := fmt.Sprintf("(%s) Upstream is gone", branch)
		res := createSelect(msg, opt)

		if res == del {
			gitDeleteBranch(branch)
		}

		if res == rem {
			gitUnsetUpstream(branch)
		}
	}
}

func checkDiverged(str string, branch string) {
	r, _ := regexp.Compile("diverged")
	if r.MatchString(str) {
		not := "Nothing"

		cmdRst := fmt.Sprintf("git reset --hard origin/%s", branch)
		rst := fmt.Sprintf("Reset WARNING: You will lose any local changes (%s)", cmdRst)

		opt := []string{not, rst}
		msg := fmt.Sprintf("(%s) Branch has diverged", branch)
		res := createSelect(msg, opt)

		if res == rst {
			gitResetHard(branch)
		}
	}
}

func checkBehind(str string, branch string) {
	r, _ := regexp.Compile("behind")
	if r.MatchString(str) {
		not := "Nothing"

		cmdPul := "git pull"
		pul := fmt.Sprintf("Pull (%s)", cmdPul)

		cmdRst := fmt.Sprintf("git reset --hard origin/%s", branch)
		rst := fmt.Sprintf("Reset (%s)", cmdRst)

		opt := []string{not, pul, rst}
		msg := fmt.Sprintf("(%s) Branch is behind", branch)
		res := createSelect(msg, opt)

		if res == pul {
			gitPull()
		}

		if res == rst {
			gitResetHard(branch)
		}
	}
}

func main() {
	gitFetch()
	gitPrune()
	brn := branches()

	for _, name := range brn {
		fmt.Printf("Checking branch %s\n", name)

		// Switch to branch
		gitCheckout(name)

		// Check the status
		out := gitStatus()
		str := string(out)

		// Perform checks
		checkUpstream(str, name)
		checkDiverged(str, name)
		checkBehind(str, name)
	}

	gitCheckout("master")
}
