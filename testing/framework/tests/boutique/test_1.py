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

    # sleep for 1 min while initial deployment is deployed.
    time.sleep(50)

    # start load testing using the locust script
    locust_command = [
        "locust",
        "-f", locust_file,      
        "--headless",           
        "-u", "500",            
        "-r", "2",            
        "--run-time", "5m",     
        "--host", "http://192.168.0.200/", 
        "--csv", save_dir + "/results/" + scheduler_name + "-results-boutique"
    ]

    print("Starting Locust Load Test...")
    subprocess.run(locust_command)