package main

import (
	"context"
	"fmt"
	"path/filepath"
	"flag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func getNodeInfo() {
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

	// 获取node信息
	nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, node := range nodeList.Items {
		fmt.Printf("=========================\nName    :%s\nPhase   :%s\nAddress :%s\nOSImage :%s\nk8sVer  :%s\nOS      :%s\nArch    :%s\n",
			node.Name,
			node.Status.Phase,
			node.Status.Addresses,
			node.Status.NodeInfo.OSImage,
			node.Status.NodeInfo.KubeletVersion,
			node.Status.NodeInfo.OperatingSystem,
			node.Status.NodeInfo.Architecture,
		)
	}
}

func main() {
	getNodeInfo()
}
