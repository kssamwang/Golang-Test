package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"golang.org/x/crypto/ssh/terminal"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
)

func initClientsetConfig() (*kubernetes.Clientset,*rest.Config,error){
	// 连接集群
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// 使用kubeconfig中的当前上下文,加载配置文件
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	//fmt.Println(reflect.TypeOf(config))
	// 创建clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset,config,err
}

func execInPod(clientset *kubernetes.Clientset,config *rest.Config,podName string,namespace string,cmdlines []string){
	req := clientset.CoreV1().RESTClient().Post().
	Resource("pods").
	Name(podName).
	Namespace(namespace).
	SubResource("exec").
	VersionedParams(&corev1.PodExecOptions{
		Command: cmdlines,
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
        outfile,err := os.Create("result.txt")
        if err != nil {
                fmt.Println("open failed.",err)
                return
        }
        if err = exec.Stream(remotecommand.StreamOptions{
                Stdin: screen,
                Stdout: outfile,
                Stderr: screen,
                Tty:    false,
        }); err != nil {
                fmt.Print(err)
        }
}

func main() {
        clientset,config,err := initClientsetConfig()
        if err != nil {
                panic(err.Error())
                return
        }
        cmdlines := [6]string{"/GraphX/bin/algo_BellmanFordGPUTest","testGraph.txt","100","2000",">>","result.txt"}
        execInPod(clientset,config,"gpu-pod","default",cmdlines[:])
}
