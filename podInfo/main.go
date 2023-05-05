package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "config文件绝对路径")
	}
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Println("err:", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	pods, _ := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	for _, n := range pods.Items {
		fmt.Println("nodename:", n.Spec.NodeName)
		fmt.Println("status:", n.Status.Conditions[0].Status)
		fmt.Println("Phase:", n.Status.Phase)
		fmt.Println("HostIP:", n.Status.HostIP)
		fmt.Println("PodIP:", n.Status.PodIP)
		fmt.Println("name:", n.ObjectMeta.Name)
		fmt.Println("GenerateName:", n.ObjectMeta.GenerateName)
		fmt.Println("Namespace:", n.ObjectMeta.Namespace)
		fmt.Println("SelfLink:", n.ObjectMeta.SelfLink)
		fmt.Println("UID:", n.ObjectMeta.UID)
		fmt.Println("ResourceVersion:", n.ObjectMeta.ResourceVersion)
		fmt.Println("Generation:", n.ObjectMeta.Generation)
		fmt.Println("CreationTimestamp:", n.ObjectMeta.CreationTimestamp.Format("2006-01-02 15:04:05"))
		fmt.Println("Labels:", n.ObjectMeta.Labels)
		if n.Status.Conditions[0].Status == "True" {
			fmt.Println("Ready:", n.Status.ContainerStatuses[0].Ready)
			fmt.Println("Started:", *n.Status.ContainerStatuses[0].Started)
			fmt.Println("重启次数:", n.Status.ContainerStatuses[0].RestartCount)
			switch true {
				case n.Status.ContainerStatuses[0].State.Running != nil:
				{
					fmt.Println("容器状态:running")
				}
				case n.Status.ContainerStatuses[0].State.Waiting != nil:
				{
					fmt.Println("容器状态:waiting")
				}
				case n.Status.ContainerStatuses[0].State.Terminated != nil:
				{
					fmt.Println("容器状态:terminated")
				}
			}
		}
	}
}
