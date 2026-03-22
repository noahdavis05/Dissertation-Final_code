import yaml
import subprocess
import requests

URL = "https://raw.githubusercontent.com/DescartesResearch/TeaStore/master/examples/kubernetes/teastore-clusterip.yaml"

RESOURCE_MAPPING = {
    "teastore-db":           {"cpu": "300m", "mem": "512Mi"},
    "teastore-registry":     {"cpu": "100m", "mem": "256Mi"},
    "teastore-persistence":  {"cpu": "300m", "mem": "512Mi"},
    "teastore-auth":         {"cpu": "200m", "mem": "256Mi"},
    "teastore-webui":        {"cpu": "300m", "mem": "512Mi"},
    "teastore-image":        {"cpu": "400m", "mem": "512Mi"},
    "teastore-recommender":  {"cpu": "400m", "mem": "512Mi"},
}

def deploy_teastore_for_research():
    print("Fetching TeaStore manifests...")
    resp = requests.get(URL)
    docs = list(yaml.safe_load_all(resp.text))
    
    for doc in docs:
        if not doc or doc.get("kind") != "Deployment":
            continue
            
        name = doc["metadata"]["name"]
        spec = doc["spec"]["template"]["spec"]
        
        spec["schedulerName"] = "topsis-scheduler"
        

        res = RESOURCE_MAPPING.get(name, {"cpu": "200m", "mem": "256Mi"})
        
        for container in spec["containers"]:
            container["resources"] = {
                "requests": {
                    "cpu": res["cpu"],
                    "memory": res["mem"]
                },
            }
        print(f"Configured {name} with CPU: {res['cpu']} and Mem: {res['mem']}")

    full_yaml = yaml.dump_all(docs)
    subprocess.run(["kubectl", "apply", "-f", "-"], input=full_yaml.encode())

if __name__ == "__main__":
    deploy_teastore_for_research()