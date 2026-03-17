from prometheus_api_client import PrometheusConnect

class TelemetryHandler:

    def __init__(self, url="http://127.0.0.1:9090"):
        # Connect to the local port-forwarded instance
        self.prom = PrometheusConnect(url=url, disable_ssl=True)

    def get_node_cpu_utilization(self):
        query = '100 - (avg by (instance) (irate(node_cpu_seconds_total{mode="idle"}[2m])) * 100)'
        result = self.prom.custom_query(query=query)
        
        # put results into dict
        utilization = {}
        for entry in result:
            node = entry['metric']['instance']
            value = float(entry['value'][1])
            utilization[node] = round(value, 2)
            
        return utilization
    
    def get_ram_utilization(self):
        query = '(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100'
        
        result = self.prom.custom_query(query=query)
        
        utilization = {}
        for entry in result:
            # Clean up node name from 'instance' label (e.g., '192.168.1.10:9100' -> '192.168.1.10')
            node = entry['metric']['instance'].split(':')[0]
            value = float(entry['value'][1])
            utilization[node] = round(value, 2)
            
        return utilization

"""
th = TelemetryHandler()

print(th.get_node_cpu_utilization())
print(th.get_ram_utilization())

"""