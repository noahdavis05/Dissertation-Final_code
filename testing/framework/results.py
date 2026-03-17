import datetime
import json


"""
Contains results from a test, either custom, default load balance, or defaul bin pack

Stores all recordings in pairs of timestamps and value

e.g. schedule event and time it happened
e.g. telemetry log event and time it happened
"""
class TestResults:

    def __init__(self, schedulerName="default"):
        self.scheduler_name = schedulerName
        self.telemetry_cpu_logs = []
        self.telemetry_ram_logs = []

        # list of dicts
        # dicts contain, node scheduled to, time
        self.schedule_event_logs = []

    def add_schedule_event(self, node_name):
        self.schedule_event_logs.append(
            {
                "timestamp": datetime.datetime.now().isoformat(),
                "node": node_name,
            }
        )

    def add_telemetry_logs(self, currentCPU, currentRAM, node_name):
        self.telemetry_cpu_logs.append(
            {
                "timestamp": datetime.datetime.now().isoformat(),
                "value": currentCPU,
                "node": node_name,
            }
        )

        self.telemetry_ram_logs.append(
            {
                "timestamp": datetime.datetime.now().isoformat(),
                "value": currentRAM,
                "node": node_name,
            }
        )

    def save_logs(self, filepath):
        data = {
            "scheduler": self.scheduler_name,
            "cpu_telemetry": self.telemetry_cpu_logs,
            "ram_telemetry": self.telemetry_ram_logs,
            "events": self.schedule_event_logs
        }

        with open(filepath, 'w') as f:
            json.dump(data, f, indent=4)
