package main

import (
	"context"
	"fmt"
	"os"
	"flag"
	"path/filepath"
	//"reflect"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/remotecommand"
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

func CopyIntoPod(clientset *kubernetes.Clientset,config *rest.Config,podName string, namespace string, containerName string, srcPath string, dstPath string) {

	// 加载本地文件
	localFile, err := os.Open(srcPath)
	if err != nil {
		fmt.Printf("Error opening local file: %s\n", err)
		return
	}
	defer localFile.Close()

	// 获取指定的Pod对象
	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error getting pod: %s\n", err)
		return
	}

	// 获取指定Pod中的指定容器
	var container *corev1.Container
	for _, c := range pod.Spec.Containers {
		if c.Name == containerName {
			container = &c
			break
		}
	}
	if container == nil {
		fmt.Printf("Container not found in pod\n")
		return
	}

	// 创建到Docker容器的文件流
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", containerName)

	req.VersionedParams(&corev1.PodExecOptions{
		Container: containerName,
		Command:   []string{"bash", "-c", "cat > " + dstPath},
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		fmt.Printf("Error creating executor: %s\n", err)
		return
	}

	// 在文件流中写入文件内容
	err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdin:  localFile,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    false,
	})
	if err != nil {
		fmt.Printf("Error streaming: %s\n", err)
		return
	}

	fmt.Printf("Local file %q has been copied to %q in the container %q of pod %q in namespace %q successfully!\n",srcPath,dstPath,containerName,podName,namespace)
}

func main() {
	clientset,config,err := initClientsetConfig()
	if err != nil {
		panic(err.Error())
		return
	}
	//fmt.Println(reflect.TypeOf(clientset))
	CopyIntoPod(clientset,config,"gpu-pod1","default","test-container-1", "./testGraph1.txt", "/root/testGraph1.txt")
	CopyIntoPod(clientset,config,"gpu-pod2","default","test-container-2", "./testGraph2.txt", "/GraphX/testGraph2.txt")
}
