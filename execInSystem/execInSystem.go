package main
import (
	"os"
	"os/exec"
	"fmt"
	"flag"
	"strings"
)

func execInSystem(command *string,args *string) {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-cmd <command>] [-args <the arguments (separated by spaces)>]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	fmt.Println("[Running] ",*command,*args)
	var argArray []string
	if *args != "" {
		argArray = strings.Split(*args, " ")
	} else {
		argArray = make([]string, 0)
	}
	cmd := exec.Command(*command, argArray...)
	buf, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "The command failed to perform: %s (Command: %s, Arguments: %s)", err, *command, *args)
		return
	}
	fmt.Fprintf(os.Stdout, "Result: %s", buf)
}

func main() {
	command := flag.String("cmd", "ping", "Set the command.")
	args := flag.String("args", "www.baidu.com -c 10","Set the args. (separated by spaces)")
	execInSystem(command,args)
}
