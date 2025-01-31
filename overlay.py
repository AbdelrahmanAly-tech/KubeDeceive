import re
import yaml
import time
import subprocess
import time
import os

def extract_valid_ips(file_path):
    valid_ips = []
    ip_pattern = re.compile(r"\b(?:\d{1,3}\.){3}\d{1,3}\b")

    with open(file_path, "r") as file:
        for line in file:
            line = line.strip()
            if ip_pattern.match(line):
                valid_ips.append(line)

    ips_string = "\n".join(valid_ips)
    return ips_string


def read_usernames(file_path):
    usernames = []

    with open(file_path, "r") as file:
        for line in file:
            username = line.strip()
            if username:
                usernames.append(username)
    usernames_string = "\n".join(usernames)
    return usernames_string

def create_conf(sec_label,usernames_list,ips_list):
    # Define the data for the ConfigMap
    data = {
        "ip-list.txt":extract_valid_ips(ips_file_path) ,
        "usernames.txt": read_usernames(usernames_file_path),
        "keys.txt":secret_label
    }

    # Create the ConfigMap object
    config_map = {
        "apiVersion": "v1",
        "kind": "ConfigMap",
        "metadata": {
            "name": "my-config"
        },
        "data": data
    }

    # Convert the ConfigMap to YAML
    config_map_yaml = yaml.safe_dump(config_map)

    # Write the YAML to a file
    with open("configmap.yaml", "w") as f:
        f.write(config_map_yaml)

    print("ConfigMap YAML written to configmap.yaml.")


#The part below is responsible for getting information about the whitlisted IP/ Username
# add_conf=input('''- Write "1" to run without whitlisting Conf  
# - Write "2" for Add Whitlisting Conf 
# - Your choise: ''')

# if add_conf != "1":
#     secret_label = input("Enter the secret label to pass from Deception: ")
#     usernames_file_path = input("Enter the trusted username file path: ")
#     ips_file_path = input("Enter the allowed IPs file path: ")
#     create_conf(secret_label,usernames_file_path,ips_file_path)

try:
    print("Update the configMap.. ")
    command8 = "sudo kubectl delete configmap my-config;sudo kubectl apply -f ./sourcecode/configmap.yaml"
    subprocess.run(command8, shell=True, check=True,stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
except subprocess.CalledProcessError as e:
    raise (e)




try:
    print("Remove the existing  MutatingWebhookConfiguration .. ")
    command1 = "sudo kubectl delete  MutatingWebhookConfiguration example-webhook"
    subprocess.run(command1, shell=True, check=False, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
except subprocess.CalledProcessError as e:
    print("Error occurred while removing the existing deployment:", e)

try:
    print("Remove the existing webhook deployment .. ")
    command2 = "sudo kubectl delete deployment example-webhook"
    subprocess.run(command2, shell=True, check=False, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
except subprocess.CalledProcessError as e:
    print("Error occurred while removing the existing deployment:", e)




try:
    print("Apply the persistant volume and persistant Volume Claim.. ")
    command3 = "sudo kubectl apply -f ./pv-pvc.yaml"
    subprocess.run(command3, shell=True, check=True,stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
except subprocess.CalledProcessError as e:
    raise (e)



try:
    print("Apply the deployment rbac .. ")
    command5 = "sudo kubectl -n default apply -f ./rbac.yaml"
    subprocess.run(command5, shell=True, check=False, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
except subprocess.CalledProcessError as e:
    print("Error occurred while Apply rbac.yaml:", e)


try:
    print("Build the updated image and push .. ")
    command6 = "sudo docker build ./sourcecode/ -t abdulrhmanaly/webhook-ksniff:v1.7"
    subprocess.run(command6, shell=True, check=True,stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    print("Build is completed and started the push process .. ")
    command7 = "sudo docker push abdulrhmanaly/webhook-ksniff:v1.7"
    subprocess.run(command6, shell=True, check=True,stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
except subprocess.CalledProcessError as e:
    raise (e)



try:
    print("Create the new deployment.. ")
    command9 = "sudo kubectl -n default apply -f ./deployment.yaml"
    subprocess.run(command9, shell=True, check=True,stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
except subprocess.CalledProcessError as e:
    raise (e)


print("Waiting for the deployment to be ready...")
while True:
    try:
        command10 = "sudo kubectl -n default wait --for=condition=available deployment/example-webhook"
        subprocess.run(command10, shell=True, check=True,stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        break  # Break the loop if the deployment is ready
    except subprocess.CalledProcessError:
        time.sleep(1)  # Wait for 1 second before checking again

try:
    print("Apply  the new MutatingWebhookConfiguration.. ")
    command11 = "sudo kubectl -n default apply -f ./webhook.yaml"
    subprocess.run(command11, shell=True, check=True,stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
except subprocess.CalledProcessError as e:
    raise (e)

'''
try:
    print("Create the demooooo.. ")
    command12 = "sudo kubectl -n default apply -f ./everything-allowed.yaml"
    subprocess.run(command12, shell=True, check=True,stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
except subprocess.CalledProcessError as e:
    raise (e)'''