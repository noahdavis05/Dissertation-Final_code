package config

import "scheduler/pkg/types"

// prometheus variables
var (
	PrometheusServiceName = "kube-prom-stack-kube-prome-prometheus"
	PrometheusNamespace   = "monitoring"
	PrometheusPort        = "9090"

	/*
		PrometheusURL = fmt.Sprintf(
			"http://%s.%s.svc.cluster.local:%s",
			PrometheusServiceName,
			PrometheusNamespace,
			PrometheusPort,
		)
	*/
	PrometheusURL = "http://localhost:9090"
)

// TOPSIS variables e.g. weights and ideals
var (
	// TOPSIS ideals
	PosCPUIdeal = types.FuzzyNumber{
		A: 75,
		B: 75,
		C: 75,
	}
	NegCPUIdeal = types.FuzzyNumber{
		A: 85,
		B: 85,
		C: 85,
	}

	PosRAMIdeal = types.FuzzyNumber{
		A: 75,
		B: 75,
		C: 75,
	}
	NegRAMIdeal = types.FuzzyNumber{
		A: 85,
		B: 85,
		C: 85,
	}

	PosCPURangeIdeal = types.FuzzyNumber{
		A: 0,
		B: 0,
		C: 0,
	}
	NegCPURangeIdeal = types.FuzzyNumber{ // negative range values shouldn't be 100 as this is awful and we should never even see this
		A: 40, // range of 40 still awful and punishes ranges of 20 - 50 better
		B: 40,
		C: 40,
	}

	PosRAMRangeIdeal = types.FuzzyNumber{
		A: 0,
		B: 0,
		C: 0,
	}
	NegRAMRangeIdeal = types.FuzzyNumber{
		A: 40,
		B: 40,
		C: 40,
	}

	// TOPSIS Weights
	CPUWeights = types.FuzzyNumber{
		A: 1,
		B: 1,
		C: 1,
	}

	RAMWeights = types.FuzzyNumber{
		A: 1,
		B: 1,
		C: 1,
	}

	CPURangeWeights = types.FuzzyNumber{
		A: 1.5,
		B: 1.5,
		C: 1.5,
	}

	RAMRangeWeights = types.FuzzyNumber{
		A: 1.5,
		B: 1.5,
		C: 1.5,
	}
)
