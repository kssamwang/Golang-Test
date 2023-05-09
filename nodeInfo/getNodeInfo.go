package main

import (
	"context"
	"fmt"
	"flag"
	"path/filepath"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

func getNodeInfo(clientset *kubernetes.Clientset) {
	// 获取node信息
	nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, node := range nodeList.Items {
		fmt.Printf("=========================\nName    : %s\nAddress : %s\nOSImage : %s\nk8sVer  : %s\nOS      : %s\nArch    : %s\nKernel  : %s\nCreated : %s\nNowtime : %s\nCPU     : %s\nFreeCPU : %s\nDocker  : %s\nStatus  : %s\nMemory  : %s\n",
			node.Name,
			node.Status.Addresses,
			node.Status.NodeInfo.OSImage,
			node.Status.NodeInfo.KubeletVersion,
			node.Status.NodeInfo.OperatingSystem,
			node.Status.NodeInfo.Architecture,
			node.Status.NodeInfo.KernelVersion,
			node.CreationTimestamp,
			node.Status.Conditions[0].LastHeartbeatTime,
			node.Status.Capacity.Cpu(),
			node.Status.Allocatable.Cpu().String(),
			node.Status.NodeInfo.ContainerRuntimeVersion,
			node.Status.Conditions[len(node.Status.Conditions)-1].Type,
			node.Status.Allocatable.Memory().String(),
		)
	}
}

func main() {
	clientset,_,err := initClientsetConfig()
	if err != nil {
		panic(err.Error())
		return
	}
	getNodeInfo(clientset)
}
