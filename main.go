package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	corev1 "k8s.io/api/core/v1"
)

func main() {
    dryRun := flag.Bool("dry-run", false, "True -> Do not kill the pod. Just generate logs.")
    interval := flag.Duration("interval", 30*time.Second, "Execution Time Window (e.g. 30s, 1m)")
    flag.Parse()

	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	fmt.Printf("KubeMonkey Activate (dry-run=%v, interval=%v)\n", *dryRun, *interval)

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	for range ticker.C {
	    runCycle(clientset, *dryRun)
	}
}

func runCycle(clientset *kubernetes.Clientset, dryRun bool) {
    pods, err := ListPods(clientset, "")
    if err != nil {
        fmt.Println("[ERROR] list pods failed:", err)
        return
    }

    filtered := FilterPods(pods)
    if len(filtered) == 0 {
        fmt.Println("len is 0, skip")
        return
    }

    err = KillRandomPod(clientset, filtered, dryRun)
    if err != nil {
        fmt.Println("[ERROR] kill failed:", err)
    }
}

func ListPods(clientset *kubernetes.Clientset, namespace string) ([]corev1.Pod, error) {
    podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    return podList.Items, nil
}

func FilterPods(pods []corev1.Pod) []corev1.Pod {
    var result []corev1.Pod

    allowedNamespaces := map[string]bool {
        "default": true,
    }

    for _, pod := range pods {
        if !allowedNamespaces[pod.Namespace] {
            continue
        }

        result = append(result, pod)
    }

    return result
}

func KillRandomPod(clientset *kubernetes.Clientset, pods []corev1.Pod, dryRun bool) error {
    if len(pods) == 0 {
        return fmt.Errorf("no pods to kill")
    }

    randomIndex := rand.Intn(len(pods))
    target := pods[randomIndex]

    if dryRun {
        fmt.Println("[DRY RUN] Would kill:", target.Namespace, "/", target.Name)
        return nil
    }

    err := clientset.CoreV1().Pods(target.Namespace).Delete(context.TODO(), target.Name, metav1.DeleteOptions{})
    if err != nil {
        return err
    }

    fmt.Println("Killed:", target.Namespace, "/", target.Name)
    return nil
}