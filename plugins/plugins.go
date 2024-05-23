package plugins

import (
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"log"
	"os"
	"strconv"
)

// 插件名称
const Name = "metrics-scheduler-plugin"

var CPUWeight = int64(1)
var MemoryWeight = int64(2)

func init() {
	var err error
	CPUWeightStr := os.Getenv("CPU_WEIGHT")
	if CPUWeightStr != "" {
		CPUWeight, err = strconv.ParseInt(os.Getenv("CPU_WEIGHT"), 10, 64)
		if err != nil {
			log.Fatalf("get CPU_WEIGHT error: %s", err)
		}
	}
	MemoryWeightStr := os.Getenv("MEMORY_WEIGHT")
	if MemoryWeightStr != "" {
		MemoryWeight, err = strconv.ParseInt(os.Getenv("MEMORY_WEIGHT"), 10, 64)
		if err != nil {
			log.Fatalf("get CPU_WEIGHT error: %s", err)
		}
	}
}

type MetricsSchedulerPlugin struct {
	handle framework.Handle
}

func (s *MetricsSchedulerPlugin) Name() string {
	return Name
}

func (s *MetricsSchedulerPlugin) Score(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) (int64, *framework.Status) {
	config, err := rest.InClusterConfig()
	log.Printf("get InclusterConfig success")
	if err != nil {
		log.Fatalf("get InclusterConfig error: %s", err)
		return -1, framework.NewStatus(framework.Error, "get InclusterConfig error")
	}
	// 创建metrics client
	metricsClient, err := metrics.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error building metrics client: %s", err.Error())
		return -1, framework.NewStatus(framework.Error, "Error building metrics client")
	}
	// 创建Kubernetes客户端集
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %s", err.Error())
		return -1, framework.NewStatus(framework.Error, "Error building kubernetes clientset")
	}
	node, err := clientset.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Error getting node info %s: %s", nodeName, err.Error())
		return -1, framework.NewStatus(framework.Error, "Error getting node info")
	}
	nodeMetrics, err := metricsClient.MetricsV1beta1().NodeMetricses().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Error getting node metrics %s: %s", nodeName, err.Error())
		return -1, framework.NewStatus(framework.Error, "Error getting node metrics")
	}
	totalMemory := node.Status.Allocatable.Memory().Value()
	usedMemory := nodeMetrics.Usage.Memory().Value()
	totalCPU := node.Status.Allocatable.Cpu().Value()
	usedCPU := nodeMetrics.Usage.Cpu().Value()
	MemoryUsedPercentage := float64(usedMemory) / float64(totalMemory)
	CPUUsedPercentage := float64(usedCPU) / float64(totalCPU)
	log.Printf("node %s memory used %d, total %d, percentage %f\n", nodeName, usedMemory, totalMemory, MemoryUsedPercentage)
	log.Printf("node %s cpu used %d, total %d, percentage %f\n", nodeName, usedCPU, totalCPU, CPUUsedPercentage)
	score := ((1-MemoryUsedPercentage)*float64(MemoryWeight) + (1-CPUUsedPercentage)*float64(CPUWeight)) / float64(MemoryWeight+CPUWeight) * 100
	log.Printf("node %s score %f intscore %d\n", nodeName, score, int64(score))
	return int64(score), framework.NewStatus(framework.Success, "")
}
func (p *MetricsSchedulerPlugin) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// type PluginFactory = func(configuration runtime.Object, f framework.Handle) (framework.Plugin, error)
func New(configuration runtime.Object, f framework.Handle) (framework.Plugin, error) {
	return &MetricsSchedulerPlugin{}, nil
}
