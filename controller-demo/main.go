package main

import (
	"flag"
	kubeInformer "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientset "k8s.io/controller-demo/pkg/generated/clientset/versioned"
	informers "k8s.io/controller-demo/pkg/generated/informers/externalversions"
	"k8s.io/controller-demo/signals"
	"k8s.io/klog/v2"
	"time"
)

var (
	masterUrl  = ""
	kubeConfig = ""
)

func main() {

	// 初始化日志
	klog.InitFlags(nil)
	// 初始化命令行参数
	flag.Parse()
	// SetupSignalHandler 创建一个线程并且返回context；线程内接收结束信号；进行退出
	ctx := signals.SetupSignalHandler()
	// 格式化日志
	logger := klog.FromContext(ctx)
	logger.Info("logger initialized success")

	//  配置kubernetes客户端
	cfg, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeConfig)
	if err != nil {
		logger.Error(err, "build config error")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	// 生成clientSet对象， 内部和自定义的外部对象
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		logger.Error(err, "build kubernetes client error")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	demoClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		logger.Error(err, "build demo client error")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	// 创建两个Informer分别缓存内部和外部的资源对象数据

	kubeInformerFactory := kubeInformer.NewSharedInformerFactory(kubeClient, time.Second*30)
	demoInformerFactory := informers.NewSharedInformerFactory(demoClient, time.Second*30)

	// 传递informer到控制器中
	controller := NewController(ctx, kubeClient, demoClient,
		kubeInformerFactory.Apps().V1().Deployments(),
		demoInformerFactory.Demo().V1().Demos(),
	)
	// 启动informer
	kubeInformerFactory.Start(ctx.Done())
	demoInformerFactory.Start(ctx.Done())

	if err = controller.Run(ctx, 2); err != nil {
		logger.Error(err, "controller run error")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
}

func init() {
	flag.StringVar(&kubeConfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterUrl, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
