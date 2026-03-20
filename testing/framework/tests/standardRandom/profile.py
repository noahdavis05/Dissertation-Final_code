import yaml 
import subprocess
import time
import random
import os

NUM_PODS = 5

current_dir = os.path.dirname(os.path.abspath(__file__))
MANIFEST_PATH = os.path.join(current_dir, "stress-template.yaml")


# hardcoded random request amounts to ensure that tests are consistent
WORKLOAD_CPU_REQUESTS = [
    107, 62, 190, 175, 164, 121, 102, 329, 94, 266, 66, 65, 97, 161, 169, 
    308, 63, 337, 151, 329, 264, 162, 279, 192, 53, 131, 266, 224, 192, 129, 
    160, 222, 102, 97, 244, 99, 233, 226, 185, 72, 285, 324, 113, 243, 90, 
    332, 200, 235, 345, 148, 85, 73, 166, 198, 90, 169, 101, 244, 192, 282, 
    236, 133, 239, 231, 157, 186, 86, 137, 323, 175, 133, 286, 244, 188, 335, 
    162, 216, 78, 167, 66, 211, 255, 187, 83, 158, 340, 211, 158, 305, 252, 
    284, 123, 185, 121, 176, 337, 325, 184, 349, 269
]

def test(schedulerName, results_object, mode):
    # open the manifest
    with open(MANIFEST_PATH) as f:
        pod = yaml.safe_load(f)

    pod["spec"]["schedulerName"] = schedulerName

    for i in range(0,NUM_PODS):
        pod_name = f"cpu-stressor-{i}"
        pod["metadata"]["name"] = pod_name

        pod["spec"]["containers"][0]["resources"]["limits"]["cpu"] = str(WORKLOAD_CPU_REQUESTS[i])+"m"
        pod["spec"]["containers"][0]["resources"]["requests"]["cpu"] = str(WORKLOAD_CPU_REQUESTS[i])+"m"

        manifest = yaml.dump(pod)
        if mode == "microk8s":
            subprocess.run(["microk8s","kubectl", "apply", "-f", "-"], input=manifest.encode())
        else:
            subprocess.run(["kubectl", "apply", "-f", "-"], input=manifest.encode())

        # get what node our pod was scheduled on
        time.sleep(1)
        if mode == "microk8s":
            cmd = [
                    "microk8s","kubectl", "get", "pod", pod_name, 
                    "-o", "jsonpath={.spec.nodeName}"
                ]
        else:
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