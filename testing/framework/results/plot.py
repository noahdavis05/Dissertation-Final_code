import json
import pandas as pd
import matplotlib.pyplot as plt
import os


NODE_MAP = {
    "192.168.0.251": "server4",
    "192.168.0.181": "server1",
    "192.168.0.68": "server2",
    "192.168.0.182": "server3"
}

def plot_cpu_and_pods(json_path):
    if not os.path.exists(json_path):
        print(f"Error: {json_path} not found.")
        return

    with open(json_path, 'r') as f:
        data = json.load(f)
    
    df_cpu = pd.DataFrame(data["cpu_telemetry"])
    df_events = pd.DataFrame(data["events"])
    
    df_cpu['timestamp'] = pd.to_datetime(df_cpu['timestamp'])
    df_events['timestamp'] = pd.to_datetime(df_events['timestamp'])
    
    global_start = min(df_cpu['timestamp'].min(), df_events['timestamp'].min())
    
    df_cpu['seconds_elapsed'] = (df_cpu['timestamp'] - global_start).dt.total_seconds()
    df_events['seconds_elapsed'] = (df_events['timestamp'] - global_start).dt.total_seconds()

    df_cpu['node'] = df_cpu['node'].replace(NODE_MAP)

    fig, ax1 = plt.subplots(figsize=(14, 8))
    ax2 = ax1.twinx() 

    all_nodes = sorted(list(set(df_cpu['node'].unique()) | set(df_events['node'].unique())))
    cmap = plt.colormaps.get_cmap('tab10')
    
    for i, node in enumerate(all_nodes):
        color = cmap(i % 10)
        
        node_cpu = df_cpu[df_cpu['node'] == node].sort_values('seconds_elapsed')
        if not node_cpu.empty:
            ax1.plot(node_cpu['seconds_elapsed'], node_cpu['value'], 
                     label=f"CPU: {node}", color=color, linewidth=2, alpha=0.7)

        node_events = df_events[df_events['node'] == node].sort_values('seconds_elapsed')
        if not node_events.empty:
            counts = range(1, len(node_events) + 1)
            ax2.plot(node_events['seconds_elapsed'], counts, 
                     label=f"Pods: {node}", color=color, 
                     linestyle='--', linewidth=2, drawstyle='steps-post')

    ax1.set_xlabel("Seconds from Start of Experiment")
    ax1.set_ylabel("CPU Utilization (%)", color='tab:blue', fontsize=12)
    ax2.set_ylabel("Cumulative Pods Scheduled", color='tab:red', fontsize=12)
    
    plt.title(f"Scheduler Performance: {data['scheduler']}\n(Dashed = Pod Count, Solid = CPU %)", fontsize=14)
    
    ax1.grid(True, linestyle=':', alpha=0.6)
    ax1.set_ylim(0, 105)
    
    max_time = max(df_cpu['seconds_elapsed'].max(), df_events['seconds_elapsed'].max())
    ax1.set_xlim(0, max_time + 10)

    h1, l1 = ax1.get_legend_handles_labels()
    h2, l2 = ax2.get_legend_handles_labels()
    ax1.legend(h1 + h2, l1 + l2, loc='upper left', ncol=2)

    plt.tight_layout()
    
    output_filename = f"analysis_{data['scheduler']}.png"
    plt.savefig(output_filename)
    print(f"Graph successfully saved as: {output_filename}")

if __name__ == "__main__":
    plot_cpu_and_pods("custom-fuzzy-topsis-scheduler2.json")