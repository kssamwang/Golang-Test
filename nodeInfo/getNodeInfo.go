package main

import (
	"context"
	"fmt"
	"flag"
	"path/filepath"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func printNodeInfo(node *corev1.Node) {
	if node == nil{
		return
	}
	fmt.Println("============================================================================")
	fmt.Println("Node Name                : ",node.Name)
	fmt.Println("Node Status              : ",node.Status.Conditions[len(node.Status.Conditions)-1].Type)
	fmt.Println("Node IP Address          : ",node.Status.Addresses)
	fmt.Println("Operating System         : ",node.Status.NodeInfo.OperatingSystem)
	fmt.Println("Hardware Architecture    : ",node.Status.NodeInfo.Architecture)
	fmt.Println("Operating System Image   : ",node.Status.NodeInfo.OSImage)
	fmt.Println("Operating System Kernel  : ",node.Status.NodeInfo.KernelVersion)
	fmt.Println("Contain Runtime Version  : ",node.Status.NodeInfo.ContainerRuntimeVersion)
	fmt.Println("Kubernetes Version       : ",node.Status.NodeInfo.KubeletVersion)
	fmt.Println("Creation Timestamp       : ",node.CreationTimestamp)
	fmt.Println("Last Heartbeat Timestamp : ",node.Status.Conditions[0].LastHeartbeatTime)
	fmt.Println("CPU Capacity Kernels     : ",node.Status.Capacity.Cpu())
	fmt.Println("Allocatable CPU Kernels  : ",node.Status.Allocatable.Cpu().String())
	fmt.Println("Allocatable Memory Size  : ",node.Status.Allocatable.Memory().String())
}

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

func getNodeInfo(clientset *kubernetes.Clientset,nodeName string)(*corev1.Node,error) {
	// 获取node信息
	nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, node := range nodeList.Items {
		if node.Name == nodeName {
			return &node,err
		}
	}
	return nil,err
}

func getNodeListInfo(clientset *kubernetes.Clientset)(*corev1.NodeList,error) {
	// 获取node信息
        nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
        if err != nil {
                panic(err)
        }
	return nodeList,err
}

func main() {
	clientset,_,err := initClientsetConfig()
	if err != nil {
		panic(err.Error())
		return
	}
	node,_ := getNodeInfo(clientset,"master")
	printNodeInfo(node)
	nodeList,_ := getNodeListInfo(clientset)
	for _,n := range nodeList.Items {
		printNodeInfo(&n)
	}
}
