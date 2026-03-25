import yaml
import subprocess
import requests

URL = "https://raw.githubusercontent.com/DescartesResearch/TeaStore/master/examples/kubernetes/teastore-clusterip.yaml"

POLICY = {
    "teastore-webui":       {"cpu_req": "200m", "cpu_lim": "500m", "mem_req": "384Mi", "mem_lim": "512Mi"},
    "teastore-auth":        {"cpu_req": "100m", "cpu_lim": "200m", "mem_req": "128Mi", "mem_lim": "256Mi"},
    "teastore-image":       {"cpu_req": "150m", "cpu_lim": "300m", "mem_req": "256Mi", "mem_lim": "512Mi"},
    "teastore-persistence": {"cpu_req": "150m", "cpu_lim": "300m", "mem_req": "384Mi", "mem_lim": "512Mi"},
    "teastore-recommender": {"cpu_req": "300m", "cpu_lim": "600m", "mem_req": "512Mi", "mem_lim": "1Gi"},
    "teastore-db":          {"cpu_req": "200m", "cpu_lim": "500m", "mem_req": "512Mi", "mem_lim": "1Gi"},
    "teastore-registry":    {"cpu_req": "50m",  "cpu_lim": "100m", "mem_req": "128Mi", "mem_lim": "256Mi"},
}

def deploy_teastore(schedulerName):
    print(f"Deploying TeaStore with scheduler: {schedulerName}")
    resp = requests.get(URL)
    docs = list(yaml.safe_load_all(resp.text))
    
    modified_docs = []

    for doc in docs:
        if not doc: continue
        
        if doc.get("kind") == "Deployment":
            name = doc["metadata"]["name"]
            spec = doc["spec"]["template"]["spec"]
            
            spec["schedulerName"] = schedulerName
            
            p = POLICY.get(name, {"cpu_req": "100m", "cpu_lim": "200m", "mem_req": "256Mi", "mem_lim": "512Mi"})
            
            for container in spec["containers"]:
                container["resources"] = {
                    "requests": {"cpu": p["cpu_req"], "memory": p["mem_req"]},
                    "limits": {"cpu": p["cpu_lim"], "memory": p["mem_lim"]}
                }
                
                # add startup probe so when pods autoscale they don't create an avalanche
                # because pods use all their CPU booting.
                port = 3306 if "db" in name else 8080
                container["startupProbe"] = {
                    "httpGet": {"path": "/", "port": port},
                    "failureThreshold": 30,
                    "periodSeconds": 10
                }
            print(f"Set resources for {name}")

        modified_docs.append(doc)

    full_yaml = yaml.dump_all(modified_docs)
    subprocess.run(["kubectl", "apply", "-f", "-"], input=full_yaml.encode())
