import yaml
import subprocess
import requests
import os

FILE_PATH = "deployment.yaml"

def deploy_online_boutique(scheduler_name):
    if not os.path.exists(FILE_PATH):
        print(f"Error: {FILE_PATH} not found")
        return

    with open(FILE_PATH, 'r') as f:
        docs = list(yaml.safe_load_all(f))
    modified_docs = []

    for doc in docs:
        if doc and doc.get("kind") == "Deployment":
            name = doc["metadata"]["name"]
            
            template_spec = doc["spec"]["template"]["spec"]
            template_spec["schedulerName"] = scheduler_name
            
            print(f"Updated Deployment: {name}")

        modified_docs.append(doc)

    full_yaml = yaml.dump_all(modified_docs)
    
    try:
        subprocess.run(
            ["microk8s","kubectl", "apply", "-f", "-"], 
            input=full_yaml.encode(), 
            check=True
        )
        print("\nDeployment successful.")
    except subprocess.CalledProcessError as e:
        print(f"Error applying YAML: {e}")
