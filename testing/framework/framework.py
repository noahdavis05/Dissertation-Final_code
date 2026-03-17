from results import TestResults
from telemetry import TelemetryHandler

from tests.standard.profile import standardTest

#import yaml 
import subprocess
import time
from datetime import datetime
import random
import threading


# globals
CUSTOM_SCHEDULER_NAME = "topsis-scheduler"
DEFAULT_SCHEDULER_NAME = "default-scheduler"

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


    def run_stress_ng_tests(self):
        ###################################################
        ## PHASE 1 - we run the test on custon scheduler ##
        ###################################################
        self.stop_event.clear()
        monitor_thread = threading.Thread(
            target=self.monitor_nodes_telemetry, 
            args=(self.custom_results,)
        )
        monitor_thread.start()

        # run test
        self.customTest(CUSTOM_SCHEDULER_NAME, self.custom_results)

        # stop monitoring
        self.stop_event.set()
        monitor_thread.join()

        print("Finished test")

        self.cleanup_default_namspace()

        ###########################################
        ## PHASE 2 - we run on default scheduler ##
        ###########################################
        self.stop_event.clear()
        monitor_thread = threading.Thread(
            target=self.monitor_nodes_telemetry, 
            args=(self.default_load_balance_results,)
        )
        monitor_thread.start()
        self.customTest(DEFAULT_SCHEDULER_NAME, self.default_load_balance_results)

        self.stop_event.set()
        monitor_thread.join()

        print("Finished test")



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

    def cleanup_default_namspace(self):
        print("cleaning default namespace")
        subprocess.run(["kubectl", "delete", "pods", "--all", "-n", "default", "--now"])


framework = SchedulerTester(standardTest)

framework.run_stress_ng_tests()

#framework.cleanup_default_namspace()


