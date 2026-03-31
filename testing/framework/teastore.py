import yaml
import subprocess
import requests

URL = "https://raw.githubusercontent.com/DescartesResearch/TeaStore/master/examples/kubernetes/teastore-clusterip.yaml"

POLICY = {
    "teastore-webui":       {"cpu_req": "600m", "cpu_lim": "800m", "mem_req": "768Mi", "mem_lim": "1Gi"},
    "teastore-auth":        {"cpu_req": "300m", "cpu_lim": "400m", "mem_req": "512Mi", "mem_lim": "1Gi"},
    "teastore-image":       {"cpu_req": "300m", "cpu_lim": "400m", "mem_req": "512Mi", "mem_lim": "1Gi"},
    "teastore-persistence": {"cpu_req": "300m", "cpu_lim": "400m", "mem_req": "512Mi", "mem_lim": "1Gi"},
    "teastore-recommender": {"cpu_req": "500m", "cpu_lim": "700m", "mem_req": "512Mi", "mem_lim": "1Gi"},
    "teastore-db":          {"cpu_req": "400m", "cpu_lim": "700m", "mem_req": "512Mi", "mem_lim": "1Gi"},
    "teastore-registry":    {"cpu_req": "150m", "cpu_lim": "300m", "mem_req": "256Mi", "mem_lim": "512Mi"},
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
                
                port = 3306 if "db" in name else 8080
                
                
                container["startupProbe"] = {
                    "tcpSocket": {"port": port},
                    "initialDelaySeconds": 30,
                    "failureThreshold": 30,   
                    "periodSeconds": 5   
                }

                
                container["readinessProbe"] = {
                    "tcpSocket": {"port": port},
                    "periodSeconds": 10
                }

            print(f"Set resources for {name}")

        modified_docs.append(doc)

    full_yaml = yaml.dump_all(modified_docs)
    subprocess.run(["microk8s","kubectl", "apply", "-f", "-"], input=full_yaml.encode())


deploy_teastore("custom-fuzzy-topsis-scheduler")