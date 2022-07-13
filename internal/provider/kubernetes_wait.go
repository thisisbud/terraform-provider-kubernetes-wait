package provider

import (
	"context"
	"fmt"
	"os"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func doIt() error {

	// userHomeDir, err := os.UserHomeDir()
	// if err != nil {
	// 	fmt.Printf("error getting user home dir: %v\n", err)
	// 	os.Exit(1)
	// }
	// kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	// fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

	config, err := clientcmd.BuildConfigFromFlags("127.0.0.1:8001", "")
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error getting Kubernetes clientset: %v\n", err)
		os.Exit(1)
	}

	pods, err := clientset.CoreV1().Pods("default").List(context.Background(), v1.ListOptions{})
	if err != nil {
		fmt.Printf("error getting pods: %v\n", err)
		os.Exit(1)
	}
	for _, pod := range pods.Items {
		fmt.Printf("Pod name: %s\n", pod.Name)
	}

	/*
		serving, err := servingv1.NewForConfig(config)
		if err != nil {
			return err
		}

		// Get services in the default namespace
		list, err := serving.Services("webhook-demo").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}
		// How to print them out.
		fmt.Printf("There are %d services in the default namespace", len(list.Items))
		for _, i := range list.Items {
			fmt.Printf("  > Service %q", i.Name)
		}

	*/

	services, err := clientset.CoreV1().Services("webhook-demo").List(context.Background(), v1.ListOptions{})

	if err != nil {
		fmt.Printf("error getting pods: %v\n", err)
		os.Exit(1)
	}
	for _, pod := range services.Items {
		fmt.Printf("Service name: %s\n", pod.Name)
	}

	return nil

}
