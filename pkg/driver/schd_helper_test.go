package driver

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
)

func lvmNode(name string, vgFree int64) apis.LVMNode {
	return apis.LVMNode{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		VolumeGroups: []apis.VolumeGroup{{
			Name: "lvmvg",
			Size: *resource.NewQuantity(100*Gi, resource.BinarySI),
			Free: *resource.NewQuantity(vgFree, resource.BinarySI),
		}},
	}
}

func TestFilterSuitableNodesVgFreeEqualsCapacity(t *testing.T) {
	const capacity = 10 * Gi
	nodes, err := filterSuitableNodes([]apis.LVMNode{
		lvmNode("node-less", capacity-Mi),
		lvmNode("node-equal", capacity),
		lvmNode("node-more", capacity+Mi),
	}, regexp.MustCompile("^lvmvg$"), strconv.FormatInt(capacity, 10))
	assert.NoError(t, err)
	assert.Equal(t, []string{"node-equal", "node-more"}, nodes)
}
