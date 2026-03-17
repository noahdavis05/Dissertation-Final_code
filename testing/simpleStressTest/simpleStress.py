


#########################################################################
# SIMPLE STRESS is a basic stress test on the k8s cluster, it uses a    #
# simple manifest which creates pods which use a fraction of the CPU,   #
# e.g. 0.1 cores. This test creates lots of these pods (of different    #
# sizes), and schedules them over time. It then monitors where each pod #
# gets scheduled in the cluster.                                        #
#########################################################################


import yaml 
import subprocess
import time
import random

NUM_PODS = 50

with open("stress-template.yaml") as f:
    pod = yaml.safe_load(f)


for i in range (0,NUM_PODS):
    pod_name = f"cpu-stressor-{i}"
    pod["metadata"]["name"] = pod_name

    # convert back to yaml
    manifest = yaml.dump(pod)

    # apply the manifest
    subprocess.run(["kubectl", "apply", "-f", "-"], input=manifest.encode())

    
    

    

    # sleep for random time until ready to schedule another
    time.sleep(random.randint(5, 20))

