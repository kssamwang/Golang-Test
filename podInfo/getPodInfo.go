package main

import (
	"context"
	"flag"
	"fmt"
	//"reflect"
	"path/filepath"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func printPodInfo(p *corev1.Pod){
	fmt.Println("=========================================================")
	fmt.Println("Pod Name          : ", p.ObjectMeta.Name)
	fmt.Println("Namespace         : ", p.ObjectMeta.Namespace)
	fmt.Println("Pod Status Phase  : ", p.Status.Phase)
	fmt.Println("Pod Status Ready  : ", p.Status.Conditions[0].Status)
	fmt.Println("Node              : ", p.Spec.NodeName)
	fmt.Println("Host IP           : ", p.Status.HostIP)
	fmt.Println("Pod  IP           : ", p.Status.PodIP)
	fmt.Println("SelfLink          : ", p.ObjectMeta.SelfLink)
	fmt.Println("UID               : ", p.ObjectMeta.UID)
	fmt.Println("Resource Version  : ", p.ObjectMeta.ResourceVersion)
	fmt.Println("Generate Name     : ", p.ObjectMeta.GenerateName)
	fmt.Println("Generation        : ", p.ObjectMeta.Generation)
	fmt.Println("Created Timestamp : ", p.ObjectMeta.CreationTimestamp.Format("2006-01-02 15:04:05"))
	fmt.Println("Labels            : ", p.ObjectMeta.Labels)
	if p.Status.Conditions[0].Status == "True" {
		fmt.Println("Ready             : ", p.Status.ContainerStatuses[0].Ready)
		fmt.Println("Started           : ", *p.Status.ContainerStatuses[0].Started)
		fmt.Println("Restart Times     : ", p.Status.ContainerStatuses[0].RestartCount)
		switch true {
			case p.Status.ContainerStatuses[0].State.Running != nil:
			{
				fmt.Println("Container Status  :  Running")
			}
			case p.Status.ContainerStatuses[0].State.Waiting != nil:
			{
				fmt.Println("Container Status  :  Waiting")
			}
			case p.Status.ContainerStatuses[0].State.Terminated != nil:
			{
				fmt.Println("Container Status  :  Terminated")
			}
		}
	}
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

func getPodListInNamespace(clientset *kubernetes.Clientset,namespace string) (*corev1.PodList,error){
	pods,err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
                panic(err.Error())
        }
	return pods,err
}

func getPodInfo(clientset *kubernetes.Clientset,podName string,namespace string) (*corev1.Pod,error){
	pod,err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	return pod,err
}

func main() {
	clientset,_,err := initClientsetConfig()
	if err != nil {
		panic(err.Error())
		return
	}
	pods,_ := getPodListInNamespace(clientset,"kube-system")
	for _,p := range pods.Items {
		printPodInfo(&p)
	}
	pod1,_ := getPodInfo(clientset,"gpu-pod1","default")
	printPodInfo(pod1)
	pod2,_ := getPodInfo(clientset,"gpu-pod2","default")
	printPodInfo(pod2)
	pod3,_ := getPodInfo(clientset,"gpu-pod-master","default")
	printPodInfo(pod3)
}
