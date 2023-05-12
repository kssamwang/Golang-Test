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
	"github.com/fsouza/go-dockerclient"
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

func getPodInfo(clientset *kubernetes.Clientset,podName string,namespace string) (*corev1.Pod,error){
	pod,err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	return pod,err
}

func updateResources(containerName string, resources *docker.UpdateContainerOptions) error {
	// 创建 Docker client
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}

	// 通过容器名获取容器对象
	container, err := client.InspectContainer(containerName)
	if err != nil {
		return err
	}

	// 更新容器资源
	err = client.UpdateContainer(container.ID, *resources)
	if err != nil {
		return err
	}

	return nil
}

func updateK8SDockerResources(containerName string,podName string,namespace string,idx int){
	clientset,_,err1 := initClientsetConfig()
	if err1 != nil {
		panic(err1.Error())
		return
	}
	/*	!!!!!!!!  注意  !!!!!!!!!!!
		通过k8s创建的docker容器的真实名称docker name
		不是 yaml 文件中或 client-go 调用函数创建容器时指定的container name
		而是一串包含了容器上层pod和namespace信息的完整名称，格式为：
		k8s_{container_name}_{pod_name}_{namespace}_{pod_uid}_{container在pod中序号,0开始}
		因为uid字段是随机生成的，不能在创建容器时就拿到，所以需要创建容器后，再取pod信息获得uid
		最后拼接出完整的容器docker name
	*/
	pod,_ := getPodInfo(clientset,podName,namespace)
	var uid = pod.ObjectMeta.UID
	var dockerNamePattern = "k8s_%s_%s_%s_%s_%d"
	var dockerName = fmt.Sprintf(dockerNamePattern,containerName,podName,namespace,uid,idx)
	res := docker.UpdateContainerOptions{
		CPUShares : int(8192),
		CPUQuota : int(100000000),
		Memory : int(20000000000),
	}
	err2 := updateResources(dockerName,&res)
	if err2 != nil {
		panic(err2.Error())
	}

}

func main(){
	updateK8SDockerResources("test-container-1","gpu-pod1","default",0)
}
