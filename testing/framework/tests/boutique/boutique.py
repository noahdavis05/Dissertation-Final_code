import yaml
import subprocess
import requests

URL = "https://raw.githubusercontent.com/GoogleCloudPlatform/microservices-demo/main/release/kubernetes-manifests.yaml"

def deploy_online_boutique(scheduler_name):
    resp = requests.get(URL)

    docs = list(yaml.safe_load_all(resp.text))
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
            ["kubectl", "apply", "-f", "-"], 
            input=full_yaml.encode(), 
            check=True
        )
        print("\nDeployment successful.")
    except subprocess.CalledProcessError as e:
        print(f"Error applying YAML: {e}")
