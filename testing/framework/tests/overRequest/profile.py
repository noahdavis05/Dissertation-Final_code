import yaml 
import subprocess
import time
import random
import os

NUM_PODS = 5

current_dir = os.path.dirname(os.path.abspath(__file__))
MANIFEST_PATH = os.path.join(current_dir, "stress-template.yaml")

def test(schedulerName, results_object):
    # open the manifest
    with open(MANIFEST_PATH) as f:
        pod = yaml.safe_load(f)

    pod["spec"]["schedulerName"] = schedulerName

    for i in range(0,NUM_PODS):
        pod_name = f"cpu-stressor-{i}"
        pod["metadata"]["name"] = pod_name

        manifest = yaml.dump(pod)
        subprocess.run(["kubectl", "apply", "-f", "-"], input=manifest.encode())

        # get what node our pod was scheduled on
        time.sleep(1)
        cmd = [
                "kubectl", "get", "pod", pod_name, 
                "-o", "jsonpath={.spec.nodeName}"
            ]
        
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        node_name = result.stdout.strip()

        results_object.add_schedule_event(node_name)

        time.sleep(random.randint(5, 20))

    # save the results
    filename = f"results_{schedulerName}.json"
    save_path = os.path.join(current_dir, filename)
    results_object.save_logs(save_path)