package integration

import (
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FindReplicaSetNewerThan finds a ReplicaSet created after or at the specified time.
// This helper logs detailed information about ReplicaSets found to help
// with debugging timestamp issues.
//
// It will return nil if no ReplicaSet is found, and will log a warning if
// multiple ReplicaSets are found that were created after or at the specified time.
// In case of multiple matches, it returns the most recently created one.
func FindReplicaSetNewerThan(
	t *testing.T,
	ctx context.Context,
	cli client.Client,
	creationTime time.Time,
	namespace string,
	labels map[string]string,
) *appsv1.ReplicaSet {
	t.Helper()

	newReplicaSetList := &appsv1.ReplicaSetList{}
	if err := cli.List(ctx, newReplicaSetList,
		client.InNamespace(namespace),
		client.MatchingLabels(labels)); err != nil {
		t.Logf("Error listing ReplicaSets: %v", err)
		return nil
	}

	// Find ReplicaSets newer than or exactly at the specified time
	var newReplicaSets []*appsv1.ReplicaSet
	t.Logf("Looking for ReplicaSets created at or after %v", creationTime)
	for i := range newReplicaSetList.Items {
		rs := &newReplicaSetList.Items[i]
		t.Logf("Found ReplicaSet %s with creation time %v", rs.Name, rs.CreationTimestamp.Time)
		// Truncate times to seconds for comparison since k8s doesn't use same precision
		rsTime := rs.CreationTimestamp.Truncate(time.Second)
		refTime := creationTime.Truncate(time.Second)
		if rsTime.Equal(refTime) || rsTime.After(refTime) {
			t.Logf("ReplicaSet %s is at or after %v", rs.Name, creationTime)
			newReplicaSets = append(newReplicaSets, rs)
		}
	}

	// No new ReplicaSets found
	if len(newReplicaSets) == 0 {
		t.Logf("No ReplicaSets found at or after %v", creationTime)
		return nil
	}

	// Multiple new ReplicaSets found - fail the test
	if len(newReplicaSets) > 1 {
		t.Errorf("Found %d ReplicaSets created at or after %v, expected exactly 1",
			len(newReplicaSets), creationTime)
		t.FailNow()
	}

	// Success: exactly one new ReplicaSet found
	return newReplicaSets[0]
}
