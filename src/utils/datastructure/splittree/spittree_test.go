// Unit test for splittree

package splittree

import (
    "testing"
)

func _TestSplittreeAllFuncs(t *testing.T) {
	t.Log(FromNodeToLeaf(0),FromNodeToLeaf(1),FromNodeToLeaf(2))
    t.Log(IsLeaf(1),IsLeaf(3),IsLeaf(4))
    t.Log(FromLeaftoNode(1),FromLeaftoNode(5),FromLeaftoNode(3))
    t.Log(Parent(1),Parent(6),Parent(8),Parent(11),Parent(12))
    t.Log(Left(2),Left(12),Left(4),Left(14))
    t.Log(Right(2),Right(12),Right(4),Right(14))
    t.Log(
        GetRootLable(1),GetRootLable(2),GetRootLable(3),GetRootLable(4),
        GetRootLable(5),GetRootLable(6),GetRootLable(7),GetRootLable(8),
        GetRootLable(9),GetRootLable(10),GetRootLable(11),GetRootLable(12),
    )
}

func TestTraverse(t *testing.T) {
    Traverse(5, func(nodeid uint32, layer uint32) {
        t.Log(layer,",",nodeid)
    })
}
