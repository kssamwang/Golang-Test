package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/remotecommand"
)

func CopyIntoPod(podName string, namespace string, containerName string, srcPath string, dstPath string) {
	// Get the default kubeconfig file
	kubeConfig := filepath.Join(homedir.HomeDir(), ".kube", "config")

	// Create a config object using the kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		fmt.Printf("Error creating config: %s\n", err)
		return
	}

	// Create a Kubernetes client
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error creating client: %s\n", err)
		return
	}

	// Open the file to copy
	localFile, err := os.Open(srcPath)
	if err != nil {
		fmt.Printf("Error opening local file: %s\n", err)
		return
	}
	defer localFile.Close()

	pod, err := client.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error getting pod: %s\n", err)
		return
	}

	// Find the container in the pod
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

	// Create a stream to the container
	req := client.CoreV1().RESTClient().Post().
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

	// Create a stream to the container
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

	fmt.Println("File copied successfully")
}

func main(){
	CopyIntoPod("gpu-pod1","default","test-container-1", "./testGraph1.txt", "/root/testGraph1.txt")
	CopyIntoPod("gpu-pod2","default","test-container-2", "./testGraph2.txt", "/GraphX/testGraph2.txt")
}