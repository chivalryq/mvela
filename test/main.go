package main

import (
	"k8s.io/klog/v2"
)

func main() {
	s := []string{"a", "b"}
	klog.Info(s)
}
