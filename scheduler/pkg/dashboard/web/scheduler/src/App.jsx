import 'bootstrap/dist/css/bootstrap.min.css'
import LiveTelemetryGraphs from './LiveTelemetryGraphs'
import LiveTelemetryCache from './LiveTelemetryCache'
import LiveSchedulerUpdates from './LiveSchedulerUpdates'
import DetailedScheduleInfo from './DetailedScheduleInfo'
import './App.css';

import { useEffect, useState } from "react";

function App() {

  const [scheduleIndex, setScheduleIndex] = useState(null);

  const [schedulerUpdates, setSchedulerUpdates] = useState([]);
  const [telemetryCache, setTelemetryCache] = useState([]);
  const [telemetryLive, setTelemetryLive] = useState([]);

  useEffect(() => {
    const ws = new WebSocket("ws://192.168.0.251:8090/ws");

    ws.onopen = () => {
      console.log("Connected to WebSocket");
    };

    ws.onmessage = (event) => {
      // event.data is plain text
      let message;
      try {
        message = JSON.parse(event.data); 
      } catch {
        console.log("Message could not be parsed");
        return;
      }

      if (message.subject == "scheduling_update"){
        setSchedulerUpdates(prev => {
          const newArray = [...prev, message];
          return newArray.slice(-500);
        });
      } else if (message.subject == "telemetry_cache"){
        setTelemetryCache(prev => {
          const newArray = [...prev, message];
          return newArray.slice(-500);
        });
      } else if (message.subject == "live_telemetry"){
        setTelemetryLive(prev => {
          const newArray = [...prev, message];
          return newArray.slice(-500);
        });
      }

    };

    ws.onerror = (err) => console.error("WebSocket error:", err);
    ws.onclose = () => console.log("WebSocket disconnected");

    return () => ws.close();
  }, []);

  return (
    <div className="app-container bg-dark text-light min-vh-100">
      <nav className="navbar navbar-dark bg-dark shadow-sm mb-3">
        <div className="container-fluid">
          <span className="navbar-brand mb-0 h1">Fuzzy TOPSIS Scheduler Monitor</span>
        </div>
      </nav>

      <div className="container-fluid">
        <div className="row gx-3" style={{ minHeight: '50vh' }}>
          <div className="col-lg-4 mb-3">
            <LiveTelemetryGraphs data={telemetryLive}/>
          </div>
          <div className="col-lg-4 mb-3">
            <LiveTelemetryCache data={telemetryCache} />
          </div>
          <div className="col-lg-4 mb-3">
            <LiveSchedulerUpdates data={schedulerUpdates} setScheduleIndex={setScheduleIndex} />
          </div>
        </div>

        <div className="row mt-3" style={{ minHeight: '40vh' }}>
          <div className="col">
            <DetailedScheduleInfo data={schedulerUpdates} index={scheduleIndex} />
          </div>
        </div>
      </div>
    </div>
  )
}

export default App
