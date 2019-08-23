package main

import (
	"fmt"
	"log"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const name = "custom-scheduler"

type Scheduler struct {
	clientSet *kubernetes.Clientset
}

func NewScheduler() Scheduler {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	return Scheduler{
		clientSet: clientSet,
	}
}

func main() {
	log.Println("Start a scheduler!")

	scheduler := NewScheduler()
	scheduler.SchedulePods()
}

func (s *Scheduler) SchedulePods() error {

	// start to watch pod
	watch, err := s.clientSet.CoreV1().Pods("").Watch(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.schedulerName=%s,spec.nodeName=", name),
	})
	if err != nil {
		log.Println("error when watching pods", err.Error())
		return err
	}

	// handle add pod event
	for event := range watch.ResultChan() {
		if event.Type != "ADDED" {
			continue
		}
		p, ok := event.Object.(*v1.Pod)
		if !ok {
			fmt.Println("unexpected type")
			continue
		}

		fmt.Println("found a pod to schedule:", p.Namespace, "/", p.Name)

		// find best node to bind
		node, err := s.findFit()
		if err != nil {
			log.Println("cannot find node that fits pod", err.Error())
			continue
		}

		// bind pod with selected node
		err = s.bindPod(p, node)
		if err != nil {
			log.Println("failed to bind pod", err.Error())
			continue
		}

		message := fmt.Sprintf("Placed pod [%s/%s] on %s\n", p.Namespace, p.Name, node.Name)

		// emit the scheduled event
		err = s.emitEvent(p, message)
		if err != nil {
			log.Println("failed to emit scheduled event", err.Error())
			continue
		}

		fmt.Println(message)
	}
	return nil
}

func (s *Scheduler) findFit() (*v1.Node, error) {
	nodes, err := s.clientSet.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	//always use the first node
	log.Println("Available node list:")
	for _, node := range nodes.Items {
		log.Println("available node name:", node.GetName())
	}
	return &nodes.Items[0], nil
}

func (s *Scheduler) bindPod(p *v1.Pod, node *v1.Node) error {
	return s.clientSet.CoreV1().Pods(p.Namespace).Bind(&v1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name,
			Namespace: p.Namespace,
		},
		Target: v1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Node",
			Name:       node.Name,
		},
	})
}

func (s *Scheduler) emitEvent(p *v1.Pod, message string) error {
	timestamp := time.Now().UTC()
	_, err := s.clientSet.CoreV1().Events(p.Namespace).Create(&v1.Event{
		Count:          1,
		Message:        message,
		Reason:         "Scheduled",
		LastTimestamp:  metav1.NewTime(timestamp),
		FirstTimestamp: metav1.NewTime(timestamp),
		Type:           "Normal",
		Source: v1.EventSource{
			Component: name,
		},
		InvolvedObject: v1.ObjectReference{
			Kind:      "Pod",
			Name:      p.Name,
			Namespace: p.Namespace,
			UID:       p.UID,
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: p.Name + "-",
		},
	})
	if err != nil {
		return err
	}
	return nil
}
