package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func printPodInfo(p *corev1.Pod){
	fmt.Println("nodename:", p.Spec.NodeName)
	fmt.Println("status:", p.Status.Conditions[0].Status)
	fmt.Println("Phase:", p.Status.Phase)
	fmt.Println("HostIP:", p.Status.HostIP)
	fmt.Println("PodIP:", p.Status.PodIP)
	fmt.Println("name:", p.ObjectMeta.Name)
	fmt.Println("GenerateName:", p.ObjectMeta.GenerateName)
	fmt.Println("Namespace:", p.ObjectMeta.Namespace)
	fmt.Println("SelfLink:", p.ObjectMeta.SelfLink)
	fmt.Println("UID:", p.ObjectMeta.UID)
	fmt.Println("ResourceVersion:", p.ObjectMeta.ResourceVersion)
	fmt.Println("Generation:", p.ObjectMeta.Generation)
	fmt.Println("CreationTimestamp:", p.ObjectMeta.CreationTimestamp.Format("2006-01-02 15:04:05"))
	fmt.Println("Labels:", p.ObjectMeta.Labels)
	if p.Status.Conditions[0].Status == "True" {
		fmt.Println("Ready:", p.Status.ContainerStatuses[0].Ready)
		fmt.Println("Started:", *p.Status.ContainerStatuses[0].Started)
		fmt.Println("重启次数:", p.Status.ContainerStatuses[0].RestartCount)
		switch true {
			case p.Status.ContainerStatuses[0].State.Running != nil:
			{
				fmt.Println("容器状态:running")
			}
			case p.Status.ContainerStatuses[0].State.Waiting != nil:
			{
				fmt.Println("容器状态:waiting")
			}
			case p.Status.ContainerStatuses[0].State.Terminated != nil:
			{
				fmt.Println("容器状态:terminated")
			}
		}
	}
}
func initClientset() (*kubernetes.Clientset,error){
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
	// 创建clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset,err
}

func getPodListInNamespace(clientset *kubernetes.Clientset,namespace string) {
	pods, _ := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	for _, p := range pods.Items {
		printPodInfo(&p)
	}
}

func getPodInfo(clientset *kubernetes.Clientset,podName string,namespace string) {
	pod,err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	} else {
		printPodInfo(pod)
	}
}

func main() {
	clientset,err := initClientset()
	if err != nil {
		panic(err.Error())
		return
	}
	getPodListInNamespace(clientset,"kube-system")
	getPodInfo(clientset,"gpu-pod1","default")
	getPodInfo(clientset,"gpu-pod2","default")
	getPodInfo(clientset,"gpu-pod-master","default")
}
