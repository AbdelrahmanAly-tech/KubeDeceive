package main 


import (
	_ "context"
	 "fmt"
	"k8s.io/api/core/v1"
	"math/rand"
	"time"
)


// When setting a container to privileged within Kubernetes,
// the container can access additional resources and kernel capabilities of the host. 


func privilege_update(pod *v1.Pod,patches []patchOperation)([]patchOperation,bool){
	prevent_execution := true
	//**** Deception technique1 --> to fail the creation or change the malcious parameters 
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(3) + 1

	if (pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.RunAsUser != nil) {
		if int(*pod.Spec.SecurityContext.RunAsUser) == 0 {
			if randomNumber == 1 {
			securityContext := map[string]interface{}{
			    "runAsUser": 20702,
			    "RunAsGroup":1000,
			    "FSGroup":2000,
			}

		patches = append(patches, patchOperation{
			    Op:    "replace",
			    Path:  "/spec/securityContext",
			    Value: securityContext,
			})
		} else if randomNumber == 2 {
			patches = append(patches, patchOperation{
				Op:    "add",
				Path:  "/spec/nodeName",
				Value: "kind-worker2",
			})
		} else if randomNumber == 3 {
			prevent_execution = false
		} 
	}
	}

	//**** Deception technique2 --> Change the node to honeypod 
	rand.Seed(time.Now().UnixNano())
	randomNumber = rand.Intn(2) + 1
	if  (pod.Spec.NodeName == "kind-control-plane") {

		if randomNumber == 1 {
			fmt.Println("Executing Node update code...")
			patches = append(patches, patchOperation{
				Op:    "replace",
				Path:  "/spec/nodeName",
				Value: "kind-worker2",
				})
		} else if randomNumber == 2 {
			prevent_execution = false
		} 
	}

	//**** Deception technique3 --> Update the Privileged parameter 
	rand.Seed(time.Now().UnixNano())
	randomNumber = rand.Intn(3) + 1
	for i := range pod.Spec.Containers {
	if pod.Spec.Containers[i].SecurityContext != nil && pod.Spec.Containers[i].SecurityContext.Privileged != nil && *pod.Spec.Containers[i].SecurityContext.Privileged {
		fmt.Println("Executing privileged update code...")
		if randomNumber == 1 {
			patches = append(patches, patchOperation{
				Op:    "replace",
				Path:  fmt.Sprintf("/spec/containers/%d/securityContext/privileged", i),
				Value: false,
				})
		} else if randomNumber == 2 {
			patches = append(patches, patchOperation{
				Op:    "add",
				Path:  "/spec/nodeName",
				Value: "kind-worker2",
			})
		} else if randomNumber == 3 {
			prevent_execution = false
		} 
	  	}
	} 
	//**** Deception technique4 --> Remove the malcious flags to disrubt the attacker view
	if (pod.Spec.HostPID ==true ) {
	    fmt.Println("Executing HostPID flag ...")
	    fmt.Println(pod.Spec.HostPID)

	patches = append(patches, patchOperation{
		Op:    "remove",
		Path:  "/spec/hostPID",
		})
	}
	if (pod.Spec.HostNetwork ==true ) {
	    fmt.Println("Executing HostNetwork flag ...")
	    fmt.Println(pod.Spec.HostNetwork)

	patches = append(patches, patchOperation{
		Op:    "remove",
		Path:  "/spec/hostNetwork",
		})
	}
	if (pod.Spec.HostIPC ==true ) {
	    fmt.Println("Executing HostIPC flag...")
	    fmt.Println(pod.Spec.HostIPC)

	patches = append(patches, patchOperation{
		Op:    "remove",
		Path:  "/spec/hostIPC",
		})
	}
	return patches,prevent_execution
  }


	