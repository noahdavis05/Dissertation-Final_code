import yaml 
import subprocess
import time
import random
import os

NUM_PODS = 30

current_dir = os.path.dirname(os.path.abspath(__file__))
MANIFEST_PATH = os.path.join(current_dir, "stress-template.yaml")


# hardcoded random request amounts to ensure that tests are consistent
WORKLOAD_CPU_REQUESTS = [
    369, 359, 411, 431, 583, 327, 358, 350, 478, 574, 517, 598, 499, 565, 544, 
    388, 384, 385, 436, 524, 537, 600, 487, 353, 440, 392, 309, 575, 547, 323, 
    594, 591, 496, 502, 550, 351, 391, 526, 399, 565, 383, 362, 334, 328, 433, 
    528, 529, 470, 434, 465, 373, 326, 468, 492, 599, 394, 510, 588, 494, 325, 
    432, 443, 562, 336, 400, 370, 578, 350, 480, 557, 475, 357, 486, 368, 549, 
    310, 353, 479, 489, 444, 452, 518, 595, 513, 554, 356, 386, 495, 407, 501, 
    573, 425, 587, 503, 436, 402, 447, 353, 475, 539
]

def test(schedulerName, results_object, mode):
    # open the manifest
    with open(MANIFEST_PATH) as f:
        pod = yaml.safe_load(f)

    pod["spec"]["schedulerName"] = schedulerName

    for i in range(0,NUM_PODS):
        cpu_m = WORKLOAD_CPU_REQUESTS[i]
        pod_name = f"cpu-stressor-{i}"
        pod["metadata"]["name"] = pod_name

        # over request resources
        pod["spec"]["containers"][0]["resources"]["requests"]["cpu"] = str(WORKLOAD_CPU_REQUESTS[i]*1.5)+"m"

        load_val = str(max(1, int(cpu_m / 10))) 
        
        args = pod["spec"]["containers"][0]["args"]
        if "--cpu-load" in args:
            idx = args.index("--cpu-load")
            args[idx + 1] = load_val

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