package main

import (
	"context"
	"fmt"
	"flag"
	"time"
	"path/filepath"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
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

func createPod(clientset *kubernetes.Clientset,podName string,namespace string,containerName string,containerImage string,cmdlines []string) {

	// pod模板
	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: containerName, Image: containerImage, Command: cmdlines},
			},
		},
	}

	// 创建pod
	pod, err := clientset.CoreV1().Pods(namespace).Create(context.Background(),newPod, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	} else {
		for {
			fmt.Println("Start to Create Pod...Wait for Pod to Initilize completely...")
			creatingpod, e := clientset.CoreV1().Pods(namespace).Get(context.TODO(),podName, metav1.GetOptions{})
			if e != nil{
				fmt.Println("Failed to Create Pod!\n")
				panic(err.Error())
				break
			} else if creatingpod.Status.Phase == "Running" {
				fmt.Println("Pod has been created successfully and its Stauts is Running.")
				fmt.Printf("Pod       Name  : %s\n",podName)
				fmt.Printf("Namespace       : %s\n",namespace)
				fmt.Printf("Container Name  : %s\n",containerName)
				fmt.Printf("Container Image : %s\n",containerImage)
				break
			} else if creatingpod.Status.Phase == "Pending" {
				fmt.Println("Pod has been created successfully and its Stauts is Pending.\n")
				fmt.Printf("Pod       Name  : %s\n",podName)
				fmt.Printf("Namespace       : %s\n",namespace)
				fmt.Printf("Container Name  : %s\n",containerName)
				fmt.Printf("Container Image : %s\n",containerImage)
				break
			} else if creatingpod.Status.Phase == "ContainerCreating" {
				fmt.Println("Pod is being created and its Status is ContainerCreating....Please Wait...\n")
				time.Sleep(1 * time.Second)
			} else if creatingpod.Status.Phase == "PodInitializing" {
				fmt.Println("Pod is being created and its Status is PodInitializing....Please Wait...\n")
				time.Sleep(1 * time.Second)
			} else {
				fmt.Printf("Some Errors Occured During Creating Pod...\nPod Status is %s\n",pod.Status.Phase)
				// 此处删除创建失败的pod
				break
			}
		}
	}
}

func main() {
	clientset,_,err := initClientsetConfig()
	if err != nil {
		panic(err.Error())
		return
	}
	cmdlines := [2]string{"sleep","10000"}
	createPod(clientset,"gpu-pod3","default","test-container-3","kssamwang/gx-plug:v3.0-GraphX",cmdlines[:]) 
}

