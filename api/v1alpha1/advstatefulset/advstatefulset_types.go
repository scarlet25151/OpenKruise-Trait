package advstatefulset

type StatefulSetUpdateStrategyType string

const (
	RollingUpdateStrategyType StatefulSetUpdateStrategyType = "RollingUpdate"

	OnDeleteStrategyType StatefulSetUpdateStrategyType = "OnDeleted"
)
