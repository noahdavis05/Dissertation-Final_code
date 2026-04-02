import time
import subprocess
import os
from .boutique import deploy_online_boutique


current_dir = os.path.dirname(os.path.abspath(__file__))
locust_file = os.path.join(current_dir, "locust.py")
results_csv = os.path.join(current_dir, "results_boutique")

def boutique_load_test(scheduler_name, save_dir):
    # deploy the whole online boutique application
    deploy_online_boutique(scheduler_name)

    # sleep for 7 min while initial deployment is deployed.
    # and locust test is manually carried out.
    time.sleep(420)