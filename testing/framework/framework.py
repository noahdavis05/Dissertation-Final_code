from results import TestResults
from telemetry import TelemetryHandler

#import yaml 
import subprocess
import time
from datetime import datetime
import random
from datetime import datetime, timezone
import threading
import json
import os


# globals
CUSTOM_SCHEDULER_NAME = "custom-fuzzy-topsis-scheduler"
DEFAULT_SCHEDULER_NAME = "default-scheduler"
STANDARD_TOPSIS_NAME = "topsis-scheduler"
STANDARD_FUZZY_TOPSIS_NAME = "fuzzy-topsis-scheduler"
DEFAULT_BIN_PACK_NAME = "bin-packing-scheduler"


SCHEDULERS = [CUSTOM_SCHEDULER_NAME, STANDARD_TOPSIS_NAME, STANDARD_FUZZY_TOPSIS_NAME, DEFAULT_SCHEDULER_NAME, DEFAULT_BIN_PACK_NAME]

MODE = "microk8s" # modes can be kind (kubernetes in docker), or microk8s
# this changes the command based on what environment we are tetsing in

CURRENT_DIR = current_dir = os.path.dirname(os.path.abspath(__file__))


"""
The role of this class is to run different kinds of tests on the 
custom scheduler, and compare this with the default scheduler.

The class can run multiple different tests.
These tests will schedule a set of pods via custom scheduler, then schedule the same set of pods via the default.
The results will be monitored throughout, and logged at the end for comparison.
"""
class SchedulerTester:
    
    def __init__(self, function):
        self.customTest = function

        # test results
        self.custom_results = TestResults(schedulerName=CUSTOM_SCHEDULER_NAME)
        self.default_load_balance_results = TestResults(schedulerName=DEFAULT_SCHEDULER_NAME)
        self.default_bin_pack_results = TestResults()

        # telemetry which will run in a background thread
        self.telemetry_handler = TelemetryHandler()
        self.stop_event = threading.Event()

        # number of scheduled pods
        self.scheduled_pods_num = 0

    
    def run_tests(self):
        if self.tp.test_type == "stress-ng":
            self.run_stress_ng_tests()


    def run_stress_tests(self):
        ###################################################
        ## iterate over all schedulers and run same test ##
        ###################################################
        for scheduler in SCHEDULERS:
            newResults = TestResults(schedulerName=scheduler)
            print("Running Test on " + scheduler)
            self.stop_event.clear()
            monitor_thread = threading.Thread(
                target=self.monitor_nodes_telemetry, 
                args=(newResults,)
            )
            monitor_thread.start()

            # run test
            self.customTest(scheduler, newResults, MODE)

            # stop monitoring
            self.stop_event.set()
            monitor_thread.join()

            print("Finished test")

            self.cleanup_default_namspace(MODE)
            if MODE == "kind":
                self.wait_for_idle(5)
            else:
                self.wait_for_idle(20)

    def run_boutique_test(self, scheduler):
        newResults = TestResults(schedulerName=scheduler)
        print("Running Test on " + scheduler)
        self.stop_event.clear()
        monitor_thread = threading.Thread(
            target=self.monitor_nodes_telemetry, 
            args=(newResults,)
        )
        monitor_thread.start()

        schedule_thread = threading.Thread(
            target=self.detect_scheduled_pod,
            args=(newResults,)
        )
        schedule_thread.start()

        # run test
        self.customTest(scheduler, CURRENT_DIR)

        # stop monitoring
        self.stop_event.set()
        monitor_thread.join()
        schedule_thread.join()

        print("Finished test")

        filepath = CURRENT_DIR + "/results/" + scheduler + ".json"

        newResults.save_logs(filepath)

        self.cleanup_default_deployments()

        self.scheduled_pods_num = 0


    def monitor_nodes_telemetry(self, results_object):
        print("Running background telemetry scraping")
        while not self.stop_event.is_set():
            cpu_data = self.telemetry_handler.get_node_cpu_utilization()
            ram_data = self.telemetry_handler.get_ram_utilization()

            #print(cpu_data, ram_data)

            # both are in dicts
            for key, value in cpu_data.items():
                # get the corresponding ram data for the key
                ram_val = ram_data[key.split(":")[0]]
                results_object.add_telemetry_logs(value, ram_val, key.split(":")[0])

            time.sleep(5)
            #print("Added values to the logs ")

    def cleanup_default_namspace(self, mode):
        print("cleaning default namespace")
        if mode == "microk8s":
            subprocess.run(["microk8s","kubectl", "delete", "pods", "--all", "-n", "default", "--now"])
        else:
            subprocess.run(["kubectl", "delete", "pods", "--all", "-n", "default", "--now"])

    def cleanup_default_deployments(self):
        print("Removign all deployments")
        subprocess.run(["microk8s","kubectl", "delete", "deployment", "--all"])

    
    def wait_for_idle(self, threshold=5.0):
        # minimum of 5 min wait to ensure all history is cleared from telemetry
        time.sleep(300)
        print(f"Waiting for nodes to drop below " + str(threshold) + " % CPU...")
        while True:
            cpu_data = self.telemetry_handler.get_node_cpu_utilization()
            if all(float(val) <= threshold for val in cpu_data.values()):
                print("Nodes are idle. Starting next test.")
                break
            time.sleep(10)


    def detect_scheduled_pod(self, results_object):
        while not self.stop_event.is_set():
            time.sleep(10)
            command = ["microk8s","kubectl", "get", "pods", "-o", "json"]
            result = subprocess.run(command, capture_output=True, text=True, check=True)
            json_result = json.loads(result.stdout)

            pods = json_result.get("items", [])
            now = datetime.now(timezone.utc)
            
            count = len(pods)
            if count > self.scheduled_pods_num:
                # means that pods have been scheduled since last check
                pod_difference = count - self.scheduled_pods_num

                sorted_pods = []

                for pod in pods:
                    node = pod.get("spec", {}).get("nodeName")
                    if not node:
                        node = "Not Scheduled"
                    creation_str = pod["metadata"].get("creationTimestamp")
                    creation_time = datetime.strptime(creation_str, "%Y-%m-%dT%H:%M:%SZ").replace(tzinfo=timezone.utc)
                    age_seconds = int((now - creation_time).total_seconds())
                    sorted_pods.append((node, age_seconds))
                
                # now sort the pods by age_seconds ascending
                sorted_pods.sort(key=lambda x: x[1])
                sorted_pods = sorted_pods[0:pod_difference]

                # now log these results
                for item in sorted_pods:
                    results_object.add_schedule_event(item[0])
                    print("Pod scheduled to node: " + item[0])
                    self.scheduled_pods_num += 1

                
        


            



## choose which library we import the test from

from tests.standardRandom.profile import test as test1
from tests.overRequestRandom.profile import test as test2

from tests.boutique.test_1 import boutique_load_test

"""
framework1 = SchedulerTester(test1)
framework2 = SchedulerTester(test2)



framework1.run_stress_ng_tests()
framework2.run_stress_ng_tests()

framework2.cleanup_default_namspace(MODE)
"""

framework = SchedulerTester(test2)
framework.run_stress_tests()
#framework.run_boutique_test("custom-fuzzy-topsis-scheduler")

