# KubeDeceive: Enhancing Kubernetes Security  

## Overview  
KubeDeceive is a security framework designed to strengthen **Kubernetes environments** by integrating **multi-class threat detection, adaptive deception, and privilege escalation monitoring**. This approach enhances resilience against unauthorized access and attacks in cloud-native infrastructure.  

### Key Components  
- **KServe** – AI-based real-time threat classification  
- **CICFlowMeter** – Network traffic analysis for behavioral detection  
- **KubeDeceive** – Deception-based security framework  
- **Webhook Security** – Monitors API interactions and applies security rules  
- **Audit Logging** – Captures and analyzes security-relevant API requests  

---

## How It Works  
1. **Captures Kubernetes API requests** to identify suspicious activities  
2. **Analyzes network traffic** using NetFlow feature extraction  
3. **Detects and mitigates privilege escalation attempts**  
4. **Applies deception techniques** to mislead and log attacker behavior  
5. **Uses AI-based classification** to determine security risks  


---



## Setup & Configuration  

### 1. SSL Configuration   
Refer to the [SSL Configuration Guide](TLS_Configurations.md) for detailed instructions.

### 2. Kubernetes Cluster Initialization  
`sudo kind create cluster --config cluster_auditconf.yaml`

### 3️. Install KServe
`sudo curl -s "https://raw.githubusercontent.com/kserve/kserve/release-0.12/hack/quick_install.sh" | sudo bash`

### 4️. Execute Overlay Script
`sudo python3 ./overlay.py`

### 5️5 Verify Deployment
`kubectl get all -A`



