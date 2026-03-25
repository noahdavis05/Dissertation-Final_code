from results import TestResults
from telemetry import TelemetryHandler

#import yaml 
import subprocess
import time
from datetime import datetime
import random
import threading


# globals
CUSTOM_SCHEDULER_NAME = "custom-fuzzy-topsis-scheduler"
DEFAULT_SCHEDULER_NAME = "default-scheduler"
STANDARD_TOPSIS_NAME = "topsis-scheduler"
STANDARD_FUZZY_TOPSIS_NAME = "fuzzy-topsis-scheduler"


SCHEDULERS = [CUSTOM_SCHEDULER_NAME, STANDARD_TOPSIS_NAME, STANDARD_FUZZY_TOPSIS_NAME, DEFAULT_SCHEDULER_NAME]

MODE = "kind" # modes can be kind (kubernetes in docker), or microk8s
# this changes the command based on what environment we are tetsing in


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

    
    def wait_for_idle(self, threshold=5.0):
        print(f"Waiting for nodes to drop below " + str(threshold) + " % CPU...")
        while True:
            cpu_data = self.telemetry_handler.get_node_cpu_utilization()
            if all(float(val) <= threshold for val in cpu_data.values()):
                print("Nodes are idle. Starting next test.")
                break
            time.sleep(10)


## choose which library we import the test from
from tests.standardRandom.profile import test as test1
from tests.overRequestRandom.profile import test as test2

framework1 = SchedulerTester(test1)
framework2 = SchedulerTester(test2)



framework1.run_stress_ng_tests()
framework2.run_stress_ng_tests()

framework2.cleanup_default_namspace(MODE)


