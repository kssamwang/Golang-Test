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
	"k8s.io/apimachinery/pkg/api/resource"
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
			/* 此处不用写tolerations 
			   只要master被打了参与pod调度的标记
			   client-go创建的pod也可以被调度到master上
			*/
			Containers: []corev1.Container{
				{
					Name: containerName,
					Image: containerImage,
					Command: cmdlines,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							/* spec.container[].resources.requests.cpu 参数值
							   这个参数值会被转化为core数，100m => 0.1，然后*1024，
							   结果作为 docker run 的 --cpu-shares 参数
							   该参数不能理解为虚拟cpu核心数，而应该理解为Docker Daemon
							   为多个运行中的Docker容器分配主机cpu资源时采用的资源配比
							   例如：
							   	容器A --cpu-shares = 1024
								容器B --cpu-shares = 2048
								则意味着Docker按 1:2 为容器A与B分配CPU计算资源
								而不是A可以使用1核，B可以使用2核
							*/
							"cpu":    resource.MustParse("4"),
							/* spec.container[].resources.requests.memory 参数值
                                                           只提供给Kubernetes调度器作为调度和管理依据
                                                           是容器上台时分配的内存，不能大于limit值
                                                           不会作为任何参数传递给Docker
                                                        */
							"memory": resource.MustParse("16000M"),
						},
						Limits: corev1.ResourceList{
							/* spec.container[].resources.limits.cpu 参数值
							   该数值 *1000 转化为millicore数，1 => 1000,100m => 100
							   然后 先 * 100000，再 / 1000
							   结果作为 docker run 的 --cpu-quota 参数
							   该参数配合 另一个参数 --cpu-period 使用，默认值100000
							   表示Docker Daemon重新为容器计量cpu使用时间的周期，单位us
							   该参数不能理解为虚拟cpu核心数，而应该理解为：
							   在 --cpu-period us 的时间内，Docker 最多为该容器分配--cpu-quota的单核cpu使用时间
							   因此，spec.container[].resources.limits.cpu = 8
							   可以解释为在 100000 us 的分配周期内，该容器最多获得 800000 us的单cpu使用时间
							   实际上是通过cpu的时间片使得容器等价于在物理上限制使用8核
							   不是同一时刻容器只能使用主机上8个核
							*/
							"cpu" :    resource.MustParse("8"),
							/* spec.container[].resources.limits.memory 参数值
							   该参数值换算成字节数
							   作为 docker run 的 --memory 参数，表示容器最多使用的内存
							   如果容器使用的内存超过这个值，容器会被kill
							   如果容器有restart策略，则会被kubelet重启
							   因此准确估算容器运行程序需要的内存非常重要
							   K/M/G分别表示十进制的KB/MB/GB，Ki/Mi/Gi分别表示二进制的KB/MB/GB
							*/
							"memory" : resource.MustParse("20000M"),
							/* GPU资源的数值 */
							"nvidia.com/gpu" : resource.MustParse("1"),
							"nvidia.com/gpumem" : resource.MustParse("3000"),
							"nvidia.com/gpucores" : resource.MustParse("90"),
						},
					},
				},
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

/* !!!!!!!! 注意 !!!!!!!!!!!
   容器中使用top或free查看cpu/内存
   实际上还是查看的主机信息
   这一块是Docker与主机隔离的不完善之处
   Docker容器真正可用的CPU资源和内存
   需要用docker inspect 容器ID 查看
*/

func main() {
	clientset,_,err := initClientsetConfig()
	if err != nil {
		panic(err.Error())
		return
	}
	cmdlines := [2]string{"sleep","10000"}
	createPod(clientset,"gpu-pod2","default","test-container-3","kssamwang/gx-plug:v3.0-GraphX",cmdlines[:])
}

