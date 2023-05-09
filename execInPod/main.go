package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"golang.org/x/crypto/ssh/terminal"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name("gpu-pod").
		Namespace("default").
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: []string{"nvidia-smi"},
			Stdin:   true,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if !terminal.IsTerminal(0) || !terminal.IsTerminal(1) {
		fmt.Errorf("stdin/stdout should be terminal")
	}
	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		fmt.Println(err)
	}
	defer terminal.Restore(0, oldState)
	screen := struct {
			io.Reader
			io.Writer
	}{os.Stdin, os.Stdout}
	if err = exec.Stream(remotecommand.StreamOptions{
		Stdin: screen,
		Stdout: screen,
		Stderr: screen,
		Tty:    false,
	}); err != nil {
		fmt.Print(err)
	}
}
