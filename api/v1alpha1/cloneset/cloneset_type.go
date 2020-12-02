package cloneset

type CloneSet struct {
	// Replicas is the desired number of replicas of the given Template.
	Replica *int32

	// ScaleStrategy indicate the ScaleStrategy that will create and delete Pods
	ScaleStrategy CloneSetScaleStrategy
}

type CloneSetScaleStrategy struct {
}
