package main

import (
	"context"
	"fmt"
	"time"
	"path/filepath"
	"flag"
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

func deletePod(clientset *kubernetes.Clientset, podName string, namespace string) {
	err := clientset.CoreV1().Pods(namespace).Delete(context.TODO(), podName, metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("Start to delete pod %q in namespace %q and wait for terminating...\n", podName, namespace)
		for {
			pod, e := clientset.CoreV1().Pods(namespace).Get(context.TODO(),podName, metav1.GetOptions{})
			if e != nil{
				break
			} else if pod.Status.Phase == "Terminating" {
				time.Sleep(1 * time.Second)
			}
		}
		fmt.Printf("Pod %q in namespace %q has been deleted safely.\n", podName, namespace)
	}
}

func main() {
	clientset,_,err := initClientsetConfig()
	if err != nil {
		panic(err.Error())
		return
	}
	deletePod(clientset,"gpu-pod1","default")
	deletePod(clientset,"gpu-pod2","default")
}

