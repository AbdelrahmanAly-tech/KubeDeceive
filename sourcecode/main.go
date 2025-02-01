package main

import (
  "net/http"
	"log"
	"net"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"strings"
	"os"
	"fmt"
	"path/filepath"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"flag"
	"strconv"
	"io/ioutil"
	"context"
	"os/exec"
	"time"
	"k8s.io/api/admission/v1beta1"
  	"errors"
	apiv1 "k8s.io/api/core/v1"
	"encoding/json"
	"bytes"
	    "encoding/csv"


)

type ServerParameters struct {
	port           int    // webhook server port
	certFile       string // path to the x509 certificate for https
	keyFile        string // path to the x509 private key matching `CertFile`
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}
var parameters ServerParameters
var (
	universalDeserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
)
var config *rest.Config
var clientSet *kubernetes.Clientset


/////////////////////////////////////////////////////////// Logs data/
type EventList struct {
   Kind  string      `json:"kind"`
   Items []EventItem `json:"items"`
}

type EventItem struct {
   Level                   string     `json:"level"`
   AuditID                 string     `json:"auditID"`
   Stage                   string     `json:"stage"`
   RequestURI              string     `json:"requestURI"`
   Verb                    string     `json:"verb"`
   User                    User       `json:"user"`
   SourceIPs               []string   `json:"sourceIPs"`
   UserAgent               string     `json:"userAgent"`
   ObjectRef               ObjectRef  `json:"objectRef"`
   RequestReceivedTimestamp string     `json:"requestReceivedTimestamp"`
   StageTimestamp          string     `json:"stageTimestamp"`
}

type User struct {
   Username string   `json:"username"`
   Groups   []string `json:"groups"`
}

type ObjectRef struct {
   Resource   string `json:"resource"`
   Namespace  string `json:"namespace"`
   Name       string `json:"name"`
   APIVersion string `json:"apiVersion"`
}
//////////////////////////////////////////////////////////////////


type DataDict struct {
    Inputs []struct {
        Name     string                 `json:"name"`
        Datatype string                 `json:"datatype"`
        Data     map[string][]string    `json:"data"`
    } `json:"inputs"`
}


func main() {
 
	useKubeConfig := os.Getenv("USE_KUBECONFIG")
	kubeConfigFilePath := os.Getenv("KUBECONFIG")
	kubeconfigContent := "empty path"
	kubeconfigFile := "/tmp/kubeconfig"
	flag.IntVar(&parameters.port, "port", 8443, "Webhook server port.")
  	flag.StringVar(&parameters.certFile, "tlsCertFile", "/etc/webhook/certs/tls.crt", "File containing the x509 Certificate for HTTPS.")
  	flag.StringVar(&parameters.keyFile, "tlsKeyFile", "/etc/webhook/certs/tls.key", "File containing the x509 private key to --tlsCertFile.")
  	flag.Parse()

	if len(useKubeConfig) == 0 {
		log.Println("i entered the usekubeconfig from cluster")
		// default to service account in cluster token
		c, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
    
}
	
	log.Println("In-cluster Kubeconfig file created at:", kubeconfigFile)
	ksniffRun(kubeconfigFile)
/////////////////////////////////////////////////////////////////////////////////////////////////////// start the call of the model
	//time.Sleep(5 * time.Minute)
	getPrediction()
//////////////////////////////////////////////////////////////////////////////////////////////////////////// end of the call of the model
		config = c
	} else {
		//load from a kube config
		var kubeconfig string

		if kubeConfigFilePath == "" {
			if home := homedir.HomeDir(); home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
			} 
		} else {
			kubeconfig = kubeConfigFilePath
		}

    fmt.Println("kubeconfig: " + kubeconfig)

		c, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}
		config = c
	}
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	clientSet = cs
	
	

	// TLS listener for Mutations
	go func() {
		http.HandleFunc("/", HandleRoot)
		http.HandleFunc("/mutate", HandleMutate)
		// Start the webhook server
		log.Println("Starting webhook server on port 8443...")
		log.Fatal(http.ListenAndServeTLS(":"+strconv.Itoa(parameters.port), parameters.certFile, parameters.keyFile, nil))
	}()

	// Non-TLS listener for Audit Logs
	go func() {
		http.HandleFunc("/audits", handleWebhookRequest)
		log.Println("Starting non-TLS server on port 8080...")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	select {}
}

/*
func getFilePathFromDirectory(directory string) (string, error) {
    files, err := ioutil.ReadDir(directory)
    if err != nil {
        return "", err
    }

    for _, file := range files {
        if !file.IsDir() {
            return filepath.Join(directory, file.Name()), nil
        }
    }

    return "", fmt.Errorf("no files found in directory")
}

*/

func ksniffRun(kubeconfigFile string){
	// Define the target pod and namespace
	targetPodName := "kube-apiserver-kind-control-plane" // Replace with the actual pod name
	targetNamespace := "kube-system"
	log.Println("Starting sniffing process")

	// Define the output file path in PVC mounted on the webhook pod
	outputDir := "/mnt/data/input" // Replace with your mounted PVC path

	timestamp := time.Now().Format("20060102_150405") // Format: YYYYMMDD_HHMMSS

	// Create a dynamic output file name using pod name, namespace, and timestamp
	outputFileName := fmt.Sprintf("%s_%s_%s.pcap", targetPodName, targetNamespace, timestamp)
	outputFilePath := fmt.Sprintf("%s/%s", outputDir, outputFileName)

	// Define the command to sniff traffic using kubectl sniff
cmd := exec.Command(
    "kubectl",
    "sniff",
    "-p", targetPodName,
    "-n", targetNamespace,
    "-o", outputFilePath,
)

// Set the KUBECONFIG environment variable
cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", kubeconfigFile))

	// Capture standard output and standard error
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	// Set the timeout duration
	timeoutDuration := 30 * time.Second

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	// Create a channel to receive errors from the command execution
	errChan := make(chan error, 1)

	// Run the command in a goroutine
	go func() {
		errChan <- cmd.Run()
	}()

	// Wait for the command to complete or the timeout to occur
	select {
	case <-ctx.Done():
		// Timeout occurred
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Command timed out")
			// Kill the process if it is still running
			if err := cmd.Process.Kill(); err != nil {
				log.Println("Failed to kill process:", err)
			}
		}
	case err := <-errChan:
		// Command completed
		if err != nil {
			log.Printf("Error running kubectl sniff: %v\n", err)
			log.Printf("Standard Output:\n%s\n", outBuf.String())
			log.Printf("Standard Error:\n%s\n", errBuf.String())
		} else {
			log.Printf("Successfully saved sniffed traffic to %s\n", outputFilePath)
			log.Printf("Command Output:\n%s\n", outBuf.String())
		}
	}
}


func csvToDict(filePath string) (map[string][]interface{}, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        return nil, err
    }

    headers := records[0]
    dataDict := make(map[string][]interface{})

    // Initialize map with headers as keys
    for _, header := range headers {
        dataDict[header] = []interface{}{}
    }

    // Populate the map with the CSV data
    for _, record := range records[1:] {
        for i, value := range record {
            // Try to convert the value to float, if possible
            floatValue, err := strconv.ParseFloat(value, 64)
            if err != nil {
                dataDict[headers[i]] = append(dataDict[headers[i]], value)
            } else {
                dataDict[headers[i]] = append(dataDict[headers[i]], floatValue)
            }
        }
    }

    return dataDict, nil
}



func HandleRoot(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("HandleRoot!"))
}
func handleWebhookRequest(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
  // Write the body to a file
    bodyString := string(body)
   err = ioutil.WriteFile("/home/request_body.json", []byte(bodyString), 0644)
   if err != nil {
      log.Printf("Failed to write request body to file: %v", err)
      http.Error(w, "Failed to write request body to file", http.StatusInternalServerError)
      return
   }

   // Read the logs from a file
   logs, err := ioutil.ReadFile("/home/request_body.json")
   if err != nil {
      fmt.Println("Error reading file:", err)
      return
   }
   var eventList EventList
   err = json.Unmarshal(logs, &eventList)
   if err != nil {
      fmt.Println("Error decoding JSON:", err)
      return
   }

   // Access the object fields and print Logs
   eventItem := eventList.Items[0]
   fmt.Println("Username/Account:", eventItem.User.Username)
   fmt.Println("Source IP:", eventItem.SourceIPs[0])
   fmt.Println("Resource Type:", eventItem.ObjectRef.Resource)
   fmt.Println("Resource Name:", eventItem.ObjectRef.Name)
   fmt.Println("Verb:", eventItem.Verb)
   fmt.Println("Time:", eventItem.RequestReceivedTimestamp)
   fmt.Println("=========================================================")

	// Send a response back to the client
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook request received successfully"))
}

func HandleMutate(w http.ResponseWriter, r *http.Request){

	var admissionReviewReq v1beta1.AdmissionReview
	var admissionReviewResponse v1beta1.AdmissionReview
	var pod apiv1.Pod
	var serviceAccount apiv1.ServiceAccount
	//var clusterRoleBinding apiv1.ClusterRoleBinding
	f1,f2,f3 :=0,0,0
	var patches []patchOperation
	prevent_execution := true
	//	delay := false


	// Write the body to shared file
	body, err := ioutil.ReadAll(r.Body)
	err = ioutil.WriteFile("/tmp/request", body, 0644)
	if err != nil {
		panic(err.Error())
	}
	// Deserialize the intercepted request
	if _, _, err := universalDeserializer.Decode(body, nil, &admissionReviewReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Errorf("could not deserialize request: %v", err)
	} else if admissionReviewReq.Request == nil {
		w.WriteHeader(http.StatusBadRequest)
		errors.New("malformed admission review: request is nil")
	}

	fmt.Printf("Type: %v \t Event: %v \t Name: %v \n",
		admissionReviewReq.Request.Kind,
		admissionReviewReq.Request.Operation,
		admissionReviewReq.Request.Name,)

	//if (admissionReviewReq.Request.operation)
	switch admissionReviewReq.Request.Kind.Kind {
		// Process ServiceAccount request	    
		case "ServiceAccount":
			err := json.Unmarshal(admissionReviewReq.Request.Object.Raw, &serviceAccount)
			if err != nil {
				fmt.Errorf("could not unmarshal ServiceAccount on admission request: %v", err)
			}

		// Process pod request
		case "Pod":
			err = json.Unmarshal(admissionReviewReq.Request.Object.Raw, &pod)
			if err != nil {
				fmt.Errorf("could not unmarshal pod on admission request: %v", err)
			}
		default:
			fmt.Println("Unsupported resource type")
		}

	// Check for Label whitelist
	labels := pod.ObjectMeta.Labels
    content, err := ioutil.ReadFile("/tmp/test/keys.txt")
    if err != nil {
        fmt.Println(err)
    	}
		lines := strings.Split(string(content), "\n")
		secret_label:=lines[0]
		if _, exists := labels[secret_label]; exists {
				f1 = 1	
			} 
	// Check for IP whitelist
	ip ,_,_ := net.SplitHostPort(r.RemoteAddr)
	fmt.Println("IP from request",ip)
	content, err = ioutil.ReadFile("/tmp/test/ip-list.txt")
	if err != nil {
		fmt.Println(err)
			}
	IPs := strings.Split(string(content), "\n")
	for _, loopip := range IPs {
		fmt.Println("IP from file",loopip)
		if ip == loopip {
			f2 = 1 
		}
	}

	// Check for Users whitelist
	userInfo := admissionReviewReq.Request.UserInfo
	if userInfo.Username == "" {
		http.Error(w, "User info not found", http.StatusBadRequest)
		return
	}
	username := userInfo.Username
	fmt.Println("Real Username from file",username)
	content, err = ioutil.ReadFile("/tmp/test/usernames.txt")
	if err != nil {
		fmt.Println(err)
			}
	usernames := strings.Split(string(content), "\n")
	for _, u := range usernames {
	fmt.Println("Real Username from request",u)
		if username == u {
			f3 =1
		}	
	}

	if f1 == 0 {
		//This check for CTF to ensure that the user must include the secret label to proceed in creation process
		prevent_execution = false
	} else if f1 == 0 || f2 == 0 || f3 ==0 {
		fmt.Println("Deception in progress")
			patches, prevent_execution = privilege_update(&pod,patches)
			fmt.Println("Deception Deceision:", prevent_execution)	
	} else {
		fmt.Println("Allowed Source and Action")
	}
		admissionReviewResponse = v1beta1.AdmissionReview{
			Response: &v1beta1.AdmissionResponse{
				UID:     admissionReviewReq.Request.UID,
				Allowed: prevent_execution,
			},
		}

	// Encode Patches and respond
	patchBytes, err := json.Marshal(patches)
	
	if err != nil {
		fmt.Errorf("could not marshal JSON patch: %v", err)
	}
	admissionReviewResponse.Response.Patch = patchBytes
	bytes, err := json.Marshal(&admissionReviewResponse)
	if err != nil {
		fmt.Errorf("marshaling response: %v", err)
	}
		_, err = w.Write(bytes)
		if err != nil {
			http.Error(w, "Error sending response", http.StatusInternalServerError)
			return
		}
}
