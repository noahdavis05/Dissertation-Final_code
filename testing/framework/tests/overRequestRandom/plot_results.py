import json
import pandas as pd
import matplotlib.pyplot as plt
import os

files = ["results_topsis-scheduler.json", "results_default-scheduler.json", "results_fuzzy-topsis-scheduler.json", "results_custom-fuzzy-topsis-scheduler.json"]

def plot_cpu_comparison(json_files):
    fig, axes = plt.subplots(len(json_files), 1, figsize=(12, 10), sharex=False)
    
    if len(json_files) == 1:
        axes = [axes]

    for i, file_path in enumerate(json_files):
        with open(file_path, 'r') as f:
            data = json.load(f)
        
        df = pd.DataFrame(data["cpu_telemetry"])
        df['timestamp'] = pd.to_datetime(df['timestamp'])
        
        start_time = df['timestamp'].min()
        df['seconds_elapsed'] = (df['timestamp'] - start_time).dt.total_seconds()
        
        df = df.sort_values(['node', 'seconds_elapsed'])
        
        for node in df['node'].unique():
            node_df = df[df['node'] == node]
            axes[i].plot(node_df['seconds_elapsed'], node_df['value'], label=f"Node {node}", marker='o', markersize=3)
        
        axes[i].set_title(f"Scheduler Strategy: {data['scheduler']}")
        axes[i].set_ylabel("CPU Utilization (%)")
        axes[i].set_xlabel("Seconds from start of test")
        axes[i].legend(loc='upper right', fontsize='small')
        axes[i].grid(True, linestyle='--', alpha=0.5)

    plt.tight_layout()
    plt.savefig("scheduler_comparison_results.png")
    print("Graph saved as scheduler_comparison_results.png")

plot_cpu_comparison(files)