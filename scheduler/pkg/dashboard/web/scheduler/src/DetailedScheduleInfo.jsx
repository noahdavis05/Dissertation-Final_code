import FuzzyDecisionMatrix from "./FuzzyDecisionMatrix";
import { NodeInfoTable } from "./NodeInfoTable";
import { TelemetryCacheTable } from "./TelemetryCacheTable";
import { NodeScoresTable } from "./NodeScoresTable";

function DetailedScheduleInfo({ data, index }) {
  if (index == null || !data) return <p>No Schedule Selected</p>;

  const payload = data[index].payload;
  const { initialFuzzyDM, filteredFuzzyDM, weightedFuzzyDM, telemetryCache, nodeScores, manualFuzzyDM, ...rest } = payload;

  return (
    <div className="card">
      <div className="card-header">
        <h2>Detailed Schedule Info</h2>
      </div>
      <div className="card-body">
        <NodeInfoTable {...rest} />
        <TelemetryCacheTable telemetryCache={telemetryCache} />
        <FuzzyDecisionMatrix matrix={initialFuzzyDM} title="Initial Fuzzy DM" />
        <FuzzyDecisionMatrix matrix={manualFuzzyDM} title="Manual additions" />
        <FuzzyDecisionMatrix matrix={filteredFuzzyDM} title="Filtered Fuzzy DM" />
        <FuzzyDecisionMatrix matrix={weightedFuzzyDM} title="Weighted Fuzzy DM" />

        
        
        <NodeScoresTable nodeScores={nodeScores} />

      </div>
    </div>
  );
}

export default DetailedScheduleInfo;